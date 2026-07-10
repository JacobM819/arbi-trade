import os
import time
import requests
import logging
from datetime import datetime
from py_clob_client_v2 import ClobClient, OrderArgs, OrderType, Side, MarketOrderArgs

# --- Configuration & Safety Limits ---
POLYMARKET_HOST = "https://clob.polymarket.com"
CHAIN_ID = 137  # Polygon Mainnet
PRIVATE_KEY = os.environ.get("PK")
SPORTSBOOK_API_URL = os.environ.get("SPORTSBOOK_API_URL")

# The wallet cap: Maximum pUSD to risk per calendar day
DAILY_SPEND_CAP_USD = float(os.environ.get("DAILY_SPEND_CAP", 50.0))
PROFIT_MARGIN_THRESHOLD = 0.05  # Trigger trade if Polymarket is 5% cheaper than the sportsbook

# --- State Management ---
current_date = datetime.now().date()
spent_today = 0.0

logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')

def initialize_polymarket():
    """Initializes the authenticated L1/L2 Polymarket CLOB client."""
    client = ClobClient(host=POLYMARKET_HOST, chain_id=CHAIN_ID, key=PRIVATE_KEY)
    creds = client.create_or_derive_api_key()
    # Fully authenticated client
    return ClobClient(host=POLYMARKET_HOST, chain_id=CHAIN_ID, key=PRIVATE_KEY, creds=creds)

def fetch_sportsbook_implied_prob(market_id):
    """
    Mock function: Fetches sharp odds from your external sportsbook.
    Returns the implied probability (0.0 to 1.0).
    """
    # response = requests.get(f"{SPORTSBOOK_API_URL}/odds?market={market_id}")
    # data = response.json()
    # return convert_american_to_implied(data['price'])
    return 0.65  # e.g., Sportsbook prices outcome at 65%

def can_trade(stake_amount):
    """Enforces the daily wallet cap, resetting at local midnight."""
    global current_date, spent_today
    
    today = datetime.now().date()
    if today > current_date:
        logging.info("New day. Resetting daily spend cap.")
        current_date = today
        spent_today = 0.0

    if spent_today + stake_amount > DAILY_SPEND_CAP_USD:
        logging.warning(f"BLOCKED: Trade of ${stake_amount} exceeds daily cap of ${DAILY_SPEND_CAP_USD}. Spent today: ${spent_today}")
        return False
        
    return True

def execute_arbitrage(client, token_id, poly_price, sharp_prob):
    """Evaluates the discrepancy and executes a market order if profitable."""
    global spent_today

    discrepancy = sharp_prob - poly_price
    
    if discrepancy >= PROFIT_MARGIN_THRESHOLD:
        # Calculate dynamic stake (e.g., fractional Kelly), capping at max $10 for this example
        stake = 10.0 
        
        if can_trade(stake):
            logging.info(f"Discrepancy found: Poly ({poly_price}) vs Book ({sharp_prob}). Executing BUY.")
            
            try:
                # Market Buy using FOK (Fill or Kill) to avoid partial fills during slippage
                resp = client.create_and_post_market_order(
                    order_args=MarketOrderArgs(
                        token_id=token_id,
                        amount=stake,
                        side=Side.BUY
                    ),
                    order_type=OrderType.FOK,
                )
                logging.info(f"Order filled: {resp}")
                spent_today += stake
                logging.info(f"Updated Daily Spend: ${spent_today} / ${DAILY_SPEND_CAP_USD}")
                
            except Exception as e:
                logging.error(f"Trade failed: {e}")
    else:
        logging.info("Markets are efficient. No trade.")

def main():
    client = initialize_polymarket()
    logging.info("Polymarket client authenticated successfully.")
    
    # Example token ID for a specific market outcome (e.g., Team A to win)
    target_token_id = "0xYourMarketTokenIdHere"
    
    while True:
        try:
            # 1. Fetch current midpoint price on Polymarket
            poly_price_data = client.get_midpoint(target_token_id)
            poly_price = float(poly_price_data.get('midpoint', 0))
            
            # 2. Fetch "Source of Truth" from Sportsbook
            sharp_prob = fetch_sportsbook_implied_prob(target_token_id)
            
            # 3. Evaluate and Trade
            if poly_price > 0:
                execute_arbitrage(client, target_token_id, poly_price, sharp_prob)
                
        except Exception as e:
            logging.error(f"Error in main loop: {e}")
            
        # Rate limit the loop to avoid API bans
        time.sleep(5)

if __name__ == "__main__":
    main()
