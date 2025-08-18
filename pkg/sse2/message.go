package sse2

type Message struct {
	Event string `json:"event"`
	Data  any    `json:"data"`
}
