package storage

type Storage struct {
}

func New() *Storage {
	return &Storage{}
}

func (s *Storage) Read(key []byte) []byte {
	return nil
}

func (s *Storage) Write(key []byte, d []byte) int {
	return -1
}

func (s *Storage) Delete(key []byte) bool {
	return false
}
