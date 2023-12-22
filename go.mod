module carizon-device-plugin

go 1.16

require (
	github.com/fsnotify/fsnotify v1.4.9
	github.com/go-resty/resty/v2 v2.4.0
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/mohae/deepcopy v0.0.0-20170603005431-491d3605edfb
	github.com/nacos-group/nacos-sdk-go v1.0.9
	github.com/pelletier/go-toml v1.2.0
	github.com/robfig/cron v1.1.0
	github.com/spf13/viper v1.3.2
	github.com/stretchr/testify v1.5.1
	golang.org/x/net v0.0.0-20220127200216-cd36cc0744dd
	golang.org/x/sys v0.0.0-20220209214540-3681064d5158 // indirect
	google.golang.org/genproto v0.0.0-20210402141018-6c239bbf2bb1 // indirect
	google.golang.org/grpc v1.38.0
	google.golang.org/protobuf v1.27.1 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/kubelet v0.18.10
	k8s.io/kubernetes v1.18.10
)

replace (
	k8s.io/api => k8s.io/api v0.18.10
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.18.10
	k8s.io/apimachinery => k8s.io/apimachinery v0.18.10
	k8s.io/apiserver => k8s.io/apiserver v0.18.10
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.18.10
	k8s.io/client-go => k8s.io/client-go v0.18.10
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.18.10
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.18.10
	k8s.io/code-generator => k8s.io/code-generator v0.18.10
	k8s.io/component-base => k8s.io/component-base v0.18.10
	k8s.io/cri-api => k8s.io/cri-api v0.18.10
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.18.10
	k8s.io/klog => k8s.io/klog v1.0.0
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.18.10
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.18.10
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.18.10
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.18.10
	k8s.io/kubectl => k8s.io/kubectl v0.18.10
	k8s.io/kubelet => k8s.io/kubelet v0.18.10
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.18.10
	k8s.io/metrics => k8s.io/metrics v0.18.10
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.18.10
)
