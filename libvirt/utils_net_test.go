package libvirt

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestRandomMACAddress(t *testing.T) {
	mac, err := randomMACAddress()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_, err = net.ParseMAC(mac)

	if err != nil {
		t.Errorf("Invalid MAC address generated: %s - %v", mac, err)
	}
}

type fileWebServer struct {
	t      *testing.T
	server *httptest.Server
	Dir    string
}

func newFileWebServer(t *testing.T) *fileWebServer {
	dir := t.TempDir()
	handler := http.NewServeMux()
	handler.Handle("/", http.FileServer(http.Dir(dir)))

	return &fileWebServer{
		t:      t,
		server: httptest.NewUnstartedServer(handler),
		Dir:    dir,
	}
}

func (fws *fileWebServer) Start() {
	fws.server.Start()
}

// Adds a file (with some content) in the directory served by the fileWebServer.
func (fws *fileWebServer) AddContent(content []byte) (string, *os.File, error) {
	tmpfile, err := os.CreateTemp(fws.Dir, "file-")
	if err != nil {
		return "", nil, err
	}

	if len(content) > 0 {
		if _, err := tmpfile.Write(content); err != nil {
			return "", nil, err
		}
	}

	return fmt.Sprintf("%s/%s", fws.server.URL, filepath.ToSlash(filepath.Base(tmpfile.Name()))), tmpfile, nil
}

// Symlinks a file into the directory server by the webserver.
func (fws *fileWebServer) AddFile(filePath string) (string, error) {
	err := os.Symlink(filePath, filepath.Join(fws.Dir, filepath.Base(filePath)))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s", fws.server.URL, filepath.ToSlash(filepath.Base(filePath))), nil
}

func (fws *fileWebServer) Close() {
	fws.server.Close()
}
