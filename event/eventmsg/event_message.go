package eventmsg

import "time"

type Message[T any] struct {
	Key       string
	Value     T
	Timestamp time.Time
	Headers   map[string][]string
}
