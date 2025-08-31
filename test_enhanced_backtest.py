#!/usr/bin/env python3
"""
Test Enhanced Backtest Endpoint
"""

import requests
import json
import time

def test_enhanced_backtest():
    url = "http://localhost:8080/api/backtest/enhanced"
    
    # Minimal test configuration
    config = {
        "instrument": "DummyTest",
        "strategy": "VWAP EMA Crossover (9,21,20)",
        "initialCapital": 50000,
        "positionSize": 1,
        "stopLoss": 2,
        "takeProfit": 4
    }
    
    print("Testing enhanced backtest endpoint...")
    print(f"Config: {json.dumps(config, indent=2)}")
    
    try:
        start_time = time.time()
        response = requests.post(url, json=config, timeout=30)
        end_time = time.time()
        
        print(f"Response time: {end_time - start_time:.2f} seconds")
        print(f"Status code: {response.status_code}")
        
        if response.status_code == 200:
            result = response.json()
            print("✅ Enhanced backtest successful!")
            print(f"Summary: {json.dumps(result.get('summary', {}), indent=2)}")
            print(f"Trade count: {len(result.get('tradeLog', []))}")
        else:
            print(f"❌ Error: {response.status_code}")
            print(f"Response: {response.text}")
            
    except requests.exceptions.Timeout:
        print("❌ Request timed out after 30 seconds")
    except Exception as e:
        print(f"❌ Error: {str(e)}")

if __name__ == "__main__":
    test_enhanced_backtest()