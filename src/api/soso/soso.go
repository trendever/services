package soso

const (
	Version   string = "0.2"
	logPrefix string = "[SOSO]"
)

type Engine struct {
	*Router
}

func (s *Engine) RunReceiver(session Session) {
	// Обработка входящих сообщений.
	for {
		if msg, err := session.Recv(); err == nil {
			go s.execute(session, msg)
			continue
		}
		Sessions.Pull(session)
		break
	}
}

func New() *Engine {
	soso := Engine{}
	soso.Router = NewRouter()
	return &soso
}

func Default() *Engine {
	return New()
}
