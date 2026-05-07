package storage

type Storage struct{}

func New() *Storage {
	return &Storage{}
}

func (s *Storage) Open(_ string) error {
	return nil
}

func (s *Storage) Close() error {
	return nil
}
