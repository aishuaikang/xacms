package parse_fsm

import (
	"context"

	"github.com/looplab/fsm"
)

type offlineHandler struct{}

// Before 离线状态前回调
func (h *offlineHandler) Before(ctx context.Context, e *fsm.Event) {
}

// Enter 离线状态进入回调
func (h *offlineHandler) Enter(ctx context.Context, e *fsm.Event) {
}

type onlineHandler struct{}

// Before 在线状态前回调
func (h *onlineHandler) Before(ctx context.Context, e *fsm.Event) {
}

// Enter 在线状态进入回调
func (h *onlineHandler) Enter(ctx context.Context, e *fsm.Event) {
}
