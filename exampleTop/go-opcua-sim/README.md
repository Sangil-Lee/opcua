# Go OPC UA PLC Simulator

Go로 구현된 OPC UA PLC Simulator로, go-lsplc-sim의 구조를 OPC UA 프로토콜로 포팅한 프로젝트입니다.

## 주요 기능

- **OPC UA 서버**: 표준 OPC UA 프로토콜을 지원하는 시뮬레이션 서버
- **PLC 태그 시스템**: 센서/액츄에이터를 위한 태그 기반 데이터 관리
- **Lua 기반 PLC 로직**: Lua 스크립트를 통한 유연한 PLC 로직 구현
- **다양한 센서 시뮬레이션**: 온도, 압력, 진동, 노이즈 등 다양한 센서 타입 지원
- **JSON 기반 설정**: 센서 및 액츄에이터 설정을 JSON으로 관리

## 프로젝트 구조

```
go-opcua-sim/
├── cmd/
│   ├── server/          # OPC UA 서버 메인
│   └── client/          # OPC UA 클라이언트 (검증용)
├── internal/
│   ├── config/          # 설정 파일 로더
│   ├── plc/             # PLC 태그 시스템 및 Lua 엔진
│   │   ├── tag.go              # 태그 정의 및 관리
│   │   ├── tag_generator.go   # 센서로부터 태그 자동 생성
│   │   └── lua_engine.go       # Lua 스크립트 실행 엔진
│   ├── sim/             # 센서 시뮬레이션
│   │   ├── sensors/            # 다양한 센서 타입 구현
│   │   ├── manager.go          # 센서 매니저
│   │   └── registry.go         # 센서 팩토리 레지스트리
│   └── opcuaserver/     # OPC UA 서버 구현
├── sensors.json         # 센서 설정 파일
├── plc_logic.lua        # PLC 로직 스크립트
└── Makefile             # 빌드 스크립트

```

## 설치

### 필요 요구사항

- Go 1.23.8 이상
- Linux/macOS/Windows

### 빌드

```bash
# 의존성 다운로드
make deps

# 모든 바이너리 빌드
make build

# 또는 개별 빌드
make server
make client
```

## 사용 방법

### 서버 실행

기본 설정으로 서버 실행:
```bash
make run-server
```

또는 직접 실행:
```bash
./bin/server -config sensors.json -script plc_logic.lua -scantime 100 -endpoint "opc.tcp://0.0.0.0:4840"
```

서버 옵션:
- `-config`: 센서 설정 파일 경로 (기본: sensors.json)
- `-script`: PLC Lua 스크립트 파일 경로 (기본: plc_logic.lua)
- `-scantime`: PLC 스캔 주기 (밀리초, 기본: 100)
- `-plc`: PLC 로직 활성화 여부 (기본: true)
- `-endpoint`: OPC UA 서버 엔드포인트 (기본: opc.tcp://0.0.0.0:4840)

### 클라이언트 실행

단일 읽기:
```bash
make run-client
```

연속 읽기:
```bash
make run-client-continuous
```

또는 직접 실행:
```bash
# 단일 읽기 (String Identifier 사용)
./bin/client -endpoint "opc.tcp://localhost:4840" -node "ns=2;s=TemperatureSensor_Tank1"

# 연속 읽기 (1초 간격)
./bin/client -endpoint "opc.tcp://localhost:4840" -node "ns=2;s=TemperatureSensor_Tank1" -continuous -interval 1000

# Boolean 노드 읽기
./bin/client -node "ns=2;s=DoorSensor_MainEntrance"

# Integer 노드 읽기
./bin/client -node "ns=2;s=MotorSpeed_Conveyor"
```

클라이언트 옵션:
- `-endpoint`: OPC UA 서버 엔드포인트 (기본: opc.tcp://localhost:4840)
- `-node`: 읽을 노드 ID (예: ns=2;s=TemperatureSensor_Tank1)
- `-continuous`: 연속 읽기 모드 활성화
- `-interval`: 읽기 간격 (밀리초, 기본: 1000)

## 센서 설정

`sensors.json` 파일에서 센서를 정의할 수 있습니다:

```json
{
  "sensors": [
    {
      "name": "TemperatureSensor_Tank1",
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
      "description": "Tank 1 temperature sensor"
    }
  ]
}
```

### 지원하는 센서 타입

- `temperature`: 온도 센서 (사인파 + 노이즈)
- `pressure`: 압력 센서 (램프 업/홀드/램프 다운 사이클)
- `sine`: 사인파 센서
- `random`: 랜덤 값 센서
- `digital`: 디지털 센서 (ON/OFF)
- `vibration`: 진동 센서
- `noise`: 소음 센서

## PLC 로직

`plc_logic.lua` 파일에서 PLC 로직을 정의할 수 있습니다:

```lua
-- 초기화 함수 (한 번만 실행)
function init()
    plc_log("PLC logic initialized")
end

-- 메인 로직 함수 (주기적으로 실행)
function run_logic()
    -- 센서 값 읽기
    local temp = Data.TemperatureSensor_Tank1

    -- 제어 로직
    if temp > 50 then
        plc_log("Temperature too high: " .. temp)
    end
end
```

### 사용 가능한 Lua 함수

- `Data.<TagName>`: 태그 값에 접근
- `plc_log(message)`: 로그 출력
- `get_tag(name)`: 태그 값 읽기
- `set_tag(name, value)`: 태그 값 쓰기
- `get_time()`: 현재 시간 (Unix timestamp)
- `sleep(ms)`: 대기 (밀리초)

## OPC UA 노드 매핑

각 센서/태그는 OPC UA 노드로 매핑됩니다:

- 노드 ID 형식: `ns=2;s=<TagName>` (String Identifier)
- 예시:
  - TemperatureSensor_Tank1 → ns=2;s=TemperatureSensor_Tank1
  - TemperatureSensor_Tank2 → ns=2;s=TemperatureSensor_Tank2
  - PressureSensor_Pump1 → ns=2;s=PressureSensor_Pump1

String identifier를 사용하므로 태그 이름을 그대로 노드 ID로 사용할 수 있어 EPICS DB 생성 시 직관적입니다.

서버 시작 시 전체 노드 매핑이 출력됩니다.

## go-lsplc-sim과의 차이점

- **프로토콜**: LS XGT FEnet → OPC UA
- **서버 구현**: LS 프로토콜 서버 → OPC UA 서버
- **클라이언트**: LS 클라이언트 → OPC UA 클라이언트
- **나머지는 동일**: PLC 태그 시스템, Lua 엔진, 센서 시뮬레이션 등 go-lsplc-sim과 동일한 구조 사용

## 빠른 시작 및 테스트

### 자동 테스트 실행 (권장)

```bash
# 빌드
make build

# 자동 테스트 실행
./test_server.sh
```

자동 테스트는 다음을 검증합니다:
- ✅ 설정 파일 유효성
- ✅ 서버 정상 시작
- ✅ 태그 생성 및 센서 시뮬레이션
- ✅ PLC 로직 실행
- ✅ 실시간 데이터 업데이트

**상세한 테스트 방법은 [TESTING.md](TESTING.md)를 참고하세요.**

## 예제

### 1. 기본 실행 (서버 + PLC 로직)

```bash
# 빌드
make build

# 서버 실행
make run-server

# 다른 터미널에서 클라이언트 실행
make run-client-continuous
```

### 2. PLC 로직 없이 센서만 실행

```bash
./bin/server -plc=false
```

### 3. 특정 노드 모니터링

```bash
# TemperatureSensor_Tank2 읽기 (String Identifier 사용)
./bin/client -node "ns=2;s=TemperatureSensor_Tank2" -continuous -interval 500

# 여러 센서 모니터링 예시
./bin/client -node "ns=2;s=PressureSensor_Pump1" -continuous -interval 500
./bin/client -node "ns=2;s=VibrationSensor_Motor1" -continuous -interval 500
```

## EPICS 통합

String Identifier를 사용하므로 EPICS DB 파일 생성 시 직관적으로 매핑할 수 있습니다:

### 예시 EPICS DB 파일

```
# 온도 센서 (Float64)
record(ai, "TANK1:TEMP") {
    field(DTYP, "opcua")
    field(INP,  "@opc.tcp://localhost:4840 ns=2;s=TemperatureSensor_Tank1")
    field(SCAN, "1 second")
    field(EGU,  "degC")
    field(PREC, "2")
}

# 도어 센서 (Boolean)
record(bi, "DOOR:STATUS") {
    field(DTYP, "opcua")
    field(INP,  "@opc.tcp://localhost:4840 ns=2;s=DoorSensor_MainEntrance")
    field(SCAN, "1 second")
    field(ZNAM, "Closed")
    field(ONAM, "Open")
}

# 모터 속도 (Int32)
record(longin, "MOTOR:SPEED") {
    field(DTYP, "opcua")
    field(INP,  "@opc.tcp://localhost:4840 ns=2;s=MotorSpeed_Conveyor")
    field(SCAN, "1 second")
    field(EGU,  "RPM")
}

# 압력 센서 (Float64)
record(ai, "PUMP1:PRESSURE") {
    field(DTYP, "opcua")
    field(INP,  "@opc.tcp://localhost:4840 ns=2;s=PressureSensor_Pump1")
    field(SCAN, "1 second")
    field(EGU,  "bar")
    field(PREC, "3")
}
```

### 주요 장점

- **직관적인 매핑**: 태그 이름이 그대로 노드 ID에 사용됨
- **자동 생성 가능**: 센서 정의에서 EPICS DB 자동 생성 가능
- **유지보수 용이**: 센서 이름 변경 시 일관성 유지 쉬움

### 클라이언트 검증

상세한 클라이언트 테스트 방법은 [CLIENT_TESTING.md](CLIENT_TESTING.md)를 참고하세요.

## 문서

- **[README.md](README.md)**: 프로젝트 개요 및 빠른 시작 (현재 문서)
- **[QUICKSTART.md](QUICKSTART.md)**: 5분 빠른 시작 가이드
- **[MANUAL.md](MANUAL.md)**: 상세 사용 매뉴얼
  - 센서 설정 가이드 (11가지 센서 타입)
  - PLC 로직 작성 가이드
  - 고급 사용법
  - 트러블슈팅
  - 실전 예제
- **[TESTING.md](TESTING.md)**: 테스트 가이드
  - 자동 테스트 실행 방법
  - 수동 테스트 시나리오
  - 성능 테스트
  - 에러 시나리오
  - 통합 테스트
- **[CLIENT_TESTING.md](CLIENT_TESTING.md)**: 클라이언트 검증 가이드
  - Go 클라이언트 사용법
  - Python 클라이언트 예제
  - UaExpert 연결 방법
  - EPICS 통합 예시

## 참고

- go-lsplc-sim의 모든 센서 시뮬레이션 로직을 그대로 사용합니다
- OPC UA 표준 프로토콜을 지원하므로 다양한 OPC UA 클라이언트와 호환됩니다
- Lua 스크립트를 통해 복잡한 PLC 로직을 유연하게 구현할 수 있습니다

## 라이센스

이 프로젝트는 go-lsplc-sim의 구조를 기반으로 하며, OPC UA 프로토콜로 포팅되었습니다.
