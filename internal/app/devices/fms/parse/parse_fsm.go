package parse_fsm

import (
	"context"

	"github.com/looplab/fsm"
)

type ParseFsm struct {
	*fsm.FSM
	offlineHandler *offlineHandler
	onlineHandler  *onlineHandler
}

// NewParseFsm 创建新的Parse状态机实例
func NewParseFsm() *ParseFsm {
	d := &ParseFsm{
		offlineHandler: &offlineHandler{},
		onlineHandler:  &onlineHandler{},
	}

	d.FSM = fsm.NewFSM(
		string(StateOffline), // 初始状态
		fsm.Events{
			{Name: string(EventToOffline), Src: []string{string(StateOnline)}, Dst: string(StateOffline)},
			{Name: string(EventToOnline), Src: []string{string(StateOffline)}, Dst: string(StateOnline)},
		},
		fsm.Callbacks{
			"before_event": func(ctx context.Context, e *fsm.Event) { d.beforeEvent(ctx, e) },
			"enter_state":  func(ctx context.Context, e *fsm.Event) { d.enterState(ctx, e) },
		},
	)

	return d
}

// beforeEvent 事件前回调
func (d *ParseFsm) beforeEvent(ctx context.Context, e *fsm.Event) {
	switch e.Dst {
	case string(StateOffline):
		d.offlineHandler.Before(ctx, e)
	case string(StateOnline):
		d.onlineHandler.Before(ctx, e)
	}
}

// enterState 进入状态回调
func (d *ParseFsm) enterState(ctx context.Context, e *fsm.Event) {
	switch e.Dst {
	case string(StateOffline):
		d.offlineHandler.Enter(ctx, e)
	case string(StateOnline):
		d.onlineHandler.Enter(ctx, e)
	}
}
