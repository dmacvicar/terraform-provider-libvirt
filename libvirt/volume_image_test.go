package libvirt

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/libvirt/libvirt-go-xml"
)

func TestLocalImageDetermineType(t *testing.T) {
	abspath, err := filepath.Abs("testdata/test.qcow2")
	if err != nil {
		t.Fatal(err)
	}

	url := fmt.Sprintf("file://%s", abspath)
	image, err := newImage(url)
	if err != nil {
		t.Errorf("Could not create local image: %v", err)
	}

	qcow2, err := image.IsQCOW2()
	if err != nil {
		t.Errorf("Can't determine image type: %v", err)
	}
	if !qcow2 {
		t.Errorf("Expected image to be recognized as QCOW2")
	}
}

func TestLocalImageDownload(t *testing.T) {
	content := []byte("this is a qcow image... well, it is not")
	tmpfile, err := ioutil.TempFile("", "test-image-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	t.Logf("Adding some content to %s", tmpfile.Name())
	if _, err := tmpfile.Write(content); err != nil {
		t.Fatal(err)
	}
	tmpfileStat, err := tmpfile.Stat()
	if err != nil {
		t.Fatal(err)
	}
	url := fmt.Sprintf("file://%s", tmpfile.Name())
	image, err := newImage(url)
	if err != nil {
		t.Errorf("Could not create local image: %v", err)
	}

	t.Logf("Importing %s", url)
	vol := newDefVolume()
	vol.Target.Timestamps = &libvirtxml.StorageVolumeTargetTimestamps{
		Mtime: fmt.Sprintf("%d.%d", tmpfileStat.ModTime().Unix(), tmpfileStat.ModTime().Nanosecond()),
	}

	copier := func(r io.Reader) error {
		t.Fatalf("ERROR: starting copy of %s... but the file is the same!", url)
		return nil
	}
	if err = image.Import(copier, vol); err != nil {
		t.Fatalf("Could not copy image from %s: %v", url, err)
	}
	t.Log("File not copied because modification time was the same")
}

func TestRemoteImageDetermineType(t *testing.T) {
	content, err := ioutil.ReadFile("testdata/test.qcow2")
	if err != nil {
		t.Fatal(err)
	}

	fws := fileWebServer{}
	if err := fws.Start(); err != nil {
		t.Fatal(err)
	}
	defer fws.Stop()

	url, _, err := fws.AddContent(content)
	if err != nil {
		t.Fatal(err)
	}

	image, err := newImage(url)
	if err != nil {
		t.Errorf("Could not create local image: %v", err)
	}

	qcow2, err := image.IsQCOW2()
	if err != nil {
		t.Errorf("Can't determine image type: %v", err)
	}
	if !qcow2 {
		t.Errorf("Expected image to be recognized as QCOW2")
	}
}

func TestRemoteImageDownloadRetry(t *testing.T) {
	content := []byte("this is a qcow image... well, it is not")

	// returns a server that returns every error from
	// errorList before returning a valid response
	newErrorServer := func(errorList []int) *httptest.Server {
		errorCount := 0
		return httptest.NewServer(
			http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					if errorCount < len(errorList) {
						t.Logf("Server serving retry %d", errorCount)
						http.Error(w, fmt.Sprintf("Error %d", errorCount), errorList[errorCount])
						errorCount = errorCount + 1
					} else {
						t.Logf("Server: success (after %d errors)", errorCount)
						http.ServeContent(w, r, "content", time.Now(), bytes.NewReader(content))
					}
				}))
	}

	copier := func(r io.Reader) error {
		_, err := ioutil.ReadAll(r)
		return err
	}

	server := newErrorServer([]int{503, 503})
	defer server.Close()
	vol := newDefVolume()
	image, err := newImage(server.URL)
	if err != nil {
		t.Errorf("Could not create image object: %v", err)
	}
	start := time.Now()
	if err = image.Import(copier, vol); err != nil {
		t.Fatalf("Expected to retry: %v", err)
	}
	if time.Since(start).Seconds() < 4 {
		t.Fatalf("Expected to retry at least 2 times x 2 seconds")
	}

	server = newErrorServer([]int{503, 404})
	defer server.Close()
	vol = newDefVolume()
	start = time.Now()
	image, err = newImage(server.URL)
	if err != nil {
		t.Errorf("Could not create image object: %v", err)
	}
	if err = image.Import(copier, vol); err == nil {
		t.Fatalf("Expected %s to fail with status 4xx", server.URL)
	}
	if time.Since(start).Seconds() < 2 {
		t.Fatalf("Expected to retry at least 1 times x 2 seconds")
	}

}

func TestRemoteImageDownload(t *testing.T) {
	content := []byte("this is a qcow image... well, it is not")
	fws := fileWebServer{}
	if err := fws.Start(); err != nil {
		t.Fatal(err)
	}
	defer fws.Stop()

	url, tmpfile, err := fws.AddContent(content)
	if err != nil {
		t.Fatal(err)
	}

	tmpfileStat, err := tmpfile.Stat()
	if err != nil {
		t.Fatal(err)
	}
	image, err := newImage(url)
	if err != nil {
		t.Errorf("Could not create local image: %v", err)
	}

	t.Logf("Importing %s", url)
	vol := newDefVolume()
	vol.Target.Timestamps = &libvirtxml.StorageVolumeTargetTimestamps{
		Mtime: fmt.Sprintf("%d.%d", tmpfileStat.ModTime().Unix(), tmpfileStat.ModTime().Nanosecond()),
	}
	copier := func(r io.Reader) error {
		t.Fatalf("ERROR: starting copy of %s... but the file is the same!", url)
		return nil
	}
	if err = image.Import(copier, vol); err != nil {
		t.Fatalf("Could not copy image from %s: %v", url, err)
	}
	t.Log("File not copied because modification time was the same")

}
