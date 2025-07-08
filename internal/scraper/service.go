package scraper

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"price-comparison-tool/internal/config"
	"price-comparison-tool/internal/matcher"
	"price-comparison-tool/internal/models"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/debug"
)

type Service struct {
	config    *config.Config
	sites     []models.SiteConfig
	collectors map[string]*colly.Collector
	matcher   *matcher.Service
	mutex     sync.RWMutex
}

func NewService(cfg *config.Config) *Service {
	s := &Service{
		config:     cfg,
		collectors: make(map[string]*colly.Collector),
		matcher:    matcher.NewService(cfg),
	}
	
	s.loadSiteConfigs()
	s.initializeCollectors()
	
	return s
}

func (s *Service) loadSiteConfigs() {
	s.sites = []models.SiteConfig{
		{
			Name:       "Amazon US",
			BaseURL:    "https://www.amazon.com",
			SearchPath: "/s?k=",
			Countries:  []string{"US"},
			Selectors: models.SiteSelectors{
				Product:  "[data-component-type='s-search-result']",
				Price:    ".a-price-whole, .a-offscreen",
				Title:    "[data-cy='title-recipe-title'] span, h2 a span",
				Link:     "h2 a",
				Currency: ".a-price-symbol",
			},
			Headers: map[string]string{
				"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
			},
			RateLimit: 2000,
		},
		{
			Name:       "Amazon Canada",
			BaseURL:    "https://www.amazon.ca",
			SearchPath: "/s?k=",
			Countries:  []string{"CA"},
			Selectors: models.SiteSelectors{
				Product:  "[data-component-type='s-search-result']",
				Price:    ".a-price-whole, .a-offscreen",
				Title:    "[data-cy='title-recipe-title'] span, h2 a span",
				Link:     "h2 a",
				Currency: ".a-price-symbol",
			},
			Headers: map[string]string{
				"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
			},
			RateLimit: 2000,
		},
		{
			Name:       "Amazon UK",
			BaseURL:    "https://www.amazon.co.uk",
			SearchPath: "/s?k=",
			Countries:  []string{"UK"},
			Selectors: models.SiteSelectors{
				Product:  "[data-component-type='s-search-result']",
				Price:    ".a-price-whole, .a-offscreen",
				Title:    "[data-cy='title-recipe-title'] span, h2 a span",
				Link:     "h2 a",
				Currency: ".a-price-symbol",
			},
			Headers: map[string]string{
				"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
			},
			RateLimit: 2000,
		},
		{
			Name:       "Amazon India",
			BaseURL:    "https://www.amazon.in",
			SearchPath: "/s?k=",
			Countries:  []string{"IN"},
			Selectors: models.SiteSelectors{
				Product:  "[data-component-type='s-search-result']",
				Price:    ".a-price-whole, .a-offscreen",
				Title:    "[data-cy='title-recipe-title'] span, h2 a span",
				Link:     "h2 a",
				Currency: ".a-price-symbol",
			},
			Headers: map[string]string{
				"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
			},
			RateLimit: 2000,
		},
		{
			Name:       "eBay US",
			BaseURL:    "https://www.ebay.com",
			SearchPath: "/sch/i.html?_nkw=",
			Countries:  []string{"US"},
			Selectors: models.SiteSelectors{
				Product:  ".s-item",
				Price:    ".s-item__price",
				Title:    ".s-item__title",
				Link:     ".s-item__link",
			},
			Headers: map[string]string{
				"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
			},
			RateLimit: 1500,
		},
		{
			Name:       "eBay Canada",
			BaseURL:    "https://www.ebay.ca",
			SearchPath: "/sch/i.html?_nkw=",
			Countries:  []string{"CA"},
			Selectors: models.SiteSelectors{
				Product:  ".s-item",
				Price:    ".s-item__price",
				Title:    ".s-item__title",
				Link:     ".s-item__link",
			},
			Headers: map[string]string{
				"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
			},
			RateLimit: 1500,
		},
		{
			Name:       "eBay UK",
			BaseURL:    "https://www.ebay.co.uk",
			SearchPath: "/sch/i.html?_nkw=",
			Countries:  []string{"UK"},
			Selectors: models.SiteSelectors{
				Product:  ".s-item",
				Price:    ".s-item__price",
				Title:    ".s-item__title",
				Link:     ".s-item__link",
			},
			Headers: map[string]string{
				"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
			},
			RateLimit: 1500,
		},
		{
			Name:       "Flipkart",
			BaseURL:    "https://www.flipkart.com",
			SearchPath: "/search?q=",
			Countries:  []string{"IN"},
			Selectors: models.SiteSelectors{
				Product:  "[data-id]",
				Price:    "._30jeq3, ._1_WHN1",
				Title:    "._4rR01T, .s1Q9rs",
				Link:     "._1fQZEK, ._2rpwqI",
			},
			Headers: map[string]string{
				"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
			},
			RateLimit: 2000,
		},
		{
			Name:       "Snapdeal",
			BaseURL:    "https://www.snapdeal.com",
			SearchPath: "/search?keyword=",
			Countries:  []string{"IN"},
			Selectors: models.SiteSelectors{
				Product:  ".product-tuple-listing",
				Price:    ".lfloat.product-price",
				Title:    ".product-title",
				Link:     ".dp-widget-link",
			},
			Headers: map[string]string{
				"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
			},
			RateLimit: 2000,
		},
		{
			Name:       "Walmart US",
			BaseURL:    "https://www.walmart.com",
			SearchPath: "/search?q=",
			Countries:  []string{"US"},
			Selectors: models.SiteSelectors{
				Product:  "[data-testid='item-stack']",
				Price:    "[data-automation-id='product-price']",
				Title:    "[data-automation-id='product-title']",
				Link:     "[data-automation-id='product-title'] a",
			},
			Headers: map[string]string{
				"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
			},
			RateLimit: 2500,
		},
		{
			Name:       "Walmart Canada",
			BaseURL:    "https://www.walmart.ca",
			SearchPath: "/search?q=",
			Countries:  []string{"CA"},
			Selectors: models.SiteSelectors{
				Product:  "[data-testid='product-tile']",
				Price:    "[data-testid='price-current']",
				Title:    "[data-testid='product-title']",
				Link:     "[data-testid='product-title'] a",
			},
			Headers: map[string]string{
				"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
			},
			RateLimit: 2500,
		},
	}
}

func (s *Service) initializeCollectors() {
	for _, site := range s.sites {
		collector := colly.NewCollector(
			colly.Debugger(&debug.LogDebugger{}),
		)
		
		collector.Limit(&colly.LimitRule{
			DomainGlob:  "*",
			Parallelism: 2,
			Delay:       time.Duration(site.RateLimit) * time.Millisecond,
		})
		
		for key, value := range site.Headers {
			collector.OnRequest(func(r *colly.Request) {
				r.Headers.Set(key, value)
			})
		}
		
		collector.OnError(func(r *colly.Response, err error) {
			log.Printf("Error scraping %s: %v", r.Request.URL, err)
		})
		
		s.collectors[site.Name] = collector
	}
}

func (s *Service) FetchPrices(ctx context.Context, country, query string) ([]models.ProductResult, error) {
	relevantSites := s.getSitesForCountry(country)
	if len(relevantSites) == 0 {
		return nil, fmt.Errorf("no supported sites for country: %s", country)
	}
	
	resultsChan := make(chan models.ScrapingResult, len(relevantSites))
	var wg sync.WaitGroup
	
	for _, site := range relevantSites {
		wg.Add(1)
		go func(site models.SiteConfig) {
			defer wg.Done()
			results := s.scrapeWebsite(ctx, site, query, country)
			resultsChan <- results
		}(site)
	}
	
	go func() {
		wg.Wait()
		close(resultsChan)
	}()
	
	var allResults []models.ProductResult
	for result := range resultsChan {
		if result.Error != nil {
			log.Printf("Error scraping %s: %v", result.Site, result.Error)
			continue
		}
		allResults = append(allResults, result.Products...)
	}
	
	// Temporarily disable LLM processing for debugging
	log.Printf("Found %d raw results before filtering", len(allResults))
	
	// Add fuzzy confidence scores
	for i := range allResults {
		allResults[i].Confidence = s.matcher.FuzzyProductMatch(query, allResults[i].ProductName)
	}
	
	return allResults, nil
}

func (s *Service) scrapeWebsite(ctx context.Context, site models.SiteConfig, query, country string) models.ScrapingResult {
	encodedQuery := url.QueryEscape(query)
	searchURL := site.BaseURL + site.SearchPath + encodedQuery
	
	var products []models.ProductResult
	var scrapeError error
	
	// Create a fresh collector for each request to avoid caching issues
	collector := colly.NewCollector()
	
	collector.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 1,
		Delay:       500 * time.Millisecond, // Reduced delay
	})
	
	// Add timeout
	collector.SetRequestTimeout(15 * time.Second)
	
	// Set headers
	for key, value := range site.Headers {
		collector.OnRequest(func(r *colly.Request) {
			r.Headers.Set(key, value)
		})
	}
	
	collector.OnHTML(site.Selectors.Product, func(e *colly.HTMLElement) {
		// Limit results to prevent infinite processing
		if len(products) >= 25 {
			return
		}
		
		title := strings.TrimSpace(e.ChildText(site.Selectors.Title))
		priceText := strings.TrimSpace(e.ChildText(site.Selectors.Price))
		linkHref := e.ChildAttr(site.Selectors.Link, "href")
		
		if title == "" || priceText == "" {
			return
		}
		
		fullLink := linkHref
		if !strings.HasPrefix(linkHref, "http") {
			fullLink = site.BaseURL + linkHref
		}
		
		currency := extractCurrency(priceText, country)
		cleanPrice := cleanPriceText(priceText)
		
		if cleanPrice != "" && title != "" {
			product := models.ProductResult{
				Link:        fullLink,
				Price:       cleanPrice,
				Currency:    currency,
				ProductName: title,
				Site:        site.Name,
				Country:     country,
				FetchedAt:   time.Now(),
			}
			products = append(products, product)
		}
	})
	
	collector.OnError(func(r *colly.Response, err error) {
		scrapeError = err
	})
	
	if err := collector.Visit(searchURL); err != nil {
		scrapeError = err
	}
	
	return models.ScrapingResult{
		Products: products,
		Site:     site.Name,
		Error:    scrapeError,
	}
}

func (s *Service) getSitesForCountry(country string) []models.SiteConfig {
	var relevantSites []models.SiteConfig
	
	for _, site := range s.sites {
		for _, supportedCountry := range site.Countries {
			if supportedCountry == country {
				relevantSites = append(relevantSites, site)
				break
			}
		}
	}
	
	return relevantSites
}

func (s *Service) GetSupportedSites() []string {
	var siteNames []string
	for _, site := range s.sites {
		siteNames = append(siteNames, site.Name)
	}
	return siteNames
}

func extractCurrency(priceText, country string) string {
	currencyMap := map[string]string{
		"US": "USD",
		"IN": "INR",
		"UK": "GBP",
		"CA": "CAD",
		"AU": "AUD",
	}
	
	if currency, exists := currencyMap[country]; exists {
		return currency
	}
	
	if strings.Contains(priceText, "$") {
		return "USD"
	} else if strings.Contains(priceText, "₹") {
		return "INR"
	} else if strings.Contains(priceText, "£") {
		return "GBP"
	}
	
	return "USD"
}

func cleanPriceText(priceText string) string {
	re := regexp.MustCompile(`[\d,]+\.?\d*`)
	matches := re.FindAllString(priceText, -1)
	
	if len(matches) > 0 {
		return strings.ReplaceAll(matches[0], ",", "")
	}
	
	return ""
}