package types

const (
	EventTypeSubmitSpend	= "treasury_submit_spend"
	EventTypeApproveSpend	= "treasury_approve_spend"
	EventTypeRejectSpend	= "treasury_reject_spend"
	EventTypeExecuteSpend	= "treasury_execute_spend"
	EventTypeCancelSpend	= "treasury_cancel_spend"
	EventTypeUpdateParams	= "treasury_update_params"

	AttributeKeyAuthority	= "authority"
	AttributeKeyActor	= "actor"
	AttributeKeySpendID	= "spend_id"
	AttributeKeyRecipient	= "recipient"
	AttributeKeyAmount	= "amount"
	AttributeKeyBucket	= "bucket"
	AttributeKeyEpoch	= "epoch"
)
