package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"carizon-device-plugin/metadata"
	"carizon-device-plugin/pkg/logger"
	"carizon-device-plugin/pkg/mapstr"

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

	//SearchAssociationInsts
	option := metadata.OpCondition{
		Update: []metadata.UpdateCondition{{InstID: 9527, InstInfo: map[string]interface{}{}}},
	}

	resp := new(metadata.SearchAssociationInstResult)
	err := CmdbApiClient.DoPut(context.Background(), CmdbServer+fmt.Sprintf(batchUpdateInstsAPI, ""), map[string]string{}, option).Into(resp)
	if err != nil {
		logger.Wrapper.Errorf("Failed to allocate device.Err: %+v,DeviceIP: %s", err, deviceIPs)
	}
}

// GetAllocateDevicesInfo is get device ip and how many pcie occupied
func (h *CarizonDeviceManager) GetAllocateDevicesInfo(deviceIPs []string) (info *[]PCIeAddressInfo, err error) {
	//data := map[string][]string{"ips": deviceIPs}
	//resp, err := CmdbApiClient.DoPost(context.Background(), CmdbServer+getAllocateDeviceInfoAPI, data, data)

	var result *HTTPRetAllocateDevicesInfo

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
	objectID := "host"
	eDevices := []*externalDevice{}
	nodeName, err := os.Hostname()
	if err != nil {
		log.Printf("Error. Fail to get hostname: %s", err.Error())
		return nil, err
	}

	log.Printf("Query %s devices on node: %s", deviceType, nodeName)

	//SearchAssociationInsts
	option := &metadata.SearchAssociationInstRequest{
		ObjID: objectID,
		Condition: mapstr.MapStr{
			"bk_inst_id": nodeName,
			"bk_obj_id":  objectID,
		},
	}

	resp := new(metadata.SearchAssociationInstResult)
	err = CmdbApiClient.DoPost(context.Background(), CmdbServer+findInstassociationAPI, map[string]string{}, option).Into(resp)
	if err != nil {
		logger.Wrapper.Errorf("Error. Failed to query %s bind devices on node %s. %v", deviceType, nodeName, err)
		return eDevices, nil
	}

	// for _, d := range resp.Data {
	// 	if d.Status == int8(DeviceOffline) {
	// 		continue
	// 	}
	// 	eDevice := externalDevice{}
	// 	eDevice.UUID = d.ID
	// 	eDevice.IP = d.IP
	// 	eDevices = append(eDevices, &eDevice)
	// }

	//SearchObjectInstances
	resp2 := new(metadata.Response)
	input := &metadata.CommonSearchFilter{
		Conditions: &metadata.CombinedRule{
			Condition: metadata.Condition("AND"),
			Rules:     []metadata.AtomRule{{Field: "bk_inst_id", Operator: metadata.Operator("in"), Value: []int{318, 319}}},
		},
		Fields: []string{},
		Page:   metadata.BasePage{Sort: "bk_inst_id", Start: 0, Limit: metadata.BKNoLimit},
	}

	err = CmdbApiClient.DoPost(context.Background(), CmdbServer+searchObjectInstsAPI, map[string]string{}, input).Into(resp2)
	if err != nil {
		logger.Wrapper.Errorf("Error. Failed to query %s bind devices on node %s. %v", deviceType, nodeName, err)
		return eDevices, nil
	}

	return eDevices, nil
}

func isDeviceHealthy(uuid int) bool {
	healthy := true
	// resp, err := CmdbApiClient.R().Get(CmdbServer + fmt.Sprintf(checkHealthAPI, uuid))
	// _, _, err = checkHTTPResponse(resp, err)
	// if err != nil {
	// 	log.Printf("Error. Failed to get device healthy. %v", err.Error())
	// 	// Default to true
	// 	return healthy
	// }

	// var ret *HTTPRetBool
	// _ = json.Unmarshal(resp.Body(), &ret)
	// healthy = ret.Data

	return healthy
}
