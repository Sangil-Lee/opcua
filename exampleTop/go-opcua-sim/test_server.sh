#!/bin/bash
# OPC UA Simulator Test Script

set -e

echo "=========================================="
echo "Go OPC UA Simulator - Test Script"
echo "=========================================="
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if server binary exists
if [ ! -f "bin/server" ]; then
    echo -e "${RED}Error: Server binary not found${NC}"
    echo "Please run 'make build' first"
    exit 1
fi

# Function to print success
print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

# Function to print error
print_error() {
    echo -e "${RED}✗${NC} $1"
}

# Function to print info
print_info() {
    echo -e "${YELLOW}ℹ${NC} $1"
}

# Test 1: Check if sensors.json exists
echo "Test 1: Checking configuration files..."
if [ -f "sensors.json" ]; then
    print_success "sensors.json found"
else
    print_error "sensors.json not found"
    exit 1
fi

if [ -f "plc_logic.lua" ]; then
    print_success "plc_logic.lua found"
else
    print_error "plc_logic.lua not found"
    exit 1
fi
echo ""

# Test 2: Validate JSON syntax
echo "Test 2: Validating JSON syntax..."
if command -v python3 &> /dev/null; then
    if python3 -c "import json; json.load(open('sensors.json'))" 2>/dev/null; then
        print_success "sensors.json is valid JSON"
    else
        print_error "sensors.json has syntax errors"
        exit 1
    fi
else
    print_info "Python3 not found, skipping JSON validation"
fi
echo ""

# Test 3: Start server in background
echo "Test 3: Starting OPC UA server..."
./bin/server -config sensors.json -script plc_logic.lua > /tmp/opcua_test.log 2>&1 &
SERVER_PID=$!
print_info "Server started with PID: $SERVER_PID"
echo ""

# Wait for server to initialize
echo "Waiting for server initialization (5 seconds)..."
sleep 5

# Test 4: Check if server is still running
echo "Test 4: Checking server status..."
if ps -p $SERVER_PID > /dev/null 2>&1; then
    print_success "Server process is running"
else
    print_error "Server process terminated unexpectedly"
    echo ""
    echo "Server log:"
    cat /tmp/opcua_test.log
    exit 1
fi
echo ""

# Test 5: Check server log for errors
echo "Test 5: Checking server logs for errors..."
if grep -i "error\|fail\|panic" /tmp/opcua_test.log > /dev/null 2>&1; then
    print_error "Errors found in server log"
    echo ""
    echo "Error messages:"
    grep -i "error\|fail\|panic" /tmp/opcua_test.log
    kill $SERVER_PID 2>/dev/null || true
    exit 1
else
    print_success "No errors in server log"
fi
echo ""

# Test 6: Verify tag generation
echo "Test 6: Verifying tag generation..."
if grep "Generated [0-9]* tags from sensor definitions" /tmp/opcua_test.log > /dev/null 2>&1; then
    TAG_COUNT=$(grep "Generated [0-9]* tags" /tmp/opcua_test.log | grep -o "[0-9]*" | head -1)
    print_success "Generated $TAG_COUNT tags"
else
    print_error "Tag generation not found in log"
    kill $SERVER_PID 2>/dev/null || true
    exit 1
fi
echo ""

# Test 7: Verify PLC initialization
echo "Test 7: Verifying PLC Lua engine..."
if grep "Lua engine started" /tmp/opcua_test.log > /dev/null 2>&1; then
    print_success "PLC Lua engine started successfully"
else
    print_error "PLC Lua engine failed to start"
    kill $SERVER_PID 2>/dev/null || true
    exit 1
fi
echo ""

# Test 8: Verify sensor simulation
echo "Test 8: Verifying sensor simulation..."
if grep "Sensor manager started" /tmp/opcua_test.log > /dev/null 2>&1; then
    print_success "Sensor manager started"
else
    print_error "Sensor manager failed to start"
    kill $SERVER_PID 2>/dev/null || true
    exit 1
fi
echo ""

# Test 9: Wait and check for sensor updates
echo "Test 9: Checking for sensor updates..."
echo "Waiting 3 seconds for sensor updates..."
sleep 3

if grep "Sensor Update" /tmp/opcua_test.log > /dev/null 2>&1; then
    UPDATE_COUNT=$(grep -c "Sensor Update" /tmp/opcua_test.log)
    print_success "Sensor updates detected ($UPDATE_COUNT updates)"
else
    print_error "No sensor updates found"
    kill $SERVER_PID 2>/dev/null || true
    exit 1
fi
echo ""

# Test 10: Verify OPC UA server
echo "Test 10: Verifying OPC UA server..."
if grep "Server is ready to accept connections" /tmp/opcua_test.log > /dev/null 2>&1; then
    print_success "OPC UA server is ready"
else
    print_error "OPC UA server not ready"
    kill $SERVER_PID 2>/dev/null || true
    exit 1
fi
echo ""

# Test 11: Check PLC scan cycles
echo "Test 11: Checking PLC scan cycles..."
if grep "Completed [0-9]* scan cycles" /tmp/opcua_test.log > /dev/null 2>&1; then
    SCAN_COUNT=$(grep "Completed [0-9]* scan cycles" /tmp/opcua_test.log | tail -1 | grep -o "[0-9]*")
    print_success "PLC completed $SCAN_COUNT scan cycles"
else
    print_info "PLC scan cycles not yet reached 100 (may need more time)"
fi
echo ""

# Display summary
echo "=========================================="
echo "Test Summary"
echo "=========================================="
print_success "All critical tests passed!"
echo ""
echo "Server Details:"
echo "  - PID: $SERVER_PID"
echo "  - Tags: $TAG_COUNT"
echo "  - Updates: $UPDATE_COUNT"
echo "  - Log: /tmp/opcua_test.log"
echo ""

# Show recent sensor values
echo "Recent Sensor Values:"
echo "---------------------"
tail -20 /tmp/opcua_test.log | grep -E "TemperatureSensor|PressureSensor|MotorSpeed" | head -5
echo ""

# Ask user to stop server
echo "=========================================="
print_info "Server is still running (PID: $SERVER_PID)"
echo ""
echo "Options:"
echo "  1. View live logs: tail -f /tmp/opcua_test.log"
echo "  2. Stop server: kill $SERVER_PID"
echo "  3. Keep running and press Ctrl+C when done"
echo ""
echo "Press Enter to stop the server, or Ctrl+C to keep it running..."
read -r

# Stop server
echo "Stopping server..."
kill $SERVER_PID 2>/dev/null || true
sleep 1

if ps -p $SERVER_PID > /dev/null 2>&1; then
    print_info "Force killing server..."
    kill -9 $SERVER_PID 2>/dev/null || true
fi

print_success "Server stopped"
echo ""
echo "Test complete!"
