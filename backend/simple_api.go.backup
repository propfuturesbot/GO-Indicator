// Simple Trading Platform API Server
// Compatible with Go 1.19

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
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
	Data   []OHLCV                `json:"data"`
	Period int                    `json:"period,omitempty"`
	Config map[string]interface{} `json:"config,omitempty"`
}

// IndicatorResponse represents the response with calculated indicator values
type IndicatorResponse struct {
	Values []float64 `json:"values"`
	Error  string    `json:"error,omitempty"`
}

// BandsIndicatorResponse represents response for multi-line indicators
type BandsIndicatorResponse struct {
	Upper  []float64 `json:"upper"`
	Middle []float64 `json:"middle"`
	Lower  []float64 `json:"lower"`
	Error  string    `json:"error,omitempty"`
}

func main() {
	r := mux.NewRouter()

	// API routes
	api := r.PathPrefix("/api").Subrouter()

	// Trend indicators
	api.HandleFunc("/indicators/sma", handleSMA).Methods("POST")
	api.HandleFunc("/indicators/ema", handleEMA).Methods("POST")
	api.HandleFunc("/indicators/rsi", handleRSI).Methods("POST")
	api.HandleFunc("/indicators/bollinger-bands", handleBollingerBands).Methods("POST")
	api.HandleFunc("/indicators/macd", handleMACD).Methods("POST")

	// Health check
	api.HandleFunc("/health", handleHealth).Methods("GET")
	api.HandleFunc("/strategies", handleGetStrategies).Methods("GET")

	// CORS middleware
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	handler := c.Handler(r)

	fmt.Println("🚀 Trading Platform API Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}

// Health check endpoint
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "OK",
		"time":   time.Now().Format(time.RFC3339),
	})
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
		period = 20
	}

	closes := make([]float64, len(req.Data))
	for i, d := range req.Data {
		closes[i] = d.Close
	}

	values := calculateSMA(closes, period)

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

	closes := make([]float64, len(req.Data))
	for i, d := range req.Data {
		closes[i] = d.Close
	}

	values := calculateEMA(closes, period)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(IndicatorResponse{Values: values})
}

// RSI
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

	closes := make([]float64, len(req.Data))
	for i, d := range req.Data {
		closes[i] = d.Close
	}

	values := calculateRSI(closes, period)

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

	closes := make([]float64, len(req.Data))
	for i, d := range req.Data {
		closes[i] = d.Close
	}

	upper, middle, lower := calculateBollingerBands(closes, period, 2.0)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(BandsIndicatorResponse{
		Upper:  upper,
		Middle: middle,
		Lower:  lower,
	})
}

// MACD
func handleMACD(w http.ResponseWriter, r *http.Request) {
	var req IndicatorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	closes := make([]float64, len(req.Data))
	for i, d := range req.Data {
		closes[i] = d.Close
	}

	macdLine, signalLine := calculateMACD(closes, 12, 26, 9)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(BandsIndicatorResponse{
		Upper:  macdLine,
		Lower:  signalLine,
		Middle: make([]float64, 0),
	})
}

// Get available strategies
func handleGetStrategies(w http.ResponseWriter, r *http.Request) {
	strategies := []string{
		"buy_and_hold",
		"rsi",
		"macd",
		"bollinger_bands",
		"moving_average_crossover",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(strategies)
}

// ======================
// Indicator Calculations
// ======================

func calculateSMA(data []float64, period int) []float64 {
	if len(data) < period {
		return []float64{}
	}

	result := make([]float64, 0, len(data)-period+1)

	for i := period - 1; i < len(data); i++ {
		sum := 0.0
		for j := 0; j < period; j++ {
			sum += data[i-j]
		}
		result = append(result, sum/float64(period))
	}

	return result
}

func calculateEMA(data []float64, period int) []float64 {
	if len(data) < period {
		return []float64{}
	}

	multiplier := 2.0 / (float64(period) + 1.0)
	result := make([]float64, 0, len(data))

	// Start with SMA for first value
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += data[i]
	}
	ema := sum / float64(period)
	result = append(result, ema)

	// Calculate EMA for remaining values
	for i := period; i < len(data); i++ {
		ema = (data[i] * multiplier) + (ema * (1 - multiplier))
		result = append(result, ema)
	}

	return result
}

func calculateRSI(data []float64, period int) []float64 {
	if len(data) < period+1 {
		return []float64{}
	}

	// Calculate price changes
	changes := make([]float64, len(data)-1)
	for i := 1; i < len(data); i++ {
		changes[i-1] = data[i] - data[i-1]
	}

	// Separate gains and losses
	gains := make([]float64, len(changes))
	losses := make([]float64, len(changes))
	for i, change := range changes {
		if change > 0 {
			gains[i] = change
		} else {
			losses[i] = -change
		}
	}

	result := make([]float64, 0, len(data)-period)

	// Calculate initial averages
	avgGain := 0.0
	avgLoss := 0.0
	for i := 0; i < period; i++ {
		avgGain += gains[i]
		avgLoss += losses[i]
	}
	avgGain /= float64(period)
	avgLoss /= float64(period)

	// Calculate RSI
	for i := period; i < len(changes); i++ {
		if avgLoss == 0 {
			result = append(result, 100.0)
		} else {
			rs := avgGain / avgLoss
			rsi := 100.0 - (100.0 / (1.0 + rs))
			result = append(result, rsi)
		}

		// Update averages (smoothed)
		avgGain = ((avgGain * float64(period-1)) + gains[i]) / float64(period)
		avgLoss = ((avgLoss * float64(period-1)) + losses[i]) / float64(period)
	}

	return result
}

func calculateBollingerBands(data []float64, period int, multiplier float64) ([]float64, []float64, []float64) {
	sma := calculateSMA(data, period)
	if len(sma) == 0 {
		return []float64{}, []float64{}, []float64{}
	}

	upper := make([]float64, len(sma))
	lower := make([]float64, len(sma))

	for i := 0; i < len(sma); i++ {
		dataIndex := i + period - 1

		// Calculate standard deviation
		sum := 0.0
		for j := 0; j < period; j++ {
			diff := data[dataIndex-j] - sma[i]
			sum += diff * diff
		}
		stdDev := math.Sqrt(sum / float64(period))

		upper[i] = sma[i] + (stdDev * multiplier)
		lower[i] = sma[i] - (stdDev * multiplier)
	}

	return upper, sma, lower
}

func calculateMACD(data []float64, fastPeriod, slowPeriod, signalPeriod int) ([]float64, []float64) {
	if len(data) < slowPeriod {
		return []float64{}, []float64{}
	}

	fastEMA := calculateEMA(data, fastPeriod)
	slowEMA := calculateEMA(data, slowPeriod)

	// Align the EMAs (slow EMA starts later)
	offset := slowPeriod - fastPeriod
	macdLine := make([]float64, len(slowEMA))

	for i := 0; i < len(slowEMA); i++ {
		macdLine[i] = fastEMA[i+offset] - slowEMA[i]
	}

	// Calculate signal line (EMA of MACD)
	signalLine := calculateEMA(macdLine, signalPeriod)

	return macdLine, signalLine
}