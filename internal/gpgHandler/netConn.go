package gpgHandler

import "net"

//go:generate mockgen -source=netConn.go -destination=../mocks/MockNetConn.go -package mocks
type NetConn interface {
	net.Conn
}
