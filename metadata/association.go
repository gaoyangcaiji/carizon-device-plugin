package metadata

import (
	"carizon-device-plugin/pkg/mapstr"
)

const (
	// AssociationFieldObjectID the association data field definition
	AssociationFieldObjectID = "bk_obj_id"
	// AssociationFieldAsstID the association data field bk_obj_asst_id
	AssociationFieldAsstID = "bk_obj_asst_id"
	// AssociationFieldSupplierAccount TODO
	// AssociationFieldObjectAttributeID the association data field definition
	// AssociationFieldObjectAttributeID = "bk_object_att_id"
	// AssociationFieldSupplierAccount the association data field definition
	AssociationFieldSupplierAccount = "bk_supplier_account"
	// AssociationFieldAssociationObjectID the association data field definition
	// AssociationFieldAssociationForward = "bk_asst_forward"
	// AssociationFieldAssociationObjectID the association data field definition
	AssociationFieldAssociationObjectID = "bk_asst_obj_id"
	// AssociationFieldAssociationId the association data field definition
	// AssociationFieldAssociationName = "bk_asst_name"
	// AssociationFieldAssociationId auto incr id
	AssociationFieldAssociationId = "id"
	// AssociationFieldAssociationKind TODO
	AssociationFieldAssociationKind = "bk_asst_id"
)

// SearchAssociationInstRequest TODO
type SearchAssociationInstRequest struct {
	Condition mapstr.MapStr `json:"condition"` // construct condition mapstr by condition.Condition
	ObjID     string        `json:"bk_obj_id"`
}

// SearchAssociationInstResult TODO
type SearchAssociationInstResult struct {
	BaseResp `json:",inline"`
	Data     []*InstAsst `json:"data"`
}

// InstAsst an association definition between instances.
type InstAsst struct {
	// sequence ID
	ID int64 `field:"id" json:"id,omitempty"`
	// inst id associate to ObjectID
	InstID int64 `field:"bk_inst_id" json:"bk_inst_id,omitempty" bson:"bk_inst_id"`
	// association source ObjectID
	ObjectID string `field:"bk_obj_id" json:"bk_obj_id,omitempty" bson:"bk_obj_id"`
	// inst id associate to AsstObjectID
	AsstInstID int64 `field:"bk_asst_inst_id" json:"bk_asst_inst_id,omitempty"  bson:"bk_asst_inst_id"`
	// association target ObjectID
	AsstObjectID string `field:"bk_asst_obj_id" json:"bk_asst_obj_id,omitempty" bson:"bk_asst_obj_id"`
	// bk_supplier_account
	OwnerID string `field:"bk_supplier_account" json:"bk_supplier_account,omitempty" bson:"bk_supplier_account"`
	// association id between two object
	ObjectAsstID string `field:"bk_obj_asst_id" json:"bk_obj_asst_id,omitempty" bson:"bk_obj_asst_id"`
	// association kind id
	AssociationKindID string `field:"bk_asst_id" json:"bk_asst_id,omitempty" bson:"bk_asst_id"`

	// BizID the business ID
	BizID int64 `field:"bk_biz_id" json:"bk_biz_id,omitempty" bson:"bk_biz_id"`
}

// InstAsstQueryCondition TODO
type InstAsstQueryCondition struct {
	Cond  QueryCondition `json:"cond"`
	ObjID string         `json:"bk_obj_id"`
}
