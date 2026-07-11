package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Default host for the Polymarket CLOB
const DefaultClobHost = "https://clob.polymarket.com"

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

type PolyMarketResponse struct {
	Question    string `json:"question"`
	ConditionID string `json:"condition_id"`
	Tokens      []struct {
		TokenID string  `json:"token_id"`
		Outcome string  `json:"outcome"`
		Price   float64 `json:"price"`
	} `json:"tokens"`
}

// MidpointResponse represents the JSON returned by /midpoint
type MidpointResponse struct {
	Mid string `json:"mid"`
}

// OrderBookResponse represents the JSON returned by /book
type OrderBookResponse struct {
	Market string  `json:"market"`
	Bids   []Level `json:"bids"`
	Asks   []Level `json:"asks"`
}

// Level represents a single price level in the order book
type Level struct {
	Price string `json:"price"`
	Size  string `json:"size"`
}

// GammaMarket maps the specific discovery data we need from Gamma
type GammaMarket struct {
	ID          string `json:"id"`
	Question    string `json:"question"`
	ConditionID string `json:"conditionId"`
	Active      bool   `json:"active"`
	Closed      bool   `json:"closed"`
}

func searchPolymarketMarkets(sportsNames []string, searchTerm string) ([]GammaMarket, error) {
	// searchPolymarketMarkets queries Gamma to find condition IDs programmatically

	client := &http.Client{Timeout: 10 * time.Second}

	baseURL := "https://gamma-api.polymarket.com/markets"

	params := url.Values{}
	if searchTerm != "" {
		params.Add("search", searchTerm)
	}
	params.Add("active", "true")
	params.Add("closed", "false")

	for _, s := range sportsNames {
		params.Add("tag_id", sports[s])
	}

	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	req, _ := http.NewRequest(http.MethodGet, fullURL, nil)
	req.Header.Set("Accept", "application/json")

	fmt.Printf("Searching Gamma API for: '%s'...\n", searchTerm)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		body, _ := io.ReadAll(resp.Body)

		var markets []GammaMarket
		if err := json.Unmarshal(body, &markets); err != nil {
			return nil, fmt.Errorf("failed to parse Gamma JSON: %w", err)
		}

		return markets, nil
	}

	return nil, fmt.Errorf("failed with status: %d", resp.StatusCode)
}

func getPolymarketMarket(conditionID string) (*PolyMarketResponse, error) {
	// Retrieve market data from a given condition ID

	client := &http.Client{Timeout: 10 * time.Second}

	url := fmt.Sprintf("%s/markets/%s", DefaultClobHost, conditionID)

	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		body, _ := io.ReadAll(resp.Body)

		var data PolyMarketResponse
		if err := json.Unmarshal(body, &data); err != nil {
			return nil, err
		}
		return &data, nil
	}

	return nil, fmt.Errorf("failed with status: %d", resp.StatusCode)
}

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
