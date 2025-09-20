package fms

import (
	"context"

	"github.com/looplab/fsm"
)

// StateHandler 用于状态机的 before/enter 回调
type StateHandler interface {
	Before(ctx context.Context, e *fsm.Event)
	Enter(ctx context.Context, e *fsm.Event)
}
