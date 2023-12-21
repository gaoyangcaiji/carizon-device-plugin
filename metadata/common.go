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

type Response struct {
	BaseResp `json:",inline" bson:",inline"`
	Data     interface{} `json:"data" bson:"data" mapstructure:"data"`
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

// Operator TODO
// *************** define operator ************************
type Operator string

// Condition TODO
// *************** define condition ************************
type Condition string

// AtomRule TODO
// *************** define rule ************************
type AtomRule struct {
	Field    string      `json:"field"`
	Operator Operator    `json:"operator"`
	Value    interface{} `json:"value"`
}

// CombinedRule TODO
// *************** define query ************************
type CombinedRule struct {
	Condition Condition  `json:"condition"`
	Rules     []AtomRule `json:"rules"`
}

// CommonSearchFilter is a common search action filter struct,
// such like search instance or instance associations.
// And the conditions must abide by query filter.
type CommonSearchFilter struct {
	// ObjectID is target model object id.
	ObjectID string `json:"bk_obj_id"`

	// Conditions is target search conditions that make up by the query filter.
	Conditions *CombinedRule `json:"conditions"`

	// 非必填，只能用来查时间，且与Condition是与关系
	TimeCondition *TimeCondition `json:"time_condition,omitempty"`

	// Fields indicates which fields should be returns, it's would be ignored if not exists.
	Fields []string `json:"fields"`

	// Page batch query action page.
	Page BasePage `json:"page"`
}

type UpdateCondition struct {
	InstID   int64                  `json:"inst_id"`
	InstInfo map[string]interface{} `json:"datas"`
}

// OpCondition the condition operation
type OpCondition struct {
	Update []UpdateCondition `json:"update"`
}
