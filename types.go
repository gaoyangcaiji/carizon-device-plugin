package main

import (
	"time"

	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

type externalDevice struct {
	UUID int
	IP   string
}

// Device ...
type Device struct {
	pluginapi.Device
	externalDevice
}

// AllocateDeviceReq defines devicemanager allocate req
type AllocateDeviceReq struct {
	DeviceIP  string `json:"ip"`
	Allocated bool   `json:"allocated"`
}

// PodReq defines to update pod devices
type PodReq struct {
	Device   string            `json:"device"`
	PCIeInfo []PCIeAddressInfo `json:"pcie_info"`
}

// DeviceInfo defines device orm info
type DeviceInfo struct {
	ID            int       `json:"id" gorm:"primary_key"`
	IP            string    `json:"ip" gorm:"column:ip;varchar(15)"`
	ChipType      string    `json:"chip_type" gorm:"column:chip_type;varchar(20)"`
	ChipNum       int8      `json:"chip_num" gorm:"column:chip_num"`
	Status        int8      `json:"status" gorm:"column(status)"`
	QueueName     string    `json:"queue_name" gorm:"column:queue_name;varchar(50)"`
	BindNode      string    `json:"bind_node" gorm:"column:bind_node;varchar(50)"`
	DeviceType    string    `json:"device_type" gorm:"column:device_type;varchar(50)"`
	SystemVersion string    `json:"system_version" gorm:"column:system_version;varchar(50)"`
	RegisterTime  time.Time `json:"register_time" gorm:"column:register_time"`
	LastAliveTime time.Time `json:"last_alive_time" gorm:"column:last_alive_time"`
	IsReserved    int8      `json:"is_reserved" gorm:"column:is_reserved"`
	Location      string    `json:"location" gorm:"column:location;varchar(20)"`
	Tags          string    `json:"tags" gorm:"column:tags;varchar(100)"`
}

// PCIeAddressInfo ...
type PCIeAddressInfo struct {
	IP     string `json:"ip"`
	VNetIP string `json:"vnet_ip"`
}

// HTTPRetCommon common HTTP response result
type HTTPRetCommon struct {
	Code       int    `json:"code"`
	ErrMsg     string `json:"err_msg"`
	ErrUserMsg string `json:"err_user_msg"`
}

// HTTPRetBool HTTP response result(bool data)
type HTTPRetBool struct {
	HTTPRetCommon
	Data bool `json:"data"`
}

// HTTPRetDeviceTypes HTTP response result of device types
type HTTPRetDeviceTypes struct {
	HTTPRetCommon
	Data map[string][]string `json:"data"`
}

// HTTPRetDeviceList HTTP response result of device list
type HTTPRetDeviceList struct {
	HTTPRetCommon
	Data map[string][]DeviceInfo `json:"data"`
}

// HTTPRetAllocateDevicesInfo is the api response
type HTTPRetAllocateDevicesInfo struct {
	HTTPRetCommon
	Data []PCIeAddressInfo `json:"data,omitempty"` //返回结果
}
