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

	"github.com/google/uuid"
	"github.com/hooklift/iso9660"
	libvirt "github.com/libvirt/libvirt-go"
)

const userDataFileName string = "user-data"
const metaDataFileName string = "meta-data"
const networkConfigFileName string = "network-config"

type defCloudInit struct {
	Name          string
	PoolName      string
	MetaData      string `yaml:"meta_data"`
	UserData      string `yaml:"user_data"`
	NetworkConfig string `yaml:"network_config"`
}

func newCloudInitDef() defCloudInit {
	return defCloudInit{}
}

// Create a ISO file based on the contents of the CloudInit instance and
// uploads it to the libVirt pool
// Returns a string holding terraform's internal ID of this resource
func (ci *defCloudInit) CreateIso() (string, error) {
	iso, err := ci.createISO()
	if err != nil {
		return "", err
	}
	return iso, err
}

func removeTmpIsoDirectory(iso string) {

	err := os.RemoveAll(filepath.Dir(iso))
	if err != nil {
		log.Printf("Error while removing tmp directory holding the ISO file: %s", err)
	}

}

func (ci *defCloudInit) UploadIso(client *Client, iso string) (string, error) {

	pool, err := client.libvirt.LookupStoragePoolByName(ci.PoolName)
	if err != nil {
		return "", fmt.Errorf("can't find storage pool '%s'", ci.PoolName)
	}
	defer pool.Free()

	client.poolMutexKV.Lock(ci.PoolName)
	defer client.poolMutexKV.Unlock(ci.PoolName)

	// Refresh the pool of the volume so that libvirt knows it is
	// not longer in use.
	waitForSuccess("Error refreshing pool for volume", func() error {
		return pool.Refresh(0)
	})

	volumeDef := newDefVolume()
	volumeDef.Name = ci.Name

	// an existing image was given, this mean we can't choose size
	img, err := newImage(iso)
	if err != nil {
		return "", err
	}

	defer removeTmpIsoDirectory(iso)

	size, err := img.Size()
	if err != nil {
		return "", err
	}

	volumeDef.Capacity.Unit = "B"
	volumeDef.Capacity.Value = size
	volumeDef.Target.Format.Type = "raw"

	volumeDefXML, err := xml.Marshal(volumeDef)
	if err != nil {
		return "", fmt.Errorf("Error serializing libvirt volume: %s", err)
	}

	// create the volume
	volume, err := pool.StorageVolCreateXML(string(volumeDefXML), 0)
	if err != nil {
		return "", fmt.Errorf("Error creating libvirt volume for cloudinit device %s: %s", ci.Name, err)
	}
	defer volume.Free()

	// upload ISO file
	err = img.Import(newCopier(client.libvirt, volume, uint64(size)), volumeDef)
	if err != nil {
		return "", fmt.Errorf("Error while uploading cloudinit %s: %s", img.String(), err)
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
	return fmt.Sprintf("%s;%s", volumeKey, uuid.New())
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
		"mkisofs",
		"-output",
		isoDestination,
		"-volid",
		"cidata",
		"-joliet",
		"-rock",
		filepath.Join(tmpDir, userDataFileName),
		filepath.Join(tmpDir, metaDataFileName),
		filepath.Join(tmpDir, networkConfigFileName))

	log.Printf("About to execute cmd: %+v", cmd)
	if err = cmd.Run(); err != nil {
		return "", fmt.Errorf("Error while starting the creation of CloudInit's ISO image: %s", err)
	}
	log.Printf("ISO created at %s", isoDestination)

	return isoDestination, nil
}

// write user-data,  meta-data network-config in tmp files and dedicated directory
// Returns a string containing the name of the temporary directory and an error
// object
func (ci *defCloudInit) createFiles() (string, error) {
	log.Print("Creating ISO contents")
	tmpDir, err := ioutil.TempDir("", "cloudinit")
	if err != nil {
		return "", fmt.Errorf("Cannot create tmp directory for cloudinit ISO generation: %s",
			err)
	}
	// user-data
	if err = ioutil.WriteFile(filepath.Join(tmpDir, userDataFileName), []byte(ci.UserData), os.ModePerm); err != nil {
		return "", fmt.Errorf("Error while writing user-data to file: %s", err)
	}
	// meta-data
	if err = ioutil.WriteFile(filepath.Join(tmpDir, metaDataFileName), []byte(ci.MetaData), os.ModePerm); err != nil {
		return "", fmt.Errorf("Error while writing meta-data to file: %s", err)
	}
	// network-config
	if err = ioutil.WriteFile(filepath.Join(tmpDir, networkConfigFileName), []byte(ci.NetworkConfig), os.ModePerm); err != nil {
		return "", fmt.Errorf("Error while writing network-config to file: %s", err)
	}

	log.Print("ISO contents created")

	return tmpDir, nil
}

// Creates a new defCloudInit object starting from a ISO volume handled by
// libvirt
func newCloudInitDefFromRemoteISO(virConn *libvirt.Connect, id string) (defCloudInit, error) {
	ci := defCloudInit{}

	key, err := getCloudInitVolumeKeyFromTerraformID(id)
	if err != nil {
		return ci, err
	}

	volume, err := virConn.LookupStorageVolByKey(key)
	if err != nil {
		return ci, fmt.Errorf("Can't retrieve volume %s: %v", key, err)
	}
	defer volume.Free()

	err = ci.setCloudInitDiskNameFromExistingVol(volume)
	if err != nil {
		return ci, err
	}

	err = ci.setCloudInitPoolNameFromExistingVol(volume)
	if err != nil {
		return ci, err
	}

	isoFile, err := downloadISO(virConn, *volume)
	if isoFile != nil {
		defer os.Remove(isoFile.Name())
		defer isoFile.Close()
	}
	if err != nil {
		return ci, err
	}

	err = ci.setCloudInitDataFromExistingCloudInitDisk(virConn, volume, isoFile)
	if err != nil {
		return ci, err
	}
	return ci, nil
}

// setCloudInitDataFromExistingCloudInitDisk read and set UserData, MetaData, and NetworkConfig from existing CloudInitDisk
func (ci *defCloudInit) setCloudInitDataFromExistingCloudInitDisk(virConn *libvirt.Connect, volume *libvirt.StorageVol, isoFile *os.File) error {
	isoReader, err := iso9660.NewReader(isoFile)
	if err != nil {
		return fmt.Errorf("Error initializing ISO reader: %s", err)
	}

	for {
		file, err := isoReader.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}
		dataBytes, err := readIso9660File(file)
		if err != nil {
			return err
		}
		// the following filenames need to be like this because in the ios9660 reader
		// joliet is not supported. https://github.com/hooklift/iso9660/blob/master/README.md#not-supported
		if file.Name() == "/user_dat." {
			ci.UserData = fmt.Sprintf("%s", dataBytes)
		}
		if file.Name() == "/meta_dat." {
			ci.MetaData = fmt.Sprintf("%s", dataBytes)
		}
		if file.Name() == "/network_." {
			ci.NetworkConfig = fmt.Sprintf("%s", dataBytes)
		}
	}
	log.Printf("[DEBUG]: Read cloud-init from file: %+v", ci)
	return nil
}

// setCloudInitPoolNameFromExistingVol retrieve poolname from an existing CloudInitDisk
func (ci *defCloudInit) setCloudInitPoolNameFromExistingVol(volume *libvirt.StorageVol) error {
	volPool, err := volume.LookupPoolByVolume()
	if err != nil {
		return fmt.Errorf("Error retrieving pool for cloudinit volume: %s", err)
	}
	defer volPool.Free()

	ci.PoolName, err = volPool.GetName()
	if err != nil {
		return fmt.Errorf("Error retrieving pool name: %s", err)
	}
	return nil
}

// setCloudInitDisklNameFromVol retrieve CloudInitname from an existing CloudInitDisk
func (ci *defCloudInit) setCloudInitDiskNameFromExistingVol(volume *libvirt.StorageVol) error {
	var err error
	ci.Name, err = volume.GetName()
	if err != nil {
		return fmt.Errorf("Error retrieving cloudinit volume name: %s", err)
	}
	return nil
}

func readIso9660File(file os.FileInfo) ([]byte, error) {
	log.Printf("ISO reader: processing file %s", file.Name())

	dataBytes, err := ioutil.ReadAll(file.Sys().(io.Reader))
	if err != nil {
		return nil, fmt.Errorf("Error while reading %s: %s", file.Name(), err)
	}
	return dataBytes, nil
}

// Downloads the ISO identified by `key` to a local tmp file.
// Returns a pointer to the ISO file. Note well: you have to close this file
// pointer when you are done.
func downloadISO(virConn *libvirt.Connect, volume libvirt.StorageVol) (*os.File, error) {
	// get Volume info (required to get size later)
	var bytesCopied int64

	info, err := volume.GetInfo()
	if err != nil {
		return nil, fmt.Errorf("Error retrieving info for volume: %s", err)
	}

	// create tmp file for the ISO
	tmpFile, err := ioutil.TempFile("", "cloudinit")
	if err != nil {
		return nil, fmt.Errorf("Cannot create tmp file: %s", err)
	}

	// download ISO file
	stream, err := virConn.NewStream(0)
	if err != nil {
		return tmpFile, err
	}

	defer func() {
		stream.Free()
	}()

	err = volume.Download(stream, 0, info.Capacity, 0)
	if err != nil {
		stream.Abort()
		return tmpFile, fmt.Errorf("Error by downloading content to libvirt volume:%s", err)
	}
	sio := NewStreamIO(*stream)

	bytesCopied, err = io.Copy(tmpFile, sio)
	if err != nil {
		return tmpFile, fmt.Errorf("Error while copying remote volume to local disk: %s", err)
	}

	if uint64(bytesCopied) != info.Capacity {
		stream.Abort()
		return tmpFile, fmt.Errorf("Error while copying remote volume to local disk, bytesCopied %d !=  %d volume.size", bytesCopied, info.Capacity)
	}

	err = stream.Finish()
	if err != nil {
		stream.Abort()
		return tmpFile, fmt.Errorf("Error by terminating libvirt stream %s", err)
	}

	tmpFile.Seek(0, 0)
	log.Printf("%d bytes downloaded", bytesCopied)

	return tmpFile, nil
}
