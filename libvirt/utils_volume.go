package libvirt

import (
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	libvirt "github.com/digitalocean/go-libvirt"
)

func newCopier(virConn *libvirt.Libvirt, volume *libvirt.StorageVol, size uint64) func(src io.Reader) error {
	copier := func(src io.Reader) error {
		r, w := io.Pipe()
		defer w.Close()

		go func() error {
			buffer := make([]byte, 4*1024*1024)
			bytesCopied, err := io.CopyBuffer(w, src, buffer)

			// if we get unexpected EOF this mean that connection was closed suddently from server side
			// the problem is not on the plugin but on server hosting currupted images
			if err == io.ErrUnexpectedEOF {
				return fmt.Errorf("error: transfer was unexpectedly closed from the server while downloading. Please try again later or check the server hosting sources")
			}
			if err != nil {
				return fmt.Errorf("error while copying source to volume %s", err)
			}

			log.Printf("%d bytes uploaded\n", bytesCopied)
			if uint64(bytesCopied) != size {
				return fmt.Errorf("error during volume Upload. BytesCopied: %d != %d volume.size", bytesCopied, size)
			}

			return w.Close()

		}()

		/*
			* FIXME: use alternate simpler implementation without pipe?
				if err := virConn.StorageVolUpload(*volume, src, 0, size, 0); err != nil {
					return fmt.Errorf("Error while uploading volume %s", err)
				}
		*/

		if err := virConn.StorageVolUpload(*volume, r, 0, size, 0); err != nil {
			return fmt.Errorf("error while uploading volume %s", err)
		}

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
