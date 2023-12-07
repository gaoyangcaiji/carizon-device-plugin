package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"carizon-device-plugin/logger"

	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

const (
	horizonDevicePcie        = "PCIe"
	horizonDevicePcieInfoEnv = "PCIE_INFO"
	horizonPcieFlagEnv       = "PCIE_FLAG"
	healthCheckEnv           = "DEVICE_HEALTH_CHECK"
	dmServerEnv              = "DEVICE_MANAGER_PORT"
	resourceDomain           = "hobot.cc/"
	resourceCount            = resourceDomain + horizonDevicePcie
	healthCheckInterval      = 10
	chipTypeJ5               = "J5"
	chipType2J5              = "2J5"
	dmAPI                    = "/api/dev_manager/v1/"
	fetchDevicesAPI          = dmAPI + "list"
	fetchDeviceTypesAPI      = dmAPI + "devtypes"
	checkHealthAPI           = dmAPI + "device/%d/healthy"
	allocateDeviceAPI        = dmAPI + "allocate"
	getAllocateDeviceInfoAPI = dmAPI + "list/devices"
)

// DeviceOffline represents offline status of the deivce
var DeviceOffline = 1

// DMServer address of the device-manager server
var DMServer = getDeviceManagerServerAddr()

// ResourceManager interface of the device resource maanger
type ResourceManager interface {
	Devices() []*Device
	CheckHealth(stop <-chan interface{}, devices []*Device, healthy, unhealthy chan<- *Device)
	GetAllocateDevicesInfo(deviceIPs []string) (info *[]PCIeAddressInfo, err error)
	Allocate(deviceIPs []string)
}

// HorizonDeviceManager horzion device manager
type HorizonDeviceManager struct {
	deviceType string
}

func getDeviceManagerServerAddr() string {
	server := strings.ToLower(strings.TrimSpace(os.Getenv(dmServerEnv)))
	if strings.HasPrefix(server, "http://") {
		return server
	}
	server = "http://" + strings.TrimPrefix(server, "tcp://")
	return server
}

// NewHorizonDeviceManager returns a new instance of HorizonDeviceManager
func NewHorizonDeviceManager(deviceType string) *HorizonDeviceManager {
	return &HorizonDeviceManager{deviceType: deviceType}
}

// Devices returns all devices
func (h *HorizonDeviceManager) Devices() []*Device {
	var devs []*Device

	devices, err := getDevices(h.deviceType)
	checkErr(err)

	for _, d := range devices {
		devs = append(devs, buildDevice(d))
	}

	return devs
}

// Allocate the device to job,then other job can not use the device
func (h *HorizonDeviceManager) Allocate(deviceIPs []string) {
	logger.Wrapper.Infof("Allocate deviceIPs: %+v", deviceIPs)
	reqData := AllocateDeviceReq{DeviceIP: strings.Join(deviceIPs, ","), Allocated: true}
	allocateDevice(reqData)
}

// GetAllocateDevicesInfo is get device ip and how many pcie occupied
func (h *HorizonDeviceManager) GetAllocateDevicesInfo(deviceIPs []string) (info *[]PCIeAddressInfo, err error) {
	data := map[string][]string{"ips": deviceIPs}
	resp, err := HTTPClient.R().SetBody(data).Post(DMServer + getAllocateDeviceInfoAPI)

	_, _, err = checkHTTPResponse(resp, err)
	if err != nil {
		return nil, err
	}

	var result *HTTPRetAllocateDevicesInfo
	err = json.Unmarshal(resp.Body(), &result)
	if err != nil {
		return nil, err
	}

	logger.Wrapper.Infof("result is %+v", result)
	return &result.Data, nil
}

func buildDevice(d *externalDevice) *Device {
	dev := Device{}
	// use device IP as the unique indication of the device
	dev.ID = d.IP
	dev.IP = d.IP
	dev.UUID = d.UUID
	dev.Health = pluginapi.Healthy
	return &dev
}

// CheckHealth checks health of the devices
func (h *HorizonDeviceManager) CheckHealth(stop <-chan interface{}, devices []*Device, healthy, unhealthy chan<- *Device) {
	checkHealth(stop, devices, healthy, unhealthy)
}

func checkHealth(stop <-chan interface{}, devices []*Device, healthy, unhealthy chan<- *Device) {
	healthCheck := strings.ToLower(os.Getenv(healthCheckEnv))
	if healthCheck == "false" {
		log.Printf("Disable device health checks")
		return
	}

	for {
		select {
		case <-stop:
			return
		default:
		}

		for _, d := range devices {
			if !isDeviceHealthy(d.UUID) {
				if d.Health == pluginapi.Healthy {
					log.Printf("Device %s become unhealthy", d.IP)
					unhealthy <- d
				}
			} else {
				if d.Health == pluginapi.Unhealthy {
					log.Printf("Device %s become healthy", d.IP)
					healthy <- d
				}
			}
		}
		time.Sleep(healthCheckInterval * time.Second)
	}
}

func getDevices(deviceType string) ([]*externalDevice, error) {
	eDevices := []*externalDevice{}
	nodeName, err := os.Hostname()
	if err != nil {
		log.Printf("Error. Fail to get hostname: %s", err.Error())
		return nil, err
	}

	log.Printf("Query %s devices on node: %s", deviceType, nodeName)

	params := map[string]string{"bind_node": nodeName, "chip_type": deviceType}
	req := HTTPClient.R().SetQueryParams(params)
	resp, err := req.Get(DMServer + fetchDevicesAPI)
	_, _, err = checkHTTPResponse(resp, err)
	if err != nil {
		log.Printf("Error. Failed to query %s bind devices on node %s. %v", deviceType, nodeName, err)
		return eDevices, nil
	}

	var ret *HTTPRetDeviceList
	_ = json.Unmarshal(resp.Body(), &ret)
	if v, ok := ret.Data["list"]; ok {
		for _, d := range v {
			if d.Status == int8(DeviceOffline) {
				continue
			}
			eDevice := externalDevice{}
			eDevice.UUID = d.ID
			eDevice.IP = d.IP
			eDevices = append(eDevices, &eDevice)
		}
	}
	return eDevices, nil
}

func isDeviceHealthy(uuid int) bool {
	healthy := true
	resp, err := HTTPClient.R().Get(DMServer + fmt.Sprintf(checkHealthAPI, uuid))
	_, _, err = checkHTTPResponse(resp, err)
	if err != nil {
		log.Printf("Error. Failed to get device healthy. %v", err.Error())
		// Default to true
		return healthy
	}

	var ret *HTTPRetBool
	_ = json.Unmarshal(resp.Body(), &ret)
	healthy = ret.Data

	return healthy
}

func getDeviceTypes() ([]string, error) {
	types := []string{}
	resp, err := HTTPClient.R().Get(DMServer + fetchDeviceTypesAPI)
	_, _, err = checkHTTPResponse(resp, err)
	if err != nil {
		log.Printf("Error. Failed to get device types. %v", err.Error())
		return types, err
	}

	var ret *HTTPRetDeviceTypes
	_ = json.Unmarshal(resp.Body(), &ret)
	if v, ok := ret.Data["types"]; ok {
		types = v
	}

	return types, nil
}

func allocateDevice(data AllocateDeviceReq) {
	resp, err := HTTPClient.R().SetBody(data).Post(DMServer + allocateDeviceAPI)

	_, _, err = checkHTTPResponse(resp, err)

	if err != nil {
		logger.Wrapper.Errorf("Failed to allocate device.Err: %+v,DeviceIP: %s", err, data.DeviceIP)
	}

}
