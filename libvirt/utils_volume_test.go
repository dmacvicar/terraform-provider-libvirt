package libvirt

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"
)

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
	modTime := UnixTimestamp{tmpfileStat.ModTime()}
	vol.Target.Timestamps = &defTimestamps{
		Modification: &modTime,
	}

	copier := func(r io.Reader) error {
		t.Fatalf("ERROR: starting copy of %s... but the file is the same!", url)
		return nil
	}
	if err = image.Import(copier, vol); err != nil {
		t.Fatal("Could not copy image from %s: %v", url, err)
	}
	t.Log("File not copied because modification time was the same")
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
	modTime := UnixTimestamp{tmpfileStat.ModTime()}
	vol.Target.Timestamps = &defTimestamps{
		Modification: &modTime,
	}
	copier := func(r io.Reader) error {
		t.Fatalf("ERROR: starting copy of %s... but the file is the same!", url)
		return nil
	}
	if err = image.Import(copier, vol); err != nil {
		t.Fatal("Could not copy image from %s: %v", url, err)
	}
	t.Log("File not copied because modification time was the same")

}
