package devices

type DeviceStarter interface {
	Start()
}

type Devices struct {
	devices []DeviceStarter
}

func NewDevices(fpvDevice *FPVDevice, parseDevice *ParseDevice) *Devices {
	return &Devices{
		devices: []DeviceStarter{
			fpvDevice,
			parseDevice,
		},
	}
}

func (d *Devices) Start() {
	for _, device := range d.devices {
		go device.Start()
	}
}
