package libvirt

import (
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/libvirt/libvirt-go"
)

func newCopier(virConn *libvirt.Connect, volume *libvirt.StorageVol, size uint64) func(src io.Reader) error {
	copier := func(src io.Reader) error {
		var bytesCopied int64

		stream, err := virConn.NewStream(0)
		if err != nil {
			return err
		}

		defer func() {
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
			stream.Abort()
			return fmt.Errorf("Error: transfer was unexpectedly closed from the server while downloading. Please try again later or check the server hosting sources")
		}
		if err != nil {
			stream.Abort()
			return fmt.Errorf("Error while copying source to volume %s", err)
		}

		log.Printf("%d bytes uploaded\n", bytesCopied)
		if uint64(bytesCopied) != size {
			stream.Abort()
			return fmt.Errorf("Error during volume Upload. BytesCopied: %d != %d volume.size", bytesCopied, size)
		}

		if err := stream.Finish(); err != nil {
			stream.Abort()
			return fmt.Errorf("Error by terminating libvirt stream %s", err)
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
