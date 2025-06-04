package request

type Request struct {
	Data    any    `json:"data"`
	TraceID string `json:"trace_id"`
}
