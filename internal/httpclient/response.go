package httpclient

type Response struct {
	Status     int                 `json:"status"`
	StatusText string              `json:"statusText"`
	Headers    map[string][]string `json:"headers,omitempty"`
	Body       string              `json:"body,omitempty"`
	DurationMs int64               `json:"durationMs"`
	SizeBytes  int                 `json:"sizeBytes"`
	Timeline   *Timeline           `json:"timeline,omitempty"`
}

type Timeline struct {
	DNSMs      int64 `json:"dnsMs"`
	ConnectMs  int64 `json:"connectMs"`
	TLSMs      int64 `json:"tlsMs"`
	TTFBMs     int64 `json:"ttfbMs"`
	DownloadMs int64 `json:"downloadMs"`
}
