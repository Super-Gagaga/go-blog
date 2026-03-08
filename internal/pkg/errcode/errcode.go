package errcode

const (
	Success        = 0
	InvalidParams  = 40001
	Unauthorized   = 40101
	Forbidden      = 40301
	NotFound       = 40401
	RateLimited    = 42901
	InternalServer = 50001
)

var messages = map[int]string{
	Success:        "ok",
	InvalidParams:  "invalid parameters",
	Unauthorized:   "unauthorized",
	Forbidden:      "forbidden",
	NotFound:       "not found",
	RateLimited:    "rate limit exceeded",
	InternalServer: "internal server error",
}

func Message(code int) string {
	if msg, ok := messages[code]; ok {
		return msg
	}
	return messages[InternalServer]
}
