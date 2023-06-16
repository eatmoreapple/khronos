package khronos

type Status int

const (
	OK Status = iota
	Pong
)

func (s Status) String() string {
	switch s {
	case OK:
		return "OK"
	case Pong:
		return "PONG"
	}
	return ""
}
