package dto

type FPVWarningData struct {
	DetectionID int    `json:"device"`
	Freq        string `json:"freq"`
	RSSI        string `json:"rssi"`
	IP          string `json:"ip"`
	Time        int64  `json:"time"`
}
