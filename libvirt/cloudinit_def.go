package libvirt

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	libvirt "github.com/digitalocean/go-libvirt"
	diskfs "github.com/diskfs/go-diskfs"
	"github.com/diskfs/go-diskfs/disk"
	"github.com/diskfs/go-diskfs/filesystem"
	"github.com/diskfs/go-diskfs/filesystem/iso9660"
	"github.com/google/uuid"
	oldIso9660 "github.com/hooklift/iso9660"
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
		log.Printf("error while removing tmp directory holding the ISO file: %s", err)
	}

}

func (ci *defCloudInit) UploadIso(client *Client, iso string) (string, error) {
	virConn := client.libvirt
	if virConn == nil {
		return "", fmt.Errorf(LibVirtConIsNil)
	}

	pool, err := virConn.StoragePoolLookupByName(ci.PoolName)
	if err != nil {
		return "", fmt.Errorf("can't find storage pool '%s'", ci.PoolName)
	}

	client.poolMutexKV.Lock(ci.PoolName)
	defer client.poolMutexKV.Unlock(ci.PoolName)

	// Refresh the pool of the volume so that libvirt knows it is
	// not longer in use.
	err = waitForSuccess("error refreshing pool for volume", func() error {
		return virConn.StoragePoolRefresh(pool, 0)
	})
	if err != nil {
		return "", err
	}

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
		return "", fmt.Errorf("error serializing libvirt volume: %s", err)
	}

	// create the volume
	volume, err := virConn.StorageVolCreateXML(pool, string(volumeDefXML), 0)
	if err != nil {
		return "", fmt.Errorf("error creating libvirt volume for cloudinit device %s: %s", ci.Name, err)
	}

	// upload ISO file
	err = img.Import(newCopier(virConn, &volume, uint64(size)), volumeDef)
	if err != nil {
		return "", fmt.Errorf("error while uploading cloudinit %s: %s", img.String(), err)
	}

	if volume.Key == "" {
		return "", fmt.Errorf("error retrieving volume key")
	}

	return ci.buildTerraformKey(volume.Key), nil
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
	isoDisk, err := diskfs.Create(isoDestination, 10*1024*1024, diskfs.Raw)
	if err != nil {
		return "", err
	}
	isoDisk.LogicalBlocksize = 2048
	fspec := disk.FilesystemSpec{Partition: 0, FSType: filesystem.TypeISO9660, VolumeLabel: "cidata"}
	fs, err := isoDisk.CreateFilesystem(fspec)
	if err != nil {
		return "", err
	}
	for _, s := range []struct {
		name string
		path string
	}{
		{name: "user-data", path: userDataFileName},
		{name: "meta-data", path: metaDataFileName},
		{name: "network-config", path: networkConfigFileName},
	} {
		fi, err := os.Open(s.path)
		if err != nil {
			return "", err
		}
		rw, err := fs.OpenFile(s.name, os.O_CREATE|os.O_RDWR)
		if err != nil && err != io.EOF {
			return "", err
		}
		buf := make([]byte, 1024)
		for {
			n, err := fi.Read(buf)
			if err != nil && err != io.EOF {
				return "", err
			}
			if n == 0 {
				break
			}

			_, err = rw.Write(buf[:n])
			if err != nil {
				return "", err
			}
		}
		err = fi.Close()
		if err != nil {
			return "", err
		}
	}
	iso, ok := fs.(*iso9660.FileSystem)
	if !ok {
		return "", fmt.Errorf("not an iso9660 filesystem")
	}
	err = iso.Finalize(iso9660.FinalizeOptions{
		RockRidge:        true,
		VolumeIdentifier: "cidata",
	})
	if err != nil {
		return "", err
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
		return "", fmt.Errorf("cannot create tmp directory for cloudinit ISO generation: %s",
			err)
	}
	// user-data
	if err = ioutil.WriteFile(filepath.Join(tmpDir, userDataFileName), []byte(ci.UserData), os.ModePerm); err != nil {
		return "", fmt.Errorf("error while writing user-data to file: %s", err)
	}
	// meta-data
	if err = ioutil.WriteFile(filepath.Join(tmpDir, metaDataFileName), []byte(ci.MetaData), os.ModePerm); err != nil {
		return "", fmt.Errorf("error while writing meta-data to file: %s", err)
	}
	// network-config
	if err = ioutil.WriteFile(filepath.Join(tmpDir, networkConfigFileName), []byte(ci.NetworkConfig), os.ModePerm); err != nil {
		return "", fmt.Errorf("error while writing network-config to file: %s", err)
	}

	log.Print("ISO contents created")

	return tmpDir, nil
}

// Creates a new defCloudInit object starting from a ISO volume handled by
// libvirt
func newCloudInitDefFromRemoteISO(virConn *libvirt.Libvirt, id string) (defCloudInit, error) {
	ci := defCloudInit{}

	key, err := getCloudInitVolumeKeyFromTerraformID(id)
	if err != nil {
		return ci, err
	}

	volume, err := virConn.StorageVolLookupByKey(key)
	if err != nil {
		return ci, fmt.Errorf("can't retrieve volume %s: %v", key, err)
	}

	err = ci.setCloudInitDiskNameFromExistingVol(virConn, volume)
	if err != nil {
		return ci, err
	}

	err = ci.setCloudInitPoolNameFromExistingVol(virConn, volume)
	if err != nil {
		return ci, err
	}

	isoFile, err := downloadISO(virConn, volume)
	if isoFile != nil {
		defer os.Remove(isoFile.Name())
		defer isoFile.Close()
	}
	if err != nil {
		return ci, err
	}

	err = ci.setCloudInitDataFromExistingCloudInitDisk(virConn, isoFile)
	if err != nil {
		return ci, err
	}
	return ci, nil
}

// setCloudInitDataFromExistingCloudInitDisk read and set UserData, MetaData, and NetworkConfig from existing CloudInitDisk
func (ci *defCloudInit) setCloudInitDataFromExistingCloudInitDisk(virConn *libvirt.Libvirt, isoFile *os.File) error {
	isoReader, err := oldIso9660.NewReader(isoFile)
	if err != nil {
		return fmt.Errorf("error initializing ISO reader: %s", err)
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
			ci.UserData = string(dataBytes)
		}
		if file.Name() == "/meta_dat." {
			ci.MetaData = string(dataBytes)
		}
		if file.Name() == "/network_." {
			ci.NetworkConfig = string(dataBytes)
		}
	}
	log.Printf("[DEBUG]: Read cloud-init from file: %+v", ci)
	return nil
}

// FIXME Consider doing this inline.
// setCloudInitPoolNameFromExistingVol retrieve poolname from an existing CloudInitDisk
func (ci *defCloudInit) setCloudInitPoolNameFromExistingVol(virConn *libvirt.Libvirt, volume libvirt.StorageVol) error {
	volPool, err := virConn.StoragePoolLookupByVolume(volume)
	if err != nil {
		return fmt.Errorf("error retrieving pool for cloudinit volume: %s", err)
	}

	if volPool.Name == "" {
		return fmt.Errorf("error retrieving pool name for cloudinit volume: %s", volume.Name)
	}
	ci.PoolName = volPool.Name
	return nil
}

// FIXME Consider doing this inline.
// setCloudInitDisklNameFromVol retrieve CloudInitname from an existing CloudInitDisk
func (ci *defCloudInit) setCloudInitDiskNameFromExistingVol(virConn *libvirt.Libvirt, volume libvirt.StorageVol) error {
	if volume.Name == "" {
		return fmt.Errorf("error retrieving cloudinit volume name for volume key: %s", volume.Key)
	}
	ci.Name = volume.Name
	return nil
}

func readIso9660File(file os.FileInfo) ([]byte, error) {
	log.Printf("ISO reader: processing file %s", file.Name())

	dataBytes, err := ioutil.ReadAll(file.Sys().(io.Reader))
	if err != nil {
		return nil, fmt.Errorf("error while reading %s: %s", file.Name(), err)
	}
	return dataBytes, nil
}

// Downloads the ISO identified by `key` to a local tmp file.
// Returns a pointer to the ISO file. Note well: you have to close this file
// pointer when you are done.
func downloadISO(virConn *libvirt.Libvirt, volume libvirt.StorageVol) (*os.File, error) {
	// get Volume info (required to get size later)
	_, size, _, err := virConn.StorageVolGetInfo(volume)
	if err != nil {
		return nil, fmt.Errorf("error retrieving info for volume: %s", err)
	}

	// create tmp file for the ISO
	tmpFile, err := ioutil.TempFile("", "cloudinit")
	if err != nil {
		return nil, fmt.Errorf("cannot create tmp file: %s", err)
	}

	w := bufio.NewWriterSize(tmpFile, int(size))

	// download ISO file
	if err := virConn.StorageVolDownload(volume, w, 0, size, 0); err != nil {
		return tmpFile, fmt.Errorf("error while downloading volume: %s", err)
	}

	bytesCopied := w.Buffered()
	err = w.Flush()
	if err != nil {
		return tmpFile, fmt.Errorf("error while copying remote volume to local disk: %s", err)
	}

	log.Printf("%d bytes downloaded", bytesCopied)
	if uint64(bytesCopied) != size {
		return tmpFile, fmt.Errorf("error while copying remote volume to local disk, bytesCopied %d !=  %d volume.size", bytesCopied, size)
	}

	tmpFile.Seek(0, 0)

	return tmpFile, nil
}
