package httpclient

type Response struct {
	Status     int                 `json:"status"`
	StatusText string              `json:"statusText"`
	Headers    map[string][]string `json:"headers,omitempty"`
	Body       string              `json:"body,omitempty"`
	DurationMs int64               `json:"durationMs"`
	SizeBytes  int                 `json:"sizeBytes"`
}
