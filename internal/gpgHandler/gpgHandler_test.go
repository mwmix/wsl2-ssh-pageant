package gpgHandler

import (
	"bytes"
	"errors"
	"io"
	"os"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/ninjasanonymous/wsl2gpggo/internal/mocks"
)

const (
	SOCKET_PATH = `C:\Users\Matt\AppData\Local\gnupg\S.gpg-agent`
)

var OriginalOpenFunc func(string) (*os.File, error) = OpenFunc
var OriginalGetBufferedReader func(rd io.Reader) BufferedReader = GetBufferedReader

func TestMain(m *testing.M) {
	OriginalOpenFunc = OpenFunc
	OriginalGetBufferedReader = GetBufferedReader

	exitCode := m.Run()

	OpenFunc = OriginalOpenFunc
	GetBufferedReader = OriginalGetBufferedReader

	os.Exit(exitCode)
}

func TestNewGPGHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockNetConn := mocks.NewMockNetConn(ctrl)

	type args struct {
		socketPath string
		netConn    NetConn
	}
	tests := []struct {
		name         string
		args         args
		mockOpenFunc func(string) (*os.File, error)
		gpgAgentPort string
		nonceValue   [16]byte
		want         *GPGHandler
		wantErr      bool
	}{
		{
			"gets new handler",
			args{
				socketPath: SOCKET_PATH,
				netConn:    mockNetConn,
			},
			func(s string) (*os.File, error) {
				return nil, nil
			},
			"123456",
			[16]byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '1', '2', '3', '4', '5', '6', '7'},
			&GPGHandler{
				gpgAgentPort: 123456,
				nonceValue:   [16]byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '1', '2', '3', '4', '5', '6', '7'},
				conn:         mockNetConn,
			},
			false,
		},
	}
	for _, tt := range tests {
		OpenFunc = tt.mockOpenFunc
		GetBufferedReader = func(rd io.Reader) BufferedReader {
			mockBufferedReader := mocks.NewMockBufferedReader(ctrl)
			mockBufferedReader.EXPECT().ReadBytes(byte('\n')).Return([]byte(tt.gpgAgentPort), nil).MaxTimes(1)
			mockBufferedReader.EXPECT().Read(gomock.Any()).DoAndReturn(func(p []byte) (int, error) {
				copy(p, tt.nonceValue[:])

				return len(tt.nonceValue[:]), nil
			}).MaxTimes(1)

			return mockBufferedReader
		}

		t.Run(tt.name, func(t *testing.T) {
			got, err := NewGPGHandler(tt.args.socketPath, tt.args.netConn)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewGPGHandler() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewGPGHandler() = %v, want %v", got, tt.want)
			}
		})

		OpenFunc = OriginalOpenFunc
		GetBufferedReader = OriginalGetBufferedReader
	}
}

func TestGetVersion(t *testing.T) {
	ctrl := gomock.NewController(t)

	tests := []struct {
		name          string
		versionString string
		want          string
		authErr       error
		readStringErr error
		wantErr       bool
	}{
		{
			"Gets version successfully",
			"D 2.4.0\n",
			"2.4.0",
			nil,
			nil,
			false,
		},
		{
			"error authenticating returns error",
			"",
			"",
			errors.New("you shall not pass"),
			nil,
			true,
		},
		{
			"error reading from socket returns error",
			"",
			"",
			nil,
			errors.New("reading is hard"),
			true,
		},
	}
	for _, tt := range tests {
		mockNetConn := mocks.NewMockNetConn(ctrl)
		mockNetConn.EXPECT().Close().AnyTimes()
		mockNetConn.EXPECT().Write(gomock.Any()).AnyTimes()
		mockNetConn.EXPECT().Read(gomock.Any()).AnyTimes()

		authenticatesBufferedReader := mocks.NewMockBufferedReader(ctrl)
		if tt.authErr == nil {
			authenticatesBufferedReader.EXPECT().ReadString(byte('\n')).Return("OK Pleased to meet you\n", nil).MaxTimes(1)
		} else {
			authenticatesBufferedReader.EXPECT().ReadString(byte('\n')).Return("", tt.authErr).MaxTimes(1)
		}

		getVersionsBufferedReader := mocks.NewMockBufferedReader(ctrl)
		if tt.readStringErr == nil {
			getVersionsBufferedReader.EXPECT().ReadString(byte('\n')).Return(tt.versionString, nil).MaxTimes(1)
		} else {
			getVersionsBufferedReader.EXPECT().ReadString(byte('\n')).Return("", tt.readStringErr).MaxTimes(1)
		}

		bufferedReaders := []BufferedReader{
			authenticatesBufferedReader,
			getVersionsBufferedReader,
		}

		GetBufferedReader = func(rd io.Reader) BufferedReader {
			var reader BufferedReader
			reader, bufferedReaders = bufferedReaders[0], bufferedReaders[1:]

			return reader
		}

		t.Run(tt.name, func(t *testing.T) {
			handler := GPGHandler{
				conn: mockNetConn,
			}
			got, err := handler.GetVersion()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetVersion() error = %v, wantErr %v", err, tt.wantErr)
			}

			if got != tt.want {
				t.Errorf("GetVersion() got = %v, want %v", got, tt.want)
			}
		})

		GetBufferedReader = OriginalGetBufferedReader
	}
}

func Test_getPortAndNonce(t *testing.T) {
	ctrl := gomock.NewController(t)

	tests := []struct {
		name           string
		wantPort       int
		wantNonce      [16]byte
		bytesToRead    []byte
		readBytesError error
		readError      error
		wantErr        bool
	}{
		{
			"gets correct port and nonce",
			123456,
			[16]byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '1', '2', '3', '4', '5', '6', '7'},
			[]byte("123456"),
			nil,
			nil,
			false,
		},
		{
			"error reading bytes returns error",
			0,
			[16]byte{},
			[]byte("0"),
			errors.New("I can't read!?"),
			nil,
			true,
		},
		{
			"error parsing port returns error",
			0,
			[16]byte{},
			[]byte("not-a-port"),
			nil,
			nil,
			true,
		},
		{
			"error getting nonce returns error",
			123456,
			[16]byte{},
			[]byte("123456"),
			nil,
			errors.New("LeVar Burton help me"),
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBufferedReader := mocks.NewMockBufferedReader(ctrl)

			if tt.readError == nil {
				mockBufferedReader.EXPECT().ReadBytes(byte('\n')).Return([]byte(tt.bytesToRead), nil).MaxTimes(1)
			} else {
				mockBufferedReader.EXPECT().ReadBytes(byte('\n')).Return(nil, tt.readError).MaxTimes(1)
			}

			if tt.readBytesError == nil {
				mockBufferedReader.EXPECT().Read(gomock.Any()).DoAndReturn(func(p []byte) (int, error) {
					copy(p, tt.wantNonce[:])

					return len(tt.wantNonce), nil
				}).MaxTimes(1)
			} else {
				mockBufferedReader.EXPECT().Read(gomock.Any()).Return(0, tt.readBytesError).MaxTimes(1)
			}

			port, nonce, err := getPortAndNonce(mockBufferedReader)

			if (err != nil) != tt.wantErr {
				t.Errorf("getPortAndNonce() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else if err != nil {
				return
			}
			if port != tt.wantPort {
				t.Errorf("getPortAndNonce() port = %v, wantPort %v", port, tt.wantPort)
			}
			if nonce != tt.wantNonce {
				t.Errorf("getPortAndNonce() nonce = %v, wantNonce %v", nonce, tt.wantNonce)
			}
		})
	}
}

func TestMustFprintf(t *testing.T) {
	ctrl := gomock.NewController(t)

	type args struct {
		format string
		a      []any
	}
	tests := []struct {
		name        string
		args        args
		wantW       string
		shouldPanic bool
	}{
		{
			"Fprints successfully",
			args{
				format: "format string",
			},
			"format string",
			false,
		},
		{
			"panics",
			args{
				format: "",
			},
			"",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic == false {
				w := &bytes.Buffer{}
				MustFprintf(w, tt.args.format, tt.args.a...)
				if gotW := w.String(); gotW != tt.wantW {
					t.Errorf("MustFprintf() = %v, want %v", gotW, tt.wantW)
				}
			} else {
				defer func() { _ = recover() }()
				mockWriter := mocks.NewMockMockWriter(ctrl)
				mockWriter.EXPECT().Write(gomock.Any()).Return(0, errors.New("Oh no")).AnyTimes()

				MustFprintf(mockWriter, "")
				t.Errorf("should have panicked")
			}
		})
	}
}

func TestGPGHandler_authenticate(t *testing.T) {
	tests := []struct {
		name    string
		handler *GPGHandler
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.handler.authenticate(); (err != nil) != tt.wantErr {
				t.Errorf("GPGHandler.authenticate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
