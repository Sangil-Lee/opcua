# Go OPC UA Simulator - Quick Start

## 5분 안에 시작하기

### 1단계: 빌드
```bash
cd go-opcua-sim
make build
```

### 2단계: PKI 설정 (최초 1회만)
```bash
./setup_pki.sh
```

이 스크립트는 OPC UA 서버에 필요한 인증서를 생성합니다.

### 3단계: 테스트 실행
```bash
./test_server.sh
```

자동 테스트가 완료되면 모든 기능이 정상 동작하는 것입니다! ✅

### 4단계: 서버 직접 실행
```bash
# 새 터미널에서
./bin/server
```

다음과 같은 출력을 볼 수 있습니다:
```
=== Go OPC UA PLC Simulation Server ===
[CONFIG] Loaded 26 sensor definitions
...
=== Sensor Update #20 ===
  TemperatureSensor_Tank1 (%DF100): 29.574
  PressureSensor_Pump1 (%DF108): 5.922
  ...
```

축하합니다! 서버가 실행 중입니다. 🎉

---

## 다음 단계

### 센서 커스터마이징

`sensors.json` 파일을 편집하여 센서를 추가/수정할 수 있습니다:

```json
{
  "sensors": [
    {
      "name": "MySensor",
      "type": "temperature",
      "enabled": true,
      "address": "%DF200",
      "updateIntervalMs": 100,
      "parameters": {
        "baseTemp": 30.0,
        "amplitude": 5.0,
        "period": 20.0,
        "noiseStdDev": 0.3,
        "minValue": 0.0,
        "maxValue": 100.0
      },
      "description": "My custom temperature sensor"
    }
  ]
}
```

### PLC 로직 작성

`plc_logic.lua` 파일을 편집하여 제어 로직을 추가할 수 있습니다:

```lua
function run_logic()
    -- 온도 읽기
    local temp = Data.MySensor

    -- 간단한 제어
    if temp > 35 then
        plc_log("Temperature is too high!")
    end
end
```

---

## 도움말

- **상세 매뉴얼**: [MANUAL.md](MANUAL.md) 참고
- **테스트 가이드**: [TESTING.md](TESTING.md) 참고
- **문제 해결**: [MANUAL.md의 트러블슈팅](MANUAL.md#트러블슈팅) 섹션 참고

---

## 자주 사용하는 명령어

```bash
# 서버 실행 (기본)
./bin/server

# 서버 실행 (PLC 로직 없이)
./bin/server -plc=false

# 서버 실행 (빠른 스캔)
./bin/server -scantime 50

# 서버 실행 (커스텀 설정)
./bin/server -config my_config.json -script my_logic.lua

# 자동 테스트
./test_server.sh

# 빌드
make build

# 클린 빌드
make clean && make build
```

---

## 예제 실행

간단한 탱크 제어 예제:

```bash
./bin/server -config examples/simple_tank.json -script examples/simple_tank.lua
```

이 예제는 다음을 시뮬레이션합니다:
- 탱크 온도 센서
- 탱크 레벨 센서
- 히터 제어 (온도 기반)
- 펌프 제어 (레벨 기반)

---

이제 시작할 준비가 되었습니다! 🚀

더 자세한 내용은 [MANUAL.md](MANUAL.md)를 참고하세요.
