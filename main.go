package main

import (
	"flag"
	"math/rand"
	"strings"
	"syscall"
	"time"

	"carizon-device-plugin/logger"

	"github.com/fsnotify/fsnotify"
	"github.com/robfig/cron"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

// HTTPClient init resty client
var HTTPClient = newHTTPClient()

func getAllPlugins() []*HorizonDevicePlugin {
	deviceTypes, err := getDeviceTypes()
	checkErr(err)

	plugins := []*HorizonDevicePlugin{}
	for _, t := range deviceTypes {
		//compatible with device manager,device manager read nacos need have this chiptype
		if t == horizonDevicePcie {
			continue
		}
		plugin := NewHorizonDevicePlugin(
			resourceDomain+t,
			NewHorizonDeviceManager(t),
			"HORIZON_DEVICE_"+t+"_IP_LIST",
			pluginapi.DevicePluginPath+"horizon_"+t+".sock")
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

	deviceManager := NewHorizonDeviceManager("")
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
	flag.Parse()

	if HTTPClient == nil {
		logger.Wrapper.Fatalln("[main] Failed to init HTTPClient")
	}
	// TODO: verify horizon device-manager accessible
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
