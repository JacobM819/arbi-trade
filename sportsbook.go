package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

// Payload structure to match FanDuel's required JSON body
type Payload struct {
	MarketIds []string `json:"marketIds"`
}

// fetchSportsbookOdds fetches real-time odds from FanDuel's internal JSON API.
func fetchSportsbookOdds(marketIds []string) ([]interface{}, error) {
	// 1. Initialize a Client to bypass Cloudflare TLS fingerprinting
	options := []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(10),
		tls_client.WithClientProfile(profiles.Chrome_120), // Impersonate Chrome perfectly
		// tls_client.WithProxyUrl("http://user:pass@host:port"), // Uncomment to add proxy
	}

	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	if err != nil {
		return nil, fmt.Errorf("failed to construct client: %w", err)
	}

	// 2. The exact FanDuel PA endpoint
	url := "https://smp.pa.sportsbook.fanduel.com/api/sports/fixedodds/readonly/v1/getMarketPrices?priceHistory=1"

	// 3. Construct the POST payload
	payloadStruct := Payload{
		MarketIds: marketIds,
	}
	payloadBytes, err := json.Marshal(payloadStruct)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	// 4. Create the request using the fhttp library (not standard net/http)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 5. Inject the mandatory headers
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-application", "FhMFpcPWXMeyZxOx")
	req.Header.Set("Origin", "https://sportsbook.fanduel.com")
	req.Header.Set("Referer", "https://sportsbook.fanduel.com/")

	fmt.Printf("Fetching FanDuel odds for markets: %v...\n", marketIds)

	// 6. Execute the POST request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// 7. Handle the response
	switch resp.StatusCode {
	case 200:
		fmt.Println("Successfully retrieved market prices!")
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read body: %w", err)
		}

		// Decode into a generic map (you can create struct models for this later for better type safety)
		var data []interface{}
		if err := json.Unmarshal(body, &data); err != nil {
			return nil, fmt.Errorf("failed to parse JSON: %w", err)
		}
		return data, nil

	case 401, 403:
		fmt.Printf("Blocked! Status: %d. The x-application key may have expired or Cloudflare blocked the IP.\n", resp.StatusCode)
		return nil, nil

	case 429:
		fmt.Println("Rate Limited! You are sending too many requests too fast.")
		return nil, nil

	default:
		fmt.Printf("Failed with status: %d\n", resp.StatusCode)
		return nil, nil
	}
}
