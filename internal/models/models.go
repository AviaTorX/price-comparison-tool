package models

import "time"

type PriceRequest struct {
	Country string `json:"country" binding:"required"`
	Query   string `json:"query" binding:"required"`
}

type ProductResult struct {
	Link        string    `json:"link"`
	Price       string    `json:"price"`
	Currency    string    `json:"currency"`
	ProductName string    `json:"productName"`
	Site        string    `json:"site"`
	Country     string    `json:"country"`
	Confidence  float64   `json:"confidence,omitempty"`
	FetchedAt   time.Time `json:"fetchedAt"`
}

type PriceResponse struct {
	Results []ProductResult `json:"results"`
	Query   string          `json:"query"`
	Country string          `json:"country"`
	Count   int             `json:"count"`
}

type SiteConfig struct {
	Name           string            `json:"name"`
	BaseURL        string            `json:"baseUrl"`
	SearchPath     string            `json:"searchPath"`
	Countries      []string          `json:"countries"`
	Selectors      SiteSelectors     `json:"selectors"`
	Headers        map[string]string `json:"headers,omitempty"`
	RateLimit      int               `json:"rateLimit,omitempty"`
	RequiresJS     bool              `json:"requiresJs,omitempty"`
}

type SiteSelectors struct {
	Product     string `json:"product"`
	Price       string `json:"price"`
	Title       string `json:"title"`
	Link        string `json:"link"`
	Currency    string `json:"currency,omitempty"`
}

type ScrapingResult struct {
	Products []ProductResult
	Site     string
	Error    error
}

type StreamingResult struct {
	Site       string          `json:"site"`
	Products   []ProductResult `json:"products,omitempty"`
	Status     string          `json:"status"` // "processing", "completed", "error"
	Error      string          `json:"error,omitempty"`
	Progress   int             `json:"progress"` // 0-100
	Message    string          `json:"message,omitempty"`
}