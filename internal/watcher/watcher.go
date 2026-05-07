package watcher

type Event struct {
	Path string `json:"path"`
	Op   string `json:"op"`
}

type Watcher struct{}

func New() *Watcher {
	return &Watcher{}
}

func (w *Watcher) Start(_ string) error {
	return nil
}

func (w *Watcher) Stop() error {
	return nil
}
