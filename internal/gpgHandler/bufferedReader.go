package gpgHandler

//go:generate mockgen -source=bufferedReader.go -destination=../mocks/MockBufferedReader.go -package mocks

// BufferedReader
// bufio.NewReader()
type BufferedReader interface {
	ReadBytes(delim byte) ([]byte, error)
	ReadString(delim byte) (string, error)
	Read(p []byte) (n int, err error)
}
