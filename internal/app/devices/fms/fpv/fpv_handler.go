package fpv_fsm

import (
	"context"

	"github.com/looplab/fsm"
)

type OfflineHandler struct{}

// Before 离线状态前回调
func (h *OfflineHandler) Before(ctx context.Context, e *fsm.Event) {
}

// Enter 离线状态进入回调
func (h *OfflineHandler) Enter(ctx context.Context, e *fsm.Event) {
}

type GazingHandler struct{}

// Before 凝视状态前回调`
func (h *GazingHandler) Before(ctx context.Context, e *fsm.Event) {

}

// Enter 凝视状态进入回调
func (h *GazingHandler) Enter(ctx context.Context, e *fsm.Event) {

}

type ScanningHandler struct{}

// Before 扫频状态前回调
func (h *ScanningHandler) Before(ctx context.Context, e *fsm.Event) {

}

// Enter 扫频状态进入回调
func (h *ScanningHandler) Enter(ctx context.Context, e *fsm.Event) {

}
