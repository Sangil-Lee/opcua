# Changelog

## 2026-01-07 - String Identifier 업데이트

### 주요 변경사항

#### 1. OPC UA 노드 ID를 String Identifier로 변경

**변경 전:**
- Numeric Identifier 사용: `ns=2;i=1000`, `ns=2;i=1001`, ...
- 노드 ID와 태그 이름 간 추가 매핑 필요

**변경 후:**
- String Identifier 사용: `ns=2;s=TemperatureSensor_Tank1`, `ns=2;s=PressureSensor_Pump1`, ...
- 태그 이름이 그대로 노드 ID로 사용됨

**장점:**
- ✅ EPICS DB 생성 시 직관적인 매핑
- ✅ 사용자가 노드 ID만 보고 센서 식별 가능
- ✅ 센서 이름 변경 시 일관성 유지 용이
- ✅ 자동화 스크립트 작성 시 편리

### 코드 변경

#### internal/opcuaserver/server.go

1. **nodeMapping 타입 변경**
   ```go
   // 변경 전
   nodeMapping map[string]uint32

   // 변경 후
   nodeMapping map[string]string  // tag name -> node ID string
   ```

2. **NodeID 생성 방식 변경**
   ```go
   // 변경 전
   ua.NodeIDNumeric{NamespaceIndex: 2, ID: nodeID}

   // 변경 후
   ua.NodeIDString{NamespaceIndex: 2, ID: tag.Name}
   ```

3. **출력 형식 변경**
   ```go
   // 변경 전
   fmt.Sprintf("ns=2;i=%d", nodeID)

   // 변경 후
   fmt.Sprintf("ns=2;s=%s", tag.Name)
   ```

### 문서 업데이트

#### README.md
- OPC UA 노드 매핑 섹션 업데이트
- 클라이언트 예제를 String Identifier로 변경
- EPICS 통합 섹션 추가 (예시 DB 파일 포함)

#### CLIENT_TESTING.md
- 완전히 재작성
- String Identifier 사용 예제 추가
- EPICS DB 파일 예제 추가
- 검증 체크리스트 업데이트

#### QUICKSTART.md
- PKI 설정 단계 추가 (setup_pki.sh)

### 사용 예시

#### 변경 전 (Numeric Identifier)
```bash
# 클라이언트 사용
./bin/client -node "ns=2;i=1000"

# EPICS DB (노드 ID 찾기 어려움)
record(ai, "TANK1:TEMP") {
    field(INP,  "@opc.tcp://localhost:4840 ns=2;i=1000")
}
```

#### 변경 후 (String Identifier)
```bash
# 클라이언트 사용 (태그 이름 그대로)
./bin/client -node "ns=2;s=TemperatureSensor_Tank1"

# EPICS DB (직관적)
record(ai, "TANK1:TEMP") {
    field(INP,  "@opc.tcp://localhost:4840 ns=2;s=TemperatureSensor_Tank1")
}
```

### 호환성

- ✅ 기존 센서 설정 파일 (sensors.json) 호환
- ✅ 기존 PLC 로직 스크립트 (plc_logic.lua) 호환
- ✅ 모든 센서 타입 정상 동작
- ⚠️ **클라이언트 코드는 String Identifier로 업데이트 필요**

### 테스트 결과

#### 검증 완료 항목
- ✅ Float64 노드 읽기 (온도, 압력, 진동 등)
- ✅ Boolean 노드 읽기 (도어, 릴레이, 모션 등)
- ✅ Int32 노드 읽기 (모터 속도, 밸브 위치 등)
- ✅ 실시간 데이터 업데이트
- ✅ 연속 읽기 모드
- ✅ 26개 센서 모두 정상 동작

#### 테스트 환경
- Go 1.23.8
- github.com/awcullen/opcua v1.4.0
- github.com/gopcua/opcua v0.8.0
- Linux (WSL2)

### 추가 파일

#### setup_pki.sh
OPC UA 서버에 필요한 PKI (인증서) 자동 생성 스크립트

```bash
./setup_pki.sh
```

생성 파일:
- `pki/server.key`: 서버 개인키
- `pki/server.crt`: 서버 인증서

### 마이그레이션 가이드

기존 numeric identifier를 사용하는 클라이언트 코드가 있다면:

1. **노드 ID 형식 변경**
   ```
   ns=2;i=1000  →  ns=2;s=TemperatureSensor_Tank1
   ns=2;i=1001  →  ns=2;s=DoorSensor_MainEntrance
   ns=2;i=1002  →  ns=2;s=MotorSpeed_Conveyor
   ```

2. **서버 로그에서 전체 매핑 확인**
   서버 시작 시 출력되는 노드 매핑 테이블 참고

3. **EPICS DB 파일 업데이트**
   자동 생성 스크립트 또는 수동으로 노드 ID 업데이트

### 향후 계획

- [ ] EPICS DB 자동 생성 도구 개발
- [ ] UaExpert 예제 스크린샷 추가
- [ ] Python 클라이언트 상세 예제 추가
- [ ] 성능 벤치마크 문서 작성

---

**참고 문서:**
- [README.md](README.md): 프로젝트 개요
- [CLIENT_TESTING.md](CLIENT_TESTING.md): 클라이언트 검증 가이드
- [MANUAL.md](MANUAL.md): 상세 매뉴얼
- [QUICKSTART.md](QUICKSTART.md): 빠른 시작 가이드
