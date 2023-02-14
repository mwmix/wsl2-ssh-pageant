package gpgHandler

import "io"

//go:generate mockgen -source=writer.go -destination=../mocks/MockWriter.go -package mocks
type MockWriter interface {
	io.Writer
}
