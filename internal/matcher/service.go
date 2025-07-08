package matcher

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"price-comparison-tool/internal/config"
	"price-comparison-tool/internal/models"
	"strings"
	"time"
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
			Timeout: 30 * time.Second,
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
			// If LLM fails, use basic scoring
			score = s.BasicProductMatch(query, product.ProductName)
		}
		
		// Only include products with reasonable confidence (lowered threshold)
		if score >= 0.1 {
			product.Confidence = score
			filteredProducts = append(filteredProducts, product)
		}
	}
	
	// If filtering removed all products, return original list with basic scores
	if len(filteredProducts) == 0 && len(products) > 0 {
		for _, product := range products {
			product.Confidence = s.BasicProductMatch(query, product.ProductName)
			filteredProducts = append(filteredProducts, product)
		}
	}
	
	return filteredProducts, nil
}

func (s *Service) scoreProductMatch(ctx context.Context, query, productName string) (float64, error) {
	prompt := fmt.Sprintf(`Rate how well this product matches the search query on a scale from 0.0 to 1.0.

Search Query: "%s"
Product Name: "%s"

Consider:
- Exact product matches (brand, model, specifications)
- Similar or alternative products
- Completely unrelated products should score very low

Respond with only a number between 0.0 and 1.0, no explanation.

Score:`, query, productName)

	response, err := s.callOllama(ctx, prompt)
	if err != nil {
		return 0, err
	}

	// Parse the score from response
	score := s.parseScore(response)
	return score, nil
}

func (s *Service) callOllama(ctx context.Context, prompt string) (string, error) {
	reqBody := OllamaRequest{
		Model:  "phi3:mini",
		Prompt: prompt,
		Stream: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.config.OllamaHost+"/api/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

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