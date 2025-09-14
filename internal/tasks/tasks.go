package tasks

type TaskExecutor interface {
	Execute()
}

type Tasks struct {
	tasks []TaskExecutor
}

func NewTasks(decryptTokenTask *DecryptTokenTask, devicesTask *DevicesTask) *Tasks {
	return &Tasks{
		tasks: []TaskExecutor{
			decryptTokenTask,
			devicesTask,
		},
	}
}

func (t *Tasks) Execute() {
	for _, executor := range t.tasks {
		executor.Execute()
	}
}
