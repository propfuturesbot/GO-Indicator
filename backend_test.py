#!/usr/bin/env python3
"""
Comprehensive Backend API Test for Technical Trading Platform
Tests all indicator endpoints and functionality
"""

import requests
import json
import sys
from datetime import datetime
import time

class TradingPlatformTester:
    def __init__(self, base_url="http://localhost:8080"):
        self.base_url = base_url
        self.tests_run = 0
        self.tests_passed = 0
        self.failed_tests = []
        
        # Sample market data for testing indicators
        self.sample_data = [
            {"time": 1692000000, "open": 19000, "high": 19100, "low": 18950, "close": 19050, "volume": 1000},
            {"time": 1692003600, "open": 19050, "high": 19200, "low": 19000, "close": 19150, "volume": 1100},
            {"time": 1692007200, "open": 19150, "high": 19300, "low": 19100, "close": 19250, "volume": 1200},
            {"time": 1692010800, "open": 19250, "high": 19400, "low": 19200, "close": 19350, "volume": 1300},
            {"time": 1692014400, "open": 19350, "high": 19500, "low": 19300, "close": 19450, "volume": 1400},
            {"time": 1692018000, "open": 19450, "high": 19600, "low": 19400, "close": 19550, "volume": 1500},
            {"time": 1692021600, "open": 19550, "high": 19700, "low": 19500, "close": 19650, "volume": 1600},
            {"time": 1692025200, "open": 19650, "high": 19800, "low": 19600, "close": 19750, "volume": 1700},
            {"time": 1692028800, "open": 19750, "high": 19900, "low": 19700, "close": 19850, "volume": 1800},
            {"time": 1692032400, "open": 19850, "high": 20000, "low": 19800, "close": 19950, "volume": 1900},
            {"time": 1692036000, "open": 19950, "high": 19900, "low": 19800, "close": 19850, "volume": 2000},
            {"time": 1692039600, "open": 19850, "high": 19800, "low": 19700, "close": 19750, "volume": 2100},
            {"time": 1692043200, "open": 19750, "high": 19700, "low": 19600, "close": 19650, "volume": 2200},
            {"time": 1692046800, "open": 19650, "high": 19600, "low": 19500, "close": 19550, "volume": 2300},
            {"time": 1692050400, "open": 19550, "high": 19500, "low": 19400, "close": 19450, "volume": 2400},
            {"time": 1692054000, "open": 19450, "high": 19600, "low": 19400, "close": 19550, "volume": 2500},
            {"time": 1692057600, "open": 19550, "high": 19700, "low": 19500, "close": 19650, "volume": 2600},
            {"time": 1692061200, "open": 19650, "high": 19800, "low": 19600, "close": 19750, "volume": 2700},
            {"time": 1692064800, "open": 19750, "high": 19900, "low": 19700, "close": 19850, "volume": 2800},
            {"time": 1692068400, "open": 19850, "high": 20000, "low": 19800, "close": 19950, "volume": 2900}
        ]

    def run_test(self, name, method, endpoint, expected_status, data=None, validate_response=None):
        """Run a single API test"""
        url = f"{self.base_url}/{endpoint}"
        headers = {'Content-Type': 'application/json'}

        self.tests_run += 1
        print(f"\n🔍 Testing {name}...")
        
        try:
            if method == 'GET':
                response = requests.get(url, headers=headers, timeout=10)
            elif method == 'POST':
                response = requests.post(url, json=data, headers=headers, timeout=10)
            else:
                raise ValueError(f"Unsupported method: {method}")

            success = response.status_code == expected_status
            
            if success:
                try:
                    response_data = response.json()
                    
                    # Additional validation if provided
                    if validate_response:
                        validation_result = validate_response(response_data)
                        if not validation_result:
                            success = False
                            print(f"❌ Failed - Response validation failed")
                        else:
                            print(f"✅ Passed - Status: {response.status_code}, Response validated")
                    else:
                        print(f"✅ Passed - Status: {response.status_code}")
                    
                    if success:
                        self.tests_passed += 1
                    else:
                        self.failed_tests.append(name)
                    
                    return success, response_data
                    
                except json.JSONDecodeError:
                    print(f"❌ Failed - Invalid JSON response")
                    self.failed_tests.append(name)
                    return False, {}
            else:
                print(f"❌ Failed - Expected {expected_status}, got {response.status_code}")
                print(f"   Response: {response.text[:200]}")
                self.failed_tests.append(name)
                return False, {}

        except requests.exceptions.RequestException as e:
            print(f"❌ Failed - Request error: {str(e)}")
            self.failed_tests.append(name)
            return False, {}
        except Exception as e:
            print(f"❌ Failed - Unexpected error: {str(e)}")
            self.failed_tests.append(name)
            return False, {}

    def test_health_check(self):
        """Test API health endpoint"""
        def validate_health(data):
            return 'status' in data and data['status'] == 'OK' and 'time' in data
        
        return self.run_test(
            "Health Check",
            "GET",
            "api/health",
            200,
            validate_response=validate_health
        )

    def test_get_strategies(self):
        """Test get available strategies endpoint"""
        def validate_strategies(data):
            return isinstance(data, list) and len(data) > 0
        
        return self.run_test(
            "Get Available Strategies",
            "GET",
            "api/strategies",
            200,
            validate_response=validate_strategies
        )

    def test_sma_indicator(self):
        """Test Simple Moving Average indicator"""
        def validate_sma(data):
            return 'values' in data and isinstance(data['values'], list) and len(data['values']) > 0
        
        return self.run_test(
            "SMA Indicator (period=5)",
            "POST",
            "api/indicators/sma",
            200,
            data={"data": self.sample_data, "period": 5},
            validate_response=validate_sma
        )

    def test_ema_indicator(self):
        """Test Exponential Moving Average indicator"""
        def validate_ema(data):
            return 'values' in data and isinstance(data['values'], list) and len(data['values']) > 0
        
        return self.run_test(
            "EMA Indicator (period=10)",
            "POST",
            "api/indicators/ema",
            200,
            data={"data": self.sample_data, "period": 10},
            validate_response=validate_ema
        )

    def test_rsi_indicator(self):
        """Test RSI indicator"""
        def validate_rsi(data):
            if 'values' not in data or not isinstance(data['values'], list):
                return False
            # RSI values should be between 0 and 100
            return all(0 <= val <= 100 for val in data['values'])
        
        return self.run_test(
            "RSI Indicator (period=14)",
            "POST",
            "api/indicators/rsi",
            200,
            data={"data": self.sample_data, "period": 14},
            validate_response=validate_rsi
        )

    def test_bollinger_bands(self):
        """Test Bollinger Bands indicator"""
        def validate_bb(data):
            required_keys = ['upper', 'middle', 'lower']
            if not all(key in data for key in required_keys):
                return False
            return all(isinstance(data[key], list) and len(data[key]) > 0 for key in required_keys)
        
        return self.run_test(
            "Bollinger Bands (period=20)",
            "POST",
            "api/indicators/bollinger-bands",
            200,
            data={"data": self.sample_data, "period": 20},
            validate_response=validate_bb
        )

    def test_macd_indicator(self):
        """Test MACD indicator"""
        def validate_macd(data):
            # MACD returns upper (MACD line) and lower (signal line)
            return 'upper' in data and 'lower' in data and \
                   isinstance(data['upper'], list) and isinstance(data['lower'], list)
        
        return self.run_test(
            "MACD Indicator",
            "POST",
            "api/indicators/macd",
            200,
            data={"data": self.sample_data},
            validate_response=validate_macd
        )

    def test_advanced_indicators(self):
        """Test advanced indicators if available"""
        advanced_tests = [
            ("DEMA", "api/indicators/dema", {"data": self.sample_data}),
            ("TEMA", "api/indicators/tema", {"data": self.sample_data}),
            ("WMA", "api/indicators/wma", {"data": self.sample_data, "period": 10}),
            ("HMA", "api/indicators/hma", {"data": self.sample_data, "period": 10}),
            ("KAMA", "api/indicators/kama", {"data": self.sample_data}),
            ("Aroon", "api/indicators/aroon", {"data": self.sample_data}),
            ("CCI", "api/indicators/cci", {"data": self.sample_data, "period": 20}),
            ("Stochastic", "api/indicators/stochastic", {"data": self.sample_data}),
            ("Williams %R", "api/indicators/williams-r", {"data": self.sample_data, "period": 14}),
            ("Awesome Oscillator", "api/indicators/awesome-oscillator", {"data": self.sample_data}),
            ("Donchian Channel", "api/indicators/donchian-channel", {"data": self.sample_data, "period": 20}),
            ("Keltner Channel", "api/indicators/keltner-channel", {"data": self.sample_data}),
            ("ATR", "api/indicators/atr", {"data": self.sample_data, "period": 14}),
            ("OBV", "api/indicators/obv", {"data": self.sample_data}),
            ("CMF", "api/indicators/cmf", {"data": self.sample_data, "period": 20}),
            ("MFI", "api/indicators/mfi", {"data": self.sample_data}),
            ("A/D", "api/indicators/ad", {"data": self.sample_data}),
            ("VWAP", "api/indicators/vwap", {"data": self.sample_data, "period": 14})
        ]
        
        results = []
        for name, endpoint, data in advanced_tests:
            def validate_indicator(response_data):
                # Basic validation - should have values or bands
                return ('values' in response_data and isinstance(response_data['values'], list)) or \
                       ('upper' in response_data and 'lower' in response_data)
            
            success, _ = self.run_test(
                f"{name} Indicator",
                "POST",
                endpoint,
                200,
                data=data,
                validate_response=validate_indicator
            )
            results.append((name, success))
        
        return results

    def test_error_handling(self):
        """Test API error handling"""
        print(f"\n🔍 Testing Error Handling...")
        
        # Test with invalid data
        success, _ = self.run_test(
            "Invalid Data Handling",
            "POST",
            "api/indicators/sma",
            400,
            data={"invalid": "data"}
        )
        
        # Test with empty data
        success2, _ = self.run_test(
            "Empty Data Handling",
            "POST",
            "api/indicators/sma",
            400,
            data={"data": []}
        )
        
        return success and success2

    def run_all_tests(self):
        """Run all backend API tests"""
        print("🚀 Starting Comprehensive Backend API Tests")
        print("=" * 60)
        
        start_time = time.time()
        
        # Core API tests
        self.test_health_check()
        self.test_get_strategies()
        
        # Basic indicator tests
        self.test_sma_indicator()
        self.test_ema_indicator()
        self.test_rsi_indicator()
        self.test_bollinger_bands()
        self.test_macd_indicator()
        
        # Advanced indicator tests
        print(f"\n📊 Testing Advanced Indicators...")
        advanced_results = self.test_advanced_indicators()
        
        # Error handling tests
        self.test_error_handling()
        
        end_time = time.time()
        
        # Print summary
        print("\n" + "=" * 60)
        print("📊 TEST SUMMARY")
        print("=" * 60)
        print(f"⏱️  Total time: {end_time - start_time:.2f} seconds")
        print(f"✅ Tests passed: {self.tests_passed}/{self.tests_run}")
        print(f"❌ Tests failed: {len(self.failed_tests)}")
        
        if self.failed_tests:
            print(f"\n❌ Failed tests:")
            for test in self.failed_tests:
                print(f"   - {test}")
        
        # Advanced indicators summary
        print(f"\n📈 Advanced Indicators Results:")
        for name, success in advanced_results:
            status = "✅" if success else "❌"
            print(f"   {status} {name}")
        
        success_rate = (self.tests_passed / self.tests_run) * 100 if self.tests_run > 0 else 0
        print(f"\n🎯 Success Rate: {success_rate:.1f}%")
        
        if success_rate >= 80:
            print("🎉 Backend API is working well!")
            return 0
        elif success_rate >= 60:
            print("⚠️  Backend API has some issues but core functionality works")
            return 1
        else:
            print("🚨 Backend API has significant issues")
            return 2

def main():
    """Main test execution"""
    print("Technical Trading Platform - Backend API Test")
    print("Testing Go API server on localhost:8080")
    print()
    
    tester = TradingPlatformTester()
    return tester.run_all_tests()

if __name__ == "__main__":
    sys.exit(main())