# Go OPC UA Simulator - Testing Guide

## Quick Start Testing

### 1. 자동 테스트 실행

가장 간단한 방법으로 전체 시스템을 검증합니다:

```bash
cd go-opcua-sim

# 빌드 (최초 1회만)
make build

# 자동 테스트 실행
./test_server.sh
```

테스트 스크립트는 다음을 자동으로 검증합니다:
- ✅ 설정 파일 존재 확인
- ✅ JSON 문법 검증
- ✅ 서버 시작
- ✅ 태그 생성
- ✅ PLC 엔진 초기화
- ✅ 센서 시뮬레이션
- ✅ OPC UA 서버 준비
- ✅ 실시간 데이터 업데이트

**예상 출력:**
```
==========================================
Go OPC UA Simulator - Test Script
==========================================

Test 1: Checking configuration files...
✓ sensors.json found
✓ plc_logic.lua found

Test 2: Validating JSON syntax...
✓ sensors.json is valid JSON

Test 3: Starting OPC UA server...
ℹ Server started with PID: 12345

...

==========================================
Test Summary
==========================================
✓ All critical tests passed!

Server Details:
  - PID: 12345
  - Tags: 26
  - Updates: 3
  - Log: /tmp/opcua_test.log
```

---

## 수동 테스트

### 테스트 1: 기본 서버 실행

```bash
# 서버 실행
./bin/server

# 또는 Makefile 사용
make run-server
```

**확인 사항:**
1. 서버가 에러 없이 시작되는가?
2. 태그가 생성되는가?
3. 센서 값이 주기적으로 업데이트되는가?

**성공 시 출력:**
```
=== Go OPC UA PLC Simulation Server ===
[CONFIG] Loaded 26 sensor definitions
[TAG] Generated 26 tags from sensor definitions

=== PLC Tag Summary ===
Total Tags: 26

By Type:
  Float64: 14 tags
  Int32:   4 tags
  Bool:    8 tags

...

[OPCUA] Server is ready to accept connections at opc.tcp://0.0.0.0:4840

=== Sensor Update #20 ===
  TemperatureSensor_Tank1 (%DF100): 29.574
  PressureSensor_Pump1 (%DF108): 5.922
  ...
```

### 테스트 2: 간단한 예제 실행

```bash
# 간단한 탱크 예제 실행
./bin/server -config examples/simple_tank.json -script examples/simple_tank.lua
```

**확인 사항:**
- 온도 센서 값이 변화하는가?
- 히터가 온도에 따라 ON/OFF되는가?
- PLC 로그가 10초마다 출력되는가?

**예상 로그:**
```
[PLC-LOGIC] Simple Tank Control System Initialized
[PLC-LOGIC] Target Temperature: 25°C
...
[PLC-LOGIC] Status Report #1:
[PLC-LOGIC]   Temperature: 26.34°C
[PLC-LOGIC]   Level: 52.3%
[PLC-LOGIC]   Heater: OFF
[PLC-LOGIC]   Pump: OFF
```

### 테스트 3: PLC 로직 없이 실행

```bash
# 센서 시뮬레이션만 실행
./bin/server -plc=false
```

**확인 사항:**
- 센서 값이 업데이트되는가?
- Lua 엔진 관련 로그가 없는가?

### 테스트 4: 다양한 스캔 주기 테스트

```bash
# 빠른 스캔 (20ms)
./bin/server -scantime 20

# 느린 스캔 (500ms)
./bin/server -scantime 500
```

**확인 사항:**
- 스캔 카운터가 적절한 속도로 증가하는가?
- 시스템 리소스 사용량이 적절한가?

---

## 센서별 동작 검증

### 온도 센서 검증

```bash
# 서버 실행 후 특정 센서만 모니터링
./bin/server 2>&1 | grep "TemperatureSensor_Tank1"
```

**예상 출력:**
```
  TemperatureSensor_Tank1 (%DF100): 25.234
  TemperatureSensor_Tank1 (%DF100): 26.891
  TemperatureSensor_Tank1 (%DF100): 28.456
  TemperatureSensor_Tank1 (%DF100): 29.123
```

**확인:**
- ✅ 값이 사인파 형태로 변화
- ✅ 노이즈 포함
- ✅ 설정된 범위 내에서 변화

### 압력 센서 검증

```bash
./bin/server 2>&1 | grep "PressureSensor_Pump1"
```

**확인:**
- ✅ 0 → 10 bar로 상승 (램프 업)
- ✅ 10 bar 유지 (홀드)
- ✅ 10 → 0 bar로 하강 (램프 다운)
- ✅ 주기적으로 반복

### 디지털 센서 검증

```bash
./bin/server 2>&1 | grep "DoorSensor_MainEntrance"
```

**확인:**
- ✅ 0 또는 1 값만 출력
- ✅ 설정된 주기로 토글

### 모터 센서 검증

```bash
./bin/server 2>&1 | grep "StepMotor_Conveyor"
```

**확인:**
- ✅ 위치 값이 변화
- ✅ 가속/감속 동작

---

## PLC 로직 검증

### 기본 로직 테스트

1. **서버 시작**
```bash
./bin/server -script plc_logic.lua
```

2. **Lua 로그 확인**
```bash
./bin/server 2>&1 | grep "PLC-LOGIC"
```

**확인 사항:**
- ✅ `init()` 함수가 1회 실행됨
- ✅ `run_logic()`이 주기적으로 실행됨
- ✅ 에러 메시지가 없음

### 커스텀 로직 테스트

**테스트용 Lua 파일 생성 (test_logic.lua):**
```lua
test_counter = test_counter or 0

function init()
    plc_log("Test logic initialized")
end

function run_logic()
    test_counter = test_counter + 1

    if test_counter % 10 == 0 then
        plc_log(string.format("Test counter: %d", test_counter))
    end
end
```

**실행:**
```bash
./bin/server -script test_logic.lua
```

**확인:**
- 10번마다 카운터 로그 출력
- 에러 없이 실행

---

## 성능 테스트

### CPU 사용률 테스트

```bash
# 서버 실행
./bin/server &
SERVER_PID=$!

# CPU 사용률 모니터링
top -p $SERVER_PID

# 또는
ps -p $SERVER_PID -o %cpu,%mem,cmd
```

**기대값:**
- CPU: < 5% (일반적인 경우)
- 메모리: < 50MB

### 메모리 누수 테스트

```bash
# 서버를 장시간 실행
./bin/server &
SERVER_PID=$!

# 주기적으로 메모리 확인
while true; do
    ps -p $SERVER_PID -o rss,vsz,cmd
    sleep 60
done
```

**확인:**
- RSS(실제 메모리)가 계속 증가하지 않는가?
- 안정적인 메모리 사용량 유지

### 스캔 주기 정확도 테스트

```bash
# 빠른 스캔으로 실행
./bin/server -scantime 50 2>&1 | grep "Completed.*scan cycles"
```

**확인:**
- 100 스캔 주기가 약 5초 (50ms × 100)에 완료되는가?

---

## 에러 시나리오 테스트

### 잘못된 설정 파일

```bash
# 존재하지 않는 파일
./bin/server -config nonexistent.json
```

**예상 출력:**
```
Failed to load sensor config: open nonexistent.json: no such file or directory
```

### 잘못된 JSON 문법

**bad_config.json:**
```json
{
  "sensors": [
    {
      "name": "Test"
      // 잘못된 JSON (쉼표 누락)
    }
  ]
}
```

```bash
./bin/server -config bad_config.json
```

**예상 출력:**
```
Failed to load sensor config: failed to parse config JSON: ...
```

### 잘못된 Lua 스크립트

**bad_logic.lua:**
```lua
function run_logic()
    undefined_function()  -- 에러!
end
```

```bash
./bin/server -script bad_logic.lua
```

**예상 출력:**
```
[LUA] Scan #1 error: run_logic execution error: ...
```

### 누락된 run_logic 함수

**no_runlogic.lua:**
```lua
function init()
    plc_log("Init only")
end
-- run_logic() 함수 없음!
```

```bash
./bin/server -script no_runlogic.lua
```

**예상 출력:**
```
Failed to initialize Lua engine: Lua script must define 'run_logic()' function
```

---

## 통합 테스트 시나리오

### 시나리오 1: 온도 제어 시스템

**목표:** 히터가 온도에 따라 자동으로 ON/OFF되는지 확인

1. **서버 시작**
```bash
./bin/server -config examples/simple_tank.json -script examples/simple_tank.lua
```

2. **모니터링**
```bash
# 다른 터미널에서
tail -f /tmp/opcua_test.log | grep -E "Temperature|Heater"
```

3. **확인 사항**
- 온도가 23°C 이하면 히터 ON
- 온도가 27°C 이상이면 히터 OFF
- 히스테리시스 동작 확인

### 시나리오 2: 센서 값 범위 확인

**목표:** 모든 센서 값이 설정된 범위 내에 있는지 확인

```bash
# 서버 실행 및 로그 저장
./bin/server > sensor_test.log 2>&1 &

# 10분 실행
sleep 600

# 로그 분석
cat sensor_test.log | grep "TemperatureSensor_Tank1" | \
  awk '{print $NF}' | \
  awk 'BEGIN {min=999; max=0}
       {if($1<min) min=$1; if($1>max) max=$1}
       END {print "Min:", min, "Max:", max}'
```

**확인:**
- 최소/최대값이 설정 범위 내인가?

### 시나리오 3: 장시간 안정성 테스트

```bash
# 서버를 1시간 실행
timeout 3600 ./bin/server

# 또는 백그라운드 실행
./bin/server > long_test.log 2>&1 &
SERVER_PID=$!

# 1시간 후 종료
sleep 3600
kill $SERVER_PID

# 로그 분석
grep -i "error\|panic\|fail" long_test.log
```

**확인:**
- 에러 없이 실행 완료
- 메모리 누수 없음
- 일정한 성능 유지

---

## 벤치마크

### 센서 수에 따른 성능

```bash
# 26개 센서 (기본)
time timeout 60 ./bin/server

# 10개 센서 (간단)
time timeout 60 ./bin/server -config examples/simple_tank.json
```

### 스캔 주기별 성능

```bash
# 20ms 스캔
time timeout 60 ./bin/server -scantime 20

# 100ms 스캔 (기본)
time timeout 60 ./bin/server -scantime 100

# 500ms 스캔
time timeout 60 ./bin/server -scantime 500
```

---

## 체크리스트

### 빌드 확인
- [ ] `make build` 성공
- [ ] `bin/server` 파일 생성
- [ ] `bin/client` 파일 생성

### 설정 파일 확인
- [ ] `sensors.json` 존재
- [ ] `plc_logic.lua` 존재
- [ ] JSON 문법 유효

### 서버 시작 확인
- [ ] 에러 없이 시작
- [ ] 태그 생성 로그 출력
- [ ] OPC UA 서버 준비 메시지 출력

### 센서 시뮬레이션 확인
- [ ] 센서 값이 실시간 변경
- [ ] 각 센서 타입이 예상대로 동작
- [ ] 업데이트 로그 주기적 출력

### PLC 로직 확인
- [ ] Lua 엔진 초기화 성공
- [ ] `init()` 함수 실행
- [ ] `run_logic()` 주기적 실행
- [ ] 스캔 카운터 증가

### 성능 확인
- [ ] CPU 사용률 < 5%
- [ ] 메모리 < 50MB
- [ ] 메모리 누수 없음
- [ ] 스캔 주기 정확도 유지

---

## 문제 해결

### 서버가 시작되지 않음

1. **빌드 확인**
```bash
make clean
make build
```

2. **설정 파일 확인**
```bash
ls -la sensors.json plc_logic.lua
```

3. **로그 확인**
```bash
./bin/server 2>&1 | tee debug.log
```

### 센서 값이 변하지 않음

1. **enabled 확인**
```json
{
  "enabled": true  // false면 동작 안 함
}
```

2. **PLC 로직 확인**
```bash
# PLC 비활성화로 테스트
./bin/server -plc=false
```

### 메모리 사용량 증가

1. **Lua 스크립트 확인**
```lua
-- 전역 변수가 계속 증가하는지 확인
```

2. **프로파일링**
```bash
go tool pprof http://localhost:6060/debug/pprof/heap
```

---

## 추가 테스트 도구

### 로그 분석 스크립트

```bash
#!/bin/bash
# analyze_logs.sh

LOG_FILE=$1

echo "=== Log Analysis ==="
echo ""

echo "Total sensor updates:"
grep -c "Sensor Update" $LOG_FILE

echo ""
echo "PLC scan cycles:"
grep "Completed.*scan cycles" $LOG_FILE | tail -1

echo ""
echo "Errors:"
grep -i "error\|fail" $LOG_FILE | wc -l

echo ""
echo "Temperature range:"
grep "TemperatureSensor_Tank1" $LOG_FILE | \
  awk '{print $NF}' | \
  awk 'BEGIN {min=999; max=0}
       {if($1<min) min=$1; if($1>max) max=$1}
       END {print "Min:", min, "Max:", max}'
```

### 성능 모니터링 스크립트

```bash
#!/bin/bash
# monitor_performance.sh

PID=$1

echo "Monitoring PID: $PID"
echo "Time,CPU%,MEM(KB)"

while ps -p $PID > /dev/null; do
    STATS=$(ps -p $PID -o %cpu,rss --no-headers)
    echo "$(date +%s),$STATS"
    sleep 1
done
```

---

## 결론

이 테스트 가이드를 통해 다음을 확인할 수 있습니다:

1. ✅ 시스템이 정상적으로 빌드되고 실행됨
2. ✅ 모든 센서 타입이 올바르게 동작함
3. ✅ PLC 로직이 예상대로 실행됨
4. ✅ 성능이 허용 범위 내에 있음
5. ✅ 장시간 안정적으로 동작함

문제가 발생하면 MANUAL.md의 트러블슈팅 섹션을 참고하세요.
