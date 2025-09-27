package detector_fsm

import (
	"context"

	"github.com/looplab/fsm"
)

// 离线状态处理器
type OfflineHandler struct {
}

// Before 离线状态前回调
func (h *OfflineHandler) Before(ctx context.Context, e *fsm.Event) {

}

// Enter 离线状态进入回调
func (h *OfflineHandler) Enter(ctx context.Context, e *fsm.Event) {
}

// 全向状态处理器
type OmniDirectionHandler struct {
}

// Before 全向状态前回调
func (h *OmniDirectionHandler) Before(ctx context.Context, e *fsm.Event) {
}

// Enter 全向状态进入回调
func (h *OmniDirectionHandler) Enter(ctx context.Context, e *fsm.Event) {
}

// 定向状态处理器
type DirectionHandler struct {
}

// Before 定向状态前回调
func (h *DirectionHandler) Before(ctx context.Context, e *fsm.Event) {

}

// Enter 定向状态进入回调
func (h *DirectionHandler) Enter(ctx context.Context, e *fsm.Event) {

}

// 频谱仪状态处理器
type SpectrumAnalyzerHandler struct {
}

// Before 频谱仪状态前回调
func (h *SpectrumAnalyzerHandler) Before(ctx context.Context, e *fsm.Event) {
}

// Enter 频谱仪状态进入回调
func (h *SpectrumAnalyzerHandler) Enter(ctx context.Context, e *fsm.Event) {
}
