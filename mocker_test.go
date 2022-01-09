package falso

import (
	"bytes"
	"testing"
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

func (c *mockedConnection) Close() error {
	return nil
}

func TestHandleRequest(t *testing.T) {
	t.Run("Proxy", func(t *testing.T) {
		fromClientData := []byte("sorin")
		fromRemoteToLocalData := []byte("sorin who ?")

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

		mocker := NewMocker(&dialer, "proxy", "localhost:0", "/tmp", uint(len(fromRemoteToLocalData)))
		mocker.HandleRequest(&localConn)
		if bytes.Compare(fromRemoteToLocalData, localConn.writeData) != 0 {
			t.Errorf("proxy expected %s but received %s", fromRemoteToLocalData, localConn.writeData)
		}

		mocker = NewMocker(&dialer, "proxy", "localhost:0", "/tmp", 1)
		mocker.HandleRequest(&localConn)
		if bytes.Compare(fromRemoteToLocalData, localConn.writeData) == 0 {
			t.Errorf("proxy expected response to not be equal but received %s", fromRemoteToLocalData)
		}
	})

	t.Run("Mock", func(t *testing.T) {
		fromClientData := []byte("sorin")
		fromRemoteToLocalData := []byte("sorin who ?")
		dataPath := "/tmp"

		// We assume the file is already created
		WriteFile(dataPath, CreateHash(fromClientData), fromRemoteToLocalData)

		localConn := mockedConnection{
			readData: fromClientData,
		}
		dialer := mockedDialer{
			local: &localConn,
		}

		mocker := NewMocker(&dialer, "mock", "", dataPath, uint(len(fromRemoteToLocalData)))
		mocker.HandleRequest(&localConn)
		if bytes.Compare(fromRemoteToLocalData, localConn.writeData) != 0 {
			t.Errorf("mock expected %s but received %s", fromRemoteToLocalData, localConn.writeData)
		}

		localConn.readData = []byte("undefined")
		mocker = NewMocker(&dialer, "mock", "", dataPath, uint(len(fromRemoteToLocalData)))
		defer func() {
			if r := recover(); r == nil {
				t.Error("mock expected code to panic when there is no file for a request")
			}
		}()
		mocker.HandleRequest(&localConn)
	})

}
