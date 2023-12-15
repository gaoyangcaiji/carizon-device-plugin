package conf

var Conf Config

// Filter 定义了过滤条件，这里使用map[string]interface{}是因为基于 YAML 数据中“id”的条件比较特殊
type Filter struct {
	NodeName string                 `yaml:"nodename"`
	ID       map[string]interface{} `yaml:"id"`
}

// Resource 描述了一种资源和它的查询条件
type Resource struct {
	ResourceName string     `yaml:"resource_name"`
	Filter       Filter     `yaml:"filter,omitempty"`
	Projection   []string   `yaml:"projection"`
	SubResource  []Resource `yaml:"sub_resource,omitempty"`
}

type Config struct {
	ResourceDevices []Resource `yaml:"resource_device_plugin"`
}
