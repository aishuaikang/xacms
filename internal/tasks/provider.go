package tasks

import "github.com/google/wire"

var TaskSet = wire.NewSet(
	NewTaskManager,
)
