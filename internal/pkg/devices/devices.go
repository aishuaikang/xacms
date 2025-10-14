package devices

type Device interface {
	Run()
}

// 设备管理器
type DeviceManager struct {
	devices []Device
}

func NewDeviceManager(parseDevice *ParseDevice, fpvDevice *FpvDevice, strikeDevice *StrikeDevice) *DeviceManager {
	return &DeviceManager{
		devices: []Device{
			// parseDevice,
			// fpvDevice,
			strikeDevice,
		},
	}
}

func (m *DeviceManager) Run() {
	for _, device := range m.devices {
		device.Run()
	}
}
