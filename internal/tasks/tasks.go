package tasks

type Executor interface {
	Execute()
}

// 任务管理器
type Tasks struct {
	tasks []Executor
}

func NewTasks(decryptTokenTask *DecryptTokenTask, devicesTask *DevicesTask, parseTask *ParseTask, fpvTask *FPVTask) *Tasks {
	return &Tasks{
		tasks: []Executor{
			decryptTokenTask,
			devicesTask,
			parseTask,
			fpvTask,
		},
	}
}

func (t *Tasks) Execute() {
	for _, executor := range t.tasks {
		executor.Execute()
	}
}
