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

func newCopier(virConn *libvirt.Libvirt, volume *libvirt.StorageVol, size uint64) func(src io.Reader) error {
	copier := func(src io.Reader) error {
		start := time.Now()
		if err := virConn.StorageVolUpload(*volume, bufio.NewReaderSize(src, copierBufferSize), 0, size, 0); err != nil {
			return fmt.Errorf("error while uploading volume %w", err)
		}
		log.Printf("[DEBUG] upload took %d ms", time.Since(start).Milliseconds())

		return nil
	}
	return copier
}

//nolint:gomnd
func timeFromEpoch(str string) time.Time {
	var s, ns int

	ts := strings.Split(str, ".")
	if len(ts) == 2 {
		ns, _ = strconv.Atoi(ts[1])
	}
	s, _ = strconv.Atoi(ts[0])

	return time.Unix(int64(s), int64(ns))
}
