/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.,
 * Copyright (C) 2017,-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the ",License",); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an ",AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package metadata

import (
	"carizon-device-plugin/pkg/errors"
	"carizon-device-plugin/pkg/mapstr"
	"fmt"
)

// BaseResp common result struct
type BaseResp struct {
	Result      bool           `json:"result" bson:"result" mapstructure:"result"`
	Code        int            `json:"bk_error_code" bson:"bk_error_code" mapstructure:"bk_error_code"`
	ErrMsg      string         `json:"bk_error_msg" bson:"bk_error_msg" mapstructure:"bk_error_msg"`
	Permissions *IamPermission `json:"permission" bson:"permission" mapstructure:"permission"`
}

// CCError 根据response返回的信息产生错误
func (br *BaseResp) CCError() errors.CCErrorCoder {
	if br.Result {
		return nil
	}
	return errors.New(br.Code, br.ErrMsg)
}

// Error 用于错误处理
// New 根据response返回的信息产生错误
func (br *BaseResp) Error() error {
	if br.Result {
		return nil
	}
	return errors.New(br.Code, br.ErrMsg)
}

// ToString TODO
func (br *BaseResp) ToString() string {
	return fmt.Sprintf("code:%d, message:%s", br.Code, br.ErrMsg)
}

// JsonStringResp defines a response that do not parse the data field to a struct
// but decode it to a json string if possible
type JsonStringResp struct {
	BaseResp
	Data string
}

// JsonCntInfoResp defines a response that do not parse the data's count and info fields
// to a struct but decode it to a json string if possible
type JsonCntInfoResp struct {
	BaseResp
	Data CntInfoString `json:"data"`
}

// CntInfoString TODO
type CntInfoString struct {
	Count int64 `json:"count"`
	// info is a json array string field.
	Info string `json:"info"`
}

// IamPermission TODO
type IamPermission struct {
	SystemID   string      `json:"system_id"`
	SystemName string      `json:"system_name"`
	Actions    []IamAction `json:"actions"`
}

// IamAction TODO
type IamAction struct {
	ID                   string            `json:"id"`
	Name                 string            `json:"name"`
	RelatedResourceTypes []IamResourceType `json:"related_resource_types"`
}

// IamResourceType TODO
type IamResourceType struct {
	SystemID   string                  `json:"system_id"`
	SystemName string                  `json:"system_name"`
	Type       string                  `json:"type"`
	TypeName   string                  `json:"type_name"`
	Instances  [][]IamResourceInstance `json:"instances,omitempty"`
	Attributes []IamResourceAttribute  `json:"attributes,omitempty"`
}

// IamResourceInstance TODO
type IamResourceInstance struct {
	Type     string `json:"type"`
	TypeName string `json:"type_name"`
	ID       string `json:"id"`
	Name     string `json:"name"`
}

// IamResourceAttribute TODO
type IamResourceAttribute struct {
	ID     string                      `json:"id"`
	Values []IamResourceAttributeValue `json:"values"`
}

// IamResourceAttributeValue TODO
type IamResourceAttributeValue struct {
	ID string `json:"id"`
}

// IamInstanceWithCreator TODO
type IamInstanceWithCreator struct {
	System    string                `json:"system"`
	Type      string                `json:"type"`
	ID        string                `json:"id"`
	Name      string                `json:"name"`
	Creator   string                `json:"creator"`
	Ancestors []IamInstanceAncestor `json:"ancestors,omitempty"`
}

// IamInstances iam instances
type IamInstances struct {
	System    string        `json:"system"`
	Type      string        `json:"type"`
	Instances []IamInstance `json:"instances"`
}

// IamInstancesWithCreator iam instances with creator
type IamInstancesWithCreator struct {
	IamInstances `json:",inline"`
	Creator      string `json:"creator"`
}

// IamInstance TODO
type IamInstance struct {
	ID        string                `json:"id"`
	Name      string                `json:"name"`
	Ancestors []IamInstanceAncestor `json:"ancestors,omitempty"`
}

// IamInstanceAncestor TODO
type IamInstanceAncestor struct {
	System string `json:"system"`
	Type   string `json:"type"`
	ID     string `json:"id"`
}

// IamCreatorActionPolicy TODO
type IamCreatorActionPolicy struct {
	Action   ActionWithID `json:"action"`
	PolicyID int64        `json:"policy_id"`
}

// ActionWithID iam creator action with only action id
type ActionWithID struct {
	ID string `json:"id"`
}

// IamBatchOperateInstanceAuthReq batch grant or revoke iam instance auth request
type IamBatchOperateInstanceAuthReq struct {
	Asynchronous bool             `json:"asynchronous"`
	Operate      IamAuthOperation `json:"operate"`
	System       string           `json:"system"`
	Actions      []ActionWithID   `json:"actions"`
	Subject      IamSubject       `json:"subject"`
	Resources    []IamInstances   `json:"resources"`
	ExpiredAt    int64            `json:"expired_at"`
}

// IamAuthOperation TODO
type IamAuthOperation string

const (
	// IamGrantOperation TODO
	IamGrantOperation = "grant"
	// IamRevokeOperation TODO
	IamRevokeOperation = "revoke"
)

// IamSubject iam subject that can be authorized, right now it represents user or user group
type IamSubject struct {
	Type string `json:"type"`
	Id   string `json:"id"`
}

// IamBatchOperateInstanceAuthRes batch operate iam instance auth response
type IamBatchOperateInstanceAuthRes struct {
	Action   ActionWithID `json:"action"`
	PolicyID int64        `json:"policy_id"`
}

// Permission  describes all the authorities that a user
// is need, when he attempts to operate a resource.
// Permission is used only when a user do not have the authority to
// access a resources with a action.
type Permission struct {
	SystemID      string `json:"system_id"`
	SystemName    string `json:"system_name"`
	ScopeType     string `json:"scope_type"`
	ScopeTypeName string `json:"scope_type_name"`
	ScopeID       string `json:"scope_id"`
	ScopeName     string `json:"scope_name"`
	ActionID      string `json:"action_id"`
	ActionName    string `json:"action_name"`
	// newly added two field.
	ResourceTypeName string `json:"resource_type_name"`
	ResourceType     string `json:"resource_type"`

	Resources [][]Resource `json:"resources"`
}

// Resource TODO
type Resource struct {
	ResourceTypeName string `json:"resource_type_name"`
	ResourceType     string `json:"resource_type"`
	ResourceName     string `json:"resource_name"`
	ResourceID       string `json:"resource_id"`
}

// NewNoPermissionResp TODO
func NewNoPermissionResp(permission *IamPermission) BaseResp {
	return BaseResp{
		Result:      false,
		Code:        errors.CCNoPermission,
		ErrMsg:      "no permissions",
		Permissions: permission,
	}
}

// SuccessBaseResp default result
var SuccessBaseResp = BaseResp{Result: true, Code: errors.CCSuccess, ErrMsg: errors.CCSuccessStr}

// SuccessResponse TODO
type SuccessResponse struct {
	BaseResp `json:",inline"`
	Data     interface{} `json:"data"`
}

// NewSuccessResponse TODO
func NewSuccessResponse(data interface{}) *SuccessResponse {
	return &SuccessResponse{
		BaseResp: SuccessBaseResp,
		Data:     data,
	}
}

// CreatedCount created count struct
type CreatedCount struct {
	Count uint64 `json:"created_count"`
}

// UpdatedCount created count struct
type UpdatedCount struct {
	Count uint64 `json:"updated_count"`
}

// UpdateAttributeIndex created bk_property_index info struct
type UpdateAttributeIndex struct {
	Id    int64 `json:"id"`
	Index int64 `json:"index"`
}

// DeletedCount created count struct
type DeletedCount struct {
	Count uint64 `json:"deleted_count"`
}

// ExceptionResult exception info
type ExceptionResult struct {
	Message     string      `json:"message"`
	Code        int64       `json:"code"`
	Data        interface{} `json:"data"`
	OriginIndex int64       `json:"origin_index"`
}

// CreatedDataResult common created result definition
type CreatedDataResult struct {
	OriginIndex int64  `json:"origin_index"`
	ID          uint64 `json:"id"`
}

// RepeatedDataResult repeated data
type RepeatedDataResult struct {
	OriginIndex int64         `json:"origin_index"`
	Data        mapstr.MapStr `json:"data"`
}

// UpdatedDataResult common update operation result
type UpdatedDataResult struct {
	OriginIndex int64  `json:"origin_index"`
	ID          uint64 `json:"id"`
}

// SetDataResult common set result definition
type SetDataResult struct {
	UpdatedCount `json:",inline"`
	CreatedCount `json:",inline"`
	Created      []CreatedDataResult `json:"created"`
	Updated      []UpdatedDataResult `json:"updated"`
	Exceptions   []ExceptionResult   `json:"exception"`
}

// CreateManyInfoResult create many function return this result struct
type CreateManyInfoResult struct {
	Created    []CreatedDataResult  `json:"created"`
	Repeated   []RepeatedDataResult `json:"repeated"`
	Exceptions []ExceptionResult    `json:"exception"`
}

// CreateManyDataResult the data struct definition in create many function result
type CreateManyDataResult struct {
	CreateManyInfoResult `json:",inline"`
}

// CreateOneDataResult the data struct definition in create one function result
type CreateOneDataResult struct {
	Created CreatedDataResult `json:"created"`
}

// SearchResp common search response
type SearchResp struct {
	BaseResp `json:",inline"`
	Data     SearchDataResult `json:"data"`
}

// SearchDataResult common search data result
type SearchDataResult struct {
	Count int64           `json:"count"`
	Info  []mapstr.MapStr `json:"info"`
}

// QueryInstAssociationResult query inst association result definition
type QueryInstAssociationResult struct {
	Count uint64     `json:"count"`
	Info  []InstAsst `json:"info"`
}

// ReadInstAssociationResult TODO
type ReadInstAssociationResult struct {
	BaseResp `json:",inline"`
	Data     QueryInstAssociationResult `json:"data"`
}

// OperaterException  result
type OperaterException struct {
	BaseResp `json:",inline"`
	Data     []ExceptionResult `json:"data"`
}

// Uint64DataResponse TODO
type Uint64DataResponse struct {
	BaseResp `json:",inline"`
	Data     uint64 `json:"data"`
}

// TransferException TODO
type TransferException struct {
	HostID []int64 `json:"bk_host_id"`
	ErrMsg string  `json:"bk_error_msg"`
}

// TransferExceptionResult TODO
type TransferExceptionResult struct {
	BaseResp `json:",inline"`
	Data     TransferException `json:"data"`
}

// SyncHostIdentifierResult sync host identifier result struct
type SyncHostIdentifierResult struct {
	SuccessList []int64 `json:"success_list"`
	FailedList  []int64 `json:"failed_list"`
	TaskID      string  `json:"task_id"`
}
