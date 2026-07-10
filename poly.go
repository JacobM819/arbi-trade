package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Default host for the Polymarket CLOB
const DefaultClobHost = "https://clob.polymarket.com"

// ---------------------------------------------------------
// 1. Client Initialization
// ---------------------------------------------------------

// ClobClient holds the configuration for Polymarket API interaction
type ClobClient struct {
	Host       string
	HTTPClient *http.Client
}

// NewClobClient initializes a new Polymarket API client
func NewClobClient(host string) *ClobClient {
	if host == "" {
		host = DefaultClobHost
	}
	return &ClobClient{
		Host: host,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// ---------------------------------------------------------
// 2. Data Structures
// ---------------------------------------------------------

// MidpointResponse represents the JSON returned by /midpoint
type MidpointResponse struct {
	Mid string `json:"mid"`
}

// OrderBookResponse represents the JSON returned by /book
type OrderBookResponse struct {
	Market string `json:"market"`
	Bids   []Level `json:"bids"`
	Asks   []Level `json:"asks"`
}

// Level represents a single price level in the order book
type Level struct {
	Price string `json:"price"`
	Size  string `json:"size"`
}

// ---------------------------------------------------------
// 3. API Fetching Methods
// ---------------------------------------------------------

// GetMidpoint retrieves the current implied probability of a token
func (c *ClobClient) GetMidpoint(tokenID string) (*MidpointResponse, error) {
	url := fmt.Sprintf("%s/midpoint?token_id=%s", c.Host, tokenID)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data MidpointResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	return &data, nil
}

// GetOrderBook retrieves the full depth of bids and asks for a token
func (c *ClobClient) GetOrderBook(tokenID string) (*OrderBookResponse, error) {
	url := fmt.Sprintf("%s/book?token_id=%s", c.Host, tokenID)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data OrderBookResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	return &data, nil
}

