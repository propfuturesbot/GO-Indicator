package main

import (
	"fmt"

	"github.com/cinar/indicator/v2/asset"
	"github.com/cinar/indicator/v2/helper"
	"github.com/cinar/indicator/v2/strategy"
	"github.com/cinar/indicator/v2/trend"
	"github.com/cinar/indicator/v2/momentum"
	"github.com/cinar/indicator/v2/volume"
	"github.com/cinar/indicator/v2/volatility"
)

// VWAPEMAStrategy implements VWAP EMA Crossover Strategy
type VWAPEMAStrategy struct {
	EMAFastPeriod    int     `json:"emaFastPeriod"`
	EMASlowPeriod    int     `json:"emaSlowPeriod"`
	VWAPPeriod       int     `json:"vwapPeriod"`
	ScalpTarget      float64 `json:"scalpTarget"`
	StopLoss         float64 `json:"stopLoss"`
	MinVolumeThreshold float64 `json:"minVolumeThreshold"`
	
	emaFast *trend.Ema[float64]
	emaSlow *trend.Ema[float64]
	vwap    *volume.Vwap[float64]
}

// NewVWAPEMAStrategy creates a new VWAP EMA Crossover strategy
func NewVWAPEMAStrategy() *VWAPEMAStrategy {
	return &VWAPEMAStrategy{
		EMAFastPeriod:      9,
		EMASlowPeriod:      21,
		VWAPPeriod:         20,
		ScalpTarget:        5.0,
		StopLoss:           10.0,
		MinVolumeThreshold: 1000,
		emaFast:            trend.NewEmaWithPeriod[float64](9),
		emaSlow:            trend.NewEmaWithPeriod[float64](21),
		vwap:               volume.NewVwapWithPeriod[float64](20),
	}
}

func (v *VWAPEMAStrategy) Name() string {
	return fmt.Sprintf("VWAP EMA Crossover (%d,%d,%d)", v.EMAFastPeriod, v.EMASlowPeriod, v.VWAPPeriod)
}

func (v *VWAPEMAStrategy) Compute(snapshots <-chan *asset.Snapshot) <-chan strategy.Action {
	snapshotsSplit := helper.Duplicate(snapshots, 4)
	
	closings := asset.SnapshotsAsClosings(snapshotsSplit[0])
	volumes := asset.SnapshotsAsVolumes(snapshotsSplit[1])
	
	closingsSplit := helper.Duplicate(closings, 3)
	volumesSplit := helper.Duplicate(volumes, 2)
	
	emaFast := v.emaFast.Compute(closingsSplit[0])
	emaSlow := v.emaSlow.Compute(closingsSplit[1])
	vwapValues := v.vwap.Compute(closingsSplit[2], volumesSplit[0])
	
	actions := helper.Operate3(emaFast, emaSlow, vwapValues,
		func(fast, slow, vwap float64) strategy.Action {
			// Simplified logic without volume check for now
			// Long conditions: EMA fast above EMA slow and price above VWAP
			if fast > slow && fast > vwap {
				return strategy.Buy
			}
			
			// Short conditions: EMA fast below EMA slow and price below VWAP
			if fast < slow && fast < vwap {
				return strategy.Sell
			}
			
			return strategy.Hold
		})
	
	// Shift for the longest period among indicators
	idlePeriod := v.EMASlowPeriod
	if v.VWAPPeriod > idlePeriod {
		idlePeriod = v.VWAPPeriod
	}
	
	actions = helper.Shift(actions, idlePeriod, strategy.Hold)
	return actions
}

func (v *VWAPEMAStrategy) Report(c <-chan *asset.Snapshot) *helper.Report {
	snapshots := helper.Duplicate(c, 3)
	dates := asset.SnapshotsAsDates(snapshots[0])
	closings := helper.Duplicate(asset.SnapshotsAsClosings(snapshots[1]), 2)
	
	actions, outcomes := strategy.ComputeWithOutcome(v, snapshots[2])
	annotations := strategy.ActionsToAnnotations(actions)
	outcomes = helper.MultiplyBy(outcomes, 100)
	
	report := helper.NewReport(v.Name(), dates)
	report.AddChart()
	
	report.AddColumn(helper.NewNumericReportColumn("Close", closings[0]))
	report.AddColumn(helper.NewAnnotationReportColumn(annotations))
	report.AddColumn(helper.NewNumericReportColumn("Outcome", outcomes), 1)
	
	return report
}

// DonchianBreakoutStrategy implements Donchian Channel Breakout Strategy
type DonchianBreakoutStrategy struct {
	Period int `json:"period"`
	dc     *volatility.DonchianChannel[float64]
}

func NewDonchianBreakoutStrategy() *DonchianBreakoutStrategy {
	return &DonchianBreakoutStrategy{
		Period: 20,
		dc:     volatility.NewDonchianChannelWithPeriod[float64](20),
	}
}

func (d *DonchianBreakoutStrategy) Name() string {
	return fmt.Sprintf("Donchian Channel Breakout (%d)", d.Period)
}

func (d *DonchianBreakoutStrategy) Compute(snapshots <-chan *asset.Snapshot) <-chan strategy.Action {
	snapshotsSplit := helper.Duplicate(snapshots, 2)
	highs := asset.SnapshotsAsHighs(snapshotsSplit[0])
	closes := asset.SnapshotsAsClosings(snapshotsSplit[1])
	
	upper, middle, lower := d.dc.Compute(highs)
	go helper.Drain(middle) // Don't need middle channel
	
	closes = helper.Skip(closes, d.dc.IdlePeriod())
	
	actions := helper.Operate3(upper, lower, closes, func(upperBound, lowerBound, close float64) strategy.Action {
		// Buy on breakout above upper channel
		if close > upperBound {
			return strategy.Buy
		}
		
		// Sell on breakdown below lower channel  
		if close < lowerBound {
			return strategy.Sell
		}
		
		return strategy.Hold
	})
	
	actions = helper.Shift(actions, d.dc.IdlePeriod(), strategy.Hold)
	return actions
}

func (d *DonchianBreakoutStrategy) Report(c <-chan *asset.Snapshot) *helper.Report {
	snapshots := helper.Duplicate(c, 3)
	dates := asset.SnapshotsAsDates(snapshots[0])
	highs := asset.SnapshotsAsHighs(snapshots[1])
	
	actions, outcomes := strategy.ComputeWithOutcome(d, snapshots[2])
	annotations := strategy.ActionsToAnnotations(actions)
	outcomes = helper.MultiplyBy(outcomes, 100)
	
	report := helper.NewReport(d.Name(), dates)
	report.AddChart()
	
	report.AddColumn(helper.NewNumericReportColumn("High", highs))
	report.AddColumn(helper.NewAnnotationReportColumn(annotations))
	report.AddColumn(helper.NewNumericReportColumn("Outcome", outcomes), 1)
	
	return report
}

// RSIDivergenceStrategy implements RSI Divergence Strategy
type RSIDivergenceStrategy struct {
	RSIPeriod     int     `json:"rsiPeriod"`
	OverboughtLevel float64 `json:"overboughtLevel"`
	OversoldLevel   float64 `json:"oversoldLevel"`
	
	rsi *momentum.Rsi[float64]
}

func NewRSIDivergenceStrategy() *RSIDivergenceStrategy {
	return &RSIDivergenceStrategy{
		RSIPeriod:       14,
		OverboughtLevel: 70,
		OversoldLevel:   30,
		rsi:             momentum.NewRsiWithPeriod[float64](14),
	}
}

func (r *RSIDivergenceStrategy) Name() string {
	return fmt.Sprintf("RSI Divergence Strategy (%d)", r.RSIPeriod)
}

func (r *RSIDivergenceStrategy) Compute(snapshots <-chan *asset.Snapshot) <-chan strategy.Action {
	closings := asset.SnapshotsAsClosings(snapshots)
	rsiValues := r.rsi.Compute(closings)
	
	actions := helper.Map(rsiValues, func(rsi float64) strategy.Action {
		// Buy when RSI is oversold
		if rsi <= r.OversoldLevel {
			return strategy.Buy
		}
		
		// Sell when RSI is overbought
		if rsi >= r.OverboughtLevel {
			return strategy.Sell
		}
		
		return strategy.Hold
	})
	
	actions = helper.Shift(actions, r.rsi.IdlePeriod(), strategy.Hold)
	return actions
}

func (r *RSIDivergenceStrategy) Report(c <-chan *asset.Snapshot) *helper.Report {
	snapshots := helper.Duplicate(c, 3)
	dates := asset.SnapshotsAsDates(snapshots[0])
	closings := helper.Duplicate(asset.SnapshotsAsClosings(snapshots[1]), 2)
	
	rsiValues := helper.Shift(r.rsi.Compute(closings[1]), r.rsi.IdlePeriod(), 0)
	
	actions, outcomes := strategy.ComputeWithOutcome(r, snapshots[2])
	annotations := strategy.ActionsToAnnotations(actions)
	outcomes = helper.MultiplyBy(outcomes, 100)
	
	report := helper.NewReport(r.Name(), dates)
	report.AddChart()
	report.AddChart()
	
	report.AddColumn(helper.NewNumericReportColumn("Close", closings[0]))
	report.AddColumn(helper.NewNumericReportColumn("RSI", rsiValues), 1)
	report.AddColumn(helper.NewAnnotationReportColumn(annotations), 0, 1)
	report.AddColumn(helper.NewNumericReportColumn("Outcome", outcomes), 2)
	
	return report
}

// MovingAverageCrossoverStrategy implements Moving Average Crossover Strategy
type MovingAverageCrossoverStrategy struct {
	FastPeriod int `json:"fastPeriod"`
	SlowPeriod int `json:"slowPeriod"`
	
	fastMA *trend.Sma[float64]
	slowMA *trend.Sma[float64]
}

func NewMovingAverageCrossoverStrategy() *MovingAverageCrossoverStrategy {
	return &MovingAverageCrossoverStrategy{
		FastPeriod: 10,
		SlowPeriod: 30,
		fastMA:     trend.NewSmaWithPeriod[float64](10),
		slowMA:     trend.NewSmaWithPeriod[float64](30),
	}
}

func (m *MovingAverageCrossoverStrategy) Name() string {
	return fmt.Sprintf("Moving Average Crossover (%d,%d)", m.FastPeriod, m.SlowPeriod)
}

func (m *MovingAverageCrossoverStrategy) Compute(snapshots <-chan *asset.Snapshot) <-chan strategy.Action {
	closings := asset.SnapshotsAsClosings(snapshots)
	closingsSplit := helper.Duplicate(closings, 2)
	
	fastMA := m.fastMA.Compute(closingsSplit[0])
	slowMA := m.slowMA.Compute(closingsSplit[1])
	
	actions := helper.Operate(fastMA, slowMA, func(fast, slow float64) strategy.Action {
		// Golden cross - fast MA crosses above slow MA
		if fast > slow {
			return strategy.Buy
		}
		
		// Death cross - fast MA crosses below slow MA
		if fast < slow {
			return strategy.Sell
		}
		
		return strategy.Hold
	})
	
	actions = helper.Shift(actions, m.SlowPeriod, strategy.Hold)
	return actions
}

func (m *MovingAverageCrossoverStrategy) Report(c <-chan *asset.Snapshot) *helper.Report {
	snapshots := helper.Duplicate(c, 3)
	dates := asset.SnapshotsAsDates(snapshots[0])
	closings := helper.Duplicate(asset.SnapshotsAsClosings(snapshots[1]), 3)
	
	fastMA := helper.Shift(m.fastMA.Compute(closings[1]), m.SlowPeriod, 0)
	slowMA := helper.Shift(m.slowMA.Compute(closings[2]), m.SlowPeriod, 0)
	
	actions, outcomes := strategy.ComputeWithOutcome(m, snapshots[2])
	annotations := strategy.ActionsToAnnotations(actions)
	outcomes = helper.MultiplyBy(outcomes, 100)
	
	report := helper.NewReport(m.Name(), dates)
	report.AddChart()
	
	report.AddColumn(helper.NewNumericReportColumn("Close", closings[0]))
	report.AddColumn(helper.NewNumericReportColumn("Fast MA", fastMA))
	report.AddColumn(helper.NewNumericReportColumn("Slow MA", slowMA))
	report.AddColumn(helper.NewAnnotationReportColumn(annotations))
	report.AddColumn(helper.NewNumericReportColumn("Outcome", outcomes), 1)
	
	return report
}

// BollingerMeanReversionStrategy implements Bollinger Band Mean Reversion Strategy
type BollingerMeanReversionStrategy struct {
	Period       int     `json:"period"`
	StdDeviation float64 `json:"stdDeviation"`
	
	bb *volatility.BollingerBands[float64]
}

func NewBollingerMeanReversionStrategy() *BollingerMeanReversionStrategy {
	return &BollingerMeanReversionStrategy{
		Period:       20,
		StdDeviation: 2.0,
		bb:           volatility.NewBollingerBandsWithPeriod[float64](20),
	}
}

func (b *BollingerMeanReversionStrategy) Name() string {
	return fmt.Sprintf("Bollinger Mean Reversion (%d)", b.Period)
}

func (b *BollingerMeanReversionStrategy) Compute(snapshots <-chan *asset.Snapshot) <-chan strategy.Action {
	closings := helper.Duplicate(asset.SnapshotsAsClosings(snapshots), 2)
	
	upper, middle, lower := b.bb.Compute(closings[0])
	go helper.Drain(middle) // Don't need middle for this strategy
	
	closings[1] = helper.Skip(closings[1], b.bb.IdlePeriod())
	
	actions := helper.Operate3(upper, lower, closings[1], func(upperBand, lowerBand, closing float64) strategy.Action {
		// Buy when price touches lower band (mean reversion)
		if closing <= lowerBand {
			return strategy.Buy
		}
		
		// Sell when price touches upper band (mean reversion)
		if closing >= upperBand {
			return strategy.Sell
		}
		
		return strategy.Hold
	})
	
	actions = helper.Shift(actions, b.bb.IdlePeriod(), strategy.Hold)
	return actions
}

func (b *BollingerMeanReversionStrategy) Report(c <-chan *asset.Snapshot) *helper.Report {
	snapshots := helper.Duplicate(c, 3)
	dates := asset.SnapshotsAsDates(snapshots[0])
	closings := helper.Duplicate(asset.SnapshotsAsClosings(snapshots[1]), 2)
	
	upper, middle, lower := b.bb.Compute(closings[1])
	upper = helper.Shift(upper, b.bb.IdlePeriod(), 0)
	middle = helper.Shift(middle, b.bb.IdlePeriod(), 0)
	lower = helper.Shift(lower, b.bb.IdlePeriod(), 0)
	
	actions, outcomes := strategy.ComputeWithOutcome(b, snapshots[2])
	annotations := strategy.ActionsToAnnotations(actions)
	outcomes = helper.MultiplyBy(outcomes, 100)
	
	report := helper.NewReport(b.Name(), dates)
	report.AddChart()
	
	report.AddColumn(helper.NewNumericReportColumn("Close", closings[0]))
	report.AddColumn(helper.NewNumericReportColumn("Upper", upper))
	report.AddColumn(helper.NewNumericReportColumn("Middle", middle))
	report.AddColumn(helper.NewNumericReportColumn("Lower", lower))
	report.AddColumn(helper.NewAnnotationReportColumn(annotations))
	report.AddColumn(helper.NewNumericReportColumn("Outcome", outcomes), 1)
	
	return report
}

// GetAllCustomStrategies returns all custom strategies
func GetAllCustomStrategies() []strategy.Strategy {
	return []strategy.Strategy{
		NewVWAPEMAStrategy(),
		NewDonchianBreakoutStrategy(),
		NewRSIDivergenceStrategy(),
		NewMovingAverageCrossoverStrategy(),
		NewBollingerMeanReversionStrategy(),
	}
}