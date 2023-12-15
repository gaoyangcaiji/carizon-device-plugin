package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"carizon-device-plugin/pkg/logger"

	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

const (
	CarizonDevicePcie        = "PCIe"
	CarizonDevicePcieInfoEnv = "PCIE_INFO"
	CarizonPcieFlagEnv       = "PCIE_FLAG"
	healthCheckEnv           = "DEVICE_HEALTH_CHECK"
	cmdbServerEnv            = "BK_CMDB_CHART_PORT"
	resourceDomain           = "carizon/"
	resourceCount            = resourceDomain + CarizonDevicePcie
	healthCheckInterval      = 10
	cmdbAPI                  = "/api/v3/"
	findInstassociationAPI   = cmdbAPI + "find/instassociation"
	searchObjectInstsAPI     = cmdbAPI + "search/instances/object/%s"
	checkHealthAPI           = cmdbAPI + "device/%d/healthy"
	batchUpdateInstsAPI      = cmdbAPI + "/updatemany/instance/object/%s"
	getAllocateDeviceInfoAPI = cmdbAPI + "list/devices"
)

// DeviceOffline represents offline status of the deivce
var DeviceOffline = 1

// CmdbServer address of the cmdb server
var CmdbServer = getCmdbServerAddr()

// ResourceManager interface of the device resource maanger
type ResourceManager interface {
	Devices() []*Device
	CheckHealth(stop <-chan interface{}, devices []*Device, healthy, unhealthy chan<- *Device)
	GetAllocateDevicesInfo(deviceIPs []string) (info *[]PCIeAddressInfo, err error)
	Allocate(deviceIPs []string)
}

// CarizonDeviceManager horzion device manager
type CarizonDeviceManager struct {
	deviceType string
}

func getCmdbServerAddr() string {
	server := strings.ToLower(strings.TrimSpace(os.Getenv(cmdbServerEnv)))
	if strings.HasPrefix(server, "http://") {
		return server
	}
	server = "http://" + strings.TrimPrefix(server, "tcp://")
	return server
}

// NewCarizonDeviceManager returns a new instance of CarizonDeviceManager
func NewCarizonDeviceManager(deviceType string) *CarizonDeviceManager {
	return &CarizonDeviceManager{deviceType: deviceType}
}

// Devices returns all devices
func (c *CarizonDeviceManager) Devices() []*Device {
	var devs []*Device

	devices, err := getDevices(c.deviceType)
	checkErr(err)

	for _, d := range devices {
		devs = append(devs, buildDevice(d))
	}

	return devs
}

// Allocate the device to job,then other job can not use the device
func (h *CarizonDeviceManager) Allocate(deviceIPs []string) {
	logger.Wrapper.Infof("Allocate deviceIPs: %+v", deviceIPs)
	reqData := AllocateDeviceReq{DeviceIP: strings.Join(deviceIPs, ","), Allocated: true}
	allocateDevice(reqData)
}

// GetAllocateDevicesInfo is get device ip and how many pcie occupied
func (h *CarizonDeviceManager) GetAllocateDevicesInfo(deviceIPs []string) (info *[]PCIeAddressInfo, err error) {
	data := map[string][]string{"ips": deviceIPs}
	resp, err := HTTPClient.R().SetBody(data).Post(CmdbServer + getAllocateDeviceInfoAPI)

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
func (h *CarizonDeviceManager) CheckHealth(stop <-chan interface{}, devices []*Device, healthy, unhealthy chan<- *Device) {
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

	params := map[string]string{"bind_node": nodeName, "device_type": deviceType}
	req := HTTPClient.R().SetQueryParams(params)
	resp, err := req.Get(CmdbServer + findInstassociationAPI)

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
	resp, err := HTTPClient.R().Get(CmdbServer + fmt.Sprintf(checkHealthAPI, uuid))
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

func allocateDevice(data AllocateDeviceReq) {
	resp, err := HTTPClient.R().SetBody(data).Post(CmdbServer + batchUpdateInstsAPI)

	_, _, err = checkHTTPResponse(resp, err)

	if err != nil {
		logger.Wrapper.Errorf("Failed to allocate device.Err: %+v,DeviceIP: %s", err, data.DeviceIP)
	}

}
