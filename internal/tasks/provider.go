package tasks

import "github.com/google/wire"

var TaskSet = wire.NewSet(
	NewDecryptTokenTask,
	NewDevicesTask,
	NewParseTask,
	NewDetectorTask,
	NewParseSyncDetectorTask,
	NewFPVTask,
	NewTasks,
)
