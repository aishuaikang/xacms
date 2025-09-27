package detector_fsm

import (
	"context"

	"github.com/looplab/fsm"
)

type DetectorFsm struct {
	*fsm.FSM
	DetectorHeartbeatUpdateTime
	offlineHandler  *OfflineHandler
	omniHandler     *OmniDirectionHandler
	directHandler   *DirectionHandler
	spectrumHandler *SpectrumAnalyzerHandler
}

// NewDetectorFsm 创建新的FPV状态机实例
func NewDetectorFsm() *DetectorFsm {

	d := &DetectorFsm{
		offlineHandler:  &OfflineHandler{},
		omniHandler:     &OmniDirectionHandler{},
		directHandler:   &DirectionHandler{},
		spectrumHandler: &SpectrumAnalyzerHandler{},
	}

	d.FSM = fsm.NewFSM(
		string(StateOffline), // 初始状态
		fsm.Events{
			{Name: string(EventToOffline), Src: []string{string(StateOmniDirection), string(StateDirection), string(StateSpectrumAnalyzer)}, Dst: string(StateOffline)},
			{Name: string(EventToOmni), Src: []string{string(StateOffline), string(StateDirection), string(StateSpectrumAnalyzer)}, Dst: string(StateOmniDirection)},
			{Name: string(EventToDirect), Src: []string{string(StateOmniDirection)}, Dst: string(StateDirection)},
			{Name: string(EventToSpectrum), Src: []string{string(StateOmniDirection), string(StateDirection)}, Dst: string(StateSpectrumAnalyzer)},
		},
		fsm.Callbacks{
			"before_event": func(ctx context.Context, e *fsm.Event) { d.beforeEvent(ctx, e) },
			"enter_state":  func(ctx context.Context, e *fsm.Event) { d.enterState(ctx, e) },
		},
	)

	return d
}

// beforeEvent 事件前回调
func (d *DetectorFsm) beforeEvent(ctx context.Context, e *fsm.Event) {
	switch e.Dst {
	case string(StateOffline):
		d.offlineHandler.Before(ctx, e)
	case string(StateOmniDirection):
		d.omniHandler.Before(ctx, e)
	case string(StateDirection):
		d.directHandler.Before(ctx, e)
	case string(StateSpectrumAnalyzer):
		d.spectrumHandler.Before(ctx, e)

	}
}

// enterState 进入状态回调
func (d *DetectorFsm) enterState(ctx context.Context, e *fsm.Event) {
	switch e.Dst {
	case string(StateOffline):
		d.offlineHandler.Enter(ctx, e)
	case string(StateOmniDirection):
		d.omniHandler.Enter(ctx, e)
	case string(StateDirection):
		d.directHandler.Enter(ctx, e)
	case string(StateSpectrumAnalyzer):
		d.spectrumHandler.Enter(ctx, e)
	}
}
