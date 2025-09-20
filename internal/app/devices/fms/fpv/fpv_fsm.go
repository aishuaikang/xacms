package fpv_fsm

import (
	"context"
	"uav_defender/internal/app/devices/conn"

	"github.com/looplab/fsm"
)

type FPVFsm struct {
	*fsm.FSM
	offlineHandler  *OfflineHandler
	gazingHandler   *GazingHandler
	scanningHandler *ScanningHandler
}

// NewFPVFsm 创建新的FPV状态机实例
func NewFPVFsm(fpvConnection *conn.FpvConnection) *FPVFsm {
	d := &FPVFsm{
		offlineHandler: &OfflineHandler{},
		gazingHandler: &GazingHandler{
			fpvConnection: fpvConnection,
		},
		scanningHandler: &ScanningHandler{},
	}

	d.FSM = fsm.NewFSM(
		string(StateOffline), // 初始状态
		fsm.Events{
			{Name: string(EventToOffline), Src: []string{string(StateGazing), string(StateScanning)}, Dst: string(StateOffline)},
			{Name: string(EventToGazing), Src: []string{string(StateScanning)}, Dst: string(StateGazing)},
			{Name: string(EventToScanning), Src: []string{string(StateOffline), string(StateGazing)}, Dst: string(StateScanning)},
		},
		fsm.Callbacks{
			"before_event": func(ctx context.Context, e *fsm.Event) { d.beforeEvent(ctx, e) },
			"enter_state":  func(ctx context.Context, e *fsm.Event) { d.enterState(ctx, e) },
		},
	)

	return d
}

// beforeEvent 事件前回调
func (d *FPVFsm) beforeEvent(ctx context.Context, e *fsm.Event) {
	switch e.Dst {
	case string(StateOffline):
		d.offlineHandler.Before(ctx, e)
	case string(StateGazing):
		d.gazingHandler.Before(ctx, e)
	case string(StateScanning):
		d.scanningHandler.Before(ctx, e)
	}
}

// enterState 进入状态回调
func (d *FPVFsm) enterState(ctx context.Context, e *fsm.Event) {
	switch e.Dst {
	case string(StateOffline):
		d.offlineHandler.Enter(ctx, e)
	case string(StateGazing):
		d.gazingHandler.Enter(ctx, e)
	case string(StateScanning):
		d.scanningHandler.Enter(ctx, e)
	}
}
