package main

import (
	"math/rand"
	"time"
)

// SimpleBacktestResult for testing
type SimpleBacktestResult struct {
	TotalTrades   int     `json:"totalTrades"`
	WinRate       float64 `json:"winRate"`
	TotalReturn   float64 `json:"totalReturn"`
	MaxDrawdown   float64 `json:"maxDrawdown"`
	ProfitFactor  float64 `json:"profitFactor"`
}

// RunSimpleBacktest for testing purposes
func RunSimpleBacktest() *SimpleBacktestResult {
	// Generate some realistic-looking test results
	rand.Seed(time.Now().UnixNano())
	
	return &SimpleBacktestResult{
		TotalTrades:  15 + rand.Intn(10),
		WinRate:      60.0 + rand.Float64()*20.0,
		TotalReturn:  5.0 + rand.Float64()*15.0,
		MaxDrawdown:  3.0 + rand.Float64()*8.0,
		ProfitFactor: 1.2 + rand.Float64()*0.8,
	}
}