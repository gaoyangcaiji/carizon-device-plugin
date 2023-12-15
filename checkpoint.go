package main

import (
	"encoding/json"
	"io/ioutil"

	"carizon-device-plugin/pkg/logger"
)

const (
	checkPointfile = "/var/lib/kubelet/device-plugins/kubelet_internal_checkpoint"
)

// PodDevicesEntry devices info map
type PodDevicesEntry struct {
	PodUID        string
	ContainerName string
	ResourceName  string
	DeviceIDs     []string
	AllocResp     []byte
}

// checkpointData ...
type checkpointData struct {
	PodDeviceEntries  []PodDevicesEntry
	RegisteredDevices map[string][]string
}

// checkpointFileData ...
type checkpointFileData struct {
	Data     checkpointData
	Checksum uint64
}

// checkpoint ...
type checkpoint struct {
	fileName   string
	podEntires []PodDevicesEntry
}

// GetCheckpoint get checkpoint
func GetCheckpoint() (ResourceClient, error) {
	logger.Wrapper.Infof("GetCheckpoint(): invoked")
	return getCheckpoint(checkPointfile)
}

// getCheckpoint get checkpoint from PodEntries
func getCheckpoint(filePath string) (ResourceClient, error) {
	cp := &checkpoint{fileName: filePath}
	err := cp.getPodEntries()
	if err != nil {
		return nil, err
	}
	return cp, nil
}

// getPodEntries ...
func (cp *checkpoint) getPodEntries() error {

	cpd := &checkpointFileData{}
	rawBytes, err := ioutil.ReadFile(cp.fileName)
	if err != nil {
		logger.Wrapper.Errorf("[getPodEntries]: error reading files %s\n%v\n", checkPointfile, err)
		return err
	}

	if err = json.Unmarshal(rawBytes, cpd); err != nil {
		logger.Wrapper.Errorf("[getPodEntries]: error unmarshall bytes %v", err)
		return err
	}

	cp.podEntires = cpd.Data.PodDeviceEntries
	return nil
}

// GetPodResourceMap ...
func (cp *checkpoint) GetPodResourceMap() (map[string]*ResourceInfo, error) {
	resourceMap := make(map[string]*ResourceInfo)

	for _, pod := range cp.podEntires {
		entry, ok := resourceMap[pod.ResourceName]
		if ok {
			// already exists; append to it
			entry.DeviceIDs = append(entry.DeviceIDs, pod.DeviceIDs...)
		} else {
			// new entry
			resourceMap[pod.ResourceName] = &ResourceInfo{DeviceIDs: pod.DeviceIDs}
		}
	}

	return resourceMap, nil
}
