package gpgHandler

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// TODO: convert these to use gomocks
var OpenFunc func(string) (*os.File, error) = os.Open
var GetBufferedReader func(rd io.Reader) BufferedReader = func(rd io.Reader) BufferedReader {
	return bufio.NewReader(rd)
}

type GPGHandler struct {
	gpgAgentPort int
	nonceValue   [16]byte

	// For unit testing purposes
	conn NetConn
}

func NewGPGHandler(socketPath string, netConn NetConn) (*GPGHandler, error) {
	gpgSocketFile, err := OpenFunc(socketPath)
	if err != nil {
		return nil, err
	}

	port, nonceValue, err := getPortAndNonce(GetBufferedReader(gpgSocketFile))
	if err != nil {
		return nil, err
	}

	if netConn == nil {
		netConn, err = net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", port), 3*time.Second)
		if err != nil {
			return nil, err
		}
	}

	return &GPGHandler{
		gpgAgentPort: port,
		nonceValue:   nonceValue,
		conn:         netConn,
	}, nil
}

func getPortAndNonce(gpgSocketReader BufferedReader) (int, [16]byte, error) {
	tmp, err := gpgSocketReader.ReadBytes('\n')
	if err != nil {
		return 0, [16]byte{}, err
	}

	port, err := strconv.Atoi(string(tmp))
	if err != nil {
		return 0, [16]byte{}, err
	}

	var nonceValue [16]byte
	_, err = gpgSocketReader.Read(nonceValue[:])
	if err != nil {
		return 0, [16]byte{}, err
	}

	return port, nonceValue, nil
}

func (handler *GPGHandler) Close() error {
	if handler.conn != nil {
		return handler.conn.Close()
	}

	return nil
}

func (handler *GPGHandler) authenticate() error {
	_, err := handler.conn.Write(handler.nonceValue[:])
	if err != nil {
		return err
	}

	//responseMessage, err := bufio.NewReader(gpgConn).ReadString('\n')
	bufferedReader := GetBufferedReader(handler.conn)
	responseMessage, err := bufferedReader.ReadString('\n')
	if err != nil {
		return err
	}

	// Could just check that we got back "OK" and ignore what comes after too. Not sure if that is better ...
	if responseMessage != "OK Pleased to meet you\n" {
		return fmt.Errorf("received unexpected reply from gpg-agent. Got: %v", responseMessage)
	}

	return err
}

func MustFprintf(w io.Writer, format string, a ...any) {
	_, err := fmt.Fprintf(w, format, a...)
	if err != nil {
		panic(err)
	}
}

func (handler *GPGHandler) GetVersion() (string, error) {
	err := handler.authenticate()
	if err != nil {
		return "", err
	}

	MustFprintf(handler.conn, "GETINFO version\n")

	responseMessage, err := GetBufferedReader(handler.conn).ReadString('\n')
	if err != nil {
		return "", err
	}

	MustFprintf(handler.conn, "BYE\n")

	// Server returns a message that looks like: "D 2.4.0\n"
	splitResponse := strings.Split(responseMessage, " ")
	trimmedResponse := strings.TrimRight(splitResponse[1], "\n")
	return trimmedResponse, nil
}

func (handler *GPGHandler) Handle() error {
	handler.authenticate()

	// Make sure this is the best way to do this ...
	// original code had the initial stdin copy be placed in a go func() for some reason.
	_, err := io.Copy(handler.conn, os.Stdin)
	if err != nil {
		return err
	}

	_, err = io.Copy(os.Stdout, handler.conn)
	if err != nil {
		return err
	}

	return nil
}
