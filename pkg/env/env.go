/*
Package env
包括服务的名称、运行环境、加载的配置文件路径、JAEGER配置
*/
package env

import (
	"log"
	"os"
	"strconv"
)

type RunMode int64

// 各类运行环境定义以及获取服务名称、运行环境的环境变量定义
const (
	LocalMode    RunMode = iota // 本地环境
	DevMode                     // 开发环境
	TestMode                    // 测试环境
	PreMode                     // 预发布环境
	ProdMode                    // 线上环境
	SaasTestMode                // 公有云测试环境
	SaasProdMode                // 公有云线上环境

	AppName         = "APP_NAME"
	APPAITCMode     = "AITC_MODE"
	APPConfPath     = "CONF_PATH"
	APPTraceAgent   = "JAEGER_TRACE_AGENT"
	APPTraceSampler = "JAEGER_TRACE_SAMPLER"
	APPClusterName  = "CLUSTER_NAME"
	APPNacosScheme  = "NACOS_SCHEME"
	APPNacosContext = "NACOS_CONTEXT"
	APPNacosIPAddr  = "NACOS_IPADDR"
	APPNacosPort    = "NACOS_PORT"
	APPLogLevel     = "LOG_LEVEL"
)

var runModeMap = map[string]RunMode{
	"local":     LocalMode,
	"dev":       DevMode,
	"test":      TestMode,
	"uat":       TestMode,
	"pre":       PreMode,
	"online":    ProdMode,
	"prod":      ProdMode,
	"saas-test": SaasTestMode,
	"saas-prod": SaasProdMode,
}

// runModelString mode->env
var runModelString = map[RunMode]string{
	LocalMode:    "local",
	DevMode:      "dev",
	TestMode:     "test",
	PreMode:      "pre",
	ProdMode:     "prod",
	SaasTestMode: "saas-test",
	SaasProdMode: "saas-prod",
}

var useUserNacosConf = true

var (
	// ServiceName 服务的名称，全PDT唯一,对应环境变量APP_NAME
	ServiceName string
	// ServiceMode 服务的运行环境，包括本地、开发、测试、预生产、生产环境, 同时区别是否为公有云环境，对应环境变量AITC_MODE
	ServiceMode RunMode
	// ConfPath 服务的配置文件存放路径
	ConfPath string
	// ClusterName 服务所在的集群信息
	ClusterName string
	// InstanceID 机器的host name
	InstanceID string
	// NacosScheme nacos的http协议
	NacosScheme string
	// NacosContext nacos的context
	NacosContext string
	// NacosIPAddr nacos服务的IP地址
	NacosIPAddr string
	// NacosPort nacos服务的端口
	NacosPort int
	// Hostname 主机名
	Hostname = "localhost"
	//LogLevel 日志等级
	LogLevel = "debug"
)

// Init 从环境变量中获取服务的基础配置信息
// TODO: 从配置中心获取服务配置
func init() {
	getDefaultConfFromEnv()
}

func getDefaultConfFromEnv() {
	Hostname, _ = os.Hostname()
	ServiceName = os.Getenv(AppName)
	if ServiceName == "" {
		log.Fatal("empty app name")
	}
	runEnv := os.Getenv(APPAITCMode)
	if runEnv == "" {
		runEnv = runModelString[DevMode]
	}

	runMode, ok := runModeMap[runEnv]
	if !ok {
		log.Fatal("invalid run mode")
	}
	ServiceMode = runMode
	ConfPath = os.Getenv(APPConfPath)
	if ConfPath == "" {
		confPath, err := os.Getwd()
		if err != nil {
			log.Fatal("conf path is empty")
		}
		ConfPath = confPath
	}

	clusterName := os.Getenv(APPClusterName)
	if clusterName == "" {
		log.Println("env CLUSTER_NAME is empty, use default idc")
		clusterName = "idc"
	}
	ClusterName = clusterName

	InstanceID, _ = os.Hostname()

	NacosScheme = os.Getenv(APPNacosScheme)
	if NacosScheme == "" {
		log.Println("env NACOS_SCHEME is empty, use default conf")
		useUserNacosConf = false
	} else {
		log.Println("env NACOS_SCHEME is not empty, use user defined conf")
		NacosContext = os.Getenv(APPNacosContext)
		NacosIPAddr = os.Getenv(APPNacosIPAddr)
		nacosPort := os.Getenv(APPNacosPort)
		nacosPortInt, err := strconv.Atoi(nacosPort)
		if err != nil {
			log.Fatalf("env NACOS_PORT: %s invalid", nacosPort)
		}
		NacosPort = nacosPortInt
	}

	logLevel := os.Getenv(APPLogLevel)
	switch logLevel {
	case "":
		log.Println("env LOG_LEVEL is empty, use default debug")
	case "trace", "debug", "info", "warn", "error", "fatal", "panic":
		LogLevel = logLevel
	default:
		log.Fatal("invaild logLevel")
	}

}

// GetDeployEnv 获取机器的环境
func GetDeployEnv(mode RunMode) string {
	return runModelString[mode]
}

// UseUserNacosConf 是否获取用户通过环境变量指定的nacos地址
func UseUserNacosConf() bool {
	return useUserNacosConf
}
