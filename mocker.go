package falso

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
)

const (
	PROXY = "proxy"
	MOCK  = "mock"
)

type Dialer interface {
	Dial(network, address string) (Connection, error)
}

type dialer struct{}

func NewDialer() Dialer {
	return &dialer{}
}

func (d *dialer) Dial(network, address string) (Connection, error) {
	return net.Dial(network, address)
}

type Connection interface {
	Read(b []byte) (n int, err error)
	Write(b []byte) (n int, err error)
	Close() error
}

type Mocker interface {
	HandleRequest(c Connection)
}

type mocker struct {
	dialer        Dialer
	bufferSize    uint
	remoteAddress string
	mode          string
	dataPath      string
	overwrite     bool
}

func NewMocker(
	dialer Dialer,
	mode string,
	remoteAddress string,
	dataPath string,
	bufferSize uint,
	overwrite bool,
) Mocker {
	return &mocker{
		dialer:        dialer,
		mode:          mode,
		remoteAddress: remoteAddress,
		dataPath:      dataPath,
		bufferSize:    bufferSize,
		overwrite:     overwrite,
	}
}

func (m *mocker) HandleRequest(c Connection) {
	defer func() {
		_ = c.Close()
	}()

	// Make a buffer to hold incoming data.
	buf := make([]byte, m.bufferSize)

	// Read the incoming connection into the buffer.
	_, err := c.Read(buf)
	if err != nil {
		log.Panicf("failed to read message from connection: %v\n", err.Error())
	}

	// SHA1 is enough for this use case
	hash := CreateHash(buf)
	path := GetFilePath(m.dataPath, hash)

	var replyFromServer []byte

	_, err = os.Stat(path)
	if m.mode == PROXY && (m.overwrite || errors.Is(err, os.ErrNotExist)) {
		replyFromServer = m.handleRequestToRemote(buf)
		WriteFile(path, replyFromServer)
	} else if m.mode == MOCK || !m.overwrite {
		replyFromServer = ReadFile(path)
	} else {
		log.Panicf("Unexpected mode")
	}

	// Send a response back to the original request.
	_, err = c.Write(replyFromServer)
	if err != nil {
		log.Panicf("failed to write message to connection: %v\n", err.Error())
	}
}

func (m *mocker) handleRequestToRemote(b []byte) []byte {
	if m.remoteAddress == "" {
		log.Panic("remote address flag is empty, please specify a remote address")
	}

	conn, err := m.dialer.Dial("tcp", m.remoteAddress)
	if err != nil {
		log.Panicf("failed to dial remote address: %v", err.Error())
	}
	defer func() {
		err := conn.Close()
		if err != nil {
			log.Printf("failed to close connection: %v\n", err.Error())
		}
	}()

	_, err = conn.Write(b)
	if err != nil {
		log.Panicf("failed to write to remote addres: %v", err.Error())
	}

	reply := make([]byte, m.bufferSize)
	_, err = conn.Read(reply)
	if err != nil {
		log.Panicf("failed to read from remote address: %v", err.Error())
	}

	return reply
}

func CreateHash(b []byte) string {
	h := sha1.New()
	h.Write(b)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func GetFilePath(path, filename string) string {
	return filepath.Join(path, filename)
}

func WriteFile(path string, b []byte) {
	err := os.WriteFile(path, b, 0644)
	if err != nil {
		log.Panicf("failed to write file: %v", err)
	}
}

func ReadFile(path string) []byte {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Panicf(
			"failed to read file, this probably means that there is no response recorded for your request, "+
				"more info:\n %v", err)
	}
	return data
}
