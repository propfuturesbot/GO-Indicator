#!/bin/bash

# Check TopStepX Professional Backtesting Platform Status

echo "📊 TopStepX Professional Backtesting Platform Status"
echo "=================================================="

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

check_service() {
    local url=$1
    local name=$2
    
    if curl -s -f "$url" >/dev/null 2>&1; then
        echo -e "${GREEN}✅ $name${NC} - Running"
        return 0
    else
        echo -e "${RED}❌ $name${NC} - Not responding"
        return 1
    fi
}

check_port() {
    local port=$1
    local name=$2
    
    if lsof -i:$port >/dev/null 2>&1; then
        local pid=$(lsof -ti:$port 2>/dev/null | head -1)
        echo -e "${GREEN}✅ Port $port${NC} - In use (PID: $pid) - $name"
        return 0
    else
        echo -e "${RED}❌ Port $port${NC} - Free - $name"
        return 1
    fi
}

echo "🔍 Checking Services:"
check_service "http://localhost:8888/api/health" "Backend API (Port 8888)"
check_service "http://localhost:3000/" "Frontend Server (Port 3000)"

echo ""
echo "🔍 Checking Ports:"
check_port 8888 "Backend API"
check_port 3000 "Frontend Server"

echo ""
echo "🔍 Running Processes:"
ps aux | grep -E "(trading-api|python3.*http.server)" | grep -v grep || echo "No related processes found"

echo ""
echo "🔍 Quick API Test:"
if curl -s http://localhost:8888/api/health 2>/dev/null | jq . 2>/dev/null; then
    echo -e "${GREEN}✅ Backend API responding correctly${NC}"
else
    echo -e "${RED}❌ Backend API not responding${NC}"
fi

STRATEGY_COUNT=$(curl -s http://localhost:8888/api/strategies/list 2>/dev/null | jq length 2>/dev/null || echo "0")
if [ "$STRATEGY_COUNT" -gt "0" ]; then
    echo -e "${GREEN}✅ Strategy API working - $STRATEGY_COUNT strategies loaded${NC}"
else
    echo -e "${RED}❌ Strategy API not working${NC}"
fi

echo ""
echo "🔗 Access URLs:"
echo "   Frontend: http://localhost:3000/"
echo "   Backend:  http://localhost:8888/api/"
echo "   Health:   http://localhost:8888/api/health"