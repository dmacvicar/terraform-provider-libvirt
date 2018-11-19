package libvirt

import (
	"fmt"
	"io"
	"io/ioutil"
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
	IsQCOW2() (bool, error)
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

func (i *localImage) IsQCOW2() (bool, error) {
	file, err := os.Open(i.path)
	defer file.Close()
	if err != nil {
		return false, fmt.Errorf("Error while opening %s: %s", i.path, err)
	}
	buf := make([]byte, 8)
	_, err = io.ReadAtLeast(file, buf, 8)
	if err != nil {
		return false, err
	}
	return isQCOW2Header(buf)
}

func (i *localImage) Import(copier func(io.Reader) error, vol libvirtxml.StorageVolume) error {
	file, err := os.Open(i.path)
	defer file.Close()
	if err != nil {
		return fmt.Errorf("Error while opening %s: %s", i.path, err)
	}

	fi, err := file.Stat()
	if err != nil {
		return err
	}
	// we can skip the upload if the modification times are the same
	if vol.Target.Timestamps != nil && vol.Target.Timestamps.Mtime != "" {
		if fi.ModTime() == timeFromEpoch(vol.Target.Timestamps.Mtime) {
			log.Printf("Modification time is the same: skipping image copy")
			return nil
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
	if response.StatusCode == 403 {
		// possibly only the HEAD method is forbidden, try a Body-less GET instead
		response, err = http.Get(i.url.String())
		if err != nil {
			return 0, err
		}

		response.Body.Close()
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

func (i *httpImage) IsQCOW2() (bool, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", i.url.String(), nil)
	req.Header.Set("Range", "bytes=0-7")
	response, err := client.Do(req)

	if err != nil {
		return false, err
	}
	defer response.Body.Close()

	if response.StatusCode != 206 {
		return false, fmt.Errorf(
			"Can't retrieve partial header of resource to determine file type: %s - %s",
			i.url.String(),
			response.Status)
	}

	header, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return false, err
	}

	if len(header) < 8 {
		return false, fmt.Errorf(
			"Can't retrieve read header of resource to determine file type: %s - %d bytes read",
			i.url.String(),
			len(header))
	}

	return isQCOW2Header(header)
}

// Whether the buffer is the header of a qcow2 file
func isQCOW2Header(buf []byte) (bool, error) {
	if len(buf) < 8 {
		return false, fmt.Errorf("Expected header of 8 bytes. Got %d", len(buf))
	}
	if buf[0] == 'Q' && buf[1] == 'F' && buf[2] == 'I' && buf[3] == 0xfb && buf[4] == 0x00 && buf[5] == 0x00 && buf[6] == 0x00 && buf[7] == 0x03 {
		return true, nil
	}
	return false, nil
}

func (i *httpImage) Import(copier func(io.Reader) error, vol libvirtxml.StorageVolume) error {
	// number of download retries on non client errors (eg. 5xx)
	const maxHTTPRetries int = 3
	// wait time between retries
	const retryWait time.Duration = 2 * time.Second

	client := &http.Client{}
	req, err := http.NewRequest("GET", i.url.String(), nil)

	if err != nil {
		log.Printf("[DEBUG:] Error creating new request for source url %s: %s", i.url.String(), err)
		return fmt.Errorf("Error while downloading %s: %s", i.url.String(), err)
	}

	if vol.Target.Timestamps != nil && vol.Target.Timestamps.Mtime != "" {
		req.Header.Set("If-Modified-Since", timeFromEpoch(vol.Target.Timestamps.Mtime).UTC().Format(http.TimeFormat))
	}

	var response *http.Response
	for retryCount := 0; retryCount < maxHTTPRetries; retryCount++ {
		response, err = client.Do(req)
		if err != nil {
			return fmt.Errorf("Error while downloading %s: %v", i.url.String(), err)
		}
		defer response.Body.Close()

		log.Printf("[DEBUG]: url resp status code %s (retry #%d)\n", response.Status, retryCount)
		if response.StatusCode == http.StatusNotModified {
			return nil
		} else if response.StatusCode == http.StatusOK {
			return copier(response.Body)
		} else if response.StatusCode < 500 {
			break
		} else {
			// The problem is not client but server side
			// retry a few times after a small wait
			if retryCount < maxHTTPRetries {
				time.Sleep(retryWait)
			}
		}
	}
	return fmt.Errorf("Error while downloading %s: %v", i.url.String(), response)
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

		if err := volume.Upload(stream, 0, size, 0); err != nil {
			stream.Abort()
			return fmt.Errorf("Error while uploading volume %s", err)
		}

		sio := NewStreamIO(*stream)

		bytesCopied, err = io.Copy(sio, src)
		// if we get unexpected EOF this mean that connection was closed suddently from server side
		// the problem is not on the plugin but on server hosting currupted images
		if err == io.ErrUnexpectedEOF {
			return fmt.Errorf("Error: transfer was unexpectedly closed from the server while downloading. Please try again later or check the server hosting sources")
		}
		if err != nil {
			return fmt.Errorf("Error while copying source to volume %s", err)
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

// removeVolume removes the volume identified by `key` from libvirt
func removeVolume(client *Client, key string) error {
	volume, err := client.libvirt.LookupStorageVolByKey(key)
	if err != nil {
		return fmt.Errorf("Can't retrieve volume %s: %v", key, err)
	}
	defer volume.Free()

	// Refresh the pool of the volume so that libvirt knows it is
	// not longer in use.
	volPool, err := volume.LookupPoolByVolume()
	if err != nil {
		return fmt.Errorf("Error retrieving pool for volume: %s", err)
	}
	defer volPool.Free()

	poolName, err := volPool.GetName()
	if err != nil {
		return fmt.Errorf("Error retrieving name of volume: %s", err)
	}

	client.poolMutexKV.Lock(poolName)
	defer client.poolMutexKV.Unlock(poolName)

	waitForSuccess("Error refreshing pool for volume", func() error {
		return volPool.Refresh(0)
	})

	// Workaround for redhat#1293804
	// https://bugzilla.redhat.com/show_bug.cgi?id=1293804#c12
	// Does not solve the problem but it makes it happen less often.
	_, err = volume.GetXMLDesc(0)
	if err != nil {
		return fmt.Errorf("Can't retrieve volume %s XML desc: %s", key, err)
	}

	err = volume.Delete(0)
	if err != nil {
		return fmt.Errorf("Can't delete volume %s: %s", key, err)
	}

	return nil
}

// tries really hard to find volume with `key`
// it will try to start the pool if it does not find it
//
// You have to call volume.Free() on the returned volume
func lookupVolumeReallyHard(client *Client, volPoolName string, key string) (*libvirt.StorageVol, error) {
	virConn := client.libvirt
	if virConn == nil {
		return nil, fmt.Errorf(LibVirtConIsNil)
	}

	volume, err := virConn.LookupStorageVolByKey(key)
	if err != nil {
		virErr := err.(libvirt.Error)
		if virErr.Code != libvirt.ERR_NO_STORAGE_VOL {
			return nil, fmt.Errorf("Can't retrieve volume %s", key)
		}
		log.Printf("[INFO] Volume %s not found, attempting to start its pool", key)

		volPool, err := virConn.LookupStoragePoolByName(volPoolName)
		if err != nil {
			return nil, fmt.Errorf("Error retrieving pool %s for volume %s: %s", volPoolName, key, err)
		}
		defer volPool.Free()

		active, err := volPool.IsActive()
		if err != nil {
			return nil, fmt.Errorf("error retrieving status of pool %s for volume %s: %s", volPoolName, key, err)
		}
		if active {
			log.Printf("Can't retrieve volume %s (and pool is active)", key)
			return nil, nil
		}

		err = volPool.Create(0)
		if err != nil {
			return nil, fmt.Errorf("error starting pool %s: %s", volPoolName, err)
		}

		// attempt a new lookup
		volume, err = virConn.LookupStorageVolByKey(key)
		if err != nil {
			virErr := err.(libvirt.Error)
			if virErr.Code != libvirt.ERR_NO_STORAGE_VOL {
				return nil, fmt.Errorf("Can't retrieve volume %s", key)
			}
			// does not exist, but no error
			return nil, nil
		}
	}
	return volume, nil
}
