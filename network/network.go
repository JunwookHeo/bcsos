package network

type Network struct {
}

func New() *Network {
	return &Network{}
}

func Send(dst []byte, data []byte) int {
	return -1
}

func Receive(buf *[]byte, size int) int {
	return -1
}
