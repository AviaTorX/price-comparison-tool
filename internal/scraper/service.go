package scraper

import (
	"context"
	"encoding/json"
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
				"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
				"Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8",
				"Accept-Language": "en-US,en;q=0.9",
				"Accept-Encoding": "gzip, deflate, br",
				"DNT": "1",
				"Connection": "keep-alive",
				"Upgrade-Insecure-Requests": "1",
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
				Product:  "[data-component-type='s-search-result'], .s-result-item, [data-asin], .sg-col-inner",
				Price:    ".a-price-whole, .a-offscreen, .a-price .a-offscreen, .a-price-range, .a-price",
				Title:    "[data-cy='title-recipe-title'] span, h2 a span, .a-size-medium, .a-size-base-plus, [data-cy='title-recipe-title']",
				Link:     "h2 a, .a-link-normal",
				Currency: ".a-price-symbol",
			},
			Headers: map[string]string{
				"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
				"Accept":     "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
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
				"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
				"Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
				"Accept-Language": "en-US,en;q=0.9",
				"Accept-Encoding": "gzip, deflate, br",
				"DNT": "1",
				"Connection": "keep-alive",
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
				Product:  "._1AtVbE, ._13oc-S, [data-id], ._1fQZEK, ._75nlfW, [data-testid='product-base'], .cPHDOP, ._2kHMtA, ._3pLy-c, .col-12-12",
				Price:    "._30jeq3, ._1_WHN1, .Nx9bqj, ._25b18c, ._3I9_wc, ._2rQ-NK, .Nx9bqj, ._30jeq3, ._1_WHN1, ._25b18c",
				Title:    "._4rR01T, .s1Q9rs, .IRpwTa, ._2WkVRV, ._3pLy-c, .col-7-12, .KzDlHZ, ._2WkVRV, ._4rR01T, .s1Q9rs",
				Link:     "._1fQZEK, ._2rpwqI, .IRpwTa, ._2WkVRV a, ._3pLy-c a, .col-7-12 a, .KzDlHZ, ._2WkVRV a",
			},
			Headers: map[string]string{
				"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
				"Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8",
				"Accept-Language": "en-IN,en-US;q=0.9,en;q=0.8,hi;q=0.7",
				"Accept-Encoding": "gzip, deflate, br",
				"DNT": "1",
				"Connection": "keep-alive",
				"Upgrade-Insecure-Requests": "1",
				"Sec-Fetch-Dest": "document",
				"Sec-Fetch-Mode": "navigate",
				"Sec-Fetch-Site": "none",
				"Sec-Fetch-User": "?1",
				"Referer": "https://www.flipkart.com/",
			},
			RateLimit: 3000,
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
				"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
				"Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8",
				"Accept-Language": "en-IN,en-US;q=0.9,en;q=0.8,hi;q=0.7",
				"Accept-Encoding": "gzip, deflate, br",
				"DNT": "1",
				"Connection": "keep-alive",
				"Upgrade-Insecure-Requests": "1",
				"Sec-Fetch-Dest": "document",
				"Sec-Fetch-Mode": "navigate",
				"Sec-Fetch-Site": "cross-site",
				"Sec-Fetch-User": "?1",
				"Cache-Control": "max-age=0",
			},
			RateLimit: 4000,
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
		// European Markets
		{
			Name:       "Amazon Germany",
			BaseURL:    "https://www.amazon.de",
			SearchPath: "/s?k=",
			Countries:  []string{"DE"},
			Selectors: models.SiteSelectors{
				Product:  "[data-component-type='s-search-result']",
				Price:    ".a-price-whole, .a-offscreen",
				Title:    "[data-cy='title-recipe-title'] span, h2 a span",
				Link:     "h2 a",
				Currency: ".a-price-symbol",
			},
			Headers: map[string]string{
				"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			},
			RateLimit: 2000,
		},
		{
			Name:       "Amazon France",
			BaseURL:    "https://www.amazon.fr",
			SearchPath: "/s?k=",
			Countries:  []string{"FR"},
			Selectors: models.SiteSelectors{
				Product:  "[data-component-type='s-search-result']",
				Price:    ".a-price-whole, .a-offscreen",
				Title:    "[data-cy='title-recipe-title'] span, h2 a span",
				Link:     "h2 a",
				Currency: ".a-price-symbol",
			},
			Headers: map[string]string{
				"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			},
			RateLimit: 2000,
		},
		{
			Name:       "Amazon Japan",
			BaseURL:    "https://www.amazon.co.jp",
			SearchPath: "/s?k=",
			Countries:  []string{"JP"},
			Selectors: models.SiteSelectors{
				Product:  "[data-component-type='s-search-result']",
				Price:    ".a-price-whole, .a-offscreen",
				Title:    "[data-cy='title-recipe-title'] span, h2 a span",
				Link:     "h2 a",
				Currency: ".a-price-symbol",
			},
			Headers: map[string]string{
				"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			},
			RateLimit: 2000,
		},
		{
			Name:       "Amazon Australia",
			BaseURL:    "https://www.amazon.com.au",
			SearchPath: "/s?k=",
			Countries:  []string{"AU"},
			Selectors: models.SiteSelectors{
				Product:  "[data-component-type='s-search-result']",
				Price:    ".a-price-whole, .a-offscreen",
				Title:    "[data-cy='title-recipe-title'] span, h2 a span",
				Link:     "h2 a",
				Currency: ".a-price-symbol",
			},
			Headers: map[string]string{
				"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			},
			RateLimit: 2000,
		},
		// Additional Indian Sites - Removed Amazon Fresh India (duplicate of Amazon India)
		{
			Name:       "Myntra",
			BaseURL:    "https://www.myntra.com",
			SearchPath: "/",
			Countries:  []string{"IN"},
			Selectors: models.SiteSelectors{
				Product:  ".product-base",
				Price:    ".product-discountedPrice",
				Title:    ".product-product",
				Link:     ".product-base a",
			},
			Headers: map[string]string{
				"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			},
			RateLimit: 3000,
		},
		// US Additional Sites
		{
			Name:       "Target US",
			BaseURL:    "https://www.target.com",
			SearchPath: "/s?searchTerm=",
			Countries:  []string{"US"},
			Selectors: models.SiteSelectors{
				Product:  "[data-test='product-card']",
				Price:    "[data-test='product-price']",
				Title:    "[data-test='product-title']",
				Link:     "[data-test='product-title'] a",
			},
			Headers: map[string]string{
				"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			},
			RateLimit: 2500,
		},
		{
			Name:       "Best Buy US",
			BaseURL:    "https://www.bestbuy.com",
			SearchPath: "/site/searchpage.jsp?st=",
			Countries:  []string{"US"},
			Selectors: models.SiteSelectors{
				Product:  ".sku-item",
				Price:    ".sr-price",
				Title:    ".sku-header a",
				Link:     ".sku-header a",
			},
			Headers: map[string]string{
				"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
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
	
	// Create timeout context for scraping
	scrapingCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	
	resultsChan := make(chan models.ScrapingResult, len(relevantSites))
	var wg sync.WaitGroup
	
	// Launch parallel goroutines for each website
	for _, site := range relevantSites {
		wg.Add(1)
		go func(site models.SiteConfig) {
			defer wg.Done()
			results := s.scrapeWebsiteParallel(scrapingCtx, site, query, country)
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
	
	log.Printf("Found %d raw results from %d sites before filtering", len(allResults), len(relevantSites))
	
	// Use parallel LLM processing with worker pool
	filteredResults, err := s.processResultsParallel(ctx, query, allResults)
	if err != nil {
		log.Printf("Parallel processing failed, using fallback: %v", err)
		// Fallback to fuzzy matching if LLM processing fails
		for i := range allResults {
			allResults[i].Confidence = s.matcher.FuzzyProductMatch(query, allResults[i].ProductName)
		}
		return allResults, nil
	}
	
	return filteredResults, nil
}

// FetchPricesStreaming provides real-time streaming of results as they become available
func (s *Service) FetchPricesStreaming(ctx context.Context, country, query string, resultsChan chan<- models.StreamingResult) {
	relevantSites := s.getSitesForCountry(country)
	if len(relevantSites) == 0 {
		resultsChan <- models.StreamingResult{
			Status: "error",
			Error:  fmt.Sprintf("no supported sites for country: %s", country),
		}
		return
	}
	
	// Send initial status
	resultsChan <- models.StreamingResult{
		Status:   "processing",
		Progress: 0,
		Message:  fmt.Sprintf("Starting to scrape %d websites for %s in %s", len(relevantSites), query, country),
	}
	
	// Create timeout context for scraping
	scrapingCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()
	
	var wg sync.WaitGroup
	siteResultsChan := make(chan models.ScrapingResult, len(relevantSites))
	completedSites := 0
	
	// Launch parallel goroutines for each website
	for _, site := range relevantSites {
		wg.Add(1)
		go func(site models.SiteConfig) {
			defer wg.Done()
			
			// Send processing status for this site
			resultsChan <- models.StreamingResult{
				Site:     site.Name,
				Status:   "processing",
				Progress: (completedSites * 100) / len(relevantSites),
				Message:  fmt.Sprintf("Scraping %s...", site.Name),
			}
			
			results := s.scrapeWebsiteParallel(scrapingCtx, site, query, country)
			siteResultsChan <- results
			
			// Send immediate results as they become available
			if results.Error != nil {
				resultsChan <- models.StreamingResult{
					Site:   site.Name,
					Status: "error",
					Error:  results.Error.Error(),
				}
			} else {
				// Process results through LLM if needed
				processedProducts := results.Products
				if len(processedProducts) > 0 {
					// Apply confidence scoring in smaller batches for streaming
					for i := range processedProducts {
						if processedProducts[i].Confidence == 0 {
							processedProducts[i].Confidence = s.matcher.FuzzyProductMatch(query, processedProducts[i].ProductName)
						}
					}
				}
				
				completedSites++
				resultsChan <- models.StreamingResult{
					Site:     site.Name,
					Products: processedProducts,
					Status:   "completed",
					Progress: (completedSites * 100) / len(relevantSites),
					Message:  fmt.Sprintf("Found %d products from %s", len(processedProducts), site.Name),
				}
			}
		}(site)
	}
	
	// Wait for all sites to complete
	go func() {
		wg.Wait()
		close(siteResultsChan)
	}()
	
	// Collect all results for final summary
	var allResults []models.ProductResult
	for result := range siteResultsChan {
		if result.Error == nil {
			allResults = append(allResults, result.Products...)
		}
	}
	
	// Send final completion status
	resultsChan <- models.StreamingResult{
		Status:   "completed",
		Progress: 100,
		Message:  fmt.Sprintf("Completed scraping. Found %d total products from %d sites", len(allResults), len(relevantSites)),
	}
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

// scrapeWebsiteParallel uses LLM-first approach for intelligent content extraction
func (s *Service) scrapeWebsiteParallel(ctx context.Context, site models.SiteConfig, query, country string) models.ScrapingResult {
	encodedQuery := url.QueryEscape(query)
	searchURL := site.BaseURL + site.SearchPath + encodedQuery
	
	var products []models.ProductResult
	var scrapeError error
	var pageContent string
	
	// Create a fresh collector for each request
	collector := colly.NewCollector()
	
	collector.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 2,
		Delay:       300 * time.Millisecond,
	})
	
	collector.SetRequestTimeout(10 * time.Second)
	
	// Set headers
	for key, value := range site.Headers {
		collector.OnRequest(func(r *colly.Request) {
			r.Headers.Set(key, value)
		})
	}
	
	// Extract the full page content instead of using CSS selectors
	collector.OnHTML("body", func(e *colly.HTMLElement) {
		// Get the main content area, skip navigation and footer
		mainContent := s.extractMainContent(e)
		pageContent = mainContent
	})
	
	collector.OnError(func(r *colly.Response, err error) {
		scrapeError = err
		log.Printf("Scraping error for %s: %v", site.Name, err)
	})
	
	log.Printf("Visiting %s: %s", site.Name, searchURL)
	if err := collector.Visit(searchURL); err != nil {
		scrapeError = err
		log.Printf("Visit error for %s: %v", site.Name, err)
		return models.ScrapingResult{
			Products: products,
			Site:     site.Name,
			Error:    scrapeError,
		}
	}
	
	// Use LLM to intelligently extract products from page content
	if pageContent != "" {
		extractedProducts, err := s.extractProductsWithLLM(ctx, pageContent, query, country, site.Name, site.BaseURL)
		if err != nil {
			log.Printf("LLM extraction failed for %s: %v", site.Name, err)
			// Fallback to CSS selector approach if LLM fails
			return s.fallbackCSSExtraction(ctx, site, query, country, searchURL)
		}
		products = extractedProducts
	}
	
	log.Printf("Site %s returned %d products via LLM extraction", site.Name, len(products))
	
	return models.ScrapingResult{
		Products: products,
		Site:     site.Name,
		Error:    scrapeError,
	}
}

// isGenericResult filters out generic/irrelevant results
func (s *Service) isGenericResult(title string) bool {
	genericTerms := []string{
		"shop on ebay",
		"visit store",
		"see all results",
		"more items",
		"sponsored",
		"advertisement",
		"shop now",
		"view all",
		"browse",
		"search results",
	}
	
	titleLower := strings.ToLower(title)
	for _, term := range genericTerms {
		if strings.Contains(titleLower, term) || titleLower == term {
			return true
		}
	}
	
	// Filter out titles that are too short or generic
	if len(strings.TrimSpace(title)) < 10 {
		return true
	}
	
	return false
}

// processResultsParallel handles LLM processing with worker pool pattern
func (s *Service) processResultsParallel(ctx context.Context, query string, allResults []models.ProductResult) ([]models.ProductResult, error) {
	if len(allResults) == 0 {
		return allResults, nil
	}
	
	// Create worker pool for parallel LLM processing
	numWorkers := 5 // Concurrent LLM evaluations
	jobs := make(chan models.ProductResult, len(allResults))
	results := make(chan models.ProductResult, len(allResults))
	
	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for product := range jobs {
				// Try LLM scoring first
				score, err := s.matcher.FilterAndScoreProducts(ctx, query, []models.ProductResult{product})
				if err != nil || len(score) == 0 {
					// Fallback to fuzzy matching
					product.Confidence = s.matcher.FuzzyProductMatch(query, product.ProductName)
				} else {
					product.Confidence = score[0].Confidence
				}
				results <- product
			}
		}()
	}
	
	// Send jobs
	go func() {
		for _, product := range allResults {
			jobs <- product
		}
		close(jobs)
	}()
	
	// Wait for workers to finish
	go func() {
		wg.Wait()
		close(results)
	}()
	
	// Collect results
	var processedResults []models.ProductResult
	for result := range results {
		// Apply confidence threshold - only include relevant results
		if result.Confidence >= 0.3 {
			processedResults = append(processedResults, result)
		}
	}
	
	log.Printf("Filtered %d results from %d total (%.1f%% relevant)", 
		len(processedResults), len(allResults), 
		float64(len(processedResults))/float64(len(allResults))*100)
	
	return processedResults, nil
}

// extractMainContent intelligently extracts the main product content area from a page
func (s *Service) extractMainContent(e *colly.HTMLElement) string {
	var content strings.Builder
	
	// Skip common navigation and footer areas
	skipSelectors := []string{
		"nav", "header", "footer", ".nav", ".header", ".footer",
		".navigation", ".breadcrumb", ".menu", ".sidebar",
		"[role='navigation']", "[role='banner']", "[role='contentinfo']",
	}
	
	// Focus on main content areas
	mainSelectors := []string{
		"main", ".main", "#main", ".content", "#content",
		".products", ".search-results", ".product-list",
		"[role='main']", ".container", ".results",
	}
	
	// Try to find main content area first
	for _, selector := range mainSelectors {
		mainArea := e.ChildText(selector)
		if len(mainArea) > 1000 { // Has substantial content
			return s.cleanContent(mainArea)
		}
	}
	
	// If no main area found, get all content but skip navigation
	e.ForEach("*", func(i int, child *colly.HTMLElement) {
		// Skip if it's a navigation element
		for _, skipSelector := range skipSelectors {
			if child.Name == strings.TrimPrefix(skipSelector, ".") ||
			   strings.Contains(child.Attr("class"), strings.TrimPrefix(skipSelector, ".")) {
				return
			}
		}
		
		// Add text content if it looks like product information
		text := strings.TrimSpace(child.Text)
		if len(text) > 20 && len(text) < 500 &&
		   (strings.Contains(strings.ToLower(text), "price") ||
		    strings.Contains(strings.ToLower(text), "$") ||
		    strings.Contains(strings.ToLower(text), "₹") ||
		    strings.Contains(strings.ToLower(text), "buy") ||
		    strings.Contains(strings.ToLower(text), "add to cart")) {
			content.WriteString(text + "\n")
		}
	})
	
	return s.cleanContent(content.String())
}

// cleanContent removes excessive whitespace and irrelevant content
func (s *Service) cleanContent(content string) string {
	// Remove excessive newlines and spaces
	re := regexp.MustCompile(`\s+`)
	cleaned := re.ReplaceAllString(content, " ")
	
	// Limit content size for LLM processing (around 8000 characters)
	if len(cleaned) > 8000 {
		cleaned = cleaned[:8000] + "..."
	}
	
	return strings.TrimSpace(cleaned)
}

// extractProductsWithLLM uses LLM to intelligently extract product information
func (s *Service) extractProductsWithLLM(ctx context.Context, content, query, country, siteName, baseURL string) ([]models.ProductResult, error) {
	prompt := fmt.Sprintf(`You are an expert e-commerce product extractor. Extract up to 25 relevant products from this webpage content that match the search query.

Search Query: "%s"
Country: %s
Website: %s

Webpage Content:
%s

Extract products in this exact JSON format:
{
  "products": [
    {
      "title": "Product name",
      "price": "numeric price only (no currency symbols)",
      "currency": "USD/INR/GBP/EUR/etc",
      "link": "relative or absolute URL",
      "confidence": 0.95
    }
  ]
}

Rules:
1. Only include products that actually match the search query
2. Extract exact product names from the content
3. Clean price to numbers only (remove currency symbols, commas)
4. Include relative URLs starting with / or absolute URLs
5. Confidence 0.9-1.0 for exact matches, 0.7-0.8 for good matches, 0.5-0.6 for related
6. Skip ads, navigation links, and irrelevant content
7. Focus on actual product listings with prices
8. Maximum 25 products

Respond only with valid JSON, no explanation.`, query, country, siteName, content)

	response, err := s.matcher.CallOllama(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("LLM call failed: %v", err)
	}

	// Parse JSON response
	var result struct {
		Products []struct {
			Title      string  `json:"title"`
			Price      string  `json:"price"`
			Currency   string  `json:"currency"`
			Link       string  `json:"link"`
			Confidence float64 `json:"confidence"`
		} `json:"products"`
	}

	err = json.Unmarshal([]byte(response), &result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %v", err)
	}

	var products []models.ProductResult
	for _, p := range result.Products {
		// Clean and validate the extracted data
		if p.Title == "" || p.Price == "" {
			continue
		}

		// Ensure absolute URL
		fullLink := p.Link
		if strings.HasPrefix(p.Link, "/") {
			fullLink = baseURL + p.Link
		} else if !strings.HasPrefix(p.Link, "http") {
			fullLink = baseURL + "/" + strings.TrimPrefix(p.Link, "/")
		}

		// Set default currency if missing
		currency := p.Currency
		if currency == "" {
			currency = extractCurrency(p.Price, country)
		}

		product := models.ProductResult{
			Link:        fullLink,
			Price:       cleanPriceText(p.Price),
			Currency:    currency,
			ProductName: strings.TrimSpace(p.Title),
			Site:        siteName,
			Country:     country,
			Confidence:  p.Confidence,
			FetchedAt:   time.Now(),
		}

		// Apply basic validation
		if !s.isGenericResult(product.ProductName) && product.Confidence >= 0.3 {
			products = append(products, product)
			log.Printf("LLM extracted from %s: %s - %s (confidence: %.2f)", 
				siteName, product.ProductName, product.Price, product.Confidence)
		}
	}

	return products, nil
}

// fallbackCSSExtraction provides fallback to CSS selector approach if LLM fails
func (s *Service) fallbackCSSExtraction(ctx context.Context, site models.SiteConfig, query, country, searchURL string) models.ScrapingResult {
	log.Printf("Using CSS fallback for %s", site.Name)
	
	var products []models.ProductResult
	var scrapeError error
	
	collector := colly.NewCollector()
	collector.SetRequestTimeout(10 * time.Second)
	
	for key, value := range site.Headers {
		collector.OnRequest(func(r *colly.Request) {
			r.Headers.Set(key, value)
		})
	}
	
	collector.OnHTML(site.Selectors.Product, func(e *colly.HTMLElement) {
		if len(products) >= 10 { // Reduced limit for fallback
			return
		}
		
		title := strings.TrimSpace(e.ChildText(site.Selectors.Title))
		priceText := strings.TrimSpace(e.ChildText(site.Selectors.Price))
		linkHref := e.ChildAttr(site.Selectors.Link, "href")
		
		if title != "" && priceText != "" && !s.isGenericResult(title) {
			fullLink := linkHref
			if !strings.HasPrefix(linkHref, "http") {
				fullLink = site.BaseURL + linkHref
			}
			
			product := models.ProductResult{
				Link:        fullLink,
				Price:       cleanPriceText(priceText),
				Currency:    extractCurrency(priceText, country),
				ProductName: title,
				Site:        site.Name,
				Country:     country,
				Confidence:  0.5, // Lower confidence for fallback
				FetchedAt:   time.Now(),
			}
			products = append(products, product)
		}
	})
	
	collector.Visit(searchURL)
	
	return models.ScrapingResult{
		Products: products,
		Site:     site.Name,
		Error:    scrapeError,
	}
}