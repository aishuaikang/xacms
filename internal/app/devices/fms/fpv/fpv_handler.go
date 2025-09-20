package fpv_fsm

import (
	"context"
	"fmt"
	conn_ "uav_defender/internal/app/devices/conn"
	"uav_defender/internal/pkg/global"

	"github.com/looplab/fsm"
	"go.uber.org/zap"
)

type OfflineHandler struct{}

// Before 离线状态前回调
func (h *OfflineHandler) Before(ctx context.Context, e *fsm.Event) {
}

// Enter 离线状态进入回调
func (h *OfflineHandler) Enter(ctx context.Context, e *fsm.Event) {
}

type GazingHandler struct {
	fpvConnection *conn_.FpvConnection
}

// Before 凝视状态前回调
func (h *GazingHandler) Before(ctx context.Context, e *fsm.Event) {

	if len(e.Args) < 2 {
		e.Cancel(fmt.Errorf("缺少参数"))
		return
	}
	addr, ok := e.Args[0].(string)
	if !ok {
		e.Cancel(fmt.Errorf("类型断言失败"))
		return
	}
	frequency, ok := e.Args[1].(int)
	if !ok {
		e.Cancel(fmt.Errorf("类型断言失败"))
		return
	}

	conn, exists := h.fpvConnection.GetConnection(addr)
	if !exists {
		e.Cancel(fmt.Errorf("FPV 连接不存在"))
		return
	}

	fpvCommand, expectedResponse := BuildFPVCommand(frequency, 9)

	// 发送命令
	if err := conn.SendCommand(fpvCommand); err != nil {
		e.Cancel(fmt.Errorf("发送命令失败: %v", err))
		return
	}

	// 等待响应
	response, err := conn.WaitResponse()
	if err != nil {
		e.Cancel(fmt.Errorf("等待响应失败: %v", err))
		return
	}

	global.Logger.Info("收到响应", zap.String("address", addr), zap.String("response", response))

	// 验证响应
	if !IsExpectedResponse(response, expectedResponse) {
		e.Cancel(fmt.Errorf("设备响应不符合预期: %q != %q", response, expectedResponse))
		return
	}

	global.Logger.Info("FPV 设备响应符合预期", zap.String("address", addr), zap.String("response", response))

	// reflectVal := reflect.ValueOf(result)
	// if reflectVal.Kind() != reflect.Ptr || reflectVal.IsNil() {
	// 	e.Cancel(fmt.Errorf("result 必须是非空指针"))
	// 	return
	// }

	// // 设置结果
	// reflectVal.Elem().Set(reflect.ValueOf(&frequency))

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
