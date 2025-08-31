#!/bin/bash

# Professional Backtesting Platform Startup Script
# TopStepX - Advanced Trading Strategy Testing

set -e  # Exit on any error

PORT=8888
FRONTEND_PORT=3000

echo "🚀 Starting TopStepX Professional Backtesting Platform"
echo "=================================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

# Function to kill processes on specific port
kill_port() {
    local port=$1
    print_step "Killing any processes running on port $port..."
    
    # Kill processes using the port
    local pids=$(lsof -ti:$port 2>/dev/null || true)
    if [ ! -z "$pids" ]; then
        echo "$pids" | xargs kill -9 2>/dev/null || true
        print_status "Killed processes on port $port: $pids"
    else
        print_status "No processes found on port $port"
    fi
    
    # Also kill by process name patterns
    pkill -f "trading-api" 2>/dev/null || true
    pkill -f "python3.*http.server" 2>/dev/null || true
    pkill -f ":$port" 2>/dev/null || true
    
    sleep 2
}

# Function to check if port is free
check_port() {
    local port=$1
    if lsof -i:$port >/dev/null 2>&1; then
        print_error "Port $port is still in use!"
        return 1
    else
        print_status "Port $port is free"
        return 0
    fi
}

# Function to wait for service to be ready
wait_for_service() {
    local url=$1
    local service_name=$2
    local max_attempts=30
    local attempt=1
    
    print_step "Waiting for $service_name to be ready..."
    
    while [ $attempt -le $max_attempts ]; do
        if curl -s -f "$url" >/dev/null 2>&1; then
            print_status "$service_name is ready!"
            return 0
        fi
        
        echo -n "."
        sleep 1
        attempt=$((attempt + 1))
    done
    
    print_error "$service_name failed to start within $max_attempts seconds"
    return 1
}

# Main startup process
main() {
    cd /app
    
    print_step "Step 1: Cleaning up existing processes"
    kill_port $PORT
    kill_port $FRONTEND_PORT
    
    # Verify ports are free
    if ! check_port $PORT; then
        print_error "Failed to free port $PORT. Exiting."
        exit 1
    fi
    
    print_step "Step 2: Installing Go (if needed)"
    if ! command -v go &> /dev/null; then
        print_warning "Go not found. Installing Go 1.22..."
        cd /usr/local
        rm -rf go 2>/dev/null || true
        wget -q https://go.dev/dl/go1.22.5.linux-arm64.tar.gz
        tar -xzf go1.22.5.linux-arm64.tar.gz
        rm go1.22.5.linux-arm64.tar.gz
        export PATH=/usr/local/go/bin:$PATH
        cd /app
    else
        export PATH=/usr/local/go/bin:$PATH
    fi
    
    print_status "Go version: $(go version)"
    
    print_step "Step 3: Updating backend configuration for port $PORT"
    
    # Update backend port in main.go
    sed -i "s/:8001/:$PORT/g" /app/backend/main.go
    sed -i "s/8001/$PORT/g" /app/backend/main.go
    
    # Update frontend API configuration
    sed -i "s/:8001/:$PORT/g" /app/frontend/js/backtesting-platform.js
    sed -i "s/8001/$PORT/g" /app/frontend/js/backtesting-platform.js
    
    print_step "Step 4: Building backend"
    cd /app/backend
    
    print_status "Running go mod tidy..."
    go mod tidy
    
    print_status "Building trading API..."
    go build -o trading-api-enhanced .
    
    if [ ! -f "trading-api-enhanced" ]; then
        print_error "Failed to build backend binary"
        exit 1
    fi
    
    print_status "Backend built successfully"
    
    print_step "Step 5: Starting backend service on port $PORT"
    ./trading-api-enhanced > backend.log 2>&1 &
    BACKEND_PID=$!
    
    print_status "Backend started with PID: $BACKEND_PID"
    
    # Wait for backend to be ready
    if ! wait_for_service "http://localhost:$PORT/api/health" "Backend API"; then
        print_error "Backend failed to start. Check logs:"
        tail -n 20 backend.log
        kill $BACKEND_PID 2>/dev/null || true
        exit 1
    fi
    
    print_step "Step 6: Starting frontend service on port $FRONTEND_PORT"
    cd /app/frontend
    
    python3 -m http.server $FRONTEND_PORT > frontend.log 2>&1 &
    FRONTEND_PID=$!
    
    print_status "Frontend started with PID: $FRONTEND_PID"
    
    # Wait for frontend to be ready
    if ! wait_for_service "http://localhost:$FRONTEND_PORT/" "Frontend Server"; then
        print_error "Frontend failed to start"
        kill $BACKEND_PID $FRONTEND_PID 2>/dev/null || true
        exit 1
    fi
    
    print_step "Step 7: Testing API connectivity"
    
    # Test backend health
    if ! curl -s http://localhost:$PORT/api/health | grep -q "OK"; then
        print_error "Backend health check failed"
        exit 1
    fi
    print_status "Backend health check: OK"
    
    # Test strategies endpoint
    STRATEGY_COUNT=$(curl -s http://localhost:$PORT/api/strategies/list | jq length 2>/dev/null || echo "0")
    print_status "Loaded $STRATEGY_COUNT strategies"
    
    # Test enhanced backtest endpoint
    print_status "Testing backtest endpoint..."
    BACKTEST_RESULT=$(curl -s -X POST http://localhost:$PORT/api/backtest/enhanced \
        -H "Content-Type: application/json" \
        -d '{"instrument": "DummyTest", "strategy": "Test", "initialCapital": 50000}' \
        | jq '.summary.totalTrades' 2>/dev/null || echo "0")
    
    if [ "$BACKTEST_RESULT" != "0" ] && [ "$BACKTEST_RESULT" != "null" ]; then
        print_status "Backtest endpoint working (executed $BACKTEST_RESULT trades)"
    else
        print_warning "Backtest endpoint may have issues"
    fi
    
    print_step "Step 8: Platform ready!"
    echo ""
    echo "🎉 TopStepX Professional Backtesting Platform is now running!"
    echo "=================================================="
    echo -e "${GREEN}Frontend URL:${NC}  http://localhost:$FRONTEND_PORT/"
    echo -e "${GREEN}Backend API:${NC}   http://localhost:$PORT/api/"
    echo -e "${GREEN}API Health:${NC}    http://localhost:$PORT/api/health"
    echo -e "${GREEN}Strategies:${NC}    http://localhost:$PORT/api/strategies/list"
    echo ""
    echo -e "${BLUE}Process IDs:${NC}"
    echo "  Backend PID:  $BACKEND_PID"
    echo "  Frontend PID: $FRONTEND_PID"
    echo ""
    echo -e "${YELLOW}To stop the platform:${NC}"
    echo "  kill $BACKEND_PID $FRONTEND_PID"
    echo "  or run: pkill -f 'trading-api|python3.*http.server'"
    echo ""
    echo -e "${YELLOW}Log files:${NC}"
    echo "  Backend:  /app/backend/backend.log"
    echo "  Frontend: /app/frontend/frontend.log"
    echo ""
    echo "✨ Ready for professional strategy backtesting!"
    
    # Keep script running and monitor services
    print_step "Monitoring services (Press Ctrl+C to stop)..."
    
    trap 'print_warning "Shutting down services..."; kill $BACKEND_PID $FRONTEND_PID 2>/dev/null; exit 0' INT TERM
    
    while true; do
        if ! kill -0 $BACKEND_PID 2>/dev/null; then
            print_error "Backend process died! Check logs."
            break
        fi
        
        if ! kill -0 $FRONTEND_PID 2>/dev/null; then
            print_error "Frontend process died!"
            break
        fi
        
        sleep 10
    done
}

# Handle script interruption
cleanup() {
    print_warning "Cleaning up..."
    kill_port $PORT
    kill_port $FRONTEND_PORT
    exit 0
}

trap cleanup EXIT

# Check if running as root or with sufficient permissions
if [ "$EUID" -ne 0 ] && ! groups | grep -q docker; then
    print_warning "This script may need root permissions for port management"
fi

# Run main function
main "$@"