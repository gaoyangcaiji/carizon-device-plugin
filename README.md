# Carizon device plugin for kubernetes

Register and monitor horizon devices in kubernetes cluster.

## Usage

- Make sure device-manager deployed
- Binding devices to CPU-node(update db table 'device_info' for now)
- Install this plugin(Optional, in the same namespace with device-manager):
```shell
kubectl apply -f deploy/horizon-device-plugin.yaml -n olympus
```

- Check node Capacity&Allocatable info:
```json
  cpu:                48
  ephemeral-storage:  936183168Ki
  hobot.cc/J2:        1
  hobot.cc/J3:        0
  hobot.cc/X2:        2
  hobot.cc/X3:        9
  hugepages-1Gi:      0
  hugepages-2Mi:      0
  memory:             263571868Ki
  pods:               110
Allocatable:
  cpu:                32
  ephemeral-storage:  861712664377
  hobot.cc/J2:        1
  hobot.cc/J3:        0
  hobot.cc/X2:        2
  hobot.cc/X3:        9
  hugepages-1Gi:      0
  hugepages-2Mi:      0
  memory:             179583388Ki
  pods:               110
```


- Apply device:
```json
    resources:
      limits:
        hobot.cc/X2: 2
        hobot.cc/X3: 3
```


## Deploy note
For now, there are 2 ENVs to control the behavior of this plugin:
- ```DEVICE_HEALTH_CHECK```
  - ```true``` Enable device health check. (Default)
  - ```false``` Disable device health check.

- ```DEVICE_MANAGER_PORT```
  - If this plugin deployed in the same namespace with device-manager, this ENV with be auto injected
  - If this plugin deployed in different namespaces with device-manager, you should config this env in ```horizon-device-plugin.yaml```

And, as we bind devices by node name(hostname), so please make sure the horizon-device-plugin pod use ```hostNetwork```.

## Maintain Info
- Online branch: master
- CI: TBD
- Maintainer: YanbingDu
