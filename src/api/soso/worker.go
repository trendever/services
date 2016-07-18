package soso

type Task struct {
	Session Session
	Message string
}

type Worker func(Task)

type Manager struct {
	Pool []Worker
	Min  int
	Max  int
	ch   chan Task
}

func NewManager(min, max int) *Manager {
	m := &Manager{
		Min: min,
		Max: max,
		ch:  make(chan Task),
	}
	return m
}

func (m *Manager) do(t Task) {

}
