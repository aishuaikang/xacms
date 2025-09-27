package tasks

type Executor interface {
	Execute()
}

// 任务管理器
type Tasks struct {
	tasks []Executor
}

func NewTasks(decryptTokenTask *DecryptTokenTask, devicesTask *DevicesTask, parseTask *ParseTask, fpvTask *FPVTask, detectorTask *DetectorTask, parseSyncDetectorTask *ParseSyncDetectorTask) *Tasks {
	return &Tasks{
		tasks: []Executor{
			decryptTokenTask,
			devicesTask,
			parseTask,
			fpvTask,
			detectorTask,
			parseSyncDetectorTask,
		},
	}
}

func (t *Tasks) Execute() {
	for _, executor := range t.tasks {
		executor.Execute()
	}
}
