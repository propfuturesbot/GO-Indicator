// Indicator calculation and management system
// This module interfaces with the Go backend API for indicator calculations

class IndicatorManager {
  constructor() {
    this.activeIndicators = new Map();
    this.indicatorSeries = new Map();
    // Will be set dynamically when needed
    this.apiBaseUrl = null;
  }

  // Get API base URL dynamically
  getApiBaseUrl() {
    if (!this.apiBaseUrl) {
      this.apiBaseUrl = `${window.location.protocol}//${window.location.hostname}:8080/api`;
      console.log('Setting API Base URL:', this.apiBaseUrl);
    }
    return this.apiBaseUrl;
  }

  // Add an indicator by calling the backend API
  async addIndicator(type, period = null, config = {}) {
    if (this.activeIndicators.has(type)) {
      console.log(`Indicator ${type} already active`);
      return;
    }

    if (!historicalData || historicalData.length === 0) {
      alert('No data available to calculate indicators');
      return;
    }

    try {
      console.log(`Adding indicator: ${type} with period: ${period}`);
      
      const requestData = {
        data: historicalData,
        period: period,
        config: config
      };

      const endpoint = this.getIndicatorEndpoint(type);
      const response = await fetch(`${this.getApiBaseUrl()}${endpoint}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(requestData)
      });

      if (!response.ok) {
        throw new Error(`API request failed: ${response.status}`);
      }

      const result = await response.json();
      
      if (result.error) {
        throw new Error(result.error);
      }

      // Display the indicator on the chart
      this.displayIndicator(type, result, period);
      
      // Store the indicator configuration
      this.activeIndicators.set(type, { 
        period: period, 
        values: result,
        config: config
      });
      
      console.log(`${type} indicator added successfully`);
    } catch (error) {
      console.error(`Error adding ${type} indicator:`, error);
      alert(`Error adding ${type} indicator: ${error.message}`);
    }
  }

  // Get the correct API endpoint for each indicator type
  getIndicatorEndpoint(type) {
    const endpoints = {
      'SMA': '/indicators/sma',
      'EMA': '/indicators/ema',
      'DEMA': '/indicators/dema',
      'TEMA': '/indicators/tema',
      'WMA': '/indicators/wma',
      'HMA': '/indicators/hma',
      'KAMA': '/indicators/kama',
      'MACD': '/indicators/macd',
      'Aroon': '/indicators/aroon',
      'CCI': '/indicators/cci',
      'RSI': '/indicators/rsi',
      'Stochastic': '/indicators/stochastic',
      'Williams': '/indicators/williams-r',
      'AwesomeOscillator': '/indicators/awesome-oscillator',
      'BollingerBands': '/indicators/bollinger-bands',
      'DonchianChannel': '/indicators/donchian-channel',
      'KeltnerChannel': '/indicators/keltner-channel',
      'ATR': '/indicators/atr',
      'OBV': '/indicators/obv',
      'CMF': '/indicators/cmf',
      'MFI': '/indicators/mfi',
      'AD': '/indicators/ad',
      'VWAP': '/indicators/vwap'
    };
    
    return endpoints[type] || `/indicators/${type.toLowerCase()}`;
  }

  // Display indicator on the chart
  displayIndicator(type, data, period) {
    if (!chart || !historicalData) return;
    
    const colors = ['#FF6B6B', '#4ECDC4', '#45B7D1', '#96CEB4', '#FECA57', '#FF9FF3', '#54A0FF'];
    const colorIndex = this.indicatorSeries.size % colors.length;
    const color = colors[colorIndex];
    
    try {
      // Handle multi-line indicators (Bollinger Bands, Donchian Channel, etc.)
      if (this.isMultiLineIndicator(type, data)) {
        this.displayMultiLineIndicator(type, data, period, color);
      } else {
        // Handle single-line indicators
        this.displaySingleLineIndicator(type, data, period, color);
      }
      
      console.log(`Displayed ${type} indicator with data points`);
      
    } catch (error) {
      console.error('Error displaying indicator:', error);
    }
  }

  // Check if indicator returns multiple lines
  isMultiLineIndicator(type, data) {
    return (type === 'BollingerBands' || type === 'DonchianChannel' || type === 'KeltnerChannel' || 
            type === 'MACD' || type === 'Aroon' || type === 'Stochastic') && 
           (data.upper || data.lower || data.middle);
  }

  // Display multi-line indicators
  displayMultiLineIndicator(type, data, period, color) {
    const prefix = this.getIndicatorPrefix(type);
    
    // Create series for each line
    if (data.upper) {
      const upperSeries = chart.addLineSeries({
        color: color,
        lineWidth: 1,
        title: `${prefix} Upper (${period || 'default'})`
      });
      this.setIndicatorData(upperSeries, data.upper);
      this.indicatorSeries.set(`${type}_upper`, upperSeries);
    }
    
    if (data.middle) {
      const middleSeries = chart.addLineSeries({
        color: color,
        lineWidth: 2,
        title: `${prefix} Middle (${period || 'default'})`
      });
      this.setIndicatorData(middleSeries, data.middle);
      this.indicatorSeries.set(`${type}_middle`, middleSeries);
    }
    
    if (data.lower) {
      const lowerSeries = chart.addLineSeries({
        color: color,
        lineWidth: 1,
        title: `${prefix} Lower (${period || 'default'})`
      });
      this.setIndicatorData(lowerSeries, data.lower);
      this.indicatorSeries.set(`${type}_lower`, lowerSeries);
    }
  }

  // Display single-line indicators
  displaySingleLineIndicator(type, data, period, color) {
    const lineSeries = chart.addLineSeries({
      color: color,
      lineWidth: 2,
      title: `${type} (${period || 'default'})`
    });
    
    const values = data.values || data;
    this.setIndicatorData(lineSeries, values);
    this.indicatorSeries.set(type, lineSeries);
  }

  // Set data on indicator series
  setIndicatorData(series, values) {
    if (!Array.isArray(values) || values.length === 0) {
      console.warn('Invalid indicator values');
      return;
    }

    const indicatorData = values.map((value, index) => {
      const dataIndex = index + (historicalData.length - values.length);
      return {
        time: historicalData[dataIndex]?.time,
        value: value
      };
    }).filter(item => item.time && !isNaN(item.value));
    
    series.setData(indicatorData);
  }

  // Get indicator prefix for display
  getIndicatorPrefix(type) {
    const prefixes = {
      'BollingerBands': 'BB',
      'DonchianChannel': 'DC',
      'KeltnerChannel': 'KC',
      'MACD': 'MACD',
      'Aroon': 'Aroon',
      'Stochastic': 'Stoch'
    };
    return prefixes[type] || type;
  }

  // Remove an indicator
  removeIndicator(indicatorType) {
    console.log('Removing indicator:', indicatorType);
    
    // Remove from active indicators
    this.activeIndicators.delete(indicatorType);
    
    // Remove series from chart
    if (this.isMultiLineIndicator(indicatorType, {})) {
      // Remove multi-line indicators
      const lines = ['upper', 'middle', 'lower'];
      lines.forEach(line => {
        const series = this.indicatorSeries.get(`${indicatorType}_${line}`);
        if (series) {
          chart.removeSeries(series);
          this.indicatorSeries.delete(`${indicatorType}_${line}`);
        }
      });
    } else {
      // Remove single-line indicators
      const series = this.indicatorSeries.get(indicatorType);
      if (series) {
        chart.removeSeries(series);
        this.indicatorSeries.delete(indicatorType);
      }
    }
    
    console.log(`${indicatorType} indicator removed successfully`);
  }

  // Update all indicators when new data arrives
  async updateIndicators(newBarData) {
    if (this.activeIndicators.size === 0 || !historicalData) return;
    
    // Update historical data with the new bar
    if (newBarData) {
      const existingIndex = historicalData.findIndex(bar => bar.time === newBarData.time);
      if (existingIndex !== -1) {
        historicalData[existingIndex] = newBarData;
      } else {
        historicalData.push(newBarData);
      }
    }
    
    // Recalculate all active indicators
    for (const [type, config] of this.activeIndicators) {
      try {
        const requestData = {
          data: historicalData,
          period: config.period,
          config: config.config
        };

        const endpoint = this.getIndicatorEndpoint(type);
        const response = await fetch(`${this.apiBaseUrl}${endpoint}`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(requestData)
        });

        if (response.ok) {
          const result = await response.json();
          this.updateIndicatorSeries(type, result);
          
          // Update stored values
          this.activeIndicators.set(type, { 
            period: config.period, 
            values: result,
            config: config.config
          });
        }
      } catch (error) {
        console.error(`Error updating ${type} indicator:`, error);
      }
    }
  }

  // Update indicator series with new data
  updateIndicatorSeries(type, data) {
    if (this.isMultiLineIndicator(type, data)) {
      // Update multi-line indicators
      const lines = [
        { key: 'upper', data: data.upper },
        { key: 'middle', data: data.middle },
        { key: 'lower', data: data.lower }
      ];
      
      lines.forEach(line => {
        if (line.data && line.data.length > 0) {
          const series = this.indicatorSeries.get(`${type}_${line.key}`);
          if (series) {
            const latestIndex = line.data.length - 1;
            const latestTime = historicalData[historicalData.length - 1]?.time;
            
            if (latestTime) {
              series.update({ 
                time: latestTime, 
                value: line.data[latestIndex] 
              });
            }
          }
        }
      });
    } else {
      // Update single-line indicators
      const values = data.values || data;
      const series = this.indicatorSeries.get(type);
      if (series && Array.isArray(values) && values.length > 0) {
        const latestValue = values[values.length - 1];
        const latestTime = historicalData[historicalData.length - 1]?.time;
        
        if (latestTime && !isNaN(latestValue)) {
          series.update({ time: latestTime, value: latestValue });
        }
      }
    }
  }

  // Clear all indicators
  clearAllIndicators() {
    this.indicatorSeries.forEach((series) => {
      try {
        chart.removeSeries(series);
      } catch (error) {
        console.error('Error removing indicator series:', error);
      }
    });
    
    this.indicatorSeries.clear();
    this.activeIndicators.clear();
  }

  // Get active indicators for display
  getActiveIndicators() {
    return Array.from(this.activeIndicators.keys());
  }
}

// Create global indicator manager instance
const indicatorManager = new IndicatorManager();

// Global functions for the UI
const addIndicator = async (indicatorType) => {
  if (!indicatorType) {
    document.getElementById('indicators').value = '';
    return;
  }
  
  const period = prompt(`Enter period for ${indicatorType} (leave empty for default):`, '');
  if (period === null) {
    document.getElementById('indicators').value = '';
    return;
  }
  
  const periodNum = period ? parseInt(period) : null;
  
  await indicatorManager.addIndicator(indicatorType, periodNum);
  updateActiveIndicatorsDisplay();
  
  document.getElementById('indicators').value = '';
};

const removeIndicator = (indicatorType) => {
  indicatorManager.removeIndicator(indicatorType);
  updateActiveIndicatorsDisplay();
};

const updateActiveIndicatorsDisplay = () => {
  const container = document.getElementById('activeIndicators');
  if (!container) return;
  
  // Clear existing display
  container.innerHTML = '';
  
  // Add each active indicator
  const activeIndicators = indicatorManager.getActiveIndicators();
  activeIndicators.forEach(type => {
    const config = indicatorManager.activeIndicators.get(type);
    const indicatorTag = document.createElement('div');
    indicatorTag.className = 'indicator-tag';
    
    const label = document.createElement('span');
    label.textContent = `${type} ${config.period ? `(${config.period})` : ''}`;
    
    const removeBtn = document.createElement('button');
    removeBtn.className = 'indicator-remove';
    removeBtn.textContent = '×';
    removeBtn.title = `Remove ${type}`;
    removeBtn.onclick = () => removeIndicator(type);
    
    indicatorTag.appendChild(label);
    indicatorTag.appendChild(removeBtn);
    container.appendChild(indicatorTag);
  });
};

const clearAllIndicators = () => {
  indicatorManager.clearAllIndicators();
  updateActiveIndicatorsDisplay();
};

const updateIndicators = async (newBarData) => {
  await indicatorManager.updateIndicators(newBarData);
};

// Update status display
const updateStatus = (text, isConnected) => {
  const statusElement = document.getElementById('status');
  if (statusElement) {
    statusElement.textContent = text;
    if (isConnected) {
      statusElement.classList.add('connected');
    } else {
      statusElement.classList.remove('connected');
    }
  }
};

// Export functions for use in other files
if (typeof module !== 'undefined' && module.exports) {
  module.exports = {
    IndicatorManager,
    addIndicator,
    removeIndicator,
    updateActiveIndicatorsDisplay,
    clearAllIndicators,
    updateIndicators,
    updateStatus
  };
}