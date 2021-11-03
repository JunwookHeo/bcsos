package network

type Network struct {
}

func Start() bool{
	return false
}

func Stop() bool{
	return false
}

func Send(dst []byte, data []byte) int {
	return -1
}

func Receive(buf *[]byte, size int) int {
	return -1
}
