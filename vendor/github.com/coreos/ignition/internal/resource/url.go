// Copyright 2016 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package resource

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/coreos/ignition/internal/log"
	"github.com/coreos/ignition/internal/systemd"

	"github.com/pin/tftp"
	"github.com/vincent-petithory/dataurl"
)

var (
	ErrSchemeUnsupported = errors.New("unsupported source scheme")
	ErrPathNotAbsolute   = errors.New("path is not absolute")
	ErrNotFound          = errors.New("resource not found")
	ErrFailed            = errors.New("failed to fetch resource")
)

const (
	oemDevicePath = "/dev/disk/by-label/OEM" // Device link where oem partition is found.
	oemDirPath    = "/usr/share/oem"         // OEM dir within root fs to consider for pxe scenarios.
	oemMountPath  = "/mnt/oem"               // Mountpoint where oem partition is mounted when present.
)

// FetchConfig fetches a raw config from the provided URL.
func FetchConfig(l *log.Logger, c *HttpClient, u url.URL) ([]byte, error) {
	return FetchConfigWithHeader(l, c, u, http.Header{})
}

// FetchConfigWithHeader fetches a raw config from the provided URL and returns
// the response body on success or nil on failure. The HTTP response must be
// OK, otherwise an empty (v.s. nil) config is returned. The provided headers
// are merged with a set of default headers.
func FetchConfigWithHeader(l *log.Logger, c *HttpClient, u url.URL, h http.Header) ([]byte, error) {
	header := http.Header{
		"Accept-Encoding": []string{"identity"},
		"Accept":          []string{"application/vnd.coreos.ignition+json; version=2.0.0, application/vnd.coreos.ignition+json; version=1; q=0.5, */*; q=0.1"},
	}
	for key, values := range h {
		header.Del(key)
		for _, value := range values {
			header.Add(key, value)
		}
	}

	data, err := FetchWithHeader(l, c, u, header)
	switch err {
	case nil:
		return data, nil
	case ErrNotFound:
		return []byte{}, nil
	default:
		return nil, err
	}
}

// Fetch fetches a resource given a URL. The supported schemes are
// http, data, tftp, and oem.
func Fetch(l *log.Logger, c *HttpClient, u url.URL) ([]byte, error) {
	return FetchWithHeader(l, c, u, http.Header{})
}

// FetchWithHeader fetches a resource given a URL. If the resource is
// of the http or https scheme, the provided header will be used when
// fetching. The supported schemes are http, data, tftp, and oem.
func FetchWithHeader(l *log.Logger, c *HttpClient, u url.URL, h http.Header) ([]byte, error) {
	if u.Scheme == "tftp" {
		return FetchFromTftp(l, u)
	}

	var data []byte

	dataReader, err := FetchAsReaderWithHeader(l, c, u, h)
	if err != nil {
		return nil, err
	}
	defer dataReader.Close()

	data, err = ioutil.ReadAll(dataReader)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// FetchFromTftp fetches a resource from a tftp server.
func FetchFromTftp(l *log.Logger, u url.URL) ([]byte, error) {
	if !strings.ContainsRune(u.Host, ':') {
		u.Host = u.Host + ":69"
	}
	c, err := tftp.NewClient(u.Host)
	if err != nil {
		return nil, err
	}
	wt, err := c.Receive(u.Path, "octet")
	if err != nil {
		return nil, err
	}
	buf := &bytes.Buffer{}
	_, err = wt.WriteTo(buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// readUnmounter calls umountOEM() when closed, in addition to closing the
// ReadCloser it wraps.
type readUnmounter struct {
	io.ReadCloser
	logger *log.Logger
}

func (f readUnmounter) Close() error {
	defer umountOEM(f.logger)
	return f.ReadCloser.Close()
}

// FetchAsReader returns a ReadCloser to the data at the URL specified.
// The caller is responsible for closing the reader.
func FetchAsReader(l *log.Logger, c *HttpClient, u url.URL) (io.ReadCloser, error) {
	return FetchAsReaderWithHeader(l, c, u, http.Header{})
}

// FetchAsReader returns a ReadCloser to the data at the URL specified.
// If the URL is of the http or https scheme, the provided header will be used
// when fetching. The caller is responsible for closing the reader.
func FetchAsReaderWithHeader(l *log.Logger, c *HttpClient, u url.URL, h http.Header) (io.ReadCloser, error) {
	switch u.Scheme {
	case "http", "https":
		dataReader, status, err := c.getReaderWithHeader(u.String(), h)
		if err != nil {
			return nil, err
		}

		switch status {
		case http.StatusOK, http.StatusNoContent:
			return dataReader, nil
		case http.StatusNotFound:
			return nil, ErrNotFound
		default:
			return nil, ErrFailed
		}

	case "data":
		url, err := dataurl.DecodeString(u.String())
		if err != nil {
			return nil, err
		}
		return ioutil.NopCloser(bytes.NewReader(url.Data)), nil

	case "oem":
		path := filepath.Clean(u.Path)
		if !filepath.IsAbs(path) {
			l.Err("oem path is not absolute: %q", u.Path)
			return nil, ErrPathNotAbsolute
		}

		// check if present under oemDirPath, if so use it.
		absPath := filepath.Join(oemDirPath, path)

		if f, err := os.Open(absPath); err == nil {
			return f, nil
		} else if !os.IsNotExist(err) {
			l.Err("failed to read oem config: %v", err)
			return nil, ErrFailed
		}

		l.Info("oem config not found in %q, trying %q",
			oemDirPath, oemMountPath)

		// try oemMountPath, requires mounting it.
		if err := mountOEM(l); err != nil {
			l.Err("failed to mount oem partition: %v", err)
			return nil, ErrFailed
		}

		absPath = filepath.Join(oemMountPath, path)
		f, err := os.Open(absPath)
		if err != nil {
			l.Err("failed to read oem config: %v", err)
			umountOEM(l)
			return nil, ErrFailed
		}

		return readUnmounter{
			logger:     l,
			ReadCloser: f,
		}, nil

	case "":
		f, err := os.Open(os.DevNull)
		if err != nil {
			l.Err("Failed to open /dev/null for writing empty files")
			return nil, ErrFailed
		}
		return f, nil

	default:
		return nil, ErrSchemeUnsupported
	}
}

// mountOEM waits for the presence of and mounts the oem partition at
// oemMountPath.
func mountOEM(l *log.Logger) error {
	dev := []string{oemDevicePath}
	if err := systemd.WaitOnDevices(dev, "oem-cmdline"); err != nil {
		l.Err("failed to wait for oem device: %v", err)
		return err
	}

	if err := os.MkdirAll(oemMountPath, 0700); err != nil {
		l.Err("failed to create oem mount point: %v", err)
		return err
	}

	if err := l.LogOp(
		func() error {
			return syscall.Mount(dev[0], oemMountPath, "ext4", 0, "")
		},
		"mounting %q at %q", oemDevicePath, oemMountPath,
	); err != nil {
		return fmt.Errorf("failed to mount device %q at %q: %v",
			oemDevicePath, oemMountPath, err)
	}

	return nil
}

// umountOEM unmounts the oem partition at oemMountPath.
func umountOEM(l *log.Logger) {
	l.LogOp(
		func() error { return syscall.Unmount(oemMountPath, 0) },
		"unmounting %q", oemMountPath,
	)
}
