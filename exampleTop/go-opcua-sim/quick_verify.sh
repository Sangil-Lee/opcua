#!/bin/bash
# Quick OPC UA Server Verification Script

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "=========================================="
echo "Quick OPC UA Server Verification"
echo "=========================================="
echo ""

# 서버 시작
echo "Starting server..."
./bin/server > /tmp/quick_verify.log 2>&1 &
PID=$!
echo -e "${YELLOW}Server PID: $PID${NC}"
echo ""

# 초기화 대기
echo "Waiting for initialization (8 seconds)..."
sleep 8

echo ""
echo "=== Verification Results ==="
echo ""

# 1. 노드 생성 확인
NODES=$(grep -c "ns=2;i=" /tmp/quick_verify.log 2>/dev/null || echo "0")
if [ "$NODES" -gt 0 ]; then
    echo -e "${GREEN}✓${NC} OPC UA nodes created: $NODES"
else
    echo -e "${RED}✗${NC} No nodes created"
fi

# 2. 센서 업데이트 확인
UPDATES=$(grep -c "Sensor Update" /tmp/quick_verify.log 2>/dev/null || echo "0")
if [ "$UPDATES" -gt 0 ]; then
    echo -e "${GREEN}✓${NC} Sensor updates detected: $UPDATES"
else
    echo -e "${YELLOW}!${NC} No sensor updates yet (may need more time)"
fi

# 3. PLC 로직 확인
if grep -q "PLC logic initialized" /tmp/quick_verify.log 2>/dev/null; then
    echo -e "${GREEN}✓${NC} PLC logic running"
else
    echo -e "${YELLOW}!${NC} PLC logic not detected"
fi

# 4. 서버 준비 확인
if grep -q "Server is ready" /tmp/quick_verify.log 2>/dev/null; then
    echo -e "${GREEN}✓${NC} OPC UA server ready"
else
    echo -e "${YELLOW}!${NC} Server status unclear"
fi

# 5. 노드 매핑 샘플
echo ""
echo "=== Node Mapping Sample ==="
grep -A 5 "Tag Name.*NodeID.*Data Type" /tmp/quick_verify.log 2>/dev/null | \
  tail -5 || echo "No mapping found"

# 6. 센서 값 샘플
echo ""
echo "=== Current Sensor Values ==="
grep "TemperatureSensor_Tank1\|PressureSensor_Pump1\|MotorSpeed_Conveyor" /tmp/quick_verify.log 2>/dev/null | \
  head -3 | sed 's/^/  /' || echo "No sensor data yet"

# 7. 에러 확인
echo ""
ERRORS=$(grep -i "error\|fail" /tmp/quick_verify.log 2>/dev/null | wc -l)
if [ "$ERRORS" -eq 0 ]; then
    echo -e "${GREEN}✓${NC} No errors detected"
else
    echo -e "${YELLOW}!${NC} Errors found: $ERRORS"
    echo "Check /tmp/quick_verify.log for details"
fi

echo ""
echo "=========================================="
echo ""
echo "Server is still running (PID: $PID)"
echo "Options:"
echo "  1. View full log: cat /tmp/quick_verify.log"
echo "  2. Monitor live: tail -f /tmp/quick_verify.log"
echo "  3. Stop server: kill $PID"
echo ""
echo "Press Enter to stop server..."
read

# 서버 종료
echo "Stopping server..."
kill $PID 2>/dev/null || true
sleep 1

if ps -p $PID > /dev/null 2>&1; then
    kill -9 $PID 2>/dev/null || true
fi

echo -e "${GREEN}✓${NC} Server stopped"
echo ""
echo "Log saved at: /tmp/quick_verify.log"
