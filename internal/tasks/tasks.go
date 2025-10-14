package tasks

type Executor interface {
	Execute()
}

// 任务管理器
type TaskManager struct {
	tasks []Executor
}

func NewTaskManager() *TaskManager {
	return &TaskManager{
		tasks: []Executor{},
	}
}

func (t *TaskManager) Execute() {
	for _, executor := range t.tasks {
		executor.Execute()
	}
}
