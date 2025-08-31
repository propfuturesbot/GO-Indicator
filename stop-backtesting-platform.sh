#!/bin/bash

# Stop TopStepX Professional Backtesting Platform

echo "🛑 Stopping TopStepX Professional Backtesting Platform"
echo "=================================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

# Kill all related processes
print_status "Stopping backend services..."
pkill -f "trading-api" 2>/dev/null || true

print_status "Stopping frontend services..."
pkill -f "python3.*http.server" 2>/dev/null || true

# Kill processes on specific ports
for port in 8888 3000 8001 8080; do
    pids=$(lsof -ti:$port 2>/dev/null || true)
    if [ ! -z "$pids" ]; then
        echo "$pids" | xargs kill -9 2>/dev/null || true
        print_status "Killed processes on port $port"
    fi
done

print_status "All services stopped successfully"
echo "✅ Platform shutdown complete"