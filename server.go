package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"carizon-device-plugin/logger"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

// HorizonDevicePlugin implements the Kubernetes device plugin API
type HorizonDevicePlugin struct {
	ResourceManager
	resourceName  string
	deviceListEnv string
	socket        string

	server        *grpc.Server
	cachedDevices []*Device
	health        chan *Device
	unhealth      chan *Device
	stop          chan interface{}
	sync.RWMutex
}

// NewHorizonDevicePlugin returns an initialized HorizonDevicePlugin
func NewHorizonDevicePlugin(resourceName string, resourceManager ResourceManager, deviceListEnv string, socket string) *HorizonDevicePlugin {

	return &HorizonDevicePlugin{
		ResourceManager: resourceManager,
		resourceName:    resourceName,
		deviceListEnv:   deviceListEnv,
		socket:          socket,

		// These will be reinitialized every
		// time the plugin server is restarted.
		cachedDevices: nil,
		server:        nil,
		health:        nil,
		unhealth:      nil,
		stop:          nil,
	}
}

func (h *HorizonDevicePlugin) initialize() {
	h.cachedDevices = h.Devices()
	h.server = grpc.NewServer([]grpc.ServerOption{}...)
	h.health = make(chan *Device)
	h.unhealth = make(chan *Device)
	h.stop = make(chan interface{})
}

// Register registers the device plugin for the given resourceName with Kubelet.
func (h *HorizonDevicePlugin) Register() error {
	conn, err := h.dial(pluginapi.KubeletSocket, 5*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pluginapi.NewRegistrationClient(conn)
	reqt := &pluginapi.RegisterRequest{
		Version:      pluginapi.Version,
		Endpoint:     path.Base(h.socket),
		ResourceName: h.resourceName,
	}

	_, err = client.Register(context.Background(), reqt)
	if err != nil {
		return err
	}
	return nil
}

func (h *HorizonDevicePlugin) cleanup() {
	close(h.stop)
	h.cachedDevices = nil
	h.server = nil
	h.health = nil
	h.unhealth = nil
	h.stop = nil
}

// Start starts the gRPC server, registers the device plugin with the Kubelet,
// and starts the device healthchecks.
func (h *HorizonDevicePlugin) Start() error {
	h.initialize()

	err := h.Serve()
	if err != nil {
		log.Printf("Could not start device plugin for '%s': %s", h.resourceName, err)
		h.cleanup()
		return err
	}
	log.Printf("Starting to serve '%s' on %s", h.resourceName, h.socket)

	err = h.Register()
	if err != nil {
		log.Printf("Could not register device plugin: %s", err)
		h.Stop()
		return err
	}
	log.Printf("Registered device plugin for '%s' with Kubelet", h.resourceName)

	go h.CheckHealth(h.stop, h.cachedDevices, h.health, h.unhealth)

	return nil
}

// Stop stops the gRPC server.
func (h *HorizonDevicePlugin) Stop() error {
	if h == nil || h.server == nil {
		return nil
	}
	log.Printf("Stopping to serve '%s' on %s", h.resourceName, h.socket)
	h.server.Stop()
	if err := os.Remove(h.socket); err != nil && !os.IsNotExist(err) {
		return err
	}
	h.cleanup()
	return nil
}

// Serve starts the gRPC server of the device plugin.
func (h *HorizonDevicePlugin) Serve() error {
	os.Remove(h.socket)
	sock, err := net.Listen("unix", h.socket)
	if err != nil {
		return err
	}

	pluginapi.RegisterDevicePluginServer(h.server, h)

	log.Printf("Starting gRPC server for '%s'", h.resourceName)
	go func() {
		err = h.server.Serve(sock)
		if err != nil {
			log.Fatalf("gRPC server for '%s' crashed with error: %v", h.resourceName, err)
		}
	}()

	// Wait for server to start by launching a blocking connecion
	conn, err := h.dial(h.socket, 5*time.Second)
	if err != nil {
		return err
	}
	conn.Close()

	return nil
}

// ListAndWatch lists devices and update that list according to the health status
func (h *HorizonDevicePlugin) ListAndWatch(e *pluginapi.Empty, s pluginapi.DevicePlugin_ListAndWatchServer) error {
	s.Send(&pluginapi.ListAndWatchResponse{Devices: h.apiDevices()})

	for {
		select {
		case <-h.stop:
			return nil
		case d := <-h.unhealth:
			d.Health = pluginapi.Unhealthy
			log.Printf("'%s' device marked unhealthy: %s", h.resourceName, d.IP)
			s.Send(&pluginapi.ListAndWatchResponse{Devices: h.apiDevices()})
		case d := <-h.health:
			d.Health = pluginapi.Healthy
			log.Printf("'%s' device marked healthy: %s", h.resourceName, d.IP)
			s.Send(&pluginapi.ListAndWatchResponse{Devices: h.apiDevices()})
		}
	}
}

// Allocate responses resource request
func (h *HorizonDevicePlugin) Allocate(ctx context.Context, reqs *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	responses := pluginapi.AllocateResponse{}

	logger.Wrapper.Infoln("----Allocating horizon device is started----")
	var (
		info     *[]PCIeAddressInfo = new([]PCIeAddressInfo)
		pcieFlag bool
		err      error
	)

	// 1. Create container requests
	for _, req := range reqs.ContainerRequests {
		logger.Wrapper.Infof("Kubelet allocate deviceIDs:%+v", req.DevicesIDs)

		if h.resourceName == resourceDomain+chipTypeJ5 {
			pcieFlag = true

			info, err = h.ResourceManager.GetAllocateDevicesInfo(req.DevicesIDs)
			if err != nil {
				logger.Wrapper.Errorf("pcieInfo err: %+v", err)
				return nil, err
			}
		}
		if h.resourceName == resourceDomain+chipType2J5 {
			pcieFlag = true
		}

		for _, id := range req.DevicesIDs {
			if !h.deviceExists(id) {
				return nil, fmt.Errorf("invalid allocation request for '%s': unknown device: %s", h.resourceName, id)
			}
		}
		//marshal device info
		infoJSON, err := json.Marshal(PodReq{Device: strings.Join(req.DevicesIDs, ","), PCIeInfo: *info})
		if err != nil {
			return nil, fmt.Errorf("marshal info %+v error '%s'", info, err.Error())
		}

		response := pluginapi.ContainerAllocateResponse{
			Envs: map[string]string{
				h.deviceListEnv:          strings.Join(req.DevicesIDs, ","),
				horizonDevicePcieInfoEnv: string(infoJSON),
				horizonPcieFlagEnv:       strconv.FormatBool(pcieFlag),
			},
		}
		logger.Wrapper.Infof("the pcieinfo %s", string(infoJSON))
		h.ResourceManager.Allocate(req.DevicesIDs)

		responses.ContainerResponses = append(responses.ContainerResponses, &response)
	}

	return &responses, nil
}

// GetDevicePluginOptions get horizonDevicePlugin options
func (h *HorizonDevicePlugin) GetDevicePluginOptions(context.Context, *pluginapi.Empty) (*pluginapi.DevicePluginOptions, error) {
	return &pluginapi.DevicePluginOptions{}, nil
}

// PreStartContainer  get horizonDevicePlugin PreStartContainer
func (h *HorizonDevicePlugin) PreStartContainer(context.Context, *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
	return &pluginapi.PreStartContainerResponse{}, nil
}

// dial establishes the gRPC communication
func (h *HorizonDevicePlugin) dial(unixSocketPath string, timeout time.Duration) (*grpc.ClientConn, error) {
	c, err := grpc.Dial(unixSocketPath, grpc.WithInsecure(), grpc.WithBlock(),
		grpc.WithTimeout(timeout),
		grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
			return net.DialTimeout("unix", addr, timeout)
		}),
	)

	if err != nil {
		return nil, err
	}

	return c, nil
}

func (h *HorizonDevicePlugin) deviceExists(id string) bool {
	for _, d := range h.cachedDevices {
		if d.ID == id {
			return true
		}
	}
	return false
}

func (h *HorizonDevicePlugin) apiDevices() []*pluginapi.Device {
	var pdevs []*pluginapi.Device
	for _, d := range h.cachedDevices {
		pdevs = append(pdevs, &d.Device)
	}
	return pdevs
}
