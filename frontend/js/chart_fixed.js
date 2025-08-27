const ACCESS_TOKEN = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJodHRwOi8vc2NoZW1hcy54bWxzb2FwLm9yZy93cy8yMDA1LzA1L2lkZW50aXR5L2NsYWltcy9uYW1laWRlbnRpZmllciI6IjE5NDY5OSIsImh0dHA6Ly9zY2hlbWFzLnhtbHNvYXAub3JnL3dzLzIwMDUvMDUvaWRlbnRpdHkvY2xhaW1zL3NpZCI6IjViZWYyYzAwLTZjZmEtNGU1Ni04NDFiLWI5MmM4ZTk4OTZiZiIsImh0dHA6Ly9zY2hlbWFzLnhtbHNvYXAub3JnL3dzLzIwMDUvMDUvaWRlbnRpdHkvY2xhaW1zL25hbWUiOiJzdW1vbmV5MSIsImh0dHA6Ly9zY2hlbWFzLm1pY3Jvc29mdC5jb20vd3MvMjAwOC8wNi9pZGVudGl0eS9jbGFpbXMvcm9sZSI6InVzZXIiLCJtc2QiOiJDTUVHUk9VUF9UT0IiLCJtZmEiOiJ2ZXJpZmllZCIsImV4cCI6MTc1Njg1NDQzOH0.ZW4Sp93ftYT9QWcndN4i334jP7IjGR-YkxaTivyqHMM";
// API URL will be set dynamically
let API_BASE_URL;

// Global chart variables
let chart = null;
let candleSeries = null;
let connection = null;
let currentResolution = '15';
let currentSubscription = null;
let historicalData = [];
let lastBarTime = null;
let currentBar = null;
let currentChartType = 'candlestick';
let heikenAshiData = [];
let renkoData = [];
let currentBrickSize = 10;
let renkoState = {
  lastBrickHigh: null,
  lastBrickLow: null,
  direction: 1,
  lastTimestamp: null
};

// Make historicalData globally accessible for indicators
window.historicalData = historicalData;

// Indicator variables
let activeIndicators = new Map();
let indicatorSeries = new Map();

// Trading signals variables
let signalMarkers = [];
let activeSignals = new Map();
let lastTouchLower = false;
let lastTouchUpper = false;

// Tick chart accumulation variables
let tickAccumulator = null;
let tickCount = 0;
let currentTicksPerBar = 10;
let lastTickTime = null;

// Function to get ticks per bar based on resolution
function getTicksPerBar(resolution) {
  if (!resolution.endsWith('T')) return 10;
  
  const tickValue = parseInt(resolution.replace('T', ''));
  
  switch (tickValue) {
    case 100: return 5;
    case 500: return 15;
    case 1000: return 25;
    case 5000: return 50;
    default: return Math.max(5, Math.min(50, Math.floor(tickValue / 20)));
  }
}

const resolutionConfig = {
  // Tick-based resolutions
  '100T': { countback: 50, displayName: '100 Ticks', symbol: 'F.US.MNQ' },
  '500T': { countback: 50, displayName: '500 Ticks', symbol: 'F.US.MNQ' },
  '1000T': { countback: 50, displayName: '1000 Ticks', symbol: 'F.US.MNQ' },
  '5000T': { countback: 50, displayName: '5000 Ticks', symbol: 'F.US.MNQ' },
  
  // Second-based resolutions
  '1S': { countback: 500, displayName: '1 Second', symbol: 'F.US.MNQ' },
  '5S': { countback: 500, displayName: '5 Seconds', symbol: 'F.US.MNQ' },
  '10S': { countback: 500, displayName: '10 Seconds', symbol: 'F.US.MNQ' },
  '15S': { countback: 500, displayName: '15 Seconds', symbol: 'F.US.MNQ' },
  '20S': { countback: 500, displayName: '20 Seconds', symbol: 'F.US.MNQ' },
  '30S': { countback: 500, displayName: '30 Seconds', symbol: 'F.US.MNQ' },
  
  // Minute-based resolutions
  '1': { countback: 500, displayName: '1 Minute', symbol: 'F.US.MNQ' },
  '2': { countback: 500, displayName: '2 Minutes', symbol: 'F.US.MNQ' },
  '3': { countback: 500, displayName: '3 Minutes', symbol: 'F.US.MNQ' },
  '4': { countback: 500, displayName: '4 Minutes', symbol: 'F.US.MNQ' },
  '5': { countback: 500, displayName: '5 Minutes', symbol: 'F.US.MNQ' },
  '10': { countback: 500, displayName: '10 Minutes', symbol: 'F.US.MNQ' },
  '15': { countback: 500, displayName: '15 Minutes', symbol: 'F.US.MNQ' },
  '20': { countback: 500, displayName: '20 Minutes', symbol: 'F.US.MNQ' },
  '30': { countback: 500, displayName: '30 Minutes', symbol: 'F.US.MNQ' },
  '45': { countback: 500, displayName: '45 Minutes', symbol: 'F.US.MNQ' },
  '60': { countback: 500, displayName: '1 Hour', symbol: 'F.US.MNQ' },
  '1D': { countback: 326, displayName: '1 Day', symbol: 'F.US.MNQ' },
  '1W': { countback: 500, displayName: '1 Week', symbol: 'F.US.MNQ' },
  '1M': { countback: 500, displayName: '1 Month', symbol: 'F.US.MNQ' }
};

async function getHistoricalData(resolution, countback, symbol = "%2FMNQ") {
  try {
    const now = Math.floor(Date.now() / 1000);
    const from = now - (86400 * 7); // 7 days ago
    const to = now;
    
    const url = `https://chartapi.topstepx.com/History/v2?Symbol=${symbol}&Resolution=${resolution}&Countback=${countback}&From=${from}&To=${to}&SessionId=extended&Live=false`;
    
    console.log('Fetching historical data:', url);
    
    const res = await fetch(url, {
      headers: {
        'Authorization': `Bearer ${ACCESS_TOKEN}`,
        'Accept': 'application/json'
      }
    });
    
    if (!res.ok) {
      const errorText = await res.text();
      console.error(`HTTP error! status: ${res.status}, response: ${errorText}`);
      throw new Error(`HTTP error! status: ${res.status}`);
    }
    
    const text = await res.text();
    console.log('Raw response (first 500 chars):', text.substring(0, 500));
    
    let data;
    try {
      data = JSON.parse(text);
    } catch (e) {
      console.error('Failed to parse JSON:', e);
      return [];
    }
    
    // Handle both array and object with data property
    let bars = Array.isArray(data) ? data : (data.data || data.bars || []);
    
    if (!Array.isArray(bars)) {
      console.error('Unexpected data format:', data);
      return [];
    }
    
    console.log(`Received ${bars.length} bars for resolution ${resolution}`);
    
    if (bars.length === 0) {
      console.warn(`No data received for resolution ${resolution}`);
      return [];
    }
    
    // Sort bars by time and convert to chart format
    const chartData = bars
      .map(bar => {
        if (!bar || typeof bar.t === 'undefined' || typeof bar.o === 'undefined') {
          console.warn('Invalid bar data:', bar);
          return null;
        }
        
        return {
          time: Math.floor(bar.t / 1000), // Convert milliseconds to seconds
          open: parseFloat(bar.o),
          high: parseFloat(bar.h),
          low: parseFloat(bar.l),
          close: parseFloat(bar.c),
          volume: parseInt(bar.v || bar.tv || 0)
        };
      })
      .filter(bar => {
        if (!bar) return false;
        return !isNaN(bar.time) && !isNaN(bar.open) && !isNaN(bar.high) && 
               !isNaN(bar.low) && !isNaN(bar.close) && bar.time > 0;
      })
      .sort((a, b) => a.time - b.time);

    // Fix duplicate timestamps for tick charts by adding incremental seconds
    if (resolution.endsWith('T') && chartData.length > 0) {
      console.log('Fixing duplicate timestamps for tick chart...');
      let duplicatesFixed = 0;
      
      for (let i = 1; i < chartData.length; i++) {
        if (chartData[i].time <= chartData[i - 1].time) {
          chartData[i].time = chartData[i - 1].time + 1;
          duplicatesFixed++;
        }
      }
      
      console.log(`Fixed ${duplicatesFixed} duplicate timestamps in tick chart historical data`);
    }
    
    console.log(`Processed ${chartData.length} valid bars`);
    
    if (chartData.length > 0) {
      lastBarTime = chartData[chartData.length - 1].time;
      currentBar = { ...chartData[chartData.length - 1] };
      console.log('Last bar time:', new Date(lastBarTime * 1000).toLocaleString());
      console.log('Last bar:', currentBar);
    }
    
    return chartData;
  } catch (error) {
    console.error('Error fetching historical data:', error);
    return [];
  }
}

function initializeChart() {
  const chartElement = document.getElementById('chart');
  
  if (typeof LightweightCharts === 'undefined') {
    console.error('LightweightCharts library not loaded');
    return;
  }
  
  if (chart) {
    chart.remove();
  }
  
  chart = LightweightCharts.createChart(chartElement, {
    width: chartElement.offsetWidth,
    height: chartElement.offsetHeight,
    layout: {
      background: { color: 'transparent' },
      textColor: '#e2e8f0',
    },
    grid: {
      vertLines: { color: 'rgba(100, 116, 139, 0.2)' },
      horzLines: { color: 'rgba(100, 116, 139, 0.2)' },
    },
    timeScale: {
      timeVisible: true,
      secondsVisible: true,
      borderVisible: false,
      borderColor: '#334155',
    },
    rightPriceScale: {
      borderVisible: false,
      borderColor: '#334155',
      textColor: '#cbd5e1',
    },
    crosshair: {
      mode: LightweightCharts.CrosshairMode.Normal,
      vertLine: {
        color: '#06d6a0',
        width: 1,
        style: LightweightCharts.LineStyle.Dashed,
      },
      horzLine: {
        color: '#06d6a0',
        width: 1,
        style: LightweightCharts.LineStyle.Dashed,
      },
    },
  });
  
  candleSeries = chart.addCandlestickSeries({
    upColor: '#10b981',
    downColor: '#ef4444',
    borderVisible: false,
    wickUpColor: '#10b981',
    wickDownColor: '#ef4444',
  });
  
  // Handle window resize
  window.addEventListener('resize', () => {
    chart.applyOptions({ 
      width: chartElement.offsetWidth,
      height: chartElement.offsetHeight 
    });
  });
}

async function setupRealTimeConnection() {
  try {
    if (connection && connection.state === signalR.HubConnectionState.Connected) {
      await connection.stop();
    }
    
    connection = new signalR.HubConnectionBuilder()
      .withUrl(`https://chartapi.topstepx.com/hubs/chart?access_token=${ACCESS_TOKEN}`, {
        skipNegotiation: true,
        transport: signalR.HttpTransportType.WebSockets
      })
      .configureLogging(signalR.LogLevel.Information)
      .withAutomaticReconnect()
      .build();
    
    connection.on("RealTimeBar", (receivedSymbol, receivedResolution, bar) => {
      console.log('RealTimeBar received:', {
        symbol: receivedSymbol,
        resolution: receivedResolution,
        bar: bar,
        currentResolution: currentResolution
      });
      
      if (receivedResolution === currentResolution) {
        handleRealTimeBar(bar);
      }
    });
    
    connection.onreconnecting(() => {
      console.log('Reconnecting to SignalR...');
      updateStatus('Reconnecting...', false);
    });
    
    connection.onreconnected(async () => {
      console.log('Reconnected to SignalR');
      updateStatus('Connected', true);
      await subscribeToResolution(currentResolution);
    });
    
    connection.onclose(() => {
      console.log('SignalR connection closed');
      updateStatus('Disconnected', false);
    });
    
    await connection.start();
    console.log('Connected to SignalR hub successfully');
    updateStatus('Connected', true);
    
    await subscribeToResolution(currentResolution);
    
  } catch (error) {
    console.error('Failed to setup real-time connection:', error);
    updateStatus('Connection Failed', false);
  }
}

function handleRealTimeBar(bar) {
  if (!bar) {
    console.warn('Invalid bar received:', bar);
    return;
  }
  
  console.log('Raw bar data:', bar);
  
  let timestamp;
  
  if (bar.timestampUnix && bar.timestampUnix > 0) {
    timestamp = bar.timestampUnix * 1000;
  } else if (bar.timestamp && typeof bar.timestamp === 'string') {
    timestamp = new Date(bar.timestamp).getTime();
  } else if (bar.t) {
    timestamp = bar.t;
  } else if (bar.time) {
    timestamp = bar.time;
  } else {
    console.warn('No valid timestamp found in bar:', bar);
    return;
  }
  
  if (isNaN(timestamp)) {
    console.warn('Invalid timestamp:', timestamp);
    return;
  }
  
  // Fix timestamp format
  if (timestamp > 10000000000000000) {
    console.log('Timestamp too large, trying division by 1000000');
    timestamp = Math.floor(timestamp / 1000000);
  } else if (timestamp > 100000000000000) {
    console.log('Converting from microseconds to milliseconds');
    timestamp = Math.floor(timestamp / 1000);
  } else if (timestamp < 946684800000) {
    if (timestamp > 946684800) {
      console.log('Converting from seconds to milliseconds');
      timestamp = timestamp * 1000;
    }
  }
  
  if (timestamp > 2000000000000) {
    console.log('Timestamp still too large, trying additional division by 1000');
    timestamp = Math.floor(timestamp / 1000);
  }
  
  const barTime = Math.floor(timestamp / 1000);
  const update = {
    time: barTime,
    open: parseFloat(bar.open),
    high: parseFloat(bar.high),
    low: parseFloat(bar.low),
    close: parseFloat(bar.close),
    volume: parseInt(bar.volume || bar.tickVolume || bar.v || bar.tv || 0)
  };
  
  // Validate the update data
  if (isNaN(update.time) || isNaN(update.open) || isNaN(update.high) || 
      isNaN(update.low) || isNaN(update.close)) {
    console.warn('Invalid bar data, skipping update:', update);
    return;
  }
  
  // Validate OHLC relationships
  if (update.high < update.low || update.high < update.open || 
      update.high < update.close || update.low > update.open ||
      update.low > update.close) {
    console.warn('Invalid OHLC relationships, skipping update:', update);
    return;
  }
  
  // Only update if the bar time is newer than or equal to the last bar time
  if (lastBarTime && barTime < lastBarTime) {
    console.warn('Received old bar data, skipping');
    return;
  }
  
  try {
    if (currentResolution.endsWith('T')) {
      handleTickChartUpdate(update, bar);
    } else {
      let displayUpdate = update;
      
      if (currentChartType === 'heikenashi' && historicalData.length > 0) {
        const lastHACandle = heikenAshiData.length > 0 ? 
          heikenAshiData[heikenAshiData.length - 1] : null;
        
        const haClose = (update.open + update.high + update.low + update.close) / 4;
        const haOpen = lastHACandle ? 
          (lastHACandle.open + lastHACandle.close) / 2 : 
          (update.open + update.close) / 2;
        const haHigh = Math.max(update.high, haOpen, haClose);
        const haLow = Math.min(update.low, haOpen, haClose);
        
        displayUpdate = {
          time: update.time,
          open: haOpen,
          high: haHigh,
          low: haLow,
          close: haClose,
          volume: update.volume
        };
      }
      
      if (!lastBarTime || barTime > lastBarTime) {
        candleSeries.update(displayUpdate);
        
        if (bar.isClosed) {
          lastBarTime = barTime;
          currentBar = { ...update };
          
          if (currentChartType === 'heikenashi') {
            heikenAshiData.push(displayUpdate);
          }
        }
        
        updateIndicators(update);
      } else if (barTime === lastBarTime) {
        candleSeries.update(displayUpdate);
        currentBar = { ...update };
        
        if (currentChartType === 'heikenashi' && heikenAshiData.length > 0) {
          heikenAshiData[heikenAshiData.length - 1] = displayUpdate;
        }
        
        updateIndicators(update);
      }
    }
  } catch (chartError) {
    console.error('Error updating chart:', chartError);
  }
}

function handleTickChartUpdate(update, bar) {
  console.log('Tick chart: Received tick data - price:', update.close, 'volume:', update.volume);
  
  const tickPrice = update.close;
  const tickVolume = update.volume || 1;
  const tickTime = update.time;
  
  let shouldCreateNewBar = false;
  
  if (!tickAccumulator) {
    shouldCreateNewBar = true;
    console.log(`Tick chart: Starting first ${currentResolution} bar (${currentTicksPerBar} ticks per bar)`);
  } else {
    if (tickCount >= currentTicksPerBar) {
      shouldCreateNewBar = true;
      console.log(`Tick chart: ${currentTicksPerBar} ticks reached for ${currentResolution}, creating new bar`);
    } else if (lastTickTime && (tickTime - lastTickTime) > 30) {
      shouldCreateNewBar = true;
      console.log('Tick chart: Time gap detected, creating new bar');
    }
  }
  
  if (shouldCreateNewBar && tickAccumulator) {
    lastBarTime = tickAccumulator.time;
    currentBar = { ...tickAccumulator };
  }
  
  if (shouldCreateNewBar) {
    let barTime = lastBarTime ? lastBarTime + 1 : tickTime;
    
    tickAccumulator = {
      open: tickPrice,
      high: tickPrice,
      low: tickPrice,
      close: tickPrice,
      volume: 0,
      time: barTime
    };
    tickCount = 0;
    console.log(`Tick chart: Started new ${currentResolution} bar at time ${new Date(barTime * 1000).toLocaleString()}`);
  }
  
  tickAccumulator.high = Math.max(tickAccumulator.high, tickPrice);
  tickAccumulator.low = Math.min(tickAccumulator.low, tickPrice);
  tickAccumulator.close = tickPrice;
  tickAccumulator.volume += tickVolume;
  tickCount++;
  lastTickTime = tickTime;
  
  let currentBarData = {
    time: tickAccumulator.time,
    open: tickAccumulator.open,
    high: tickAccumulator.high,
    low: tickAccumulator.low,
    close: tickAccumulator.close,
    volume: tickAccumulator.volume
  };
  
  try {
    candleSeries.update(currentBarData);
  } catch (error) {
    console.error('Error updating tick chart:', error);
  }
}

async function subscribeToResolution(resolution) {
  try {
    const config = resolutionConfig[resolution];
    const symbol = config.symbol;
    
    if (currentSubscription && connection) {
      try {
        const prevConfig = resolutionConfig[currentSubscription];
        await connection.invoke("UnsubscribeBars", prevConfig.symbol, currentSubscription);
        console.log(`Unsubscribed from ${prevConfig.symbol} ${currentSubscription}`);
      } catch (unsubError) {
        console.warn('Error unsubscribing:', unsubError);
      }
    }
    
    if (connection && connection.state === signalR.HubConnectionState.Connected) {
      await connection.invoke("SubscribeBars", symbol, resolution);
      console.log(`Subscribed to ${symbol} ${resolution}`);
      currentSubscription = resolution;
    }
    
  } catch (error) {
    console.error('Failed to subscribe:', error);
  }
}

function calculateHeikenAshi(data) {
  if (!data || data.length === 0) return [];
  
  const haData = [];
  let prevHACandle = null;
  
  for (let i = 0; i < data.length; i++) {
    const candle = data[i];
    let haCandle = {};
    
    haCandle.close = (candle.open + candle.high + candle.low + candle.close) / 4;
    
    if (i === 0 || !prevHACandle) {
      haCandle.open = (candle.open + candle.close) / 2;
    } else {
      haCandle.open = (prevHACandle.open + prevHACandle.close) / 2;
    }
    
    haCandle.high = Math.max(candle.high, haCandle.open, haCandle.close);
    haCandle.low = Math.min(candle.low, haCandle.open, haCandle.close);
    haCandle.time = candle.time;
    haCandle.volume = candle.volume;
    
    haData.push(haCandle);
    prevHACandle = haCandle;
  }
  
  return haData;
}

// ATR calculation for Renko
function calculateATR(data, period = 14) {
  if (!data || data.length < period) return 50; // Default to 50 if can't calculate
  
  const trueRanges = [];
  
  for (let i = 1; i < data.length; i++) {
    const current = data[i];
    const previous = data[i - 1];
    
    const highLow = current.high - current.low;
    const highClose = Math.abs(current.high - previous.close);
    const lowClose = Math.abs(current.low - previous.close);
    
    const trueRange = Math.max(highLow, Math.max(highClose, lowClose));
    trueRanges.push(trueRange);
  }
  
  // Calculate Simple Moving Average of True Range
  if (trueRanges.length < period) return 50;
  
  let sum = 0;
  for (let i = trueRanges.length - period; i < trueRanges.length; i++) {
    sum += trueRanges[i];
  }
  
  return sum / period;
}

// Renko brick calculation functions
function convertToRenko(data, brickSize = null, useATR = true) {
  if (!data || data.length === 0) return [];
  
  // Calculate brick size using ATR if not provided
  if (brickSize === null || useATR) {
    const atrValue = calculateATR(data);
    brickSize = Math.max(1, Math.round(atrValue * 0.5));
    console.log('Calculated ATR brick size:', brickSize);
  }
  
  // Ensure brick size is valid
  if (!brickSize || brickSize <= 0 || isNaN(brickSize)) {
    brickSize = 10; // Default fallback
  }
  
  const renkoBricks = [];
  let brickIndex = 0; // Counter for unique timestamps
  
  // Initialize with first candle
  const firstPrice = data[0].close;
  const firstTime = data[0].time;
  
  let lastBrickHigh = firstPrice;
  let lastBrickLow = firstPrice;
  let direction = 1; // 1 for up, -1 for down
  
  for (let i = 1; i < data.length; i++) {
    const price = data[i].close;
    const baseTime = data[i].time;
    
    // Validate price data
    if (!price || isNaN(price)) continue;
    
    // Check for upward movement
    while (price >= lastBrickHigh + brickSize) {
      // Create green (up) brick
      const brickOpen = lastBrickHigh;
      const brickClose = lastBrickHigh + brickSize;
      
      // Ensure unique timestamp by adding brick index
      const brickTime = baseTime + brickIndex;
      brickIndex++;
      
      renkoBricks.push({
        time: brickTime,
        open: parseFloat(brickOpen.toFixed(2)),
        high: parseFloat(brickClose.toFixed(2)),
        low: parseFloat(brickOpen.toFixed(2)),
        close: parseFloat(brickClose.toFixed(2)),
        volume: data[i].volume || 0,
        color: 'green',
        direction: 1
      });
      
      lastBrickHigh = brickClose;
      lastBrickLow = brickOpen;
      direction = 1;
    }
    
    // Check for downward movement
    while (price <= lastBrickLow - brickSize) {
      // Create red (down) brick
      const brickOpen = lastBrickLow;
      const brickClose = lastBrickLow - brickSize;
      
      // Ensure unique timestamp by adding brick index
      const brickTime = baseTime + brickIndex;
      brickIndex++;
      
      renkoBricks.push({
        time: brickTime,
        open: parseFloat(brickOpen.toFixed(2)),
        high: parseFloat(brickOpen.toFixed(2)),
        low: parseFloat(brickClose.toFixed(2)),
        close: parseFloat(brickClose.toFixed(2)),
        volume: data[i].volume || 0,
        color: 'red',
        direction: -1
      });
      
      lastBrickLow = brickClose;
      lastBrickHigh = brickOpen;
      direction = -1;
    }
  }
  
  // Validate and clean up the bricks
  const validBricks = renkoBricks.filter(brick => {
    return brick &&
           typeof brick.time === 'number' &&
           typeof brick.open === 'number' &&
           typeof brick.high === 'number' &&
           typeof brick.low === 'number' &&
           typeof brick.close === 'number' &&
           !isNaN(brick.time) &&
           !isNaN(brick.open) &&
           !isNaN(brick.high) &&
           !isNaN(brick.low) &&
           !isNaN(brick.close) &&
           brick.time > 0 &&
           brick.high >= brick.low &&
           (brick.direction === 1 ? (brick.high === brick.close && brick.low === brick.open) : 
                                   (brick.low === brick.close && brick.high === brick.open));
  });
  
  // Ensure timestamps are sequential and unique
  if (validBricks.length > 1) {
    for (let i = 1; i < validBricks.length; i++) {
      if (validBricks[i].time <= validBricks[i-1].time) {
        validBricks[i].time = validBricks[i-1].time + 1;
      }
    }
  }
  
  // Update global Renko state
  if (validBricks.length > 0) {
    const lastBrick = validBricks[validBricks.length - 1];
    renkoState.lastBrickHigh = lastBrick.direction === 1 ? lastBrick.close : lastBrick.open;
    renkoState.lastBrickLow = lastBrick.direction === -1 ? lastBrick.close : lastBrick.open;
    renkoState.direction = lastBrick.direction;
    renkoState.lastTimestamp = lastBrick.time;
  }
  
  console.log(`Generated ${validBricks.length} valid Renko bricks with size ${brickSize}`);
  return validBricks;
}

function updateRenkoWithNewBar(newBar, brickSize) {
  if (!newBar || !renkoState.lastBrickHigh || !renkoState.lastBrickLow || !brickSize || brickSize <= 0) {
    return [];
  }
  
  const price = newBar.close;
  const baseTime = newBar.time;
  const newBricks = [];
  
  // Validate price
  if (!price || isNaN(price)) return [];
  
  let lastBrickHigh = renkoState.lastBrickHigh;
  let lastBrickLow = renkoState.lastBrickLow;
  
  // Start timestamp calculation from the last known timestamp
  let nextTimestamp = renkoState.lastTimestamp ? Math.max(renkoState.lastTimestamp + 1, baseTime) : baseTime;
  
  // Check for upward movement
  while (price >= lastBrickHigh + brickSize) {
    // Create green (up) brick
    const brickOpen = lastBrickHigh;
    const brickClose = lastBrickHigh + brickSize;
    
    newBricks.push({
      time: nextTimestamp,
      open: parseFloat(brickOpen.toFixed(2)),
      high: parseFloat(brickClose.toFixed(2)),
      low: parseFloat(brickOpen.toFixed(2)),
      close: parseFloat(brickClose.toFixed(2)),
      volume: newBar.volume || 0,
      color: 'green',
      direction: 1
    });
    
    lastBrickHigh = brickClose;
    lastBrickLow = brickOpen;
    renkoState.direction = 1;
    nextTimestamp++; // Ensure next brick has a later timestamp
  }
  
  // Check for downward movement
  while (price <= lastBrickLow - brickSize) {
    // Create red (down) brick
    const brickOpen = lastBrickLow;
    const brickClose = lastBrickLow - brickSize;
    
    newBricks.push({
      time: nextTimestamp,
      open: parseFloat(brickOpen.toFixed(2)),
      high: parseFloat(brickOpen.toFixed(2)),
      low: parseFloat(brickClose.toFixed(2)),
      close: parseFloat(brickClose.toFixed(2)),
      volume: newBar.volume || 0,
      color: 'red',
      direction: -1
    });
    
    lastBrickLow = brickClose;
    lastBrickHigh = brickOpen;
    renkoState.direction = -1;
    nextTimestamp++; // Ensure next brick has a later timestamp
  }
  
  // Update global state
  renkoState.lastBrickHigh = lastBrickHigh;
  renkoState.lastBrickLow = lastBrickLow;
  
  // Update last timestamp if we created new bricks
  if (newBricks.length > 0) {
    renkoState.lastTimestamp = newBricks[newBricks.length - 1].time;
  }
  
  // Validate bricks before returning
  const validBricks = newBricks.filter(brick => {
    return brick &&
           typeof brick.time === 'number' &&
           !isNaN(brick.time) &&
           typeof brick.open === 'number' &&
           !isNaN(brick.open) &&
           typeof brick.high === 'number' &&
           !isNaN(brick.high) &&
           typeof brick.low === 'number' &&
           !isNaN(brick.low) &&
           typeof brick.close === 'number' &&
           !isNaN(brick.close) &&
           brick.high >= brick.low;
  });
  
  console.log(`Real-time: Generated ${validBricks.length} new Renko bricks, last timestamp: ${renkoState.lastTimestamp}`);
  return validBricks;
}

async function changeResolution(resolution) {
  console.log(`=== Changing resolution to "${resolution}" ===`);
  currentResolution = resolution;
  lastBarTime = null;
  currentBar = null;
  
  tickAccumulator = null;
  tickCount = 0;
  lastTickTime = null;
  currentTicksPerBar = getTicksPerBar(resolution);
  
  if (candleSeries) {
    candleSeries.setData([]);
  }
  
  heikenAshiData = [];
  clearAllIndicators();
  
  const config = resolutionConfig[resolution];
  console.log(`Loading data for resolution "${resolution}" with config:`, config);
  historicalData = await getHistoricalData(resolution, config.countback);
  
  if (historicalData && historicalData.length > 0) {
    console.log(`Setting ${historicalData.length} bars on chart`);
    
    const validData = historicalData.filter(d => {
      return d && 
             typeof d.time === 'number' && 
             typeof d.open === 'number' && 
             typeof d.high === 'number' && 
             typeof d.low === 'number' && 
             typeof d.close === 'number' &&
             !isNaN(d.time) && 
             !isNaN(d.open) && 
             !isNaN(d.high) && 
             !isNaN(d.low) && 
             !isNaN(d.close) &&
             d.time > 0 &&
             d.high >= d.low &&
             d.high >= d.open &&
             d.high >= d.close &&
             d.low <= d.open &&
             d.low <= d.close;
    });
    
    if (validData.length > 0) {
      try {
        validData.sort((a, b) => a.time - b.time);
        historicalData = validData;
        
        // Update global reference for indicators
        window.historicalData = historicalData;
        
        if (currentChartType === 'heikenashi') {
          heikenAshiData = calculateHeikenAshi(validData);
          console.log('Calculated Heiken Ashi data with', heikenAshiData.length, 'bars');
        }
        
        let dataToDisplay = validData;
        if (currentChartType === 'heikenashi') {
          dataToDisplay = heikenAshiData;
        }
        
        candleSeries.setData(dataToDisplay);
        chart.timeScale().fitContent();
        
        if (currentResolution.endsWith('T') && dataToDisplay.length > 10) {
          const lastTime = dataToDisplay[dataToDisplay.length - 1].time;
          const firstTime = dataToDisplay[Math.max(0, dataToDisplay.length - 100)].time;
          chart.timeScale().setVisibleRange({ from: firstTime, to: lastTime });
        }
        
        console.log('Chart data set successfully');
      } catch (chartError) {
        console.error('Error setting chart data:', chartError);
      }
    } else {
      console.warn('No valid data after filtering');
      updateStatus('No data available', false);
    }
  } else {
    console.warn('No historical data received');
    updateStatus('No data available', false);
  }
  
  if (connection && connection.state === signalR.HubConnectionState.Connected) {
    await subscribeToResolution(resolution);
  }
}

function changeChartType(chartType) {
  console.log('Changing chart type to:', chartType);
  currentChartType = chartType;
  
  const renkoBrickSizeDiv = document.getElementById('renkoBrickSize');
  if (chartType === 'renko') {
    renkoBrickSizeDiv.style.display = 'flex';
    renkoBrickSizeDiv.style.alignItems = 'center';
  } else {
    renkoBrickSizeDiv.style.display = 'none';
  }
  
  let title = 'TradePro Platform';
  if (chartType === 'heikenashi') {
    title = 'TradePro Platform - Heiken Ashi';
  } else if (chartType === 'renko') {
    title = `TradePro Platform - Renko (${currentBrickSize})`;
  }
  document.querySelector('h1').textContent = title;
  
  updateChartData();
}

function updateChartData() {
  if (!candleSeries || !historicalData || historicalData.length === 0) {
    console.warn('Cannot update chart: missing data or chart series');
    return;
  }
  
  try {
    let dataToDisplay = historicalData;
    
    if (currentChartType === 'heikenashi') {
      heikenAshiData = calculateHeikenAshi(historicalData);
      dataToDisplay = heikenAshiData;
      console.log('Displaying Heiken Ashi data with', dataToDisplay.length, 'bars');
    } else {
      console.log('Displaying Candlestick data with', dataToDisplay.length, 'bars');
    }
    
    if (!dataToDisplay || dataToDisplay.length === 0) {
      console.warn('No data to display on chart');
      return;
    }
    
    candleSeries.setData(dataToDisplay);
    chart.timeScale().fitContent();
    
    if (currentResolution.endsWith('T') && dataToDisplay.length > 10) {
      const lastTime = dataToDisplay[dataToDisplay.length - 1].time;
      const firstTime = dataToDisplay[Math.max(0, dataToDisplay.length - 100)].time;
      chart.timeScale().setVisibleRange({ from: firstTime, to: lastTime });
    }
    
    console.log('Chart updated successfully with', dataToDisplay.length, 'data points');
    
  } catch (error) {
    console.error('Error updating chart data:', error);
    updateStatus('Chart update error', false);
  }
}

function updateRenkoBrickSize() {
  try {
    const brickSizeInput = document.getElementById('brickSizeInput');
    const newBrickSize = parseFloat(brickSizeInput.value);
    
    if (isNaN(newBrickSize) || newBrickSize <= 0) {
      alert('Please enter a valid brick size greater than 0');
      brickSizeInput.value = currentBrickSize;
      return;
    }
    
    if (newBrickSize > 1000) {
      alert('Brick size too large. Please enter a value less than 1000');
      brickSizeInput.value = currentBrickSize;
      return;
    }
    
    currentBrickSize = newBrickSize;
    console.log('Updated brick size to:', currentBrickSize);
    
    document.querySelector('h1').textContent = `TradePro Platform - Renko (${currentBrickSize})`;
    
    renkoState = {
      lastBrickHigh: null,
      lastBrickLow: null,
      direction: 1,
      lastTimestamp: null
    };
    
    if (currentChartType === 'renko') {
      try {
        updateChartData();
      } catch (error) {
        console.error('Error updating Renko chart:', error);
        alert('Error updating chart. Please try a different brick size.');
      }
    }
  } catch (error) {
    console.error('Error updating brick size:', error);
    alert('Error updating brick size. Please try again.');
  }
}

async function main() {
  console.log('=== Initializing Trading Platform ===');
  
  // Set API URL dynamically
  API_BASE_URL = `${window.location.protocol}//${window.location.hostname}:8080/api`;
  console.log('API Base URL:', API_BASE_URL);
  
  if (typeof LightweightCharts === 'undefined') {
    console.error('LightweightCharts library not loaded!');
    return;
  }
  
  if (typeof signalR === 'undefined') {
    console.error('SignalR library not loaded!');
    return;
  }
  
  console.log('All required libraries loaded successfully');
  
  console.log('Initializing chart...');
  initializeChart();
  
  if (!chart) {
    console.error('Failed to initialize chart');
    return;
  }
  
  console.log('Chart initialized successfully');
  
  const initialResolution = document.getElementById('resolution').value;
  currentResolution = initialResolution;
  currentTicksPerBar = getTicksPerBar(initialResolution);
  
  console.log(`Loading initial data for resolution: "${initialResolution}"`);
  const config = resolutionConfig[initialResolution];
  
  if (!config) {
    console.error(`No configuration found for resolution: ${initialResolution}`);
    return;
  }
  
  historicalData = await getHistoricalData(initialResolution, config.countback);
  
  // Update global reference for indicators
  window.historicalData = historicalData;
  
  if (historicalData && historicalData.length > 0) {
    console.log(`Setting ${historicalData.length} initial bars on chart`);
    
    const validData = historicalData.filter(d => {
      return d && 
             typeof d.time === 'number' && 
             typeof d.open === 'number' && 
             typeof d.high === 'number' && 
             typeof d.low === 'number' && 
             typeof d.close === 'number' &&
             !isNaN(d.time) && 
             !isNaN(d.open) && 
             !isNaN(d.high) && 
             !isNaN(d.low) && 
             !isNaN(d.close) &&
             d.time > 0 &&
             d.high >= d.low &&
             d.high >= d.open &&
             d.high >= d.close &&
             d.low <= d.open &&
             d.low <= d.close;
    });
    
    console.log(`Filtered data: ${validData.length} valid bars out of ${historicalData.length}`);
    
    if (validData.length > 0) {
      try {
        validData.sort((a, b) => a.time - b.time);
        historicalData = validData;
        
        // Update global reference for indicators
        window.historicalData = historicalData;
        
        if (currentChartType === 'heikenashi') {
          heikenAshiData = calculateHeikenAshi(validData);
        }
        
        let dataToDisplay = validData;
        if (currentChartType === 'heikenashi') {
          dataToDisplay = heikenAshiData;
        }
        
        candleSeries.setData(dataToDisplay);
        chart.timeScale().fitContent();
      } catch (chartError) {
        console.error('Error setting initial chart data:', chartError);
      }
    }
  } else {
    console.warn('No initial data loaded');
    updateStatus('No data available', false);
  }
  
  console.log('Setting up real-time connection...');
  await setupRealTimeConnection();
}

// Make functions available globally
window.addIndicator = addIndicator;
window.changeResolution = changeResolution;
window.changeChartType = changeChartType;
window.updateRenkoBrickSize = updateRenkoBrickSize;

// Start the application when DOM is ready
if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', main);
} else {
  main();
}