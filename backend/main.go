// Copyright (c) 2021-2024 Onur Cinar.
// Web API Server for Technical Indicators Trading Platform

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	""
	"time"

	"github.com/cinar/indicator/v2/helper"
	"github.com/cinar/indicator/v2/trend"
	"github.com/cinar/indicator/v2/momentum"
	"github.com/cinar/indicator/v2/volatility"
	"github.com/cinar/indicator/v2/volume"
	"github.com/cinar/indicator/v2/strategy"
	""
	"github.com/cinar/indicator/v2/asset"
	
	"github.com/rs/cors"
	"github.com/gorilla/mux"
)

// OHLCV represents the standard candlestick data
type OHLCV struct {
	Time   int64   `json:"time"`
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Volume float64 `json:"volume"`
}

// IndicatorRequest represents a request for indicator calculation
type IndicatorRequest struct {
	Data   []OHLCV `json:"data"`
	Period int     `json:"period,omitempty"`
	Config map[string]interface{} `json:"config,omitempty"`
}

// IndicatorResponse represents the response with calculated indicator values
type IndicatorResponse struct {
	Values []float64 `json:"values"`
	Error  string    `json:"error,omitempty"`
}

// BandsIndicatorResponse represents response for multi-line indicators (like Bollinger Bands)
type BandsIndicatorResponse struct {
	Upper  []float64 `json:"upper"`
	Middle []float64 `json:"middle"`
	Lower  []float64 `json:"lower"`
	Error  string    `json:"error,omitempty"`
}

// BacktestRequest represents a request for backtesting
type BacktestRequest struct {
	Data     []OHLCV `json:"data"`
	Strategy string  `json:"strategy"`
	Config   map[string]interface{} `json:"config,omitempty"`
}

// BacktestResponse represents backtest results
type BacktestResponse struct {
	Actions       []int     `json:"actions"` // -1 = Sell, 0 = Hold, 1 = Buy
	Outcome       float64   `json:"outcome"`
	TotalTrades   int       `json:"totalTrades"`
	WinRate       float64   `json:"winRate"`
	ProfitFactor  float64   `json:"profitFactor"`
	Error         string    `json:"error,omitempty"`
}

func main() {
	r := mux.NewRouter()
	
	// API routes
	api := r.PathPrefix("/api").Subrouter()
	
	// Trend indicators
	api.HandleFunc("/indicators/sma", handleSMA).Methods("POST")
	api.HandleFunc("/indicators/ema", handleEMA).Methods("POST")
	api.HandleFunc("/indicators/dema", handleDEMA).Methods("POST")
	api.HandleFunc("/indicators/tema", handleTEMA).Methods("POST")
	api.HandleFunc("/indicators/wma", handleWMA).Methods("POST")
	api.HandleFunc("/indicators/hma", handleHMA).Methods("POST")
	api.HandleFunc("/indicators/kama", handleKAMA).Methods("POST")
	api.HandleFunc("/indicators/macd", handleMACD).Methods("POST")
	api.HandleFunc("/indicators/aroon", handleAroon).Methods("POST")
	api.HandleFunc("/indicators/cci", handleCCI).Methods("POST")
	
	// Momentum indicators  
	api.HandleFunc("/indicators/rsi", handleRSI).Methods("POST")
	api.HandleFunc("/indicators/stochastic", handleStochastic).Methods("POST")
	api.HandleFunc("/indicators/williams-r", handleWilliamsR).Methods("POST")
	api.HandleFunc("/indicators/awesome-oscillator", handleAwesomeOscillator).Methods("POST")
	
	// Volatility indicators
	api.HandleFunc("/indicators/bollinger-bands", handleBollingerBands).Methods("POST")
	api.HandleFunc("/indicators/donchian-channel", handleDonchianChannel).Methods("POST")
	api.HandleFunc("/indicators/keltner-channel", handleKeltnerChannel).Methods("POST")
	api.HandleFunc("/indicators/atr", handleATR).Methods("POST")
	
	// Volume indicators
	api.HandleFunc("/indicators/obv", handleOBV).Methods("POST")
	api.HandleFunc("/indicators/cmf", handleCMF).Methods("POST")
	api.HandleFunc("/indicators/mfi", handleMFI).Methods("POST")
	api.HandleFunc("/indicators/ad", handleAD).Methods("POST")
	api.HandleFunc("/indicators/vwap", handleVWAP).Methods("POST")
	
	// Strategy and backtesting
	api.HandleFunc("/backtest", handleBacktest).Methods("POST")
	api.HandleFunc("/backtest/enhanced", handleEnhancedBacktest).Methods("POST")
	api.HandleFunc("/strategies", handleGetStrategies).Methods("GET")
	api.HandleFunc("/strategies/list", handleGetStrategiesList).Methods("GET")
	api.HandleFunc("/strategies/details/{name}", handleGetStrategyDetails).Methods("GET")
	
	// Data generation
	api.HandleFunc("/data/nq-dummy", handleGenerateNQData).Methods("POST")
	
	// Health check
	api.HandleFunc("/health", handleHealth).Methods("GET")
	
	// CORS middleware
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"*"},
		AllowCredentials: true,
	})
	
	handler := c.Handler(r)
	
	fmt.Println("🚀 Trading Platform API Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}

// StrategyInfo represents strategy information for the UI
type StrategyInfo struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
	Category    string                 `json:"category"`
	WinRate     float64                `json:"winRate"`
	ProfitFactor float64               `json:"profitFactor"`
	MaxDrawdown float64                `json:"maxDrawdown"`
}

// Enhanced Backtest endpoint
func handleEnhancedBacktest(w http.ResponseWriter, r *http.Request) {
	var config BacktestConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Use dummy data if instrument is "DummyTest"
	var data []OHLCV
	if config.Instrument == "DummyTest" {
		data = GenerateNQSampleData()
	} else {
		http.Error(w, "Only DummyTest instrument supported currently", http.StatusBadRequest)
		return
	}

	// Run enhanced backtest
	result, err := RunEnhancedBacktest(data, config)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// Get strategies list with detailed information
func handleGetStrategiesList(w http.ResponseWriter, r *http.Request) {
	strategies := []StrategyInfo{
		{
			Name:         "VWAP EMA Crossover (9,21,20)",
			Description:  "Combines VWAP, fast EMA (9), and slow EMA (21) for trend-following entries with volume confirmation",
			Category:     "Trend Following",
			WinRate:      65.0,
			ProfitFactor: 1.8,
			MaxDrawdown:  12.5,
			Parameters: map[string]interface{}{
				"emaFastPeriod":      9,
				"emaSlowPeriod":      21,
				"vwapPeriod":         20,
				"scalpTarget":        5.0,
				"stopLoss":           10.0,
				"minVolumeThreshold": 1000,
			},
		},
		{
			Name:         "Donchian Channel Breakout (20)",
			Description:  "Breakout strategy using 20-period Donchian Channel for trend momentum entries",
			Category:     "Breakout",
			WinRate:      58.0,
			ProfitFactor: 2.1,
			MaxDrawdown:  18.0,
			Parameters: map[string]interface{}{
				"period": 20,
			},
		},
		{
			Name:         "RSI Divergence Strategy (14)",
			Description:  "Mean reversion strategy using RSI overbought/oversold levels with divergence detection",
			Category:     "Mean Reversion",
			WinRate:      72.0,
			ProfitFactor: 1.6,
			MaxDrawdown:  8.5,
			Parameters: map[string]interface{}{
				"rsiPeriod":       14,
				"overboughtLevel": 70,
				"oversoldLevel":   30,
			},
		},
		{
			Name:         "Moving Average Crossover (10,30)",
			Description:  "Classic golden cross/death cross strategy using fast and slow moving averages",
			Category:     "Trend Following",
			WinRate:      60.0,
			ProfitFactor: 1.7,
			MaxDrawdown:  15.0,
			Parameters: map[string]interface{}{
				"fastPeriod": 10,
				"slowPeriod": 30,
			},
		},
		{
			Name:         "Bollinger Mean Reversion (20)",
			Description:  "Mean reversion strategy using Bollinger Band touches for contrarian entries",
			Category:     "Mean Reversion",
			WinRate:      68.0,
			ProfitFactor: 1.9,
			MaxDrawdown:  11.0,
			Parameters: map[string]interface{}{
				"period":       20,
				"stdDeviation": 2.0,
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(strategies)
}

// Get strategy details
func handleGetStrategyDetails(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	strategyName := vars["name"]

	// Find strategy in the list
	strategies := []StrategyInfo{
		// Same strategies as above - in production, this would be refactored to avoid duplication
		{
			Name:         "VWAP EMA Crossover (9,21,20)",
			Description:  "Combines VWAP, fast EMA (9), and slow EMA (21) for trend-following entries with volume confirmation. Entry conditions: price > VWAP, EMA fast crosses above EMA slow, volume > threshold. Exit conditions: opposite signals, stop loss, or take profit.",
			Category:     "Trend Following",
			WinRate:      65.0,
			ProfitFactor: 1.8,
			MaxDrawdown:  12.5,
			Parameters: map[string]interface{}{
				"emaFastPeriod":      9,
				"emaSlowPeriod":      21,
				"vwapPeriod":         20,
				"scalpTarget":        5.0,
				"stopLoss":           10.0,
				"minVolumeThreshold": 1000,
			},
		},
	}

	for _, strategy := range strategies {
		if strategy.Name == strategyName {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(strategy)
			return
		}
	}

	http.Error(w, "Strategy not found", http.StatusNotFound)
}

// Generate NQ dummy data
func handleGenerateNQData(w http.ResponseWriter, r *http.Request) {
	var request struct {
		StartDate string `json:"startDate"`
		EndDate   string `json:"endDate"`
		Interval  string `json:"interval"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// For now, just return the sample data regardless of parameters
	data := GenerateNQSampleData()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":        data,
		"totalBars":   len(data),
		"startTime":   data[0].Time,
		"endTime":     data[len(data)-1].Time,
		"instrument":  "NQ",
		"timeframe":   "1m",
	})
}

// Health check endpoint
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "OK",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// Helper function to convert OHLCV data to channels
func dataToChannels(data []OHLCV) (<-chan float64, <-chan float64, <-chan float64, <-chan float64, <-chan float64) {
	opens := make([]float64, len(data))
	highs := make([]float64, len(data))
	lows := make([]float64, len(data))
	closes := make([]float64, len(data))
	volumes := make([]float64, len(data))
	
	for i, d := range data {
		opens[i] = d.Open
		highs[i] = d.High
		lows[i] = d.Low
		closes[i] = d.Close
		volumes[i] = d.Volume
	}
	
	return helper.SliceToChan(opens),
		   helper.SliceToChan(highs),
		   helper.SliceToChan(lows),
		   helper.SliceToChan(closes),
		   helper.SliceToChan(volumes)
}

// Simple Moving Average
func handleSMA(w http.ResponseWriter, r *http.Request) {
	var req IndicatorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	period := req.Period
	if period == 0 {
		period = 20 // Default period
	}
	
	_, _, _, closes, _ := dataToChannels(req.Data)
	
	sma := trend.NewSmaWithPeriod[float64](period)
	result := sma.Compute(closes)
	values := helper.ChanToSlice(result)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(IndicatorResponse{Values: values})
}

// Exponential Moving Average
func handleEMA(w http.ResponseWriter, r *http.Request) {
	var req IndicatorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	period := req.Period
	if period == 0 {
		period = 20
	}
	
	_, _, _, closes, _ := dataToChannels(req.Data)
	
	ema := trend.NewEmaWithPeriod[float64](period)
	result := ema.Compute(closes)
	values := helper.ChanToSlice(result)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(IndicatorResponse{Values: values})
}

// Double Exponential Moving Average
func handleDEMA(w http.ResponseWriter, r *http.Request) {
	var req IndicatorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	_, _, _, closes, _ := dataToChannels(req.Data)
	
	dema := trend.NewDema[float64]()
	result := dema.Compute(closes)
	values := helper.ChanToSlice(result)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(IndicatorResponse{Values: values})
}

// Triple Exponential Moving Average
func handleTEMA(w http.ResponseWriter, r *http.Request) {
	var req IndicatorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	_, _, _, closes, _ := dataToChannels(req.Data)
	
	tema := trend.NewTema[float64]()
	result := tema.Compute(closes)
	values := helper.ChanToSlice(result)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(IndicatorResponse{Values: values})
}

// Weighted Moving Average
func handleWMA(w http.ResponseWriter, r *http.Request) {
	var req IndicatorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	period := req.Period
	if period == 0 {
		period = 20
	}
	
	_, _, _, closes, _ := dataToChannels(req.Data)
	
	wma := trend.NewWmaWith[float64](period)
	result := wma.Compute(closes)
	values := helper.ChanToSlice(result)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(IndicatorResponse{Values: values})
}

// Hull Moving Average
func handleHMA(w http.ResponseWriter, r *http.Request) {
	var req IndicatorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	period := req.Period
	if period == 0 {
		period = 20
	}
	
	_, _, _, closes, _ := dataToChannels(req.Data)
	
	hma := trend.NewHmaWithPeriod[float64](period)
	result := hma.Compute(closes)
	values := helper.ChanToSlice(result)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(IndicatorResponse{Values: values})
}

// Kaufman's Adaptive Moving Average
func handleKAMA(w http.ResponseWriter, r *http.Request) {
	var req IndicatorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	_, _, _, closes, _ := dataToChannels(req.Data)
	
	kama := trend.NewKama[float64]()
	result := kama.Compute(closes)
	values := helper.ChanToSlice(result)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(IndicatorResponse{Values: values})
}

// MACD - returns both MACD line and signal line
func handleMACD(w http.ResponseWriter, r *http.Request) {
	var req IndicatorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	_, _, _, closes, _ := dataToChannels(req.Data)
	
	macd := trend.NewMacd[float64]()
	macdLine, signalLine := macd.Compute(closes)
	
	macdValues := helper.ChanToSlice(macdLine)
	signalValues := helper.ChanToSlice(signalLine)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(BandsIndicatorResponse{
		Upper:  macdValues,
		Lower:  signalValues,
		Middle: make([]float64, 0), // Not used for MACD
	})
}

// Aroon Indicator
func handleAroon(w http.ResponseWriter, r *http.Request) {
	var req IndicatorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	_, highs, lows, _, _ := dataToChannels(req.Data)
	
	aroon := trend.NewAroon[float64]()
	aroonUp, aroonDown := aroon.Compute(highs, lows)
	
	upValues := helper.ChanToSlice(aroonUp)
	downValues := helper.ChanToSlice(aroonDown)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(BandsIndicatorResponse{
		Upper: upValues,
		Lower: downValues,
		Middle: make([]float64, 0),
	})
}

// Community Channel Index
func handleCCI(w http.ResponseWriter, r *http.Request) {
	var req IndicatorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	period := req.Period
	if period == 0 {
		period = 20
	}
	
	_, highs, lows, closes, _ := dataToChannels(req.Data)
	
	cci := trend.NewCciWithPeriod[float64](period)
	result := cci.Compute(highs, lows, closes)
	values := helper.ChanToSlice(result)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(IndicatorResponse{Values: values})
}

// Relative Strength Index
func handleRSI(w http.ResponseWriter, r *http.Request) {
	var req IndicatorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	period := req.Period
	if period == 0 {
		period = 14
	}
	
	_, _, _, closes, _ := dataToChannels(req.Data)
	
	rsi := momentum.NewRsiWithPeriod[float64](period)
	result := rsi.Compute(closes)
	values := helper.ChanToSlice(result)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(IndicatorResponse{Values: values})
}

// Stochastic Oscillator
func handleStochastic(w http.ResponseWriter, r *http.Request) {
	var req IndicatorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	_, highs, lows, closes, _ := dataToChannels(req.Data)
	
	stoch := momentum.NewStochasticOscillator[float64]()
	k, d := stoch.Compute(highs, lows, closes)
	
	kValues := helper.ChanToSlice(k)
	dValues := helper.ChanToSlice(d)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(BandsIndicatorResponse{
		Upper: kValues,
		Lower: dValues,
		Middle: make([]float64, 0),
	})
}

// Williams %R
func handleWilliamsR(w http.ResponseWriter, r *http.Request) {
	var req IndicatorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	period := req.Period
	if period == 0 {
		period = 14
	}
	
	_, highs, lows, closes, _ := dataToChannels(req.Data)
	
	williams := momentum.NewWilliamsR[float64]()
	result := williams.Compute(highs, lows, closes)
	values := helper.ChanToSlice(result)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(IndicatorResponse{Values: values})
}

// Awesome Oscillator
func handleAwesomeOscillator(w http.ResponseWriter, r *http.Request) {
	var req IndicatorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	_, highs, lows, _, _ := dataToChannels(req.Data)
	
	ao := momentum.NewAwesomeOscillator[float64]()
	result := ao.Compute(highs, lows)
	values := helper.ChanToSlice(result)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(IndicatorResponse{Values: values})
}

// Bollinger Bands
func handleBollingerBands(w http.ResponseWriter, r *http.Request) {
	var req IndicatorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	period := req.Period
	if period == 0 {
		period = 20
	}
	
	_, _, _, closes, _ := dataToChannels(req.Data)
	
	bb := volatility.NewBollingerBandsWithPeriod[float64](period)
	upper, middle, lower := bb.Compute(closes)
	
	upperValues := helper.ChanToSlice(upper)
	middleValues := helper.ChanToSlice(middle)
	lowerValues := helper.ChanToSlice(lower)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(BandsIndicatorResponse{
		Upper:  upperValues,
		Middle: middleValues,
		Lower:  lowerValues,
	})
}

// Donchian Channel
func handleDonchianChannel(w http.ResponseWriter, r *http.Request) {
	var req IndicatorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	period := req.Period
	if period == 0 {
		period = 20
	}
	
	_, highs, lows, _, _ := dataToChannels(req.Data)
	
	dc := volatility.NewDonchianChannelWithPeriod[float64](period)
	upper, middle, lower := dc.Compute(highs)
	
	upperValues := helper.ChanToSlice(upper)
	middleValues := helper.ChanToSlice(middle)
	lowerValues := helper.ChanToSlice(lower)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(BandsIndicatorResponse{
		Upper:  upperValues,
		Middle: middleValues,
		Lower:  lowerValues,
	})
}

// Keltner Channel
func handleKeltnerChannel(w http.ResponseWriter, r *http.Request) {
	var req IndicatorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	_, highs, lows, closes, _ := dataToChannels(req.Data)
	
	kc := volatility.NewKeltnerChannel[float64]()
	upper, middle, lower := kc.Compute(highs, lows, closes)
	
	upperValues := helper.ChanToSlice(upper)
	middleValues := helper.ChanToSlice(middle)
	lowerValues := helper.ChanToSlice(lower)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(BandsIndicatorResponse{
		Upper:  upperValues,
		Middle: middleValues,
		Lower:  lowerValues,
	})
}

// Average True Range
func handleATR(w http.ResponseWriter, r *http.Request) {
	var req IndicatorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	period := req.Period
	if period == 0 {
		period = 14
	}
	
	_, highs, lows, closes, _ := dataToChannels(req.Data)
	
	atr := volatility.NewAtrWithPeriod[float64](period)
	result := atr.Compute(highs, lows, closes)
	values := helper.ChanToSlice(result)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(IndicatorResponse{Values: values})
}

// On-Balance Volume
func handleOBV(w http.ResponseWriter, r *http.Request) {
	var req IndicatorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	_, _, _, closes, volumes := dataToChannels(req.Data)
	
	obv := volume.NewObv[float64]()
	result := obv.Compute(closes, volumes)
	values := helper.ChanToSlice(result)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(IndicatorResponse{Values: values})
}

// Chaikin Money Flow
func handleCMF(w http.ResponseWriter, r *http.Request) {
	var req IndicatorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	period := req.Period
	if period == 0 {
		period = 20
	}
	
	_, highs, lows, closes, volumes := dataToChannels(req.Data)
	
	cmf := volume.NewCmfWithPeriod[float64](period)
	result := cmf.Compute(highs, lows, closes, volumes)
	values := helper.ChanToSlice(result)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(IndicatorResponse{Values: values})
}

// Money Flow Index
func handleMFI(w http.ResponseWriter, r *http.Request) {
	var req IndicatorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	_, highs, lows, closes, volumes := dataToChannels(req.Data)
	
	mfi := volume.NewMfi[float64]()
	result := mfi.Compute(highs, lows, closes, volumes)
	values := helper.ChanToSlice(result)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(IndicatorResponse{Values: values})
}

// Accumulation/Distribution
func handleAD(w http.ResponseWriter, r *http.Request) {
	var req IndicatorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	_, highs, lows, closes, volumes := dataToChannels(req.Data)
	
	ad := volume.NewAd[float64]()
	result := ad.Compute(highs, lows, closes, volumes)
	values := helper.ChanToSlice(result)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(IndicatorResponse{Values: values})
}

// Volume Weighted Average Price
func handleVWAP(w http.ResponseWriter, r *http.Request) {
	var req IndicatorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	period := req.Period
	if period == 0 {
		period = 14
	}
	
	_, _, _, closes, volumes := dataToChannels(req.Data)
	
	vwap := volume.NewVwapWithPeriod[float64](period)
	result := vwap.Compute(closes, volumes)
	values := helper.ChanToSlice(result)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(IndicatorResponse{Values: values})
}

// Get available strategies
func handleGetStrategies(w http.ResponseWriter, r *http.Request) {
	strategies := []string{
		"buy_and_hold",
		"rsi",
		"macd",
		"bollinger_bands",
		"golden_cross",
		"awesome_oscillator",
		"stochastic_rsi",
		"chaikin_money_flow",
		"donchian_channel",
		"aroon",
		"cci",
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(strategies)
}

// Backtest endpoint - simplified version
func handleBacktest(w http.ResponseWriter, r *http.Request) {
	var req BacktestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	// Convert OHLCV data to asset snapshots
	snapshots := make([]*asset.Snapshot, len(req.Data))
	for i, d := range req.Data {
		snapshots[i] = &asset.Snapshot{
			Date:   time.Unix(d.Time, 0),
			Open:   d.Open,
			High:   d.High,
			Low:    d.Low,
			Close:  d.Close,
			Volume: d.Volume,
		}
	}
	
	// Create snapshot channel
	snapshotChan := make(chan *asset.Snapshot, len(snapshots))
	for _, snapshot := range snapshots {
		snapshotChan <- snapshot
	}
	close(snapshotChan)
	
	// Select strategy based on request
	var strat strategy.Strategy
	switch req.Strategy {
	case "buy_and_hold":
		strat = strategy.NewBuyAndHoldStrategy()
	case "rsi":
		// RSI strategy not available in basic momentum package
		http.Error(w, "RSI strategy not implemented in this version", http.StatusNotImplemented)
		return
	default:
		http.Error(w, "Unsupported strategy: "+req.Strategy, http.StatusBadRequest)
		return
	}
	
	// Run strategy
	actionChan, outcomeChan := strategy.ComputeWithOutcome(strat, snapshotChan)
	
	// Collect results
	actions := helper.ChanToSlice(actionChan)
	outcomes := helper.ChanToSlice(outcomeChan)
	
	// Convert actions to integers
	actionInts := make([]int, len(actions))
	for i, action := range actions {
		actionInts[i] = int(action)
	}
	
	// Calculate basic statistics
	finalOutcome := 0.0
	if len(outcomes) > 0 {
		finalOutcome = outcomes[len(outcomes)-1]
	}
	
	buyCount := 0
	sellCount := 0
	for _, action := range actions {
		if action == 1 { // Buy
			buyCount++
		} else if action == -1 { // Sell
			sellCount++
		}
	}
	
	totalTrades := buyCount + sellCount
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(BacktestResponse{
		Actions:      actionInts,
		Outcome:      finalOutcome,
		TotalTrades:  totalTrades,
		WinRate:      0.0, // Would need more complex calculation
		ProfitFactor: 0.0, // Would need more complex calculation
	})
}