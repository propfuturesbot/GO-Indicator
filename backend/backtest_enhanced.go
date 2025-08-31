package main

import (
        "fmt"
        "math"
        "time"

        "github.com/cinar/indicator/v2/asset"
        "github.com/cinar/indicator/v2/helper"
        "github.com/cinar/indicator/v2/strategy"
)

// BacktestConfig represents configuration for backtesting
type BacktestConfig struct {
        Instrument     string                 `json:"instrument"`
        ChartType      string                 `json:"chartType"`
        Timeframe      string                 `json:"timeframe"`
        Strategy       string                 `json:"strategy"`
        FromDate       string                 `json:"fromDate"`
        ToDate         string                 `json:"toDate"`
        InitialCapital float64                `json:"initialCapital"`
        PositionSize   float64                `json:"positionSize"`
        MaxPositions   int                    `json:"maxPositions"`
        StopLoss       float64                `json:"stopLoss"`
        TakeProfit     float64                `json:"takeProfit"`
        MaxDrawdown    float64                `json:"maxDrawdown"`
        CommissionPer  float64                `json:"commissionPer"`
        Slippage       float64                `json:"slippage"`
        Parameters     map[string]interface{} `json:"parameters"`
}

// Trade represents a single trade
type Trade struct {
        EntryTime   string  `json:"entryTime"`
        ExitTime    string  `json:"exitTime"`
        Side        string  `json:"side"`
        EntryPrice  float64 `json:"entryPrice"`
        ExitPrice   float64 `json:"exitPrice"`
        Quantity    float64 `json:"quantity"`
        PnL         float64 `json:"pnl"`
        PnLPercent  float64 `json:"pnlPercent"`
        Duration    string  `json:"duration"`
        EntryReason string  `json:"entryReason"`
        ExitReason  string  `json:"exitReason"`
}

// EquityPoint represents a point on the equity curve
type EquityPoint struct {
        Time   string  `json:"time"`
        Equity float64 `json:"equity"`
}

// DrawdownPoint represents a point on the drawdown curve
type DrawdownPoint struct {
        Time     string  `json:"time"`
        Drawdown float64 `json:"drawdown"`
}

// MonthlyReturn represents monthly return data
type MonthlyReturn struct {
        Month  string  `json:"month"`
        Return float64 `json:"return"`
}

// DetailedBacktestResult represents comprehensive backtest results
type DetailedBacktestResult struct {
        Summary      BacktestSummary `json:"summary"`
        EquityCurve  []EquityPoint   `json:"equityCurve"`
        DrawdownCurve []DrawdownPoint `json:"drawdownCurve"`
        MonthlyReturns []MonthlyReturn `json:"monthlyReturns"`
        TradeLog     []Trade         `json:"tradeLog"`
        Actions      []int           `json:"actions"`
        Config       BacktestConfig  `json:"config"`
}

// BacktestSummary represents summary statistics
type BacktestSummary struct {
        TotalTrades           int     `json:"totalTrades"`
        WinningTrades         int     `json:"winningTrades"`
        LosingTrades          int     `json:"losingTrades"`
        WinRate               float64 `json:"winRate"`
        ProfitFactor          float64 `json:"profitFactor"`
        TotalPnL              float64 `json:"totalPnL"`
        MaxDrawdown           float64 `json:"maxDrawdown"`
        MaxDrawdownDuration   int     `json:"maxDrawdownDuration"`
        SharpeRatio           float64 `json:"sharpeRatio"`
        SortinoRatio          float64 `json:"sortinoRatio"`
        AverageWin            float64 `json:"averageWin"`
        AverageLoss           float64 `json:"averageLoss"`
        LargestWin            float64 `json:"largestWin"`
        LargestLoss           float64 `json:"largestLoss"`
        ConsecutiveWins       int     `json:"consecutiveWins"`
        ConsecutiveLosses     int     `json:"consecutiveLosses"`
        MaxConsecutiveWins    int     `json:"maxConsecutiveWins"`
        MaxConsecutiveLosses  int     `json:"maxConsecutiveLosses"`
        FinalEquity           float64 `json:"finalEquity"`
        TotalReturnPercent    float64 `json:"totalReturnPercent"`
}

// RunEnhancedBacktest performs detailed backtesting with comprehensive results
func RunEnhancedBacktest(data []OHLCV, config BacktestConfig) (*DetailedBacktestResult, error) {
        // Limit data size for testing to prevent hanging
        if len(data) > 1000 {
                data = data[:1000]
        }
        
        // For initial testing, use a simple buy and hold strategy
        actions := make([]strategy.Action, len(data))
        for i := range actions {
                if i == 0 {
                        actions[i] = strategy.Buy // Buy at the start
                } else if i == len(data)-1 {
                        actions[i] = strategy.Sell // Sell at the end
                } else {
                        actions[i] = strategy.Hold
                }
        }

        // Convert actions to integers
        actionInts := make([]int, len(actions))
        for i, action := range actions {
                actionInts[i] = int(action)
        }

        // Generate detailed analysis
        result := &DetailedBacktestResult{
                Actions: actionInts,
                Config:  config,
        }

        // Calculate trades and performance metrics
        result.TradeLog = calculateTrades(data, actions, config)
        result.Summary = calculateSummaryStatistics(result.TradeLog, config.InitialCapital)
        result.EquityCurve = calculateEquityCurve(result.TradeLog, config.InitialCapital)
        result.DrawdownCurve = calculateDrawdownCurve(result.EquityCurve)
        result.MonthlyReturns = calculateMonthlyReturns(result.EquityCurve)

        return result, nil
}

// calculateTrades generates trade log from actions
func calculateTrades(data []OHLCV, actions []strategy.Action, config BacktestConfig) []Trade {
        var trades []Trade
        var currentTrade *Trade
        
        for i, action := range actions {
                if i >= len(data) {
                        break
                }
                
                bar := data[i]
                timestamp := time.Unix(bar.Time, 0)
                
                switch action {
                case strategy.Buy:
                        if currentTrade == nil {
                                currentTrade = &Trade{
                                        EntryTime:   timestamp.Format("2006-01-02 15:04:05"),
                                        Side:        "Long",
                                        EntryPrice:  bar.Close,
                                        Quantity:    config.PositionSize,
                                        EntryReason: "Strategy Signal",
                                }
                        }
                        
                case strategy.Sell:
                        if currentTrade != nil {
                                // Close long position
                                currentTrade.ExitTime = timestamp.Format("2006-01-02 15:04:05")
                                currentTrade.ExitPrice = bar.Close
                                currentTrade.PnL = (currentTrade.ExitPrice - currentTrade.EntryPrice) * currentTrade.Quantity
                                currentTrade.PnLPercent = (currentTrade.PnL / (currentTrade.EntryPrice * currentTrade.Quantity)) * 100
                                
                                entryTime, _ := time.Parse("2006-01-02 15:04:05", currentTrade.EntryTime)
                                duration := timestamp.Sub(entryTime)
                                currentTrade.Duration = formatDuration(duration)
                                currentTrade.ExitReason = "Strategy Signal"
                                
                                trades = append(trades, *currentTrade)
                                currentTrade = nil
                        } else {
                                // Open short position
                                currentTrade = &Trade{
                                        EntryTime:   timestamp.Format("2006-01-02 15:04:05"),
                                        Side:        "Short",
                                        EntryPrice:  bar.Close,
                                        Quantity:    config.PositionSize,
                                        EntryReason: "Strategy Signal",
                                }
                        }
                }
                
                // Handle stop loss and take profit
                if currentTrade != nil {
                        var exitPrice float64
                        var exitReason string
                        
                        if currentTrade.Side == "Long" {
                                // Check stop loss
                                if config.StopLoss > 0 && bar.Low <= currentTrade.EntryPrice*(1-config.StopLoss/100) {
                                        exitPrice = currentTrade.EntryPrice * (1 - config.StopLoss/100)
                                        exitReason = "Stop Loss"
                                }
                                // Check take profit
                                if config.TakeProfit > 0 && bar.High >= currentTrade.EntryPrice*(1+config.TakeProfit/100) {
                                        exitPrice = currentTrade.EntryPrice * (1 + config.TakeProfit/100)
                                        exitReason = "Take Profit"
                                }
                        } else { // Short position
                                // Check stop loss
                                if config.StopLoss > 0 && bar.High >= currentTrade.EntryPrice*(1+config.StopLoss/100) {
                                        exitPrice = currentTrade.EntryPrice * (1 + config.StopLoss/100)
                                        exitReason = "Stop Loss"
                                }
                                // Check take profit
                                if config.TakeProfit > 0 && bar.Low <= currentTrade.EntryPrice*(1-config.TakeProfit/100) {
                                        exitPrice = currentTrade.EntryPrice * (1 - config.TakeProfit/100)
                                        exitReason = "Take Profit"
                                }
                        }
                        
                        if exitPrice > 0 {
                                currentTrade.ExitTime = timestamp.Format("2006-01-02 15:04:05")
                                currentTrade.ExitPrice = exitPrice
                                if currentTrade.Side == "Long" {
                                        currentTrade.PnL = (currentTrade.ExitPrice - currentTrade.EntryPrice) * currentTrade.Quantity
                                } else {
                                        currentTrade.PnL = (currentTrade.EntryPrice - currentTrade.ExitPrice) * currentTrade.Quantity
                                }
                                currentTrade.PnLPercent = (currentTrade.PnL / (currentTrade.EntryPrice * currentTrade.Quantity)) * 100
                                
                                entryTime, _ := time.Parse("2006-01-02 15:04:05", currentTrade.EntryTime)
                                duration := timestamp.Sub(entryTime)
                                currentTrade.Duration = formatDuration(duration)
                                currentTrade.ExitReason = exitReason
                                
                                trades = append(trades, *currentTrade)
                                currentTrade = nil
                        }
                }
        }
        
        return trades
}

// calculateSummaryStatistics computes performance metrics
func calculateSummaryStatistics(trades []Trade, initialCapital float64) BacktestSummary {
        summary := BacktestSummary{
                TotalTrades: len(trades),
        }
        
        if len(trades) == 0 {
                return summary
        }
        
        var totalPnL float64
        var winningTrades, losingTrades int
        var totalWinAmount, totalLossAmount float64
        var largestWin, largestLoss float64
        var consecutiveWins, consecutiveLosses int
        var maxConsecutiveWins, maxConsecutiveLosses int
        var returns []float64
        
        for _, trade := range trades {
                totalPnL += trade.PnL
                
                if trade.PnL > 0 {
                        winningTrades++
                        totalWinAmount += trade.PnL
                        if trade.PnL > largestWin {
                                largestWin = trade.PnL
                        }
                        consecutiveWins++
                        consecutiveLosses = 0
                        if consecutiveWins > maxConsecutiveWins {
                                maxConsecutiveWins = consecutiveWins
                        }
                } else if trade.PnL < 0 {
                        losingTrades++
                        totalLossAmount += math.Abs(trade.PnL)
                        if math.Abs(trade.PnL) > largestLoss {
                                largestLoss = math.Abs(trade.PnL)
                        }
                        consecutiveLosses++
                        consecutiveWins = 0
                        if consecutiveLosses > maxConsecutiveLosses {
                                maxConsecutiveLosses = consecutiveLosses
                        }
                }
                
                returns = append(returns, trade.PnLPercent/100)
        }
        
        summary.WinningTrades = winningTrades
        summary.LosingTrades = losingTrades
        summary.TotalPnL = totalPnL
        summary.LargestWin = largestWin
        summary.LargestLoss = largestLoss
        summary.MaxConsecutiveWins = maxConsecutiveWins
        summary.MaxConsecutiveLosses = maxConsecutiveLosses
        summary.FinalEquity = initialCapital + totalPnL
        summary.TotalReturnPercent = (totalPnL / initialCapital) * 100
        
        if summary.TotalTrades > 0 {
                summary.WinRate = (float64(winningTrades) / float64(summary.TotalTrades)) * 100
        }
        
        if winningTrades > 0 {
                summary.AverageWin = totalWinAmount / float64(winningTrades)
        }
        
        if losingTrades > 0 {
                summary.AverageLoss = totalLossAmount / float64(losingTrades)
        }
        
        if totalLossAmount > 0 {
                summary.ProfitFactor = totalWinAmount / totalLossAmount
        }
        
        // Calculate Sharpe and Sortino ratios
        if len(returns) > 0 {
                meanReturn := calculateMean(returns)
                stdDev := calculateStdDev(returns, meanReturn)
                
                if stdDev > 0 {
                        summary.SharpeRatio = (meanReturn * math.Sqrt(252)) / stdDev // Assuming daily returns
                }
                
                // Sortino ratio (only downside deviation)
                downsideReturns := make([]float64, 0)
                for _, ret := range returns {
                        if ret < 0 {
                                downsideReturns = append(downsideReturns, ret)
                        }
                }
                
                if len(downsideReturns) > 0 {
                        downsideStdDev := calculateStdDev(downsideReturns, 0)
                        if downsideStdDev > 0 {
                                summary.SortinoRatio = (meanReturn * math.Sqrt(252)) / downsideStdDev
                        }
                }
        }
        
        return summary
}

// calculateEquityCurve generates equity curve data
func calculateEquityCurve(trades []Trade, initialCapital float64) []EquityPoint {
        var equityCurve []EquityPoint
        
        currentEquity := initialCapital
        equityCurve = append(equityCurve, EquityPoint{
                Time:   "Start",
                Equity: currentEquity,
        })
        
        for _, trade := range trades {
                currentEquity += trade.PnL
                equityCurve = append(equityCurve, EquityPoint{
                        Time:   trade.ExitTime,
                        Equity: currentEquity,
                })
        }
        
        return equityCurve
}

// calculateDrawdownCurve generates drawdown curve
func calculateDrawdownCurve(equityCurve []EquityPoint) []DrawdownPoint {
        var drawdownCurve []DrawdownPoint
        var peak float64
        var maxDrawdown float64
        
        for _, point := range equityCurve {
                if point.Equity > peak {
                        peak = point.Equity
                }
                
                drawdown := ((peak - point.Equity) / peak) * 100
                if drawdown > maxDrawdown {
                        maxDrawdown = drawdown
                }
                
                drawdownCurve = append(drawdownCurve, DrawdownPoint{
                        Time:     point.Time,
                        Drawdown: drawdown,
                })
        }
        
        return drawdownCurve
}

// calculateMonthlyReturns generates monthly return data
func calculateMonthlyReturns(equityCurve []EquityPoint) []MonthlyReturn {
        var monthlyReturns []MonthlyReturn
        
        if len(equityCurve) < 2 {
                return monthlyReturns
        }
        
        // Simple implementation - could be enhanced with proper monthly grouping
        monthlyReturns = append(monthlyReturns, MonthlyReturn{
                Month:  "Total",
                Return: ((equityCurve[len(equityCurve)-1].Equity - equityCurve[0].Equity) / equityCurve[0].Equity) * 100,
        })
        
        return monthlyReturns
}

// Helper functions
func calculateMean(values []float64) float64 {
        if len(values) == 0 {
                return 0
        }
        
        var sum float64
        for _, v := range values {
                sum += v
        }
        
        return sum / float64(len(values))
}

func calculateStdDev(values []float64, mean float64) float64 {
        if len(values) <= 1 {
                return 0
        }
        
        var sum float64
        for _, v := range values {
                sum += math.Pow(v-mean, 2)
        }
        
        return math.Sqrt(sum / float64(len(values)-1))
}

func formatDuration(duration time.Duration) string {
        hours := int(duration.Hours())
        minutes := int(duration.Minutes()) % 60
        
        if hours > 0 {
                return fmt.Sprintf("%dh %dm", hours, minutes)
        }
        return fmt.Sprintf("%dm", minutes)
}

func getStrategyByName(name string) (strategy.Strategy, error) {
        strategies := GetAllCustomStrategies()
        
        for _, strat := range strategies {
                if strat.Name() == name {
                        return strat, nil
                }
        }
        
        return nil, fmt.Errorf("strategy not found: %s", name)
}