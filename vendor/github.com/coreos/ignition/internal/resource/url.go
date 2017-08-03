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
	"compress/gzip"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/coreos/ignition/config/types"
	"github.com/coreos/ignition/internal/log"
	"github.com/coreos/ignition/internal/systemd"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/pin/tftp"
	"github.com/vincent-petithory/dataurl"
)

var (
	ErrSchemeUnsupported      = errors.New("unsupported source scheme")
	ErrPathNotAbsolute        = errors.New("path is not absolute")
	ErrNotFound               = errors.New("resource not found")
	ErrFailed                 = errors.New("failed to fetch resource")
	ErrCompressionUnsupported = errors.New("compression is not supported with that scheme")

	// ConfigHeaders are the HTTP headers that should be used when the Ignition
	// config is being fetched
	ConfigHeaders = http.Header{
		"Accept-Encoding": []string{"identity"},
		"Accept":          []string{"application/vnd.coreos.ignition+json; version=2.1.0, application/vnd.coreos.ignition+json; version=1; q=0.5, */*; q=0.1"},
	}
)

// ErrHashMismatch is returned when the calculated hash for a fetched object
// doesn't match the expected sum of the object.
type ErrHashMismatch struct {
	Calculated string
	Expected   string
}

func (e ErrHashMismatch) Error() string {
	return fmt.Sprintf("hash verification failed (calculated %s but expected %s)",
		e.Calculated, e.Expected)
}

const (
	oemDevicePath = "/dev/disk/by-label/OEM" // Device link where oem partition is found.
	oemDirPath    = "/usr/share/oem"         // OEM dir within root fs to consider for pxe scenarios.
	oemMountPath  = "/mnt/oem"               // Mountpoint where oem partition is mounted when present.
)

// Fetcher holds settings for fetching resources from URLs
type Fetcher struct {
	// The logger object to use when logging information.
	Logger *log.Logger

	// Client is the http client that will be used when fetching http(s)
	// resources. If left nil, one will be created and used, but this means any
	// timeouts Ignition was configured to used will be ignored.
	Client *HttpClient

	// The AWS Session to use when fetching resources from S3. If left nil, the
	// first S3 object that is fetched will initialize the field. This can be
	// used to set credentials.
	AWSSession *session.Session
}

type FetchOptions struct {
	// Headers are the http headers that will be used when fetching http(s)
	// resources. They have no effect on other fetching schemes.
	Headers http.Header

	// Hash is the hash to use when calculating a fetched resource's hash. If
	// left as nil, no hash will be calculated.
	Hash hash.Hash

	// The expected sum to be produced by the given hasher. If the Hash field is
	// nil, this field is ignored.
	ExpectedSum []byte

	// Compression specifies the type of compression to use when decompressing
	// the fetched object. If left empty, no decompression will be used.
	Compression string
}

// FetchToBuffer will fetch the given url into a temporrary file, and then read
// in the contents of the file and delete it. It will return the downloaded
// contents, or an error if one was encountered.
func (f *Fetcher) FetchToBuffer(u url.URL, opts FetchOptions) ([]byte, error) {
	file, err := ioutil.TempFile("", "ignition")
	if err != nil {
		return nil, err
	}
	defer os.Remove(file.Name())
	defer file.Close()
	err = f.Fetch(u, file, opts)
	if err != nil {
		return nil, err
	}
	_, err = file.Seek(0, os.SEEK_SET)
	if err != nil {
		return nil, err
	}
	res, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// Fetch calls the appropriate FetchFrom* function based on the scheme of the
// given URL. The results will be decompressed if compression is set in opts,
// and written into dest. If opts.Hash is set the data stream will also be
// hashed and compared against opts.ExpectedSum, and any match failures will
// result in an error being returned.
func (f *Fetcher) Fetch(u url.URL, dest *os.File, opts FetchOptions) error {
	switch u.Scheme {
	case "http", "https":
		return f.FetchFromHTTP(u, dest, opts)
	case "tftp":
		return f.FetchFromTFTP(u, dest, opts)
	case "data":
		return f.FetchFromDataURL(u, dest, opts)
	case "oem":
		return f.FetchFromOEM(u, dest, opts)
	case "s3":
		return f.FetchFromS3(u, dest, opts)
	case "":
		return nil
	default:
		return ErrSchemeUnsupported
	}
}

// FetchFromTFTP fetches a resource from u via TFTP into dest, returning an
// error if one is encountered.
func (f *Fetcher) FetchFromTFTP(u url.URL, dest *os.File, opts FetchOptions) error {
	if !strings.ContainsRune(u.Host, ':') {
		u.Host = u.Host + ":69"
	}
	c, err := tftp.NewClient(u.Host)
	if err != nil {
		return err
	}
	wt, err := c.Receive(u.Path, "octet")
	if err != nil {
		return err
	}
	// The TFTP library takes an io.Writer to send data in to, but to decompress
	// the stream the gzip library wraps an io.Reader, so let's create a pipe to
	// connect these two things
	pReader, pWriter := io.Pipe()
	doneChan := make(chan error, 2)

	checkForDoneChanErr := func(err error) error {
		// If an error is encountered while decompressing or copying data out of
		// the pipe, there's probably an error from writing into the pipe that
		// will better describe what went wrong. This function does a
		// non-blocking read of doneChan, overriding the returned error val if
		// there's anything in doneChan.
		select {
		case writeErr := <-doneChan:
			if writeErr != nil {
				return writeErr
			}
			return err
		default:
			return err
		}
	}

	// A goroutine is used to handle writing the fetched data into the pipe
	// while also copying it out of the pipe concurrently
	go func() {
		_, err := wt.WriteTo(pWriter)
		doneChan <- err
		err = pWriter.Close()
		doneChan <- err
	}()
	err = f.decompressCopyHashAndVerify(dest, pReader, opts)
	if err != nil {
		return checkForDoneChanErr(err)
	}
	// receive the error from wt.WriteTo()
	err = <-doneChan
	if err != nil {
		return err
	}
	// receive the error from pWriter.Close()
	err = <-doneChan
	if err != nil {
		return err
	}
	return nil
}

// FetchFromHTTP fetches a resource from u via HTTP(S) into dest, returning an
// error if one is encountered.
func (f *Fetcher) FetchFromHTTP(u url.URL, dest *os.File, opts FetchOptions) error {
	if f.Client == nil {
		f.Logger.Warning("Fetcher http client not initialized, ignoring any possible timeouts")
		c := NewHttpClient(f.Logger, types.Timeouts{})
		f.Client = &c
	}
	dataReader, status, err := f.Client.getReaderWithHeader(u.String(), opts.Headers)
	if err != nil {
		return err
	}
	defer dataReader.Close()

	switch status {
	case http.StatusOK, http.StatusNoContent:
		break
	case http.StatusNotFound:
		return ErrNotFound
	default:
		return ErrFailed
	}

	return f.decompressCopyHashAndVerify(dest, dataReader, opts)
}

// FetchFromDataURL writes the data stored in the dataurl u into dest, returning
// an error if one is encountered.
func (f *Fetcher) FetchFromDataURL(u url.URL, dest *os.File, opts FetchOptions) error {
	if opts.Compression != "" {
		return ErrCompressionUnsupported
	}
	url, err := dataurl.DecodeString(u.String())
	if err != nil {
		return err
	}

	return f.decompressCopyHashAndVerify(dest, bytes.NewBuffer(url.Data), opts)
}

// FetchFromOEM gets data off the oem partition as described by u and writes it
// into dest, returning an error if one is encountered.
func (f *Fetcher) FetchFromOEM(u url.URL, dest *os.File, opts FetchOptions) error {
	path := filepath.Clean(u.Path)
	if !filepath.IsAbs(path) {
		f.Logger.Err("oem path is not absolute: %q", u.Path)
		return ErrPathNotAbsolute
	}

	// check if present under oemDirPath, if so use it.
	absPath := filepath.Join(oemDirPath, path)

	if fi, err := os.Open(absPath); err == nil {
		defer fi.Close()
		return f.decompressCopyHashAndVerify(dest, fi, opts)
	} else if !os.IsNotExist(err) {
		f.Logger.Err("failed to read oem config: %v", err)
		return ErrFailed
	}

	f.Logger.Info("oem config not found in %q, trying %q",
		oemDirPath, oemMountPath)

	// try oemMountPath, requires mounting it.
	if err := f.mountOEM(); err != nil {
		f.Logger.Err("failed to mount oem partition: %v", err)
		return ErrFailed
	}
	defer f.umountOEM()

	absPath = filepath.Join(oemMountPath, path)
	fi, err := os.Open(absPath)
	if err != nil {
		f.Logger.Err("failed to read oem config: %v", err)
		return ErrFailed
	}
	defer fi.Close()

	return f.decompressCopyHashAndVerify(dest, fi, opts)
}

// FetchFromS3 gets data from an S3 bucket as described by u and writes it into
// dest, returning an error if one is encountered. It will attempt to acquire
// IAM credentials from the EC2 metadata service, and if this fails will attempt
// to fetch the object with anonymous credentials.
func (f *Fetcher) FetchFromS3(u url.URL, dest *os.File, opts FetchOptions) error {
	if opts.Compression != "" {
		return ErrCompressionUnsupported
	}
	ctx := context.Background()
	if f.Client != nil && f.Client.timeout != 0 {
		var cancelFn context.CancelFunc
		ctx, cancelFn = context.WithTimeout(ctx, f.Client.timeout)
		defer cancelFn()
	}

	if f.AWSSession == nil {
		var err error
		f.AWSSession, err = session.NewSession(&aws.Config{
			Credentials: credentials.AnonymousCredentials,
		})
		if err != nil {
			return err
		}
	}
	sess := f.AWSSession.Copy()

	// Determine the region this bucket is in
	region, err := s3manager.GetBucketRegion(ctx, sess, u.Host, "us-east-1")
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == "NotFound" {
			return fmt.Errorf("couldn't determine the region for bucket %q: %v", u.Host, err)
		}
		return err
	}

	sess.Config.Region = aws.String(region)

	input := &s3.GetObjectInput{
		Bucket: &u.Host,
		Key:    &u.Path,
	}
	err = f.fetchFromS3WithCreds(ctx, dest, input, sess)
	if err != nil {
		return err
	}
	if opts.Hash != nil {
		opts.Hash.Reset()
		_, err = dest.Seek(0, os.SEEK_SET)
		if err != nil {
			return err
		}
		_, err = io.Copy(opts.Hash, dest)
		if err != nil {
			return err
		}
		calculatedSum := opts.Hash.Sum(nil)
		if !bytes.Equal(calculatedSum, opts.ExpectedSum) {
			return ErrHashMismatch{
				Calculated: hex.EncodeToString(calculatedSum),
				Expected:   hex.EncodeToString(opts.ExpectedSum),
			}
		}
		f.Logger.Debug("file matches expected sum of: %s", hex.EncodeToString(opts.ExpectedSum))
	}
	return nil
}

func (f *Fetcher) fetchFromS3WithCreds(ctx context.Context, dest *os.File, input *s3.GetObjectInput, sess *session.Session) error {
	downloader := s3manager.NewDownloader(sess)
	_, err := downloader.DownloadWithContext(ctx, dest, input)
	if err != nil {
		if awserrval, ok := err.(awserr.Error); ok && awserrval.Code() == "EC2RoleRequestError" {
			// If this error was due to an EC2 role request error, try again
			// with the anonymous credentials.
			sess.Config.Credentials = credentials.AnonymousCredentials
			return f.fetchFromS3WithCreds(ctx, dest, input, sess)
		}
		return err
	}
	return nil
}

// uncompress will wrap the given io.Reader in a decompresser specified in the
// FetchOptions, and return an io.ReadCloser with the decompressed data stream.
func (f *Fetcher) uncompress(r io.Reader, opts FetchOptions) (io.ReadCloser, error) {
	switch opts.Compression {
	case "":
		return ioutil.NopCloser(r), nil
	case "gzip":
		return gzip.NewReader(r)
	default:
		return nil, types.ErrCompressionInvalid
	}
}

// decompressCopyHashAndVerify will decompress src if necessary, copy src into
// dest until src returns an io.EOF while also calculating a hash if one is set,
// and will return an error if there's any problems with any of this or if the
// hash doesn't match the expected hash in the opts.
func (f *Fetcher) decompressCopyHashAndVerify(dest io.Writer, src io.Reader, opts FetchOptions) error {
	decompressor, err := f.uncompress(src, opts)
	if err != nil {
		return err
	}
	defer decompressor.Close()
	if opts.Hash != nil {
		opts.Hash.Reset()
		dest = io.MultiWriter(dest, opts.Hash)
	}
	_, err = io.Copy(dest, decompressor)
	if err != nil {
		return err
	}
	if opts.Hash != nil {
		calculatedSum := opts.Hash.Sum(nil)
		if !bytes.Equal(calculatedSum, opts.ExpectedSum) {
			return ErrHashMismatch{
				Calculated: hex.EncodeToString(calculatedSum),
				Expected:   hex.EncodeToString(opts.ExpectedSum),
			}
		}
		f.Logger.Debug("file matches expected sum of: %s", hex.EncodeToString(opts.ExpectedSum))
	}
	return nil
}

// mountOEM waits for the presence of and mounts the oem partition at
// oemMountPath.
func (f *Fetcher) mountOEM() error {
	dev := []string{oemDevicePath}
	if err := systemd.WaitOnDevices(dev, "oem-cmdline"); err != nil {
		f.Logger.Err("failed to wait for oem device: %v", err)
		return err
	}

	if err := os.MkdirAll(oemMountPath, 0700); err != nil {
		f.Logger.Err("failed to create oem mount point: %v", err)
		return err
	}

	if err := f.Logger.LogOp(
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
func (f *Fetcher) umountOEM() {
	f.Logger.LogOp(
		func() error { return syscall.Unmount(oemMountPath, 0) },
		"unmounting %q", oemMountPath,
	)
}
