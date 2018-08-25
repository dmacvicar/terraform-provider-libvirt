package libvirt

import (
	"fmt"
	"io"
	"io/ioutil"
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

	url, _, err := fws.AddFile(content)
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

func TestRemoteImageDownload(t *testing.T) {
	content := []byte("this is a qcow image... well, it is not")
	fws := fileWebServer{}
	if err := fws.Start(); err != nil {
		t.Fatal(err)
	}
	defer fws.Stop()

	url, tmpfile, err := fws.AddFile(content)
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

func TestTimeFromEpoch(t *testing.T) {
	if ts := timeFromEpoch(""); ts.UnixNano() > 0 {
		t.Fatalf("expected timestamp '0.0', got %v.%v", ts.Unix(), ts.Nanosecond())
	}

	if ts := timeFromEpoch("abc"); ts.UnixNano() > 0 {
		t.Fatalf("expected timestamp '0.0', got %v.%v", ts.Unix(), ts.Nanosecond())
	}

	if ts := timeFromEpoch("123"); ts.UnixNano() != time.Unix(123, 0).UnixNano() {
		t.Fatalf("expected timestamp '123.0', got %v.%v", ts.Unix(), ts.Nanosecond())
	}

	if ts := timeFromEpoch("123.456"); ts.UnixNano() != time.Unix(123, 456).UnixNano() {
		t.Fatalf("expected timestamp '123.456', got %v.%v", ts.Unix(), ts.Nanosecond())
	}
}
