package request

type Request struct {
	Data    any    `json:"data"`
	TraceId string `json:"trace_id"`
}
