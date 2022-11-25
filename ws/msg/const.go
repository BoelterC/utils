package msg

const (
	MSG_HEAD     = "socket:connect."
	MSG_TAIL     = ".end"
	HEAD_MIN_LEN = 32
)

const (
	HEAD_INCOMPLETE = 0
	HEAD_COMPLETE   = 1
	HEAD_ILLEGAL    = -1
)
