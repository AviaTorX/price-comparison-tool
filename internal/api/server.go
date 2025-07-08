package api

import (
	"context"
	"io"
	"net/http"
	"price-comparison-tool/internal/config"
	"price-comparison-tool/internal/models"
	"price-comparison-tool/internal/scraper"
	"sort"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Server struct {
	config  *config.Config
	scraper *scraper.Service
	router  *gin.Engine
}

func NewServer(cfg *config.Config) *Server {
	gin.SetMode(gin.ReleaseMode)
	
	s := &Server{
		config:  cfg,
		scraper: scraper.NewService(cfg),
		router:  gin.Default(),
	}
	
	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	s.router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"*"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	api := s.router.Group("/api/v1")
	{
		api.GET("/health", s.healthCheck)
		api.POST("/prices", s.getPrices)
		api.GET("/prices/stream", s.getPricesStream)
		api.GET("/sites", s.getSupportedSites)
	}

	s.router.Static("/static", "./web/static")
	s.router.LoadHTMLGlob("web/templates/*")
	s.router.GET("/", s.indexHandler)
}

func (s *Server) Start() error {
	return s.router.Run(":" + s.config.Port)
}

func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
		"service":   "price-comparison-tool",
	})
}

func (s *Server) getPrices(c *gin.Context) {
	var req models.PriceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Add timeout context
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	results, err := s.scraper.FetchPrices(ctx, req.Country, req.Query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	sortResultsByConfidenceAndPrice(results)

	response := models.PriceResponse{
		Results: results,
		Query:   req.Query,
		Country: req.Country,
		Count:   len(results),
	}

	c.JSON(http.StatusOK, response)
}

func (s *Server) getPricesStream(c *gin.Context) {
	// Get parameters from query string for EventSource compatibility
	country := c.Query("country")
	query := c.Query("query")
	
	if country == "" || query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "country and query parameters are required"})
		return
	}
	
	req := models.PriceRequest{
		Country: country,
		Query:   query,
	}

	// Set headers for SSE (Server-Sent Events)
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// Create channel for streaming results
	resultsChan := make(chan models.StreamingResult, 10)
	
	// Add timeout context
	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	// Start price fetching in goroutine
	go func() {
		defer close(resultsChan)
		s.scraper.FetchPricesStreaming(ctx, req.Country, req.Query, resultsChan)
	}()

	// Stream results as they come in
	c.Stream(func(w io.Writer) bool {
		select {
		case result, ok := <-resultsChan:
			if !ok {
				// Channel closed, send final message
				c.SSEvent("complete", gin.H{
					"message": "All results fetched",
					"query": req.Query,
					"country": req.Country,
				})
				return false
			}
			
			// Send individual result
			c.SSEvent("result", result)
			return true
		case <-ctx.Done():
			c.SSEvent("error", gin.H{"error": "Request timeout"})
			return false
		}
	})
}

func (s *Server) getSupportedSites(c *gin.Context) {
	sites := s.scraper.GetSupportedSites()
	c.JSON(http.StatusOK, gin.H{
		"sites": sites,
		"count": len(sites),
	})
}

func (s *Server) indexHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title": "Price Comparison Tool",
	})
}

func sortResultsByConfidenceAndPrice(results []models.ProductResult) {
	sort.Slice(results, func(i, j int) bool {
		// First, sort by confidence score (descending - higher confidence first)
		if results[i].Confidence != results[j].Confidence {
			return results[i].Confidence > results[j].Confidence
		}
		
		// If confidence is the same, sort by price (ascending - lower price first)
		priceI, errI := parsePrice(results[i].Price)
		priceJ, errJ := parsePrice(results[j].Price)
		
		if errI != nil || errJ != nil {
			return false
		}
		
		return priceI < priceJ
	})
}

func parsePrice(priceStr string) (float64, error) {
	cleaned := ""
	for _, char := range priceStr {
		if (char >= '0' && char <= '9') || char == '.' {
			cleaned += string(char)
		}
	}
	return strconv.ParseFloat(cleaned, 64)
}