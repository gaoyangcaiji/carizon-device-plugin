package main

import (
	"carizon-device-plugin/conf"
	httpclient "carizon-device-plugin/pkg/client"
	"carizon-device-plugin/pkg/logger"
	"carizon-device-plugin/pkg/nacos"
	"math/rand"
	"strings"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/robfig/cron"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

// HTTPClient init resty client
var CmdbApiClient = httpclient.NewClient()

func getAllPlugins() []*CarizonDevicePlugin {
	plugins := []*CarizonDevicePlugin{}
	for _, t := range conf.Conf.ResourceDevices {
		// TODO: 对子资源也进行处理
		plugin := NewCarizonDevicePlugin(
			resourceDomain+t.ResourceName,
			NewCarizonDeviceManager(t.ResourceName),
			"CARIZON_DEVICE_"+t.ResourceName+"_IP_LIST",
			pluginapi.DevicePluginPath+"carizon_"+t.ResourceName+".sock")
		plugins = append(plugins, plugin)
	}
	return plugins
}

func startCron() {
	c := cron.New()
	if err := c.AddFunc("@every 3m", func() { // execute every 3 min
		refreshDeviceReserved()
	}); err != nil {
		logger.Wrapper.Errorf("strat cron job error:%s", err.Error())
	}
	//random sleep 0-10min to start job,avoid flow flood at the same moment
	r := rand.Intn(10)
	time.Sleep(time.Duration(r) * time.Minute)
	c.Start()
}

func refreshDeviceReserved() {
	var deviceList []string

	deviceManager := NewCarizonDeviceManager("")
	client, _ := GetResourceClient("")
	resourceInfos, _ := client.GetPodResourceMap()

	for name, item := range resourceInfos {
		//Composed chip type allocate directly,because device count is small
		//will not make pressure to device manager microservice
		if strings.Contains(name, "4J5") {
			deviceManager.Allocate(item.DeviceIDs)
		} else {
			deviceList = append(deviceList, item.DeviceIDs...)
		}
	}

	if len(deviceList) == 0 {
		return
	}

	deviceManager.Allocate(deviceList)

}

func main() {
	nacos.Init("model", "DEFAULT_GROUP", "carizon.cmdb", "config", &conf.Conf)

	if CmdbApiClient == nil {
		logger.Wrapper.Fatalln("[main] Failed to init CmdbApiClient")
	}
	// TODO: verify Carizon device-manager accessible
	logger.Wrapper.Infoln("[main] Starting FS watcher.")
	watcher, err := newFSWatcher(pluginapi.DevicePluginPath)
	if err != nil {
		logger.Wrapper.Fatalln("[main] Create FS watcher failed.")
	}
	defer watcher.Close()

	logger.Wrapper.Infoln("[main] Starting OS watcher.")
	sigs := newOSWatcher(syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	logger.Wrapper.Infoln("[main] Retrieving plugins.")
	plugins := getAllPlugins()

	go startCron()

restart:
	pluginStartError := make(chan struct{})
	started := 0
	for _, plugin := range plugins {
		plugin.Stop()

		if err := plugin.Start(); err != nil {
			logger.Wrapper.Infoln("[main] Start plugin error: %s", err.Error())
			close(pluginStartError)
			goto events
		}
		started++
	}

	if started == 0 {
		logger.Wrapper.Infoln("[main] No device found. Waiting indefinitely.")
	}

events:
	for {
		select {
		case <-pluginStartError:
			goto restart
		case event := <-watcher.Events:
			if event.Name == pluginapi.KubeletSocket && event.Op&fsnotify.Create == fsnotify.Create {
				logger.Wrapper.Infoln("[main][event] inotify: %s created, restarting.", pluginapi.KubeletSocket)
				goto restart
			}
		case err := <-watcher.Errors:
			logger.Wrapper.Infoln("[main][event] inotify: %s", err)
		case s := <-sigs:
			switch s {
			case syscall.SIGHUP:
				logger.Wrapper.Infoln("[main][event] Received SIGHUP, restarting.")
				goto restart
			default:
				logger.Wrapper.Infof("[main][event] Received signal \"%v\", shutting down.", s)
				for _, p := range plugins {
					p.Stop()
				}
				break events
			}
		}
	}
}
