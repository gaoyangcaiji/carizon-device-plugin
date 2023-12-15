package main

import (
	"net/url"
	"os"
	"time"

	"carizon-device-plugin/pkg/logger"

	"golang.org/x/net/context"

	"k8s.io/kubernetes/pkg/kubelet/apis/podresources"
	podresourcesapi "k8s.io/kubernetes/pkg/kubelet/apis/podresources/v1alpha1"
	"k8s.io/kubernetes/pkg/kubelet/util"
)

const (
	defaultKubeletSocketFile   = "kubelet.sock"
	defaultPodResourcesMaxSize = 1024 * 1024 * 16 // 16 Mb
	defaultPodResourcesPath    = "/var/lib/kubelet/pod-resources"
)

// ResourceInfo is struct to hold Pod device allocation information
type ResourceInfo struct {
	Index     int
	DeviceIDs []string
}

// ResourceClient provides a kubelet Pod resource handle
type ResourceClient interface {
	// GetPodResourceMap returns an instance of a map of Pod ResourceInfo given a (Pod name, namespace) tuple
	GetPodResourceMap() (map[string]*ResourceInfo, error)
}

// GetResourceClient get initialized with Pod resource information
func GetResourceClient(kubeletSocket string) (ResourceClient, error) {
	if kubeletSocket == "" {
		kubeletSocket, _ = util.LocalEndpoint(defaultPodResourcesPath, podresources.Socket)
	}
	logger.Wrapper.Infof("[GetResourceClient]: using Kubelet resource API endpoint")
	// If Kubelet resource API endpoint exist use that by default
	if hasKubeletAPIEndpoint(kubeletSocket) {
		return getKubeletClient(kubeletSocket)
	}

	return GetCheckpoint()
}

// getKubeletClient ...
func getKubeletClient(kubeletSocket string) (ResourceClient, error) {
	newClient := &kubeletClient{}
	if kubeletSocket == "" {
		kubeletSocket, _ = util.LocalEndpoint(defaultPodResourcesPath, podresources.Socket)
	}

	client, conn, err := podresources.GetClient(kubeletSocket, 10*time.Second, defaultPodResourcesMaxSize)
	if err != nil {
		logger.Wrapper.Errorf("[getKubeletClient]: get grpc client error: %v\n", err)
		return nil, err
	}
	defer conn.Close()

	if err := newClient.getPodResources(client); err != nil {
		logger.Wrapper.Errorf("[getKubeletClient]: get pod resources from client error: %v\n", err)
		return nil, err
	}

	return newClient, nil
}

// kubeletClient ...
type kubeletClient struct {
	resources []*podresourcesapi.PodResources
}

// getPodResources ...
func (rc *kubeletClient) getPodResources(client podresourcesapi.PodResourcesListerClient) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := client.List(ctx, &podresourcesapi.ListPodResourcesRequest{})
	if err != nil {
		logger.Wrapper.Errorf("[getPodResources]: failed to list pod resources, %v.Get(_) = _, %v", client, err)
		return err
	}

	rc.resources = resp.PodResources
	return nil
}

// GetPodResourceMap ...
func (rc *kubeletClient) GetPodResourceMap() (map[string]*ResourceInfo, error) {
	resourceMap := make(map[string]*ResourceInfo)

	for _, pr := range rc.resources {
		for _, cnt := range pr.Containers {
			for _, dev := range cnt.Devices {
				if rInfo, ok := resourceMap[dev.ResourceName]; ok {
					rInfo.DeviceIDs = append(rInfo.DeviceIDs, dev.DeviceIds...)
				} else {
					resourceMap[dev.ResourceName] = &ResourceInfo{DeviceIDs: dev.DeviceIds}
				}
			}
		}
	}

	return resourceMap, nil
}

// hasKubeletAPIEndpoint ...
func hasKubeletAPIEndpoint(endpoint string) bool {
	u, err := url.Parse(endpoint)
	if err != nil {
		logger.Wrapper.Infof("[hasKubeletAPIEndpoint]: parse ep err: %v", err)
		return false
	}
	// Check for kubelet resource API socket file
	if _, err := os.Stat(u.Path); err != nil {
		logger.Wrapper.Infof("[hasKubeletAPIEndpoint]: looking up kubelet resource api socket file error: %q", err)
		return false
	}
	return true
}
