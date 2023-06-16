package khronos

type contextKey struct {
	name string
}

func (k *contextKey) String() string {
	return "khronos context value " + k.name
}

var (
	ServerContextKey = &contextKey{"khronos-server"}

	QueueContextKey = &contextKey{"khronos-queue"}
)
