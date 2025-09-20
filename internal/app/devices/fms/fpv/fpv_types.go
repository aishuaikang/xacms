package fpv_fsm

type FPVState string
type FPVEvent string

const (
	// 状态定义
	StateOffline  FPVState = "Offline"  // 离线
	StateGazing   FPVState = "Gazing"   // 凝视
	StateScanning FPVState = "Scanning" // 扫频
)

const (
	// 事件定义
	EventToOffline  FPVEvent = "to_offline"  // 切换到离线
	EventToGazing   FPVEvent = "to_gazing"   // 切换到凝视
	EventToScanning FPVEvent = "to_scanning" // 切换到扫频
)
