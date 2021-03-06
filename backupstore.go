package backupstore

import (
	"fmt"
	"net/url"

	"github.com/rancher/backupstore/util"
)

type Volume struct {
	Name           string
	Size           int64 `json:",string"`
	CreatedTime    string
	LastBackupName string
	LastBackupAt   string
	BlockCount     int64 `json:",string"`
}

type Snapshot struct {
	Name        string
	CreatedTime string
}

type Backup struct {
	Name              string
	VolumeName        string
	SnapshotName      string
	SnapshotCreatedAt string
	CreatedTime       string
	Size              int64 `json:",string"`
	Labels            map[string]string

	Blocks     []BlockMapping `json:",omitempty"`
	SingleFile BackupFile     `json:",omitempty"`
}

var (
	backupstoreBase = "backupstore"
)

func SetBackupstoreBase(base string) {
	backupstoreBase = base
}

func GetBackupstoreBase() string {
	return backupstoreBase
}

func addVolume(volume *Volume, driver BackupStoreDriver) error {
	if volumeExists(volume.Name, driver) {
		return nil
	}

	if !util.ValidateName(volume.Name) {
		return fmt.Errorf("Invalid volume name %v", volume.Name)
	}

	if err := saveVolume(volume, driver); err != nil {
		log.Error("Fail add volume ", volume.Name)
		return err
	}
	log.Debug("Added backupstore volume ", volume.Name)

	return nil
}

func removeVolume(volumeName string, driver BackupStoreDriver) error {
	if !util.ValidateName(volumeName) {
		return fmt.Errorf("Invalid volume name %v", volumeName)
	}

	if !volumeExists(volumeName, driver) {
		return fmt.Errorf("Volume %v doesn't exist in backupstore", volumeName)
	}

	volumeDir := getVolumePath(volumeName)
	if err := driver.Remove(volumeDir); err != nil {
		return err
	}
	log.Debug("Removed volume directory in backupstore: ", volumeDir)
	log.Debug("Removed backupstore volume ", volumeName)

	return nil
}

func encodeBackupURL(backupName, volumeName, destURL string) string {
	v := url.Values{}
	v.Add("volume", volumeName)
	v.Add("backup", backupName)
	return destURL + "?" + v.Encode()
}

func decodeBackupURL(backupURL string) (string, string, error) {
	u, err := url.Parse(backupURL)
	if err != nil {
		return "", "", err
	}
	v := u.Query()
	volumeName := v.Get("volume")
	backupName := v.Get("backup")
	if !util.ValidateName(volumeName) || !util.ValidateName(backupName) {
		return "", "", fmt.Errorf("Invalid name parsed, got %v and %v", backupName, volumeName)
	}
	return backupName, volumeName, nil
}

func LoadVolume(backupURL string) (*Volume, error) {
	_, volumeName, err := decodeBackupURL(backupURL)
	if err != nil {
		return nil, err
	}
	driver, err := GetBackupStoreDriver(backupURL)
	if err != nil {
		return nil, err
	}
	return loadVolume(volumeName, driver)
}

func GetBackupFromBackupURL(backupURL string) (string, error) {
	backup, _, err := decodeBackupURL(backupURL)
	return backup, err
}

func GetVolumeFromBackupURL(backupURL string) (string, error) {
	_, volume, err := decodeBackupURL(backupURL)
	return volume, err
}
