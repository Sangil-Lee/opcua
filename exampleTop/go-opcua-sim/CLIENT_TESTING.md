# OPC UA Client 검증 가이드

## 📌 현재 구현 상태

go-opcua-sim의 OPC UA 서버는 **완전한 OPC UA 프로토콜 서버**입니다:
- ✅ 태그 시스템 완전 구현
- ✅ 센서 시뮬레이션 동작
- ✅ PLC 로직 실행
- ✅ OPC UA 노드 매핑 (String Identifier 사용)
- ✅ 실제 OPC UA 네트워크 서버 구현
- ✅ 클라이언트 연결 및 데이터 읽기 지원

**String Identifier 사용**으로 태그 이름을 그대로 노드 ID로 사용할 수 있어 EPICS DB 생성 시 직관적입니다.

---

## 🔍 방법 1: 내장 Go 클라이언트 (권장)

가장 간단하고 확실한 검증 방법입니다.

### 1단계: 서버 실행

```bash
cd /home/ctrluser/GoProject/go-opcua-sim
./bin/server
```

서버가 시작되면 다음과 같은 노드 매핑이 출력됩니다:

```
=== OPC UA Server Tag Mapping ===
Tag Name                                 NodeID                                             Data Type
---------------------------------------------------------------------------------------------------
TemperatureSensor_Tank1                  ns=2;s=TemperatureSensor_Tank1                     Double
TemperatureSensor_Tank2                  ns=2;s=TemperatureSensor_Tank2                     Double
PressureSensor_Pump1                     ns=2;s=PressureSensor_Pump1                        Double
DoorSensor_MainEntrance                  ns=2;s=DoorSensor_MainEntrance                     Boolean
MotorSpeed_Conveyor                      ns=2;s=MotorSpeed_Conveyor                         Int32
...
```

### 2단계: 새 터미널에서 클라이언트 실행

#### 단일 읽기

```bash
# 온도 센서 읽기
./bin/client -node "ns=2;s=TemperatureSensor_Tank1"

# 출력 예시:
# === OPC UA Client Test ===
# Connecting to: opc.tcp://localhost:4840
# Connected successfully!
# Reading node: ns=2;s=TemperatureSensor_Tank1
# [2026-01-07 13:15:02.900] Value: 28.687373421007536 (Type: float64)
```

#### Boolean 노드 읽기

```bash
./bin/client -node "ns=2;s=DoorSensor_MainEntrance"

# 출력 예시:
# [2026-01-07 13:15:21.499] Value: true (Type: bool)
```

#### Integer 노드 읽기

```bash
./bin/client -node "ns=2;s=MotorSpeed_Conveyor"

# 출력 예시:
# [2026-01-07 13:15:21.499] Value: 42 (Type: int32)
```

#### 연속 읽기 (실시간 모니터링)

```bash
# 1초 간격으로 연속 읽기
./bin/client -node "ns=2;s=TemperatureSensor_Tank1" -continuous -interval 1000

# 출력 예시:
# === OPC UA Client Test ===
# Connecting to: opc.tcp://localhost:4840
# Connected successfully!
# Reading node ns=2;s=TemperatureSensor_Tank1 every 1000 ms (Press Ctrl+C to stop)
# [2026-01-07 13:15:22.499] Value: 30.294411188779772 (Type: float64)
# [2026-01-07 13:15:23.499] Value: 32.748103143005515 (Type: float64)
# [2026-01-07 13:15:24.499] Value: 33.4933029037037 (Type: float64)
# ...
```

### 3단계: 다양한 센서 테스트

```bash
# 압력 센서
./bin/client -node "ns=2;s=PressureSensor_Pump1"

# 진동 센서
./bin/client -node "ns=2;s=VibrationSensor_Motor1"

# 레벨 센서
./bin/client -node "ns=2;s=LevelSensor_Tank1"

# 릴레이 액츄에이터 (Boolean)
./bin/client -node "ns=2;s=RelayActuator_Pump1"

# 밸브 위치 (Int32)
./bin/client -node "ns=2;s=ValvePosition_MainFlow"
```

---

## 🔧 방법 2: Python OPC UA 클라이언트

Python opcua 라이브러리를 사용한 검증도 가능합니다.

### 설치

```bash
pip install opcua
```

### Python 스크립트 실행

```bash
python3 test_opcua_client.py
```

또는 빠른 테스트:

```bash
python3 test_opcua_client.py --quick
```

### 커스텀 Python 스크립트 예시

```python
#!/usr/bin/env python3
from opcua import Client

# 서버 연결
client = Client("opc.tcp://localhost:4840")
client.connect()

# String Identifier를 사용한 노드 읽기
temp_node = client.get_node("ns=2;s=TemperatureSensor_Tank1")
value = temp_node.get_value()
print(f"Temperature: {value}°C")

# Boolean 노드 읽기
door_node = client.get_node("ns=2;s=DoorSensor_MainEntrance")
value = door_node.get_value()
print(f"Door: {'Open' if value else 'Closed'}")

# Integer 노드 읽기
motor_node = client.get_node("ns=2;s=MotorSpeed_Conveyor")
value = motor_node.get_value()
print(f"Motor Speed: {value} RPM")

client.disconnect()
```

---

## 🎯 방법 3: UaExpert (GUI 클라이언트)

UaExpert는 Unified Automation에서 제공하는 무료 OPC UA 클라이언트입니다.

### 설치

1. [Unified Automation 웹사이트](https://www.unified-automation.com/downloads/opc-ua-clients.html)에서 다운로드
2. 설치 및 실행

### 연결 설정

1. **Add Server** 클릭
2. Endpoint URL 입력: `opc.tcp://localhost:4840`
3. Security Mode: None
4. Connect

### 노드 탐색

1. **Address Space** 탭에서 `Objects` 폴더 열기
2. 모든 센서 노드가 String Identifier로 표시됩니다:
   - `ns=2;s=TemperatureSensor_Tank1`
   - `ns=2;s=PressureSensor_Pump1`
   - `ns=2;s=DoorSensor_MainEntrance`
   - 등...

3. 노드를 더블클릭하면 현재 값을 볼 수 있습니다
4. **Data Access View**로 드래그하여 실시간 모니터링 가능

---

## 📊 검증 체크리스트

### 기본 기능

- [ ] 서버가 정상적으로 시작되는가?
- [ ] 모든 26개 센서 노드가 생성되는가?
- [ ] String Identifier 형식이 올바른가? (ns=2;s=TagName)
- [ ] 클라이언트가 연결에 성공하는가?

### 데이터 타입

- [ ] Float64 노드를 읽을 수 있는가? (온도, 압력, 진동 등)
- [ ] Boolean 노드를 읽을 수 있는가? (도어, 모션, 릴레이 등)
- [ ] Int32 노드를 읽을 수 있는가? (모터 속도, 밸브 위치 등)

### 실시간 업데이트

- [ ] 값이 실시간으로 변경되는가?
- [ ] 온도 센서가 사인파 패턴을 보이는가?
- [ ] 압력 센서가 램프 패턴을 보이는가?
- [ ] 디지털 센서가 ON/OFF를 전환하는가?

### PLC 로직

- [ ] PLC 로직이 실행되는가?
- [ ] 태그 값이 PLC 로직에 의해 변경되는가?
- [ ] 제어 로직이 정상 동작하는가?

---

## 🛠️ 트러블슈팅

### 연결 실패 (Connection Timeout)

**증상:**
```
Failed to connect: dial tcp 127.0.0.1:4840: i/o timeout
```

**해결 방법:**
1. 서버가 실행 중인지 확인: `ps aux | grep server`
2. 포트가 열려있는지 확인: `netstat -an | grep 4840`
3. PKI 인증서가 존재하는지 확인: `ls -la pki/`
4. 필요시 PKI 재생성: `./setup_pki.sh`

### 노드를 찾을 수 없음

**증상:**
```
Node not found: ns=2;s=SensorName
```

**해결 방법:**
1. 서버 로그에서 노드 매핑 확인
2. 노드 ID 형식 확인 (String Identifier 사용)
3. 태그 이름 정확히 입력 (대소문자 구분)

### 값이 업데이트되지 않음

**해결 방법:**
1. 센서 시뮬레이션이 실행 중인지 확인
2. 센서의 `enabled` 속성 확인 (sensors.json)
3. 서버 로그에서 "Sensor Update" 메시지 확인

---

## 📝 EPICS 통합 준비

String Identifier를 사용하므로 EPICS DB 파일 생성 시 다음과 같이 직접 매핑할 수 있습니다:

```
record(ai, "TANK1:TEMP") {
    field(DTYP, "opcua")
    field(INP,  "@opc.tcp://localhost:4840 ns=2;s=TemperatureSensor_Tank1")
    field(SCAN, "1 second")
}

record(bi, "DOOR:STATUS") {
    field(DTYP, "opcua")
    field(INP,  "@opc.tcp://localhost:4840 ns=2;s=DoorSensor_MainEntrance")
    field(SCAN, "1 second")
}

record(longin, "MOTOR:SPEED") {
    field(DTYP, "opcua")
    field(INP,  "@opc.tcp://localhost:4840 ns=2;s=MotorSpeed_Conveyor")
    field(SCAN, "1 second")
}
```

태그 이름이 그대로 노드 ID에 사용되므로 매핑이 매우 직관적입니다.

---

## 🔗 추가 리소스

- **서버 매뉴얼**: [MANUAL.md](MANUAL.md)
- **테스트 가이드**: [TESTING.md](TESTING.md)
- **빠른 시작**: [QUICKSTART.md](QUICKSTART.md)
- **README**: [README.md](README.md)

---

**검증 성공 시 확인사항:**
✅ OPC UA 서버 정상 동작
✅ 26개 센서 노드 모두 접근 가능
✅ 실시간 데이터 업데이트 확인
✅ EPICS 통합 준비 완료
