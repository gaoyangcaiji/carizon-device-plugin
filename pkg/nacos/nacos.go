/*
Package nacos

包括从nacos配置中心拉取服务配置、服务配置变更后的自动更新
*/
package nacos

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/pelletier/go-toml"

	"gopkg.in/yaml.v2"

	"carizon-device-plugin/pkg/env"

	"github.com/fsnotify/fsnotify"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/spf13/viper"
)

const (
	// typeOfYAML yaml类型配置文件
	typeOfYAML = "yaml"
	// typeOfJSON json类型配置文件
	typeOfJSON = "json"
	// typeOfTOML toml类型配置文件
	typeOfTOML = "toml"

	yamlConfigSuffix = ".yaml"
	jsonConfigSuffix = ".json"
	tomlConfigSuffix = ".toml"
)

var (
	DevSC = constant.ServerConfig{
		Scheme:      "http",
		ContextPath: "/nacos",
		IpAddr:      "nacos.aidi-dev.hobot.cc",
		Port:        80,
	}
	TestSC = constant.ServerConfig{
		Scheme:      "http",
		ContextPath: "/nacos",
		IpAddr:      "nacos.aidi-test.hobot.cc",
		Port:        80,
	}
	PreSC = constant.ServerConfig{
		Scheme:      "http",
		ContextPath: "/nacos",
		IpAddr:      "nacos.aidi.hobot.cc",
		Port:        80,
	}
	ProdSC = constant.ServerConfig{
		Scheme:      "http",
		ContextPath: "/nacos",
		IpAddr:      "nacos.aidi.hobot.cc",
		Port:        80,
	}
	SaasTestSC = constant.ServerConfig{
		Scheme:      "https",
		ContextPath: "/nacos",
		IpAddr:      "nacos.aidi-test.horizon.ai",
		Port:        443,
	}
	SaasProdSC = constant.ServerConfig{
		Scheme:      "https",
		ContextPath: "/nacos",
		IpAddr:      "nacos.aidi.horizon.ai",
		Port:        443,
	}
	SCMap = map[env.RunMode]constant.ServerConfig{
		env.DevMode:      DevSC,
		env.TestMode:     TestSC,
		env.PreMode:      PreSC,
		env.ProdMode:     ProdSC,
		env.SaasTestMode: SaasTestSC,
		env.SaasProdMode: SaasProdSC,
	}
)

var v *viper.Viper

// 监听本地配置文件
func watchLocalFile(fileName, configType, configSuffix string, conf interface{}) {
	v = viper.New()
	v.SetConfigName(fileName)

	v.SetConfigType(configType)
	v.AddConfigPath(env.ConfPath)
	if err := v.ReadInConfig(); err != nil {
		log.Fatal("viper read config error: " + err.Error())
	}
	v.AutomaticEnv()
	v.WatchConfig()
	v.OnConfigChange(func(in fsnotify.Event) {
		if conf != nil {
			err := unmarshalStruct(filepath.Join(env.ConfPath, fileName+configSuffix), conf)
			if err != nil {
				log.Fatal("viper unmarshall error: " + err.Error())
			}
		}
	})
}

func unmarshalStruct(fileName string, conf interface{}) error {
	confType := reflect.TypeOf(conf).Elem()
	if confType.NumField() == 0 {
		return nil
	}
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}
	firstField := confType.Field(0)
	if firstField.Tag.Get("mapstructure") != "" {
		return v.Unmarshal(conf)
	} else if firstField.Tag.Get("yaml") != "" {
		return yaml.Unmarshal(content, conf)
	} else if firstField.Tag.Get("json") != "" {
		return json.Unmarshal(content, conf)
	} else if firstField.Tag.Get("toml") != "" {
		return toml.Unmarshal(content, conf)
	}
	return errors.New("not supported struct tag")
}

func FetchConfig(namespace, group, dataID string) (cli config_client.IConfigClient, content string, err error) {
	// 拉取nacos远程配置
	cc := &constant.ClientConfig{
		NamespaceId:         namespace,
		AppName:             env.AppName,
		NotLoadCacheAtStart: true,
		TimeoutMs:           3000,
		LogLevel:            "info",
		LogDir:              "/tmp/nacos/log",
		CacheDir:            "/tmp/nacos/cache",
		Username:            "nacos_readonly",
		Password:            "nacos_password",
	}
	var serverConf constant.ServerConfig
	if env.UseUserNacosConf() {
		serverConf = constant.ServerConfig{
			Scheme:      env.NacosScheme,
			ContextPath: env.NacosContext,
			IpAddr:      env.NacosIPAddr,
			Port:        uint64(env.NacosPort),
		}
	} else {
		sc, ok := SCMap[env.ServiceMode]
		if !ok {
			log.Fatal("get nacos service config error")
		}
		serverConf = sc
	}
	log.Printf("using nacos service config: %+v", serverConf)
	cli, err = clients.NewConfigClient(vo.NacosClientParam{
		ClientConfig:  cc,
		ServerConfigs: []constant.ServerConfig{serverConf},
	})
	if err != nil {
		log.Fatal("get nacos config client error: " + err.Error())
	}
	content, err = cli.GetConfig(vo.ConfigParam{
		DataId: dataID,
		Group:  group,
	})
	return cli, content, err
}

// Init 初始化服务配置，从nacos中读取配置并反序列化到配置对象中
// namespace、group和dataID分别为nacos中配置所在的命名空间、组和配置集唯一标识
// fileName为从nacos同步到本地的配置文件名称，不包含文件后缀，支持yaml、json、toml格式
// conf为自定义的配置对象的指针
// 传入conf对象会自动映射，如果是其他格式改对象可以不传(nil)
func Init(namespace, group, dataID string, fileName string, conf interface{}) {

	configType := typeOfYAML // 默认yaml格式文件
	configSuffix := yamlConfigSuffix
	switch {
	case strings.HasSuffix(dataID, typeOfJSON):
		configType = typeOfJSON
		configSuffix = jsonConfigSuffix
	case strings.HasSuffix(dataID, typeOfTOML):
		configType = typeOfTOML
		configSuffix = tomlConfigSuffix
	}
	// 本地配置文件
	filename := path.Join(env.ConfPath, fileName+configSuffix)
	if env.ServiceMode == env.LocalMode {
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			// 如果文件不存，生成空配置文件
			err = ioutil.WriteFile(filename, []byte(""), 0644)
			if err != nil {
				log.Fatal("write config file error: " + err.Error())
			}
		}
		watchLocalFile(fileName, configType, configSuffix, conf)
		if conf != nil {
			err := unmarshalStruct(filename, conf)
			if err != nil {
				log.Fatal("viper unmarshall error: " + err.Error())
			}
		}
		return
	}

	os.Mkdir("/tmp/nacos", 0755)

	cli, content, err := FetchConfig(namespace, group, dataID)
	if err != nil {
		log.Fatal("get nacos config error: " + err.Error())
	}

	// 同步到本地配置文件，并使用viper开启file watch
	err = ioutil.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		log.Fatal("write config file error: " + err.Error())
	}

	watchLocalFile(fileName, configType, configSuffix, conf)

	// 将配置文件内容反序列化到配置对象中
	if conf != nil {
		err := unmarshalStruct(filename, conf)
		if err != nil {
			log.Fatal("viper unmarshall error: " + err.Error())
		}
	}

	// 开启本地与nacos的自动同步
	err = cli.ListenConfig(vo.ConfigParam{
		DataId: dataID,
		Group:  group,

		OnChange: func(namespace, group, dataId, data string) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("nacos on change panic: %v", r)
				}
			}()

			if data == "" {
				log.Print("nacos on change is empty")
				return
			}

			err = ioutil.WriteFile(filename, []byte(data), 0644)
			if err != nil {
				log.Fatal("write config file error: " + err.Error())
			}
		},
	})
	if err != nil {
		log.Fatal("nacos listen config error: " + err.Error())
	}
	return
}

// Get 获取配置
func Get(key string) interface{} {
	return v.Get(key)
}

// GetString 获取配置
func GetString(key string) (value string) {
	value = v.GetString(key)
	return
}

// GetStrings 以,分割字符串
// a,b,c,d   ===>  ["a", "b", "c", "d"]
func GetStrings(key string) (s []string) {
	value := GetString(key)
	if value == "" {
		return
	}
	for _, v := range strings.Split(value, ",") {
		s = append(s, v)
	}
	return
}

// GetInt32 获取 int32 配置
func GetInt32(key string) (value int32) {
	return v.GetInt32(key)
}

// GetInt64 获取 int64 配置
func GetInt64(key string) (value int64) {
	return v.GetInt64(key)
}

// GetBool 获取bool类型配置
func GetBool(key string) bool {
	return v.GetBool(key)
}

// GetFloat64 获取float64类型配置
func GetFloat64(key string) float64 {
	return v.GetFloat64(key)
}
