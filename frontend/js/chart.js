const ACCESS_TOKEN = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJodHRwOi8vc2NoZW1hcy54bWxzb2FwLm9yZy93cy8yMDA1LzA1L2lkZW50aXR5L2NsYWltcy9uYW1laWRlbnRpZmllciI6IjE5NDY5OSIsImh0dHA6Ly9zY2hlbWFzLnhtbHNvYXAub3JnL3dzLzIwMDUvMDUvaWRlbnRpdHkvY2xhaW1zL3NpZCI6IjQ1MTI3ODhlLTdmZWMtNDVmYS1hNWZiLTRhODYyY2VhOTVjZiIsImh0dHA6Ly9zY2hlbWFzLnhtbHNvYXAub3JnL3dzLzIwMDUvMDUvaWRlbnRpdHkvY2xhaW1zL25hbWUiOiJzdW1vbmV5MSIsImh0dHA6Ly9zY2hlbWFzLm1pY3Jvc29mdC5jb20vd3MvMjAwOC8wNi9pZGVudGl0eS9jbGFpbXMvcm9sZSI6InVzZXIiLCJtc2QiOiJDTUVHUk9VUF9UT0IiLCJtZmEiOiJ2ZXJpZmllZCIsImV4cCI6MTc1NjIyMDU2NH0.eBMx4sD3Mp9z3Nto74Fj0-Mg-3Rm9nj8k9J934LUYQ8";
const API_BASE_URL = "http://localhost:8080/api";

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
  direction: 1, // 1 for up, -1 for down
  lastTimestamp: null // Track the last timestamp used
};

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
let currentTicksPerBar = 10; // Dynamic based on resolution
let lastTickTime = null;

// Function to get ticks per bar based on resolution
const getTicksPerBar = (resolution) => {
  if (!resolution.endsWith('T')) return 10;
  
  const tickValue = parseInt(resolution.replace('T', ''));
  
  // Scale the accumulation based on tick resolution
  // For smaller resolutions, accumulate fewer individual ticks
  // For larger resolutions, accumulate more individual ticks
  switch (tickValue) {
    case 100: return 5;   // 100T: accumulate 5 individual ticks per bar
    case 500: return 15;  // 500T: accumulate 15 individual ticks per bar
    case 1000: return 25; // 1000T: accumulate 25 individual ticks per bar
    case 5000: return 50; // 5000T: accumulate 50 individual ticks per bar
    default: return Math.max(5, Math.min(50, Math.floor(tickValue / 20))); // Dynamic scaling
  }
};

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

const getHistoricalData = async (resolution, countback, symbol = "%2FMNQ") => {
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
        // Validate bar data
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
        // Ensure all required fields are valid numbers
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
};

const initializeChart = () => {
  const chartElement = document.getElementById('chart');
  
  // Check if LightweightCharts is available
  if (typeof LightweightCharts === 'undefined') {
    console.error('LightweightCharts library not loaded');
    return;
  }
  
  // Clear existing chart if any
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
};

// Rest of your existing chart.js code (setupRealTimeConnection, handleRealTimeBar, etc.)
// ... keeping all the existing functions from your original code ...

const setupRealTimeConnection = async () => {
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
      
      // Check if this update is for our current resolution
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
    
    // Subscribe to the current resolution
    await subscribeToResolution(currentResolution);
    
  } catch (error) {
    console.error('Failed to setup real-time connection:', error);
    updateStatus('Connection Failed', false);
  }
};

// Keep all your existing functions for handleRealTimeBar, Renko, Heiken Ashi etc.
// ... (include all the rest of your original chart.js code here)

const handleRealTimeBar = (bar) => {
  // Your existing handleRealTimeBar implementation
  // ... keep all the existing code
};

// Include all other existing functions from your chart.js
// ... (all the Renko, Heiken Ashi, chart management functions)

const main = async () => {
  console.log('=== Initializing Trading Platform ===');
  
  // Check if required libraries are loaded
  if (typeof LightweightCharts === 'undefined') {
    console.error('LightweightCharts library not loaded!');
    return;
  }
  
  if (typeof signalR === 'undefined') {
    console.error('SignalR library not loaded!');
    return;
  }
  
  console.log('All required libraries loaded successfully');
  
  // Initialize the chart
  console.log('Initializing chart...');
  initializeChart();
  
  if (!chart) {
    console.error('Failed to initialize chart');
    return;
  }
  
  console.log('Chart initialized successfully');
  
  // Load initial data
  const initialResolution = document.getElementById('resolution').value;
  currentResolution = initialResolution;
  
  // Set initial ticks per bar
  currentTicksPerBar = getTicksPerBar(initialResolution);
  
  console.log(`Loading initial data for resolution: "${initialResolution}"`);
  console.log('Initial resolution config:', resolutionConfig[initialResolution]);
  const config = resolutionConfig[initialResolution];
  
  if (!config) {
    console.error(`No configuration found for resolution: ${initialResolution}`);
    return;
  }
  
  historicalData = await getHistoricalData(initialResolution, config.countback);
  
  if (historicalData && historicalData.length > 0) {
    console.log(`Setting ${historicalData.length} initial bars on chart`);
    
    // Ensure we have valid candlestick data
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
        // Sort by time to ensure proper order
        validData.sort((a, b) => a.time - b.time);
        
        // Store the valid data
        historicalData = validData;
        
        // Calculate Heiken Ashi data if that chart type is selected
        if (currentChartType === 'heikenashi') {
          heikenAshiData = calculateHeikenAshi(validData);
        } else if (currentChartType === 'renko') {
          renkoData = convertToRenko(validData, currentBrickSize, false);
        }
        
        // Determine which data to display
        let dataToDisplay = validData;
        if (currentChartType === 'heikenashi') {
          dataToDisplay = heikenAshiData;
        } else if (currentChartType === 'renko') {
          dataToDisplay = renkoData;
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
  
  // Setup real-time connection
  console.log('Setting up real-time connection...');
  await setupRealTimeConnection();
};

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