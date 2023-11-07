package libvirt

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"libvirt.org/go/libvirtxml"
)

func TestNewImage(t *testing.T) {

	fixtures := []struct {
		Name    string
		Size    uint64
		IsQCOW2 bool
	}{
		{"test.qcow2", 196616, true},
		{"tcl.iso", 16834560, false},
	}

	testdata, err := filepath.Abs("testdata")
	if err != nil {
		t.Fatal(err)
	}

	fws := newFileWebServer(t)
	fws.Start()
	defer fws.Close()

	for _, fixture := range fixtures {
		localPath := filepath.Join(testdata, fixture.Name)

		var fileURLStr string
		if runtime.GOOS == "windows" {
			fileURLStr = "file:///" + localPath
		} else {
			fileURLStr = "file://" + localPath
		}

		httpURLStr, err := fws.AddFile(localPath)
		if err != nil {
			t.Fatal(err)
		}

		httpURL, err := url.Parse(httpURLStr)
		if err != nil {
			t.Fatal(err)
		}

		results := []struct {
			Source   string
			Image    image
			AsString string
		}{
			{localPath, &localImage{path: localPath}, localPath},
			{fileURLStr, &localImage{path: localPath}, localPath},
			{httpURLStr, &httpImage{url: httpURL}, httpURLStr},
		}

		for _, ex := range results {
			img, err := newImage(ex.Source)
			if err != nil {
				t.Error(err)
				continue
			}
			assert.Equal(t, ex.Image, img)
			assert.Equal(t, ex.AsString, img.String(), ex.Source)
			isQCOW2, err := img.IsQCOW2()
			if err != nil {
				t.Error(err)
				continue
			}
			assert.Equal(t, fixture.IsQCOW2, isQCOW2)

			size, err := img.Size()
			if err != nil {
				t.Error(err)
				continue
			}
			assert.Equal(t, fixture.Size, size)
		}
	}
}

func TestLocalImageDownload(t *testing.T) {
	content := []byte("this is a qcow image... well, it is not")
	tmpfile, err := os.CreateTemp(t.TempDir(), "test-image-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	t.Logf("Adding some content to %s", tmpfile.Name())
	if _, err := tmpfile.Write(content); err != nil {
		t.Fatal(err)
	}
	defer tmpfile.Close()

	tmpfileStat, err := tmpfile.Stat()
	if err != nil {
		t.Fatal(err)
	}

	url := "file:///" + tmpfile.Name()
	image, err := newImage(url)
	if err != nil {
		t.Fatalf("Could not create local image: %v", err)
	}

	t.Logf("Importing %s", tmpfile.Name())
	vol := newDefVolume()
	vol.Target.Timestamps = &libvirtxml.StorageVolumeTargetTimestamps{
		Mtime: fmt.Sprintf("%d.%d", tmpfileStat.ModTime().Unix(), tmpfileStat.ModTime().Nanosecond()),
	}

	copier := func(r io.Reader) error {
		require.FailNow(t, fmt.Sprintf("This should not be run, as image has not changed. url: %s", url))
		return nil
	}

	if err = image.Import(copier, vol); err != nil {
		require.NoError(t, err, "As the image was not modified and not copied, no error was expected. url: %s", tmpfile.Name())
	}

	t.Log("As expected, image not copied because modification time was the same")
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
						errorCount++
					} else {
						t.Logf("Server: success (after %d errors)", errorCount)
						http.ServeContent(w, r, "content", time.Now(), bytes.NewReader(content))
					}
				}))
	}

	copier := func(r io.Reader) error {
		_, err := io.ReadAll(r)
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

	fws := newFileWebServer(t)
	fws.Start()
	defer fws.Close()

	url, tmpfile, err := fws.AddContent(content)
	if err != nil {
		t.Fatal(err)
	}
	defer tmpfile.Close()

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


