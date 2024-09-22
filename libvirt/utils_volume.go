package libvirt

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	libvirt "github.com/digitalocean/go-libvirt"
)

const (
	copierBufferSize = 4 * 1024 * 1024
)

func newVolumeUploader(virConn *libvirt.Libvirt, volume *libvirt.StorageVol, size uint64) func(src io.Reader) error {
	return func(src io.Reader) error {
		start := time.Now()
		if err := virConn.StorageVolUpload(*volume, bufio.NewReaderSize(src, copierBufferSize), 0, size, 0); err != nil {
			return fmt.Errorf("error while uploading volume %w", err)
		}
		log.Printf("[DEBUG] upload took %d ms", time.Since(start).Milliseconds())

		return nil
	}
}

// returns a function you can give a writer to download the volume content
// the function will return the downloaded size
func newVolumeDownloader(virConn *libvirt.Libvirt, volume *libvirt.StorageVol) func(src io.Writer) error {
	return func(dst io.Writer) error {
		start := time.Now()

		bufdst := bufio.NewWriter(dst)
		if err := virConn.StorageVolDownload(*volume, bufdst, 0, 0, 0); err != nil {
			return fmt.Errorf("error while downloading volume: %w", err)
		}

		log.Printf("[DEBUG] download took %d ms", time.Since(start).Milliseconds())

		return bufdst.Flush()
	}
}

//nolint:mnd
func timeFromEpoch(str string) time.Time {
	var s, ns int

	ts := strings.Split(str, ".")
	if len(ts) == 2 {
		ns, _ = strconv.Atoi(ts[1])
	}
	s, _ = strconv.Atoi(ts[0])

	return time.Unix(int64(s), int64(ns))
}
