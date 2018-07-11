// Copyright 2017 CoreOS, Inc.
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

package blackbox

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pin/tftp"

	config "github.com/coreos/ignition/config/v2_3_experimental"
	"github.com/coreos/ignition/tests/register"
	"github.com/coreos/ignition/tests/types"

	// Register the tests
	_ "github.com/coreos/ignition/tests/registry"
)

var (
	// testTimeout controls how long a given test is allowed to run before being
	// cancelled.
	testTimeout = time.Second * 30
)

// HTTP Server
func (server *HTTPServer) Config(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`{
	"ignition": { "version": "2.0.0" },
	"storage": {
		"files": [{
		  "filesystem": "root",
		  "path": "/foo/bar",
		  "contents": { "source": "data:,example%20file%0A" }
		}]
	}
}`))
}

func (server *HTTPServer) Contents(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`asdf
fdsa`))
}

type HTTPServer struct{}

func (server *HTTPServer) Start() {
	http.HandleFunc("/contents", server.Contents)
	http.HandleFunc("/config", server.Config)

	s := &http.Server{Addr: ":8080"}
	go s.ListenAndServe()
}

// TFTP Server
func (server *TFTPServer) ReadHandler(filename string, rf io.ReaderFrom) error {
	var buf *bytes.Reader
	if strings.Contains(filename, "contents") {
		buf = bytes.NewReader([]byte(`asdf
fdsa`))
	} else if strings.Contains(filename, "config") {
		buf = bytes.NewReader([]byte(`{
        "ignition": { "version": "2.0.0" },
        "storage": {
                "files": [{
                  "filesystem": "root",
                  "path": "/foo/bar",
                  "contents": { "source": "data:,example%20file%0A" }
                }]
        }
}`))
	} else {
		return fmt.Errorf("no such file %q", filename)
	}

	_, err := rf.ReadFrom(buf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return err
	}
	return nil
}

type TFTPServer struct{}

func (server *TFTPServer) Start() {
	s := tftp.NewServer(server.ReadHandler, nil)
	s.SetTimeout(5 * time.Second)
	go s.ListenAndServe(":69")
}

func TestMain(m *testing.M) {
	httpServer := &HTTPServer{}
	httpServer.Start()
	tftpServer := &TFTPServer{}
	tftpServer.Start()

	os.Exit(m.Run())
}

func TestIgnitionBlackBox(t *testing.T) {
	for _, test := range register.Tests[register.PositiveTest] {
		t.Run(test.Name, func(t *testing.T) {
			err := outer(t, test, false)
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func TestIgnitionBlackBoxNegative(t *testing.T) {
	for _, test := range register.Tests[register.NegativeTest] {
		t.Run(test.Name, func(t *testing.T) {
			err := outer(t, test, true)
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func outer(t *testing.T, test types.Test, negativeTests bool) error {
	t.Log(test.Name)

	ctx, cancelFunc := context.WithDeadline(context.Background(), time.Now().Add(testTimeout))
	defer cancelFunc()

	originalTmpDir := os.Getenv("TMPDIR")
	if originalTmpDir == "" {
		err := os.Setenv("TMPDIR", "/var/tmp")
		if err != nil {
			return fmt.Errorf("couldn't initialize TMPDIR: %v", err)
		}
	}

	tmpDirectory, err := ioutil.TempDir("", "ignition-blackbox-")
	if err != nil {
		return fmt.Errorf("failed to create a temp dir: %v", err)
	}
	// the tmpDirectory must be 0755 or the tests will fail as the tool will
	// not have permissions to perform some actions in the mounted folders
	err = os.Chmod(tmpDirectory, 0755)
	if err != nil {
		return fmt.Errorf("failed to change mode of temp dir: %v", err)
	}
	err = os.Setenv("TMPDIR", tmpDirectory)
	if err != nil {
		return fmt.Errorf("failed to set TMPDIR environment var: %v", err)
	}
	defer os.Setenv("TMPDIR", originalTmpDir)
	defer os.RemoveAll(tmpDirectory)

	oemLookasideDir := filepath.Join(os.TempDir(), "oem-lookaside")
	systemConfigDir := filepath.Join(os.TempDir(), "system")
	var rootLocation string

	// Setup
	err = createFilesFromSlice(t, oemLookasideDir, test.OEMLookasideFiles)
	// Defer before the error handling because the createFilesFromSlice function
	// can fail after partially-creating things
	defer os.RemoveAll(oemLookasideDir)
	if err != nil {
		return err
	}
	err = createFilesFromSlice(t, systemConfigDir, test.SystemDirFiles)
	defer os.RemoveAll(systemConfigDir)
	if err != nil {
		return err
	}
	for i, disk := range test.In {
		// Set image file path
		disk.ImageFile = filepath.Join(os.TempDir(), fmt.Sprintf("hd%d", i))
		test.Out[i].ImageFile = disk.ImageFile

		// There may be more partitions created by Ignition, so look at the
		// expected output instead of the input to determine image size
		imageSize := test.Out[i].CalculateImageSize()

		// Finish data setup
		for _, part := range disk.Partitions {
			if part.GUID == "" {
				part.GUID, err = generateUUID(t, ctx)
				if err != nil {
					return err
				}
			}
			err := updateTypeGUID(t, part)
			if err != nil {
				return err
			}
		}

		disk.SetOffsets()
		for _, part := range test.Out[i].Partitions {
			err := updateTypeGUID(t, part)
			if err != nil {
				return err
			}
		}
		test.Out[i].SetOffsets()

		// Creation
		err = createVolume(t, ctx, i, disk.ImageFile, imageSize, 20, 16, 63, disk.Partitions)
		// Move value into the local scope, because disk.ImageFile will change
		// by the time this runs
		imageFile := disk.ImageFile
		defer func() {
			err := os.Remove(imageFile)
			if err != nil {
				t.Errorf("couldn't remove %s: %v", imageFile, err)
			}
		}()
		if err != nil {
			return err
		}

		disk.Device, err = setDevices(t, ctx, disk.ImageFile, disk.Partitions)
		device := disk.Device
		defer func() {
			err := destroyDevice(t, device)
			if err != nil {
				t.Errorf("couldn't destroy device: %v", err)
			}
		}()
		if err != nil {
			return err
		}
		test.Out[i].Device = disk.Device

		rootMounted, err := mountRootPartition(t, ctx, disk.Partitions)
		partitions := disk.Partitions
		defer func() {
			err := unmountRootPartition(t, partitions)
			if err != nil {
				t.Errorf("couldn't unmount root partition: %v", err)
			}
		}()
		if err != nil {
			return err
		}

		if rootMounted && strings.Contains(test.Config, "passwd") {
			err = prepareRootPartitionForPasswd(t, ctx, disk.Partitions)
			if err != nil {
				return err
			}
		}
		err = mountPartitions(t, ctx, disk.Partitions)
		if err != nil {
			// mountPartitions may have partially succeded, so at least try to
			// unmount things.
			unmountPartitions(t, disk.Partitions)
			return err
		}
		err = createFilesForPartitions(t, disk.Partitions)
		// unmount even if createFilesForPartitions failed
		errUnmount := unmountPartitions(t, disk.Partitions)
		if err != nil {
			if errUnmount != nil {
				t.Errorf("couldn't unmount partitions: %v", errUnmount)
			}
			return err
		} else if errUnmount != nil {
			return errUnmount
		}

		// Mount device name substitution
		for _, d := range test.MntDevices {
			device := pickPartition(t, disk.Device, disk.Partitions, d.Label)
			// The device may not be on this disk, if it's not found here let's
			// assume we'll find it on another one and keep going
			if device != "" {
				test.Config = strings.Replace(test.Config, d.Substitution, device, -1)
			}
		}

		// Replace any instance of $disk<num> with the actual loop device
		// that got assigned to it
		test.Config = strings.Replace(test.Config, fmt.Sprintf("$disk%d", i), disk.Device, -1)

		if rootLocation == "" {
			rootLocation = getRootLocation(disk.Partitions)
		}
	}

	// Validation and cleanup deferral
	for i, disk := range test.Out {
		// Update out structure with mount points & devices
		setExpectedPartitionsDrive(test.In[i].Partitions, disk.Partitions)
	}

	if rootLocation == "" {
		return fmt.Errorf("ROOT filesystem not found! A partition labeled ROOT is requred")
	}

	// Let's make sure that all of the devices we needed to substitute names in
	// for were found
	for _, d := range test.MntDevices {
		if strings.Contains(test.Config, d.Substitution) {
			return fmt.Errorf("Didn't find a drive with label: %s", d.Substitution)
		}
	}

	// If we're not expecting the config to be bad, make sure it passes
	// validation.
	if !test.ConfigShouldBeBad {
		_, rpt, err := config.Parse([]byte(test.Config))
		if rpt.IsFatal() {
			return fmt.Errorf("test has bad config: %s", rpt.String())
		}
		if err != nil {
			return fmt.Errorf("error parsing config: %v", err)
		}
	}

	// Ignition config
	if err := ioutil.WriteFile(filepath.Join(tmpDirectory, "config.ign"), []byte(test.Config), 0666); err != nil {
		return fmt.Errorf("error writing config: %v", err)
	}

	// Ignition
	appendEnv := []string{
		"IGNITION_OEM_DEVICE=" + test.In[0].Partitions.GetPartition("OEM").Device,
		"IGNITION_OEM_LOOKASIDE_DIR=" + oemLookasideDir,
		"IGNITION_SYSTEM_CONFIG_DIR=" + systemConfigDir,
	}
	disksErr := runIgnition(t, ctx, "disks", rootLocation, tmpDirectory, appendEnv)
	if !negativeTests && disksErr != nil {
		return disksErr
	}
	filesErr := runIgnition(t, ctx, "files", rootLocation, tmpDirectory, appendEnv)
	if !negativeTests && filesErr != nil {
		return err
	}
	if negativeTests && disksErr == nil && filesErr == nil {
		return fmt.Errorf("Expected failure and ignition succeeded")
	}

	for _, disk := range test.Out {
		if !negativeTests {
			// Validation
			err = mountPartitions(t, ctx, disk.Partitions)
			if err != nil {
				// mountPartitions may have partially succeeded, so try to
				// unmount everything before returning
				unmountPartitions(t, disk.Partitions)
				return err
			}
			partitions := disk.Partitions
			defer func() {
				err := unmountPartitions(t, partitions)
				if err != nil {
					t.Errorf("error unmounting partitions: %v", err)
				}
			}()
			t.Log(disk.ImageFile)
			err = validateDisk(t, disk, disk.ImageFile)
			if err != nil {
				return err
			}
			err = validateFilesystems(t, disk.Partitions, disk.ImageFile)
			if err != nil {
				return err
			}
			validateFilesDirectoriesAndLinks(t, disk.Partitions)
		}
	}
	return nil
}
