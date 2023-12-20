package metadata

import (
	"carizon-device-plugin/pkg/mapstr"
	"time"
)

// ResponseInstData TODO
type ResponseInstData struct {
	BaseResp `json:",inline"`
	Data     InstDataInfo `json:"data"`
}

// InstDataInfo response instance data result Data field
type InstDataInfo struct {
	Count int             `json:"count"`
	Info  []mapstr.MapStr `json:"info"`
}

// TimeConditionItem TODO
type TimeConditionItem struct {
	Field string     `json:"field" bson:"field"`
	Start *time.Time `json:"start" bson:"start"`
	End   *time.Time `json:"end" bson:"end"`
}

// TimeCondition TODO
type TimeCondition struct {
	Operator string              `json:"oper" bson:"oper"`
	Rules    []TimeConditionItem `json:"rules" bson:"rules"`
}
