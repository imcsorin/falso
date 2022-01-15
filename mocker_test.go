package falso

import (
	"bytes"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type mockedDialer struct {
	local  Connection
	remote Connection
}

func (m *mockedDialer) Dial(_, _ string) (Connection, error) {
	return m.remote, nil
}

type mockedConnection struct {
	readData  []byte
	writeData []byte
}

func (c *mockedConnection) Read(b []byte) (n int, err error) {
	return bytes.NewReader(c.readData).Read(b)
}

func (c *mockedConnection) Write(b []byte) (n int, err error) {
	c.writeData = b
	return 0, nil
}

var seededRand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

func randString(n int) string {
	letterBytes := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[seededRand.Intn(len(letterBytes))]
	}
	return string(b)
}

func (c *mockedConnection) Close() error {
	return nil
}

func TestHandleRequest(t *testing.T) {
	t.Run("Proxy", func(t *testing.T) {
		fromRemoteToLocalData := []byte("sorin who ?")
		fromClientData := make([]byte, len(fromRemoteToLocalData))
		tmpPath := filepath.Join("/tmp", randString(10))
		_ = os.Mkdir(tmpPath, 0755)

		localConn := mockedConnection{
			readData: fromClientData,
		}
		remoteConn := mockedConnection{
			readData: fromRemoteToLocalData,
		}
		dialer := mockedDialer{
			local:  &localConn,
			remote: &remoteConn,
		}

		mocker := NewMocker(&dialer, "proxy", "localhost:0", tmpPath, uint(len(fromRemoteToLocalData)), false)
		mocker.HandleRequest(&localConn)
		if bytes.Compare(fromRemoteToLocalData, localConn.writeData) != 0 {
			t.Errorf("proxy expected %s but received %s", fromRemoteToLocalData, localConn.writeData)
		}

		mocker = NewMocker(&dialer, "proxy", "localhost:0", tmpPath, 1, true)
		mocker.HandleRequest(&localConn)
		if bytes.Compare(fromRemoteToLocalData, localConn.writeData) == 0 {
			t.Errorf("proxy expected response to not be equal but received %s", fromRemoteToLocalData)
		}

		fromRemoteToLocalData = []byte("undefined")
		remoteConn = mockedConnection{
			readData: fromRemoteToLocalData,
		}
		dialer = mockedDialer{
			local:  &localConn,
			remote: &remoteConn,
		}
		mocker = NewMocker(&dialer, "proxy", "localhost:0", tmpPath, uint(len(fromRemoteToLocalData)), false)
		mocker.HandleRequest(&localConn)
		if bytes.Compare(fromRemoteToLocalData, localConn.writeData) != 0 {
			t.Errorf("proxy expected response to not be equal but received %s", fromRemoteToLocalData)
		}
	})

	t.Run("Mock", func(t *testing.T) {
		fromRemoteToLocalData := []byte("sorin who ?")
		fromClientData := make([]byte, len(fromRemoteToLocalData))
		tmpPath := filepath.Join("/tmp", randString(10))
		_ = os.Mkdir(tmpPath, 0755)

		// We assume the file is already created
		WriteFile(GetFilePath(tmpPath, CreateHash(fromClientData)), fromRemoteToLocalData)

		localConn := mockedConnection{
			readData: fromClientData,
		}
		dialer := mockedDialer{
			local: &localConn,
		}

		mocker := NewMocker(&dialer, "mock", "", tmpPath, uint(len(fromRemoteToLocalData)), false)
		mocker.HandleRequest(&localConn)
		if bytes.Compare(fromRemoteToLocalData, localConn.writeData) != 0 {
			t.Errorf("mock expected %s but received %s", fromRemoteToLocalData, localConn.writeData)
		}

		localConn.readData = []byte("undefined")
		mocker = NewMocker(&dialer, "mock", "", tmpPath, uint(len(fromRemoteToLocalData)), false)
		defer func() {
			if r := recover(); r == nil {
				t.Error("mock expected code to panic when there is no file for a request")
			}
		}()
		mocker.HandleRequest(&localConn)
	})

}
