package devices

type Starter interface {
	Start()
}

type Devices struct {
	devices []Starter
}

func NewDevices(fpvDevice *FPVDevice, parseDevice *ParseDevice, detectorDevice *DetectorDevice) *Devices {
	return &Devices{
		devices: []Starter{
			fpvDevice,
			parseDevice,
			detectorDevice,
		},
	}
}

func (d *Devices) Start() {
	for _, device := range d.devices {
		go device.Start()
	}
}
