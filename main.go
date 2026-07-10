package main

import (
	//"encoding/json"
	"fmt"
	//"log"
)

func main() {
	// //Fandual
	// testMarketIds := []string{"742.175813577", "742.175813576", "742.175813584", "742.175813581", "742.175813586", "742.175813580"}
	//
	// oddsData, err := fetchSportsbookOdds(testMarketIds)
	// if err != nil {
	// 	log.Fatalf("Fatal error: %v", err)
	// }
	//
	// if oddsData != nil {
	// 	// Print a small snippet of the response to verify it works
	// 	jsonData, err := json.MarshalIndent(oddsData, "", "  ")
	// 	if err != nil {
	// 		log.Fatalf("Error marshaling JSON for print: %v", err)
	// 	}
	//
	// 	output := string(jsonData)
	// 	if len(output) > 1000 {
	// 		output = output[:1000] + "\n... [truncated]"
	// 	}
	// 	fmt.Println(output)
	// }

	// Polymarket

	client := NewClobClient("")

	// Example target: Replace with an active Token ID
	testTokenID := "89453416559360575370389185597493481042047220643472797002101793932914111599485"

	// Fetch 1: Midpoint (Implied Probability)
	fmt.Printf("Fetching Midpoint for %s...\n", testTokenID[:10]+"...")
	midpoint, err := client.GetMidpoint(testTokenID)
	if err != nil {
		fmt.Printf("Error fetching midpoint: %v\n", err)
	} else {
		fmt.Printf("Current Midpoint Price: %s\n\n", midpoint.Mid)
	}

	// Fetch 2: Full Order Book (Liquidity/Slippage Check)
	fmt.Println("Fetching Order Book...")
	orderBook, err := client.GetOrderBook(testTokenID)
	if err != nil {
		fmt.Printf("Error fetching order book: %v\n", err)
	} else {
		fmt.Println("Top 3 Asks (Prices you can buy at):")
		limit := 3
		if len(orderBook.Asks) < 3 {
			limit = len(orderBook.Asks)
		}
		for i := 0; i < limit; i++ {
			fmt.Printf("Price: %s | Available Size: %s shares\n", orderBook.Asks[i].Price, orderBook.Asks[i].Size)
		}
	}
}
