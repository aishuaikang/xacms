package detector_fsm

type DetectorState string
type FPVEvent string

const (
	// 状态定义
	StateOffline          DetectorState = "Offline"          // 离线
	StateOmniDirection    DetectorState = "OmniDirection"    // 全向侦测
	StateDirection        DetectorState = "Direction"        // 定向侦测
	StateSpectrumAnalyzer DetectorState = "SpectrumAnalyzer" // 频谱分析
)

const (
	// 事件定义
	EventToOffline  FPVEvent = "to_offline"  // 切换到离线
	EventToOmni     FPVEvent = "to_omni"     // 切换到全向侦测
	EventToDirect   FPVEvent = "to_direct"   // 切换到定向侦测
	EventToSpectrum FPVEvent = "to_spectrum" // 切换到频谱分析
)
