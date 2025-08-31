package main

import (
	"math"
	"math/rand"
	"time"
)

// NQDataGenerator generates realistic NQ futures data
type NQDataGenerator struct {
	BasePrice     float64
	Volatility    float64
	TrendStrength float64
	VolumeBase    float64
}

// NewNQDataGenerator creates a new NQ data generator
func NewNQDataGenerator() *NQDataGenerator {
	return &NQDataGenerator{
		BasePrice:     18500.0, // Starting around current NQ levels
		Volatility:    0.02,    // 2% daily volatility
		TrendStrength: 0.1,     // Trend component
		VolumeBase:    2000,    // Base volume
	}
}

// GenerateData generates realistic NQ futures data for the specified period
func (g *NQDataGenerator) GenerateData(startDate, endDate time.Time, interval time.Duration) []OHLCV {
	var data []OHLCV
	
	// Seed random number generator
	rand.Seed(time.Now().UnixNano())
	
	currentTime := startDate
	currentPrice := g.BasePrice
	var prevClose float64 = g.BasePrice
	
	// Market session parameters
	rthStart := 8.5  // 8:30 AM CT (Regular Trading Hours start)
	rthEnd := 15.0   // 3:00 PM CT (Regular Trading Hours end)
	
	for currentTime.Before(endDate) {
		// Skip weekends
		if currentTime.Weekday() == time.Saturday || currentTime.Weekday() == time.Sunday {
			currentTime = currentTime.Add(interval)
			continue
		}
		
		// Determine if we're in RTH or overnight session
		hour := float64(currentTime.Hour()) + float64(currentTime.Minute())/60.0
		isRTH := hour >= rthStart && hour < rthEnd
		
		// Generate gap for first bar of the day
		if hour < 0.5 { // Around midnight, start of new session
			gap := g.generateGap()
			currentPrice = prevClose * (1 + gap/100)
		}
		
		// Generate OHLC for this bar
		bar := g.generateBar(currentPrice, isRTH, currentTime)
		
		// Add market microstructure effects
		g.addMarketEffects(&bar, isRTH, hour)
		
		data = append(data, bar)
		
		// Update for next bar
		prevClose = bar.Close
		currentPrice = bar.Close
		currentTime = currentTime.Add(interval)
	}
	
	return data
}

// generateBar creates a single OHLCV bar
func (g *NQDataGenerator) generateBar(basePrice float64, isRTH bool, timestamp time.Time) OHLCV {
	// Adjust volatility based on session
	volMultiplier := 1.0
	if !isRTH {
		volMultiplier = 0.3 // Lower overnight volatility
	}
	
	// Generate price movement
	drift := g.generateDrift()
	volatility := g.Volatility * volMultiplier
	
	// Random walk component
	random := rand.NormFloat64() * volatility
	
	// Calculate price change
	priceChange := (drift + random) * basePrice
	
	// Generate OHLC
	open := basePrice
	close := basePrice + priceChange
	
	// High and low based on intrabar volatility
	intraBarVol := volatility * 0.5
	high := math.Max(open, close) + math.Abs(rand.NormFloat64())*intraBarVol*basePrice
	low := math.Min(open, close) - math.Abs(rand.NormFloat64())*intraBarVol*basePrice
	
	// Ensure OHLC relationship
	high = math.Max(high, math.Max(open, close))
	low = math.Min(low, math.Min(open, close))
	
	// Generate volume
	volume := g.generateVolume(isRTH)
	
	return OHLCV{
		Time:   timestamp.Unix(),
		Open:   roundToTick(open),
		High:   roundToTick(high),
		Low:    roundToTick(low),
		Close:  roundToTick(close),
		Volume: volume,
	}
}

// generateDrift creates trend component
func (g *NQDataGenerator) generateDrift() float64 {
	// Simple mean-reverting trend with some persistence
	return rand.NormFloat64() * g.TrendStrength * 0.001
}

// generateGap creates opening gaps
func (g *NQDataGenerator) generateGap() float64 {
	// Generate gaps between -50 to +50 points, with occasional larger gaps
	if rand.Float64() < 0.05 { // 5% chance of large gap
		return (rand.Float64() - 0.5) * 200 // ±100 points
	}
	return (rand.Float64() - 0.5) * 100 // ±50 points
}

// generateVolume creates realistic volume patterns
func (g *NQDataGenerator) generateVolume(isRTH bool) float64 {
	baseVolume := g.VolumeBase
	
	// Higher volume during RTH
	if isRTH {
		baseVolume *= 3.0
	}
	
	// Add randomness
	volumeMultiplier := 0.5 + rand.Float64()*1.5 // 0.5x to 2.0x multiplier
	
	return math.Floor(baseVolume * volumeMultiplier)
}

// addMarketEffects adds realistic market microstructure effects
func (g *NQDataGenerator) addMarketEffects(bar *OHLCV, isRTH bool, hour float64) {
	// Volume spikes during news events (simulate)
	if rand.Float64() < 0.02 { // 2% chance of news event
		bar.Volume *= 2.0 + rand.Float64()*3.0
		
		// Price volatility spike
		priceSpike := (rand.Float64() - 0.5) * 50 // ±25 point spike
		bar.High += math.Abs(priceSpike)
		bar.Low -= math.Abs(priceSpike)
	}
	
	// Opening and closing auction effects
	if hour >= 8.5 && hour < 9.0 { // Opening auction
		bar.Volume *= 1.5
	}
	if hour >= 14.5 && hour < 15.0 { // Closing auction
		bar.Volume *= 1.3
	}
	
	// Holiday/reduced trading effects
	if isHolidayPeriod(time.Unix(bar.Time, 0)) {
		bar.Volume *= 0.3
	}
}

// roundToTick rounds price to NQ tick size (0.25)
func roundToTick(price float64) float64 {
	tickSize := 0.25
	return math.Round(price/tickSize) * tickSize
}

// isHolidayPeriod checks if date is near a holiday
func isHolidayPeriod(date time.Time) bool {
	month := date.Month()
	day := date.Day()
	
	// Simple holiday detection (US markets)
	holidays := [][2]int{
		{1, 1},   // New Year's Day
		{7, 4},   // Independence Day
		{11, 22}, // Thanksgiving (approximate)
		{12, 25}, // Christmas
	}
	
	for _, holiday := range holidays {
		if int(month) == holiday[0] && day == holiday[1] {
			return true
		}
	}
	
	return false
}

// GenerateNQSampleData generates sample NQ data for testing
func GenerateNQSampleData() []OHLCV {
	generator := NewNQDataGenerator()
	
	// Generate just 1 week of hourly data for testing
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -7) // 1 week ago
	
	// Generate hourly bars (much less data)
	interval := time.Hour
	
	data := generator.GenerateData(startDate, endDate, interval)
	
	// Limit to 200 bars maximum for testing
	if len(data) > 200 {
		data = data[:200]
	}
	
	return data
}

// addRealisticPatterns adds trend periods and consolidation to the data
func addRealisticPatterns(data []OHLCV) {
	if len(data) < 100 {
		return
	}
	
	// Add trend periods
	trendPeriods := []struct {
		start, end int
		strength   float64
	}{
		{100, 300, 0.02},   // Bullish trend
		{800, 1200, -0.015}, // Bearish trend
		{2000, 2500, 0.025}, // Strong bullish trend
	}
	
	for _, trend := range trendPeriods {
		if trend.end >= len(data) {
			continue
		}
		
		for i := trend.start; i < trend.end && i < len(data); i++ {
			if i > 0 {
				trendMove := data[i-1].Close * trend.strength * 0.01
				data[i].Open = data[i-1].Close
				data[i].Close = data[i].Open + trendMove
				data[i].High = math.Max(data[i].Open, data[i].Close) + math.Abs(rand.NormFloat64()*10)
				data[i].Low = math.Min(data[i].Open, data[i].Close) - math.Abs(rand.NormFloat64()*10)
				
				// Ensure tick size compliance
				data[i].Open = roundToTick(data[i].Open)
				data[i].High = roundToTick(data[i].High)
				data[i].Low = roundToTick(data[i].Low)
				data[i].Close = roundToTick(data[i].Close)
			}
		}
	}
}