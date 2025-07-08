package matcher

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"price-comparison-tool/internal/config"
	"price-comparison-tool/internal/models"
	"regexp"
	"strings"
	"time"

	"github.com/agnivade/levenshtein"
	"github.com/xrash/smetrics"
)

type Service struct {
	config     *config.Config
	httpClient *http.Client
}

type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type OllamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

func NewService(cfg *config.Config) *Service {
	return &Service{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 90 * time.Second,
		},
	}
}

func (s *Service) FilterAndScoreProducts(ctx context.Context, query string, products []models.ProductResult) ([]models.ProductResult, error) {
	if len(products) == 0 {
		return products, nil
	}

	var filteredProducts []models.ProductResult
	
	for _, product := range products {
		score, err := s.scoreProductMatch(ctx, query, product.ProductName)
		if err != nil {
			// If LLM fails, use fuzzy scoring
			score = s.FuzzyProductMatch(query, product.ProductName)
		}
		
		// Only include products with reasonable confidence (lowered threshold)
		if score >= 0.1 {
			product.Confidence = score
			filteredProducts = append(filteredProducts, product)
		}
	}
	
	// If filtering removed all products, return original list with fuzzy scores
	if len(filteredProducts) == 0 && len(products) > 0 {
		for _, product := range products {
			product.Confidence = s.FuzzyProductMatch(query, product.ProductName)
			filteredProducts = append(filteredProducts, product)
		}
	}
	
	return filteredProducts, nil
}

func (s *Service) scoreProductMatch(ctx context.Context, query, productName string) (float64, error) {
	// Create timeout context for LLM call
	llmCtx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()
	prompt := fmt.Sprintf(`You are a product matching expert. Rate how well this product matches the search query on a scale from 0.0 to 1.0.

Search Query: "%s"
Product Name: "%s"

Scoring Guidelines:
- 1.0: Perfect match (exact product, brand, model, specs)
- 0.8-0.9: Excellent match (same product, minor spec differences)
- 0.6-0.7: Good match (same brand/category, different model/version)
- 0.4-0.5: Moderate match (related products, accessories, or alternatives)
- 0.2-0.3: Weak match (same category but different brand/purpose)
- 0.0-0.1: No match (completely unrelated products)

Examples:
- Query: "iPhone 15 128GB" vs "Apple iPhone 15 - 128GB Black" = 1.0
- Query: "iPhone 15" vs "iPhone 14 Pro" = 0.7  
- Query: "iPhone 15" vs "iPhone Case for 15" = 0.4
- Query: "iPhone 15" vs "Samsung Galaxy S24" = 0.2
- Query: "iPhone 15" vs "Laptop Charger" = 0.0

Respond with only the numeric score (0.0-1.0), no explanation.

Score:`, query, productName)

	response, err := s.CallOllama(llmCtx, prompt)
	if err != nil {
		return 0, err
	}

	// Parse the score from response
	score := s.parseScore(response)
	return score, nil
}

func (s *Service) CallOllama(ctx context.Context, prompt string) (string, error) {
	ollamaURL := s.config.OllamaHost + "/api/generate"
	startTime := time.Now()
	log.Printf("🔗 Attempting LLM connection to: %s", ollamaURL)
	
	reqBody := OllamaRequest{
		Model:  "phi3:mini",
		Prompt: prompt,
		Stream: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", ollamaURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		elapsedTime := time.Since(startTime)
		log.Printf("❌ LLM connection failed to %s: %v (took: %.2fs)", ollamaURL, err, elapsedTime.Seconds())
		
		// Retry logic with exponential backoff
		if elapsedTime < 60*time.Second {
			log.Printf("🔄 Retrying LLM call after brief delay...")
			time.Sleep(2 * time.Second)
			
			// Retry once
			retryStart := time.Now()
			resp, retryErr := s.httpClient.Do(req)
			if retryErr == nil {
				defer resp.Body.Close()
				retryElapsed := time.Since(retryStart)
				log.Printf("✅ LLM retry successful (took: %.2fs)", retryElapsed.Seconds())
				
				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					return "", err
				}
				
				var ollamaResp OllamaResponse
				if err := json.Unmarshal(body, &ollamaResp); err != nil {
					return "", err
				}
				
				return ollamaResp.Response, nil
			} else {
				log.Printf("❌ LLM retry also failed: %v", retryErr)
			}
		}
		
		return "", err
	}
	defer resp.Body.Close()

	elapsedTime := time.Since(startTime)
	log.Printf("✅ LLM connection successful to %s (status: %d, took: %.2fs)", ollamaURL, resp.StatusCode, elapsedTime.Seconds())

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var ollamaResp OllamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return "", err
	}

	return ollamaResp.Response, nil
}

func (s *Service) parseScore(response string) float64 {
	// Clean the response and try to extract a number
	cleaned := strings.TrimSpace(response)
	cleaned = strings.ReplaceAll(cleaned, "\n", "")
	
	// Try to parse common formats
	var score float64
	_, err := fmt.Sscanf(cleaned, "%f", &score)
	if err != nil {
		// If parsing fails, look for patterns
		if strings.Contains(strings.ToLower(cleaned), "0.") {
			if strings.Contains(cleaned, "0.9") || strings.Contains(cleaned, "0.8") {
				return 0.85
			} else if strings.Contains(cleaned, "0.7") || strings.Contains(cleaned, "0.6") {
				return 0.65
			} else if strings.Contains(cleaned, "0.5") || strings.Contains(cleaned, "0.4") {
				return 0.45
			} else {
				return 0.25
			}
		}
		return 0.5 // Default if we can't parse
	}
	
	// Clamp score between 0 and 1
	if score < 0 {
		score = 0
	} else if score > 1 {
		score = 1
	}
	
	return score
}

func (s *Service) BasicProductMatch(query, productName string) float64 {
	queryLower := strings.ToLower(query)
	productLower := strings.ToLower(productName)
	
	// Split query into words
	queryWords := strings.Fields(queryLower)
	matchCount := 0
	
	for _, word := range queryWords {
		if len(word) < 3 {
			continue // Skip very short words
		}
		if strings.Contains(productLower, word) {
			matchCount++
		}
	}
	
	if len(queryWords) == 0 {
		return 0
	}
	
	score := float64(matchCount) / float64(len(queryWords))
	
	// Boost score if it contains key brand/model info
	if strings.Contains(queryLower, "iphone") && strings.Contains(productLower, "iphone") {
		score += 0.3
	}
	if strings.Contains(queryLower, "pro") && strings.Contains(productLower, "pro") {
		score += 0.2
	}
	if strings.Contains(queryLower, "128gb") && strings.Contains(productLower, "128") {
		score += 0.2
	}
	
	// Clamp score
	if score > 1 {
		score = 1
	}
	
	return score
}

// FuzzyProductMatch uses fuzzy string matching with semantic bonuses for better product matching
func (s *Service) FuzzyProductMatch(query, productName string) float64 {
	if query == "" || productName == "" {
		return 0.0
	}

	queryLower := strings.ToLower(strings.TrimSpace(query))
	productLower := strings.ToLower(strings.TrimSpace(productName))

	// Stage 1: Calculate base fuzzy similarity using Jaro-Winkler
	jaroWinkler := smetrics.JaroWinkler(queryLower, productLower, 0.7, 4)
	
	// Also calculate Levenshtein-based similarity for comparison
	maxLen := len(queryLower)
	if len(productLower) > maxLen {
		maxLen = len(productLower)
	}
	if maxLen == 0 {
		return 0.0
	}
	
	levenshteinDist := levenshtein.ComputeDistance(queryLower, productLower)
	levenshteinSim := 1.0 - float64(levenshteinDist)/float64(maxLen)
	
	// Use the higher of the two similarity scores as base
	baseSimilarity := jaroWinkler
	if levenshteinSim > jaroWinkler {
		baseSimilarity = levenshteinSim
	}

	// Stage 2: Add semantic bonuses
	score := baseSimilarity
	
	// Brand matching bonus
	brandBonus := s.calculateBrandBonus(queryLower, productLower)
	score += brandBonus
	
	// Model/number matching bonus  
	modelBonus := s.calculateModelBonus(queryLower, productLower)
	score += modelBonus
	
	// Storage/specification matching bonus
	specBonus := s.calculateSpecBonus(queryLower, productLower)
	score += specBonus
	
	// Stage 3: Apply penalties for clearly irrelevant products
	penalty := s.calculateRelevancePenalty(queryLower, productLower)
	score -= penalty

	// Ensure score stays within 0.0 to 1.0 range
	if score > 1.0 {
		score = 1.0
	} else if score < 0.0 {
		score = 0.0
	}

	return score
}

func (s *Service) calculateBrandBonus(query, product string) float64 {
	brands := []string{"apple", "iphone", "samsung", "galaxy", "google", "pixel", "oneplus", "xiaomi", "huawei", "oppo", "vivo", "realme"}
	
	queryBrand := ""
	productBrand := ""
	
	// Find brand in query and product
	for _, brand := range brands {
		if strings.Contains(query, brand) {
			queryBrand = brand
		}
		if strings.Contains(product, brand) {
			productBrand = brand
		}
	}
	
	// Special handling for iPhone (Apple product)
	if strings.Contains(query, "iphone") && strings.Contains(product, "apple") {
		queryBrand = "apple"
		productBrand = "apple"
	}
	if strings.Contains(query, "apple") && strings.Contains(product, "iphone") {
		queryBrand = "apple"
		productBrand = "apple"
	}
	
	if queryBrand != "" && queryBrand == productBrand {
		return 0.3 // Strong brand match bonus
	} else if queryBrand != "" && productBrand != "" && queryBrand != productBrand {
		return -0.2 // Different brand penalty
	}
	
	return 0.0
}

func (s *Service) calculateModelBonus(query, product string) float64 {
	// Extract numbers/models from both strings
	numRegex := regexp.MustCompile(`\d+`)
	queryNumbers := numRegex.FindAllString(query, -1)
	productNumbers := numRegex.FindAllString(product, -1)
	
	if len(queryNumbers) == 0 || len(productNumbers) == 0 {
		return 0.0
	}
	
	matchCount := 0
	for _, qNum := range queryNumbers {
		for _, pNum := range productNumbers {
			if qNum == pNum {
				matchCount++
				break
			}
		}
	}
	
	if matchCount > 0 {
		similarity := float64(matchCount) / float64(len(queryNumbers))
		return similarity * 0.2 // Model number match bonus
	}
	
	return 0.0
}

func (s *Service) calculateSpecBonus(query, product string) float64 {
	bonus := 0.0
	
	// Storage matching
	storageRegex := regexp.MustCompile(`(\d+)\s*(gb|tb)`)
	queryStorage := storageRegex.FindAllString(query, -1)
	productStorage := storageRegex.FindAllString(product, -1)
	
	for _, qStorage := range queryStorage {
		for _, pStorage := range productStorage {
			if strings.EqualFold(qStorage, pStorage) {
				bonus += 0.15
			}
		}
	}
	
	// Color matching
	colors := []string{"black", "white", "red", "blue", "green", "yellow", "purple", "pink", "gold", "silver", "gray", "grey"}
	for _, color := range colors {
		if strings.Contains(query, color) && strings.Contains(product, color) {
			bonus += 0.05
		}
	}
	
	// Condition matching
	conditions := []string{"new", "used", "refurbished", "renewed", "good", "excellent", "fair"}
	for _, condition := range conditions {
		if strings.Contains(query, condition) && strings.Contains(product, condition) {
			bonus += 0.05
		}
	}
	
	return bonus
}

func (s *Service) calculateRelevancePenalty(query, product string) float64 {
	penalty := 0.0
	
	// If query is for a phone but product is clearly an accessory
	phoneTerms := []string{"iphone", "galaxy", "pixel", "phone", "smartphone"}
	accessoryTerms := []string{"case", "cover", "charger", "cable", "screen protector", "tempered glass", "stand", "holder", "adapter"}
	
	queryIsPhone := false
	productIsAccessory := false
	
	for _, term := range phoneTerms {
		if strings.Contains(query, term) {
			queryIsPhone = true
			break
		}
	}
	
	for _, term := range accessoryTerms {
		if strings.Contains(product, term) {
			productIsAccessory = true
			break
		}
	}
	
	if queryIsPhone && productIsAccessory {
		penalty += 0.4 // Heavy penalty for phone query matching accessories
	}
	
	return penalty
}