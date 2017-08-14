package libvirt

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	libvirt "github.com/libvirt/libvirt-go"
	"github.com/libvirt/libvirt-go-xml"
)

// network transparent image
type image interface {
	Size() (uint64, error)
	Import(func(io.Reader) error, libvirtxml.StorageVolume) error
	String() string
}

type localImage struct {
	path string
}

func (i *localImage) String() string {
	return i.path
}

func (i *localImage) Size() (uint64, error) {
	file, err := os.Open(i.path)
	if err != nil {
		return 0, err
	}

	fi, err := file.Stat()
	if err != nil {
		return 0, err
	}
	return uint64(fi.Size()), nil
}

func (i *localImage) Import(copier func(io.Reader) error, vol libvirtxml.StorageVolume) error {
	file, err := os.Open(i.path)
	defer file.Close()
	if err != nil {
		return fmt.Errorf("Error while opening %s: %s", i.path, err)
	}

	if fi, err := file.Stat(); err != nil {
		return err
	} else {
		// we can skip the upload if the modification times are the same
		if vol.Target.Timestamps != nil && vol.Target.Timestamps.Mtime != "" {
			if fi.ModTime() == timeFromEpoch(vol.Target.Timestamps.Mtime) {
				log.Printf("Modification time is the same: skipping image copy")
				return nil
			}
		}
	}

	return copier(file)
}

type httpImage struct {
	url *url.URL
}

func (i *httpImage) String() string {
	return i.url.String()
}

func (i *httpImage) Size() (uint64, error) {
	response, err := http.Head(i.url.String())
	if err != nil {
		return 0, err
	}
	if response.StatusCode != 200 {
		return 0,
			fmt.Errorf(
				"Error accessing remote resource: %s - %s",
				i.url.String(),
				response.Status)
	}

	length, err := strconv.Atoi(response.Header.Get("Content-Length"))
	if err != nil {
		err = fmt.Errorf(
			"Error while getting Content-Length of \"%s\": %s - got %s",
			i.url.String(),
			err,
			response.Header.Get("Content-Length"))
		return 0, err
	}
	return uint64(length), nil
}

func (i *httpImage) Import(copier func(io.Reader) error, vol libvirtxml.StorageVolume) error {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", i.url.String(), nil)

	if vol.Target.Timestamps != nil && vol.Target.Timestamps.Mtime != "" {
		req.Header.Set("If-Modified-Since", timeFromEpoch(vol.Target.Timestamps.Mtime).UTC().Format(http.TimeFormat))
	}
	response, err := client.Do(req)
	defer response.Body.Close()

	if err != nil {
		return fmt.Errorf("Error while downloading %s: %s", i.url.String(), err)
	}
	if response.StatusCode == http.StatusNotModified {
		return nil
	}

	return copier(response.Body)
}

func newImage(source string) (image, error) {
	url, err := url.Parse(source)
	if err != nil {
		return nil, fmt.Errorf("Can't parse source '%s' as url: %s", source, err)
	}

	if strings.HasPrefix(url.Scheme, "http") {
		return &httpImage{url: url}, nil
	} else if url.Scheme == "file" || url.Scheme == "" {
		return &localImage{path: url.Path}, nil
	} else {
		return nil, fmt.Errorf("Don't know how to read from '%s': %s", url.String(), err)
	}
}

func newCopier(virConn *libvirt.Connect, volume *libvirt.StorageVol, size uint64) func(src io.Reader) error {
	copier := func(src io.Reader) error {
		var bytesCopied int64

		stream, err := virConn.NewStream(0)
		if err != nil {
			return err
		}

		defer func() {
			if uint64(bytesCopied) != size {
				stream.Abort()
			} else {
				stream.Finish()
			}
			stream.Free()
		}()

		volume.Upload(stream, 0, size, 0)

		sio := NewStreamIO(*stream)

		bytesCopied, err = io.Copy(sio, src)
		if err != nil {
			return err
		}
		log.Printf("%d bytes uploaded\n", bytesCopied)
		return nil
	}
	return copier
}

func timeFromEpoch(str string) time.Time {
	var s, ns int

	ts := strings.Split(str, ".")
	if len(ts) == 2 {
		ns, _ = strconv.Atoi(ts[1])
	}
	s, _ = strconv.Atoi(ts[0])

	return time.Unix(int64(s), int64(ns))
}
