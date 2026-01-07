package logic

// LogicProvider는 가상 PLC의 개별 노드가 동작하는 방식을 정의합니다.
type LogicProvider interface {
	ID() string             // 노드 ID (e.g., "ns=2;s=Temperature")
	BrowseName() string     // 노드 이름
	NextValue() interface{} // 다음 시뮬레이션 값 계산
}
