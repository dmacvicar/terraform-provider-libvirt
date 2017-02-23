package libvirt

import (
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	libvirt "github.com/dmacvicar/libvirt-go"
	"github.com/hooklift/iso9660"
	"github.com/mitchellh/packer/common/uuid"
	"gopkg.in/yaml.v2"
)

// names of the files expected by cloud-init
const USERDATA string = "user-data"
const METADATA string = "meta-data"

type defCloudInit struct {
	Name     string
	PoolName string
	Metadata struct {
		LocalHostname string `yaml:"local-hostname,omitempty"`
		InstanceID    string `yaml:"instance-id"`
	}
	UserDataRaw string `yaml:"user_data"`
	UserData    struct {
		SSHAuthorizedKeys []string `yaml:"ssh_authorized_keys"`
	}
}

// Creates a new cloudinit with the defaults
// the provider uses
func newCloudInitDef() defCloudInit {
	ci := defCloudInit{}
	ci.Metadata.InstanceID = fmt.Sprintf("created-at-%s", time.Now().String())

	return ci
}

// Create a ISO file based on the contents of the CloudInit instance and
// uploads it to the libVirt pool
// Returns a string holding terraform's internal ID of this resource
func (ci *defCloudInit) CreateAndUpload(virConn *libvirt.VirConnection) (string, error) {
	iso, err := ci.createISO()
	if err != nil {
		return "", err
	}

	pool, err := virConn.LookupStoragePoolByName(ci.PoolName)
	if err != nil {
		return "", fmt.Errorf("can't find storage pool '%s'", ci.PoolName)
	}
	defer pool.Free()

	PoolSync.AcquireLock(ci.PoolName)
	defer PoolSync.ReleaseLock(ci.PoolName)

	// Refresh the pool of the volume so that libvirt knows it is
	// not longer in use.
	WaitForSuccess("Error refreshing pool for volume", func() error {
		return pool.Refresh(0)
	})

	volumeDef := newDefVolume()
	volumeDef.Name = ci.Name

	// an existing image was given, this mean we can't choose size
	img, err := newImage(iso)
	if err != nil {
		return "", err
	}
	defer func() {
		// Remove the tmp directory holding the ISO
		if err = os.RemoveAll(filepath.Dir(iso)); err != nil {
			log.Printf("Error while removing tmp directory holding the ISO file: %s", err)
		}
	}()

	size, err := img.Size()
	if err != nil {
		return "", err
	}

	volumeDef.Capacity.Unit = "B"
	volumeDef.Capacity.Amount = size
	volumeDef.Target.Format.Type = "raw"

	volumeDefXml, err := xml.Marshal(volumeDef)
	if err != nil {
		return "", fmt.Errorf("Error serializing libvirt volume: %s", err)
	}

	// create the volume
	volume, err := pool.StorageVolCreateXML(string(volumeDefXml), 0)
	if err != nil {
		return "", fmt.Errorf("Error creating libvirt volume for cloudinit device %s: %s", ci.Name, err)
	}
	defer volume.Free()

	// upload ISO file
	stream, err := libvirt.NewVirStream(virConn, 0)
	if err != nil {
		return "", err
	}
	defer stream.Close()

	volume.Upload(stream, 0, uint64(volumeDef.Capacity.Amount), 0)
	err = img.WriteToStream(stream)
	if err != nil {
		return "", err
	}

	key, err := volume.GetKey()
	if err != nil {
		return "", fmt.Errorf("Error retrieving volume key: %s", err)
	}

	return ci.buildTerraformKey(key), nil
}

// create a unique ID for terraform use
// The ID is made by the volume ID (the internal one used by libvirt)
// joined by the ";" with a UUID
func (ci *defCloudInit) buildTerraformKey(volumeKey string) string {
	return fmt.Sprintf("%s;%s", volumeKey, uuid.TimeOrderedUUID())
}

func getCloudInitVolumeKeyFromTerraformID(id string) (string, error) {
	s := strings.SplitN(id, ";", 2)
	if len(s) != 2 {
		return "", fmt.Errorf("%s is not a valid key", id)
	}
	return s[0], nil
}

// Create the ISO holding all the cloud-init data
// Returns a string with the full path to the ISO file
func (ci *defCloudInit) createISO() (string, error) {
	log.Print("Creating new ISO")
	tmpDir, err := ci.createFiles()
	if err != nil {
		return "", err
	}

	isoDestination := filepath.Join(tmpDir, ci.Name)
	cmd := exec.Command(
		"genisoimage",
		"-output",
		isoDestination,
		"-volid",
		"cidata",
		"-joliet",
		"-rock",
		filepath.Join(tmpDir, USERDATA),
		filepath.Join(tmpDir, METADATA))

	log.Print("About to execute cmd: %+v", cmd)
	if err = cmd.Run(); err != nil {
		return "", fmt.Errorf("Error while starting the creation of CloudInit's ISO image: %s", err)
	}
	log.Printf("ISO created at %s", isoDestination)

	return isoDestination, nil
}

// Dumps the userdata and the metadata into two dedicated yaml files.
// The files are created inside of a temporary directory
// Returns a string containing the name of the temporary directory and an error
// object
func (ci *defCloudInit) createFiles() (string, error) {
	log.Print("Creating ISO contents")
	tmpDir, err := ioutil.TempDir("", "cloudinit")
	if err != nil {
		return "", fmt.Errorf("Cannot create tmp directory for cloudinit ISO generation: %s",
			err)
	}

	// Create files required by ISO file
	userdata := ""
	if len(ci.UserDataRaw) > 0 {
		userdata = ci.UserDataRaw
	} else {
		userdata = "#cloud-config\n"
	}

	// append the extra user data flags
	if userdata_extra, err := yaml.Marshal(&ci.UserData); err != nil {
		return "", fmt.Errorf("Error dumping cloudinit's user data: %s", err)
	} else {
		userdata = fmt.Sprintf("%s\n%s", userdata, string(userdata_extra))
	}

	if err = ioutil.WriteFile(
		filepath.Join(tmpDir, USERDATA),
		[]byte(userdata),
		os.ModePerm); err != nil {
		return "", fmt.Errorf("Error while writing user-data to file: %s", err)
	}

	metadata, err := yaml.Marshal(&ci.Metadata)
	if err != nil {
		return "", fmt.Errorf("Error dumping cloudinit's meta data: %s", err)
	}
	if err = ioutil.WriteFile(filepath.Join(tmpDir, METADATA), metadata, os.ModePerm); err != nil {
		return "", fmt.Errorf("Error while writing meta-data to file: %s", err)
	}

	log.Print("ISO contents created")

	return tmpDir, nil
}

// Creates a new defCloudInit object starting from a ISO volume handled by
// libvirt
func newCloudInitDefFromRemoteISO(virConn *libvirt.VirConnection, id string) (defCloudInit, error) {
	ci := defCloudInit{}

	key, err := getCloudInitVolumeKeyFromTerraformID(id)
	if err != nil {
		return ci, err
	}

	volume, err := virConn.LookupStorageVolByKey(key)
	if err != nil {
		return ci, fmt.Errorf("Can't retrieve volume %s", key)
	}
	defer volume.Free()

	ci.Name, err = volume.GetName()
	if err != nil {
		return ci, fmt.Errorf("Error retrieving volume name: %s", err)
	}

	volPool, err := volume.LookupPoolByVolume()
	if err != nil {
		return ci, fmt.Errorf("Error retrieving pool for volume: %s", err)
	}
	defer volPool.Free()

	ci.PoolName, err = volPool.GetName()
	if err != nil {
		return ci, fmt.Errorf("Error retrieving pool name: %s", err)
	}

	file, err := downloadISO(virConn, volume)
	if file != nil {
		defer os.Remove(file.Name())
		defer file.Close()
	}
	if err != nil {
		return ci, err
	}

	// read ISO contents
	isoReader, err := iso9660.NewReader(file)
	if err != nil {
		return ci, fmt.Errorf("Error initializing ISO reader: %s", err)
	}

	for {
		f, err := isoReader.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return ci, err
		}

		log.Printf("ISO reader: processing file %s", f.Name())

		//TODO: the iso9660 has a bug...
		if f.Name() == "/user_dat." {
			data, err := ioutil.ReadAll(f.Sys().(io.Reader))
			if err != nil {
				return ci, fmt.Errorf("Error while reading %s: %s", USERDATA, err)
			}
			if err := yaml.Unmarshal(data, &ci.UserData); err != nil {
				return ci, fmt.Errorf("Error while unmarshalling user-data: %s", err)
			}
		}

		//TODO: the iso9660 has a bug...
		if f.Name() == "/meta_dat." {
			data, err := ioutil.ReadAll(f.Sys().(io.Reader))
			if err != nil {
				return ci, fmt.Errorf("Error while reading %s: %s", METADATA, err)
			}
			if err := yaml.Unmarshal(data, &ci.Metadata); err != nil {
				return ci, fmt.Errorf("Error while unmarshalling user-data: %s", err)
			}
		}
	}

	log.Printf("Read cloud-init from file: %+v", ci)

	return ci, nil
}

// Downloads the ISO identified by `key` to a local tmp file.
// Returns a pointer to the ISO file. Note well: you have to close this file
// pointer when you are done.
func downloadISO(virConn *libvirt.VirConnection, volume libvirt.VirStorageVol) (*os.File, error) {
	// get Volume info (required to get size later)
	info, err := volume.GetInfo()
	if err != nil {
		return nil, fmt.Errorf("Error retrieving info for volume: %s", err)
	}

	// create tmp file for the ISO
	file, err := ioutil.TempFile("", "cloudinit")
	if err != nil {
		return nil, fmt.Errorf("Cannot create tmp file: %s", err)
	}

	// download ISO file
	stream, err := libvirt.NewVirStream(virConn, 0)
	if err != nil {
		return file, err
	}
	defer stream.Close()

	volume.Download(stream, 0, info.GetCapacityInBytes(), 0)

	n, err := io.Copy(file, stream)
	if err != nil {
		return file, fmt.Errorf("Error while copying remote volume to local disk: %s", err)
	}
	file.Seek(0, 0)
	log.Printf("%d bytes downloaded", n)

	return file, nil
}
