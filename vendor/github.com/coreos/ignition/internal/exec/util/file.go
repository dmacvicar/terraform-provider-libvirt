// Copyright 2015 CoreOS, Inc.
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

package util

import (
	"bufio"
	"compress/gzip"
	"encoding/hex"
	"hash"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"

	"github.com/coreos/ignition/config/types"
	"github.com/coreos/ignition/internal/log"
	"github.com/coreos/ignition/internal/resource"

	"golang.org/x/net/context"
)

const (
	DefaultDirectoryPermissions os.FileMode = 0755
	DefaultFilePermissions      os.FileMode = 0644
)

type File struct {
	io.ReadCloser
	hash.Hash
	Path        string
	Mode        os.FileMode
	Uid         int
	Gid         int
	expectedSum string
}

func (f File) Verify() error {
	if f.Hash == nil {
		return nil
	}
	sum := f.Sum(nil)
	encodedSum := make([]byte, hex.EncodedLen(len(sum)))
	hex.Encode(encodedSum, sum)

	if string(encodedSum) != f.expectedSum {
		return ErrHashMismatch{
			Calculated: string(encodedSum),
			Expected:   f.expectedSum,
		}
	}
	return nil
}

// newHashedReader returns a new ReadCloser that also writes to the provided hash.
func newHashedReader(reader io.ReadCloser, hasher hash.Hash) io.ReadCloser {
	return struct {
		io.Reader
		io.Closer
	}{
		Reader: io.TeeReader(reader, hasher),
		Closer: reader,
	}
}

// RenderFile returns a *File with a Reader that downloads, hashes, and decompresses the incoming data.
// It returns nil if f had invalid options. Errors reading/verifying/decompressing the file will
// present themselves when the Reader is actually read from.
func RenderFile(l *log.Logger, c *resource.HttpClient, f types.File) *File {
	var reader io.ReadCloser
	var err error
	var expectedSum string

	// explicitly ignoring the error here because the config should already be
	// validated by this point
	u, _ := url.Parse(f.Contents.Source)

	reader, err = resource.FetchAsReader(l, c, context.Background(), *u)
	if err != nil {
		l.Crit("Error fetching file %q: %v", f.Path, err)
		return nil
	}

	fileHash, err := GetHasher(f.Contents.Verification)
	if err != nil {
		l.Crit("Error verifying file %q: %v", f.Path, err)
		return nil
	}

	if fileHash != nil {
		reader = newHashedReader(reader, fileHash)
		// explicitly ignoring the error here because the config should already
		// be validated by this point
		_, expectedSum, _ = f.Contents.Verification.HashParts()
	}

	reader, err = decompressFileStream(l, f, reader)
	if err != nil {
		l.Crit("Error decompressing file %q: %v", f.Path, err)
		return nil
	}

	return &File{
		Path:        f.Path,
		ReadCloser:  reader,
		Hash:        fileHash,
		Mode:        os.FileMode(f.Mode),
		Uid:         f.User.ID,
		Gid:         f.Group.ID,
		expectedSum: expectedSum,
	}
}

// gzipReader is a wrapper for gzip's reader that closes the stream it wraps as well
// as itself when Close() is called.
type gzipReader struct {
	*gzip.Reader //actually a ReadCloser
	source       io.Closer
}

func newGzipReader(reader io.ReadCloser) (io.ReadCloser, error) {
	gzReader, err := gzip.NewReader(reader)
	if err != nil {
		return nil, err
	}
	return gzipReader{
		Reader: gzReader,
		source: reader,
	}, nil
}

func (gz gzipReader) Close() error {
	if err := gz.Reader.Close(); err != nil {
		return err
	}
	if err := gz.source.Close(); err != nil {
		return err
	}
	return nil
}

func decompressFileStream(l *log.Logger, f types.File, contents io.ReadCloser) (io.ReadCloser, error) {
	switch f.Contents.Compression {
	case "":
		return contents, nil
	case "gzip":
		return newGzipReader(contents)
	default:
		return nil, types.ErrCompressionInvalid
	}
}

func (u Util) WriteLink(s types.Link) error {
	path := u.JoinPath(s.Path)

	if err := MkdirForFile(path); err != nil {
		return err
	}

	if s.Hard {
		return os.Link(s.Target, path)
	}
	return os.Symlink(s.Target, path)
}

// WriteFile creates and writes the file described by f using the provided context.
func (u Util) WriteFile(f *File) error {
	defer f.Close()
	var err error

	path := u.JoinPath(string(f.Path))

	if err := MkdirForFile(path); err != nil {
		return err
	}

	// Create a temporary file in the same directory to ensure it's on the same filesystem
	var tmp *os.File
	if tmp, err = ioutil.TempFile(filepath.Dir(path), "tmp"); err != nil {
		return err
	}

	defer func() {
		tmp.Close()
		if err != nil {
			os.Remove(tmp.Name())
		}
	}()

	fileWriter := bufio.NewWriter(tmp)

	if _, err = io.Copy(fileWriter, f); err != nil {
		return err
	}
	fileWriter.Flush()

	if err = f.Verify(); err != nil {
		return err
	}

	// XXX(vc): Note that we assume to be operating on the file we just wrote, this is only guaranteed
	// by using syscall.Fchown() and syscall.Fchmod()

	// Ensure the ownership and mode are as requested (since WriteFile can be affected by sticky bit)
	if err = os.Chown(tmp.Name(), f.Uid, f.Gid); err != nil {
		return err
	}

	if err = os.Chmod(tmp.Name(), f.Mode); err != nil {
		return err
	}

	if err = os.Rename(tmp.Name(), path); err != nil {
		return err
	}

	return nil
}

// MkdirForFile helper creates the directory components of path.
func MkdirForFile(path string) error {
	return os.MkdirAll(filepath.Dir(path), DefaultDirectoryPermissions)
}
