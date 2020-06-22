package flexi

type Handler struct {
}

const (
	TypeTask     = "task"
	TypeInit     = "bind"
	TypeReady    = "ready"
	TypeRegister = "register"
	TypeInput    = "input"
	TypeOK       = "ok"
	TypeErr      = "error"
	TypeStatus   = "status"
)

func (h *Handler) Handle(ctx context.Context, msg *Msg) {
	switch msg.Type {
	case TypeTask:
		// TODO: init queue, reply with bind message.
	case TypeReady:
		// TODO: spawn process.
	default:
		// TODO: log & discard.
	}
}
