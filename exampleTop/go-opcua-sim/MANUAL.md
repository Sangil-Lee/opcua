# Go OPC UA Simulator - 사용 매뉴얼

## 목차
1. [빠른 시작](#빠른-시작)
2. [서버 실행 및 검증](#서버-실행-및-검증)
3. [센서 설정 가이드](#센서-설정-가이드)
4. [PLC 로직 작성 가이드](#plc-로직-작성-가이드)
5. [고급 사용법](#고급-사용법)
6. [트러블슈팅](#트러블슈팅)
7. [실전 예제](#실전-예제)

---

## 빠른 시작

### 1. 프로젝트 빌드

```bash
cd go-opcua-sim

# 의존성 다운로드
make deps

# 전체 빌드
make build
```

빌드 후 `bin/` 디렉토리에 다음 바이너리가 생성됩니다:
- `bin/server`: OPC UA 시뮬레이터 서버
- `bin/client`: OPC UA 테스트 클라이언트

### 2. 서버 실행

```bash
# 기본 설정으로 실행
make run-server

# 또는 직접 실행
./bin/server
```

### 3. 서버 동작 확인

서버가 시작되면 다음과 같은 출력을 확인할 수 있습니다:

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

=== OPC UA Server Tag Mapping ===
Tag Name                                 NodeID               Data Type
------------------------------------------------------------
TemperatureSensor_Tank1                  ns=2;i=1000          Double
TemperatureSensor_Tank2                  ns=2;i=1012          Double
...

[OPCUA] Server is ready to accept connections at opc.tcp://0.0.0.0:4840
```

---

## 서버 실행 및 검증

### 서버 실행 옵션

```bash
./bin/server [옵션]
```

**사용 가능한 옵션:**

| 옵션 | 기본값 | 설명 |
|------|--------|------|
| `-config` | sensors.json | 센서 설정 파일 경로 |
| `-script` | plc_logic.lua | PLC Lua 스크립트 파일 경로 |
| `-scantime` | 100 | PLC 스캔 주기 (밀리초) |
| `-plc` | true | PLC 로직 활성화 여부 |
| `-endpoint` | opc.tcp://0.0.0.0:4840 | OPC UA 서버 엔드포인트 |

**실행 예제:**

```bash
# 1. 기본 설정으로 실행
./bin/server

# 2. PLC 로직 비활성화 (센서만 실행)
./bin/server -plc=false

# 3. 빠른 스캔 주기 (50ms)
./bin/server -scantime 50

# 4. 커스텀 설정 파일 사용
./bin/server -config my_sensors.json -script my_logic.lua

# 5. 다른 포트로 실행
./bin/server -endpoint "opc.tcp://0.0.0.0:5000"
```

### 서버 동작 검증 방법

#### 방법 1: 로그 확인

서버 실행 후 주기적으로 센서 업데이트 로그가 출력됩니다:

```
=== Sensor Update #100 ===
  TemperatureSensor_Tank1 (%DF100): 29.574
  PressureSensor_Pump1 (%DF108): 5.922
  DoorSensor_MainEntrance (%MW0): 1.000
  MotorSpeed_Conveyor (%DW200): 50.000
  ...
```

**확인 항목:**
- ✅ 센서 값이 실시간으로 변경되는가?
- ✅ PLC 스캔 카운터가 증가하는가? (`[LUA] Completed 100 scan cycles`)
- ✅ 에러 메시지가 없는가?

#### 방법 2: 센서 값 변화 모니터링

특정 센서의 값이 올바르게 변화하는지 확인:

```bash
# 서버 로그를 grep으로 필터링
./bin/server 2>&1 | grep "TemperatureSensor_Tank1"
```

출력 예시:
```
  TemperatureSensor_Tank1 (%DF100): 25.234
  TemperatureSensor_Tank1 (%DF100): 27.891
  TemperatureSensor_Tank1 (%DF100): 30.456
  TemperatureSensor_Tank1 (%DF100): 32.123
```

#### 방법 3: 태그 매핑 확인

서버 시작 시 모든 태그의 OPC UA 노드 매핑을 확인:

```
=== OPC UA Server Tag Mapping ===
Tag Name                                 NodeID               Data Type
------------------------------------------------------------
TemperatureSensor_Tank1                  ns=2;i=1000          Double
LevelSensor_Tank1                        ns=2;i=1001          Double
RelayActuator_Pump1                      ns=2;i=1002          Boolean
...
```

**노드 ID 규칙:**
- 네임스페이스: `ns=2`
- 인덱스: `i=1000`부터 순차적으로 할당
- 데이터 타입: Float64 → Double, Bool → Boolean, Int32 → Int32

---

## 센서 설정 가이드

### sensors.json 파일 구조

```json
{
  "sensors": [
    {
      "name": "센서명",
      "type": "센서타입",
      "enabled": true,
      "address": "PLC주소",
      "updateIntervalMs": 100,
      "parameters": {
        // 센서 타입별 파라미터
      },
      "description": "설명"
    }
  ]
}
```

### 지원하는 센서 타입

#### 1. 온도 센서 (temperature)

사인파 + 가우시안 노이즈로 실제 온도 센서 시뮬레이션

```json
{
  "name": "TemperatureSensor_Tank1",
  "type": "temperature",
  "enabled": true,
  "address": "%DF100",
  "updateIntervalMs": 100,
  "parameters": {
    "baseTemp": 25.0,      // 기준 온도 (°C)
    "amplitude": 10.0,     // 변동 폭 (°C)
    "period": 30.0,        // 주기 (초)
    "noiseStdDev": 0.5,    // 노이즈 표준편차
    "minValue": 0.0,       // 최소값
    "maxValue": 100.0      // 최대값
  },
  "description": "Tank 1 온도 센서"
}
```

**동작:**
- `value = baseTemp + amplitude * sin(2π * t / period) + noise`
- 예: 25°C 기준, ±10°C 변동, 30초 주기

#### 2. 압력 센서 (pressure)

램프 업/홀드/램프 다운 사이클

```json
{
  "name": "PressureSensor_Pump1",
  "type": "pressure",
  "enabled": true,
  "address": "%DF108",
  "updateIntervalMs": 100,
  "parameters": {
    "minPressure": 0.0,     // 최소 압력 (bar)
    "maxPressure": 10.0,    // 최대 압력 (bar)
    "rampUpTime": 20.0,     // 상승 시간 (초)
    "holdTime": 10.0,       // 유지 시간 (초)
    "rampDownTime": 15.0,   // 하강 시간 (초)
    "noiseStdDev": 0.1      // 노이즈
  },
  "description": "펌프 1 압력 센서"
}
```

**사이클:**
1. 0 → 10 bar (20초 동안 상승)
2. 10 bar 유지 (10초)
3. 10 → 0 bar (15초 동안 하강)
4. 반복

#### 3. 사인파 센서 (sine)

순수 사인파 신호 생성

```json
{
  "name": "LevelSensor_Tank1",
  "type": "sine",
  "enabled": true,
  "address": "%DF116",
  "updateIntervalMs": 100,
  "parameters": {
    "offset": 50.0,         // 오프셋 (%)
    "amplitude": 30.0,      // 진폭 (%)
    "frequency": 0.1,       // 주파수 (Hz)
    "phase": 0.0            // 위상 (라디안)
  },
  "description": "Tank 1 레벨 센서"
}
```

#### 4. 랜덤 센서 (random)

Smooth한 랜덤 값 생성

```json
{
  "name": "FlowMeter_Pipe1",
  "type": "random",
  "enabled": true,
  "address": "%DF120",
  "updateIntervalMs": 100,
  "parameters": {
    "minValue": 0.0,        // 최소값 (L/min)
    "maxValue": 10.0,       // 최대값 (L/min)
    "changeRate": 1.0       // 변화율 (0.0-1.0)
  },
  "description": "파이프 1 유량계"
}
```

#### 5. 디지털 센서 (digital)

ON/OFF 신호 생성 (도어, 모션, 리미트 스위치)

```json
{
  "name": "DoorSensor_MainEntrance",
  "type": "digital",
  "enabled": true,
  "address": "%MW0",
  "updateIntervalMs": 100,
  "parameters": {
    "pattern": "toggle",    // toggle, pulse, random
    "togglePeriod": 10.0,   // 토글 주기 (초)
    "pulseWidth": 1.0,      // 펄스 폭 (초, pulse 패턴용)
    "pulsePeriod": 5.0,     // 펄스 주기 (초, pulse 패턴용)
    "randomProb": 0.5       // 랜덤 확률 (random 패턴용)
  },
  "description": "정문 도어 센서"
}
```

**패턴 설명:**
- `toggle`: 일정 주기로 ON/OFF 반복
- `pulse`: 짧은 펄스 신호 생성
- `random`: 랜덤하게 ON/OFF

#### 6. 릴레이 액츄에이터 (relay)

제어 가능한 디지털 출력

```json
{
  "name": "RelayActuator_Pump1",
  "type": "relay",
  "enabled": true,
  "address": "%MW10",
  "updateIntervalMs": 100,
  "parameters": {
    "defaultState": false,  // 기본 상태
    "autoToggle": false,    // 자동 토글 여부
    "togglePeriod": 10.0    // 토글 주기 (초)
  },
  "description": "펌프 1 제어 릴레이"
}
```

#### 7. 정수 액츄에이터 (integer)

0-100% 범위의 제어 신호

```json
{
  "name": "MotorSpeed_Conveyor",
  "type": "integer",
  "enabled": true,
  "address": "%DW200",
  "updateIntervalMs": 100,
  "parameters": {
    "minValue": 0,          // 최소값
    "maxValue": 100,        // 최대값
    "defaultValue": 0,      // 기본값
    "autoMode": true,       // 자동 모드
    "autoPattern": "ramp",  // ramp, step, sine
    "rampRate": 1.0,        // 램프 속도 (units/sec)
    "stepPeriod": 10.0      // 스텝 주기 (초)
  },
  "description": "컨베이어 모터 속도"
}
```

#### 8. 진동 센서 (vibration)

다중 고조파 + 스파이크 노이즈

```json
{
  "name": "VibrationSensor_Motor1",
  "type": "vibration",
  "enabled": true,
  "address": "%DF124",
  "updateIntervalMs": 100,
  "parameters": {
    "baseLevel": 2.0,       // 기준 레벨 (mm/s RMS)
    "amplitude": 1.0,       // 진폭
    "frequency": 50.0,      // 기본 주파수 (Hz)
    "harmonics": 3,         // 고조파 개수
    "spikeProb": 0.01,      // 스파이크 확률
    "spikeAmp": 5.0,        // 스파이크 진폭
    "noiseStdDev": 0.2,     // 노이즈
    "minValue": 0.0,
    "maxValue": 50.0
  },
  "description": "모터 1 진동 센서"
}
```

#### 9. 소음 센서 (noise)

주변 소음 + 피크 노이즈 시뮬레이션

```json
{
  "name": "NoiseSensor_FactoryFloor",
  "type": "noise",
  "enabled": true,
  "address": "%DF132",
  "updateIntervalMs": 100,
  "parameters": {
    "ambientLevel": 55.0,   // 주변 소음 (dB)
    "peakLevel": 85.0,      // 피크 소음 (dB)
    "cyclePeriod": 30.0,    // 사이클 주기 (초)
    "dutyCycle": 0.6,       // 듀티 사이클 (0.0-1.0)
    "noiseStdDev": 2.0,     // 노이즈
    "spikeProb": 0.02,      // 스파이크 확률
    "spikeLevel": 95.0,     // 스파이크 레벨 (dB)
    "minValue": 30.0,
    "maxValue": 120.0
  },
  "description": "공장 바닥 소음 센서"
}
```

#### 10. 스텝 모터 (stepmotor)

위치 제어 모터

```json
{
  "name": "StepMotor_Conveyor",
  "type": "stepmotor",
  "enabled": true,
  "address": "%DF140",
  "updateIntervalMs": 100,
  "parameters": {
    "maxSpeed": 1000.0,     // 최대 속도 (steps/sec)
    "acceleration": 5000.0, // 가속도 (steps/sec²)
    "stepsPerRev": 200,     // 1회전당 스텝 수
    "autoMode": true,       // 자동 모드
    "autoPattern": "oscillate", // oscillate, rotate
    "autoPeriod": 20.0      // 자동 주기 (초)
  },
  "description": "컨베이어 스텝 모터"
}
```

#### 11. 서보 모터 (servomotor)

속도/위치 제어 모터

```json
{
  "name": "ServoMotor_Spindle",
  "type": "servomotor",
  "enabled": true,
  "address": "%DF148",
  "updateIntervalMs": 100,
  "parameters": {
    "maxVelocity": 3000.0,  // 최대 속도 (RPM)
    "maxTorque": 10.0,      // 최대 토크 (Nm)
    "inertia": 0.001,       // 관성 (kg·m²)
    "damping": 0.01,        // 댐핑
    "outputMode": "velocity", // velocity, position, torque
    "autoMode": true,       // 자동 모드
    "autoPattern": "sine",  // sine, step, ramp
    "autoPeriod": 30.0,     // 자동 주기 (초)
    "kp": 0.5,              // PID P 게인
    "ki": 0.1,              // PID I 게인
    "kd": 0.01              // PID D 게인
  },
  "description": "스핀들 서보 모터"
}
```

### PLC 주소 규칙

| 주소 형식 | 데이터 타입 | 용도 | 예시 |
|----------|-----------|------|------|
| %DF### | Float64 (64-bit) | 아날로그 센서/액츄에이터 | %DF100, %DF104 |
| %DW### | Int32 (32-bit) | 정수 액츄에이터 | %DW200, %DW201 |
| %MW### | Bool | 디지털 센서/릴레이 | %MW0, %MW10 |

**주의사항:**
- 주소는 중복되면 안 됩니다
- Float64는 4워드 간격 권장 (%DF100, %DF104, %DF108...)

---

## PLC 로직 작성 가이드

### plc_logic.lua 파일 구조

```lua
-- ===================================================================
-- 전역 변수 (스캔 간 유지)
-- ===================================================================
scan_count = scan_count or 0
last_temp = last_temp or 0

-- ===================================================================
-- 초기화 함수 (서버 시작 시 1회 실행)
-- ===================================================================
function init()
    plc_log("PLC 로직 초기화 완료")
end

-- ===================================================================
-- 메인 로직 함수 (매 스캔 주기마다 실행)
-- ===================================================================
function run_logic()
    scan_count = scan_count + 1

    -- 여기에 제어 로직 작성
end
```

### 사용 가능한 Lua 함수

#### 1. 태그 접근

```lua
-- 방법 1: Data 테이블 사용 (권장)
local temp = Data.TemperatureSensor_Tank1
Data.RelayActuator_Pump1 = true

-- 방법 2: 함수 사용
local temp, err = get_tag("TemperatureSensor_Tank1")
local ok, err = set_tag("RelayActuator_Pump1", true)
```

#### 2. 로깅

```lua
plc_log("메시지")
plc_log(string.format("온도: %.2f", temp))
```

#### 3. 시간 함수

```lua
local timestamp = get_time()  -- Unix timestamp
sleep(100)  -- 100ms 대기 (비권장)
```

### 제어 로직 예제

#### 예제 1: 온도 기반 히터 제어

```lua
function run_logic()
    -- 온도 센서 읽기
    local temp = Data.TemperatureSensor_Tank1

    -- 임계값 기반 제어
    if temp < 20 then
        Data.RelayActuator_Heater = true   -- 히터 ON
        Data.HeaterPower_Tank1 = 100       -- 최대 파워
        plc_log("온도 낮음 - 히터 ON")
    elseif temp > 30 then
        Data.RelayActuator_Heater = false  -- 히터 OFF
        Data.HeaterPower_Tank1 = 0         -- 파워 0
        plc_log("온도 높음 - 히터 OFF")
    else
        -- 비례 제어
        local power = (30 - temp) / 10 * 100
        Data.HeaterPower_Tank1 = math.floor(power)
    end
end
```

#### 예제 2: 압력 안전 로직

```lua
-- 전역 변수
alarm_active = alarm_active or false
alarm_count = alarm_count or 0

function run_logic()
    local pressure = Data.PressureSensor_Pump1

    -- 과압 감지
    if pressure > 9.0 then
        alarm_count = alarm_count + 1

        -- 3번 연속 과압 시 알람
        if alarm_count >= 3 then
            Data.AlarmIndicator_System = true
            Data.RelayActuator_Pump1 = false  -- 펌프 정지
            alarm_active = true
            plc_log("알람: 과압 감지! 펌프 정지")
        end
    else
        alarm_count = 0

        -- 압력 정상 복귀
        if alarm_active and pressure < 5.0 then
            Data.AlarmIndicator_System = false
            alarm_active = false
            plc_log("알람 해제: 압력 정상")
        end
    end
end
```

#### 예제 3: 시퀀스 제어

```lua
-- 전역 변수
sequence_step = sequence_step or 0
step_timer = step_timer or 0

function run_logic()
    step_timer = step_timer + 1

    if sequence_step == 0 then
        -- Step 0: 대기
        Data.RelayActuator_Pump1 = false
        Data.ValveActuator_Tank1 = false

        if Data.DoorSensor_MainEntrance == 1 then
            sequence_step = 1
            step_timer = 0
            plc_log("시퀀스 시작")
        end

    elseif sequence_step == 1 then
        -- Step 1: 밸브 열기
        Data.ValveActuator_Tank1 = true

        if step_timer > 50 then  -- 5초 후
            sequence_step = 2
            step_timer = 0
            plc_log("밸브 열림")
        end

    elseif sequence_step == 2 then
        -- Step 2: 펌프 시작
        Data.RelayActuator_Pump1 = true

        if step_timer > 100 then  -- 10초 후
            sequence_step = 3
            step_timer = 0
            plc_log("펌프 시작")
        end

    elseif sequence_step == 3 then
        -- Step 3: 완료 대기
        if Data.LevelSensor_Tank1 > 80 then
            sequence_step = 0
            step_timer = 0
            plc_log("시퀀스 완료")
        end
    end
end
```

#### 예제 4: 모터 속도 램핑

```lua
-- 전역 변수
target_speed = target_speed or 0
current_speed = current_speed or 0

function run_logic()
    -- 도어 열리면 속도 증가
    if Data.DoorSensor_MainEntrance == 1 then
        target_speed = 100
    else
        target_speed = 0
    end

    -- 부드러운 가감속 (램핑)
    local ramp_rate = 2  -- 스캔당 2% 변화

    if current_speed < target_speed then
        current_speed = current_speed + ramp_rate
        if current_speed > target_speed then
            current_speed = target_speed
        end
    elseif current_speed > target_speed then
        current_speed = current_speed - ramp_rate
        if current_speed < target_speed then
            current_speed = target_speed
        end
    end

    Data.MotorSpeed_Conveyor = math.floor(current_speed)
end
```

### 헬퍼 함수

```lua
-- 허용 오차 내 비교
function fuzzy_equal(a, b, tolerance)
    return math.abs(a - b) < tolerance
end

-- 값 제한
function clamp(value, min_val, max_val)
    if value < min_val then return min_val end
    if value > max_val then return max_val end
    return value
end

-- 범위 매핑
function map_value(value, in_min, in_max, out_min, out_max)
    return (value - in_min) * (out_max - out_min) / (in_max - in_min) + out_min
end

-- 사용 예제
function run_logic()
    local temp = Data.TemperatureSensor_Tank1

    -- 온도를 팬 속도로 매핑 (20-40°C → 0-100%)
    local fan_speed = map_value(temp, 20, 40, 0, 100)
    fan_speed = clamp(fan_speed, 0, 100)

    Data.FanSpeed_Cooling = math.floor(fan_speed)
end
```

---

## 고급 사용법

### 1. 커스텀 센서 추가

새로운 센서 타입을 추가하려면:

1. `internal/sim/sensors/` 디렉토리에 새 파일 생성
2. `Sensor` 인터페이스 구현
3. `internal/sim/registry.go`에 팩토리 등록

예제: 간단한 카운터 센서

```go
// internal/sim/sensors/counter.go
package sensors

import "time"

type CounterSensor struct {
    BaseSensor
    CountPerSec float64
    CurrentCount float64
}

func NewCounterSensor(name, address string, enabled bool,
    updateIntervalMs int, countPerSec float64, description string) *CounterSensor {
    return &CounterSensor{
        BaseSensor: BaseSensor{
            Name: name,
            Address: address,
            Enabled: enabled,
            UpdateIntervalMs: updateIntervalMs,
            Description: description,
        },
        CountPerSec: countPerSec,
        CurrentCount: 0,
    }
}

func (c *CounterSensor) Update(deltaTime time.Duration) float64 {
    if !c.IsEnabled() {
        return c.CurrentCount
    }

    c.AddElapsedTime(deltaTime)
    c.CurrentCount += c.CountPerSec * deltaTime.Seconds()

    return c.CurrentCount
}

func (c *CounterSensor) Reset() {
    c.BaseSensor.Reset()
    c.CurrentCount = 0
}
```

registry.go에 등록:

```go
RegisterSensor("counter", func(def config.SensorDefinition) (sensors.Sensor, error) {
    countPerSec := config.GetFloat64Param(def.Parameters, "countPerSec", 1.0)
    return sensors.NewCounterSensor(
        def.Name, def.Address, def.Enabled, def.UpdateIntervalMs,
        countPerSec, def.Description,
    ), nil
})
```

### 2. 멀티 인스턴스 실행

여러 시뮬레이터를 동시 실행:

```bash
# 터미널 1: 공장 A
./bin/server -config factory_a.json -endpoint "opc.tcp://0.0.0.0:4840"

# 터미널 2: 공장 B
./bin/server -config factory_b.json -endpoint "opc.tcp://0.0.0.0:4841"

# 터미널 3: 테스트 환경
./bin/server -config test.json -endpoint "opc.tcp://0.0.0.0:4842"
```

### 3. 로그 출력 관리

```bash
# 로그 파일로 저장
./bin/server 2>&1 | tee server.log

# 특정 센서만 모니터링
./bin/server 2>&1 | grep "TemperatureSensor"

# 에러만 확인
./bin/server 2>&1 | grep -i "error\|fail"

# PLC 로직 로그만 확인
./bin/server 2>&1 | grep "PLC-LOGIC"
```

---

## 트러블슈팅

### 문제 1: 서버가 시작되지 않음

**증상:**
```
Failed to load sensor config: open sensors.json: no such file or directory
```

**해결:**
```bash
# 현재 디렉토리 확인
pwd

# sensors.json이 있는지 확인
ls -la sensors.json

# 없다면 정확한 경로 지정
./bin/server -config /full/path/to/sensors.json
```

### 문제 2: Lua 스크립트 에러

**증상:**
```
Failed to initialize Lua engine: Lua script file not found: plc_logic.lua
```

**해결:**
```bash
# plc_logic.lua 확인
ls -la plc_logic.lua

# 또는 다른 스크립트 사용
./bin/server -script /path/to/my_logic.lua
```

### 문제 3: 센서 값이 변하지 않음

**원인:**
- `enabled: false`로 설정
- PLC 로직에서 값 덮어쓰기

**해결:**
```json
// sensors.json에서 확인
{
  "enabled": true,  // false면 true로 변경
  ...
}
```

### 문제 4: Unknown sensor type

**증상:**
```
Failed to create sensor manager: unknown sensor type: relay
```

**원인:**
- registry.go에 센서 타입이 등록되지 않음

**해결:**
```bash
# 빌드 다시 실행
make clean
make build
```

### 문제 5: 메모리 사용량 증가

**원인:**
- Lua 스크립트에서 무한히 증가하는 전역 변수 사용

**해결:**
```lua
-- 나쁜 예
function run_logic()
    table.insert(history, value)  -- 계속 증가!
end

-- 좋은 예
function run_logic()
    if #history > 100 then
        table.remove(history, 1)  -- 100개로 제한
    end
    table.insert(history, value)
end
```

### 문제 6: PLC 스캔이 느림

**증상:**
```
[LUA] Scan #100 error: execution timeout
```

**원인:**
- Lua 스크립트에 무거운 연산이나 sleep() 사용

**해결:**
```lua
-- 나쁜 예
function run_logic()
    sleep(1000)  -- 1초 대기 - 절대 사용 금지!
end

-- 좋은 예
scan_count = scan_count or 0
function run_logic()
    scan_count = scan_count + 1
    if scan_count % 10 == 0 then  -- 10번에 1번만 실행
        -- 무거운 작업
    end
end
```

---

## 실전 예제

### 예제 1: 간단한 온도 제어 시스템

**목표:** 탱크 온도를 25°C로 유지

**sensors.json:**
```json
{
  "sensors": [
    {
      "name": "TankTemp",
      "type": "temperature",
      "enabled": true,
      "address": "%DF100",
      "updateIntervalMs": 100,
      "parameters": {
        "baseTemp": 25.0,
        "amplitude": 10.0,
        "period": 30.0,
        "noiseStdDev": 0.5,
        "minValue": 0.0,
        "maxValue": 100.0
      },
      "description": "탱크 온도"
    },
    {
      "name": "Heater",
      "type": "relay",
      "enabled": true,
      "address": "%MW10",
      "updateIntervalMs": 100,
      "parameters": {
        "defaultState": false
      },
      "description": "히터"
    }
  ]
}
```

**plc_logic.lua:**
```lua
-- 히스테리시스 제어
function run_logic()
    local temp = Data.TankTemp

    -- 23°C 이하면 히터 ON
    if temp < 23 then
        Data.Heater = true
    -- 27°C 이상이면 히터 OFF
    elseif temp > 27 then
        Data.Heater = false
    end
    -- 23-27°C 범위는 현재 상태 유지
end
```

**테스트:**
```bash
./bin/server -config sensors.json -script plc_logic.lua
```

### 예제 2: 컨베이어 벨트 시퀀스 제어

**목표:** 도어 센서 감지 → 컨베이어 시작 → 물체 통과 → 정지

**sensors.json:**
```json
{
  "sensors": [
    {
      "name": "DoorSensor",
      "type": "digital",
      "enabled": true,
      "address": "%MW0",
      "updateIntervalMs": 100,
      "parameters": {
        "pattern": "toggle",
        "togglePeriod": 20.0
      },
      "description": "입구 도어"
    },
    {
      "name": "LimitSwitch",
      "type": "digital",
      "enabled": true,
      "address": "%MW1",
      "updateIntervalMs": 100,
      "parameters": {
        "pattern": "pulse",
        "pulseWidth": 2.0,
        "pulsePeriod": 25.0
      },
      "description": "끝단 스위치"
    },
    {
      "name": "ConveyorSpeed",
      "type": "integer",
      "enabled": true,
      "address": "%DW200",
      "updateIntervalMs": 100,
      "parameters": {
        "minValue": 0,
        "maxValue": 100,
        "defaultValue": 0
      },
      "description": "컨베이어 속도"
    }
  ]
}
```

**plc_logic.lua:**
```lua
-- 상태 변수
state = state or "IDLE"
timer = timer or 0

function run_logic()
    if state == "IDLE" then
        Data.ConveyorSpeed = 0
        if Data.DoorSensor == 1 then
            state = "RUNNING"
            timer = 0
            plc_log("컨베이어 시작")
        end

    elseif state == "RUNNING" then
        Data.ConveyorSpeed = 50

        -- 리미트 스위치 감지 또는 타임아웃
        if Data.LimitSwitch == 1 then
            state = "STOPPING"
            timer = 0
            plc_log("물체 감지 - 정지 시작")
        elseif timer > 200 then  -- 20초 타임아웃
            state = "STOPPING"
            plc_log("타임아웃 - 정지")
        end
        timer = timer + 1

    elseif state == "STOPPING" then
        Data.ConveyorSpeed = 0

        if timer > 30 then  -- 3초 대기
            state = "IDLE"
            plc_log("준비 완료")
        end
        timer = timer + 1
    end
end
```

### 예제 3: 다중 탱크 레벨 제어

**목표:** 3개 탱크의 레벨을 독립적으로 제어

**plc_logic.lua:**
```lua
function control_tank(level_tag, pump_tag, valve_tag, setpoint)
    local level = Data[level_tag]

    if level < setpoint - 5 then
        Data[pump_tag] = true    -- 펌프 ON
        Data[valve_tag] = false  -- 밸브 닫기
    elseif level > setpoint + 5 then
        Data[pump_tag] = false   -- 펌프 OFF
        Data[valve_tag] = true   -- 밸브 열기 (배출)
    end
end

function run_logic()
    control_tank("LevelSensor_Tank1", "Pump1", "Valve1", 50)
    control_tank("LevelSensor_Tank2", "Pump2", "Valve2", 60)
    control_tank("LevelSensor_Tank3", "Pump3", "Valve3", 70)
end
```

---

## 성능 최적화

### 1. 스캔 주기 조정

```bash
# 느린 프로세스 (500ms)
./bin/server -scantime 500

# 일반적인 설정 (100ms, 기본값)
./bin/server -scantime 100

# 빠른 제어 (20ms)
./bin/server -scantime 20
```

### 2. 센서 업데이트 간격

불필요하게 빠른 업데이트 방지:

```json
{
  "name": "SlowSensor",
  "updateIntervalMs": 1000,  // 1초마다 업데이트
  ...
}
```

### 3. Lua 로직 최적화

```lua
-- 나쁜 예: 매 스캔마다 문자열 생성
function run_logic()
    plc_log(string.format("Temp: %.2f", Data.Temp))  -- 부하!
end

-- 좋은 예: 주기적으로만 로깅
scan_count = scan_count or 0
function run_logic()
    scan_count = scan_count + 1
    if scan_count % 100 == 0 then
        plc_log(string.format("Temp: %.2f", Data.Temp))
    end
end
```

---

## 부록: 빠른 참조

### Makefile 명령어

```bash
make build              # 전체 빌드
make server             # 서버만 빌드
make client             # 클라이언트만 빌드
make clean              # 빌드 정리
make run-server         # 서버 실행
make run-server-no-plc  # PLC 로직 없이 실행
make deps               # 의존성 다운로드
make fmt                # 코드 포맷팅
```

### 주요 파일 위치

```
go-opcua-sim/
├── sensors.json         # 센서 설정
├── plc_logic.lua        # PLC 로직
├── bin/server           # 서버 실행 파일
├── bin/client           # 클라이언트 실행 파일
└── internal/
    ├── plc/             # PLC 태그 및 Lua 엔진
    ├── sim/             # 센서 시뮬레이션
    ├── opcuaserver/     # OPC UA 서버
    └── config/          # 설정 로더
```

### 지원 문의

이슈 리포트: go-lsplc-sim 프로젝트 기반
문서 버전: 1.0
최종 업데이트: 2026-01-07
