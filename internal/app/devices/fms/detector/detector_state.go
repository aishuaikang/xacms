package detector_fsm

// DetectorHeartbeatUpdateTime 心跳更新时间
type DetectorHeartbeatUpdateTime int64

// GetDetectorHeartbeatUpdateTime 获取心跳更新时间
func (d *DetectorHeartbeatUpdateTime) GetDetectorHeartbeatUpdateTime() int64 {
	return int64(*d)
}

// SetDetectorHeartbeatUpdateTime 设置心跳更新时间
func (d *DetectorHeartbeatUpdateTime) SetDetectorHeartbeatUpdateTime(t int64) {
	*d = DetectorHeartbeatUpdateTime(t)
}
