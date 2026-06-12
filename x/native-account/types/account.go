package types

const (
	AccountVersionV1	= uint64(1)
	AccountVersionV2	= uint64(2)
	CurrentAccountVersion	= AccountVersionV2

	AccountStatusInactive	= "inactive"
	AccountStatusActive	= "active"
	AccountStatusFrozen	= "frozen"
	AccountStatusRecovered	= "recovered"
	AccountStatusArchived	= "archived"
	AccountStatusClosed	= "closed"

	AccountFeatureInternalMessagesV2	= "internal_messages_v2"
	AccountFeatureRecoveryPolicyV2		= "recovery_policy_v2"
	AccountFeatureMetadataV2		= "metadata_v2"

	MaxMetadataHashBytes	= 128
	MaxDisplayNameHashBytes	= 128
	MaxDomainAliasBytes	= 253
	MaxPubKeyTextBytes	= 256
	MaxAuthPolicyModeBytes	= 64
	MaxFeatureFlagBytes	= 64
	MaxReputationIDBytes	= 128
)

type Account struct {
	Version			uint64		`protobuf:"varint,1,opt,name=version,proto3" json:"version"`
	AddressUser		string		`protobuf:"bytes,2,opt,name=address_user,json=addressUser,proto3" json:"address_user"`
	AddressRaw		string		`protobuf:"bytes,3,opt,name=address_raw,json=addressRaw,proto3" json:"address_raw"`
	PubKeys			[]string	`protobuf:"bytes,4,rep,name=pubkeys,proto3" json:"pubkeys,omitempty"`
	AccountNumber		uint64		`protobuf:"varint,5,opt,name=account_number,json=accountNumber,proto3" json:"account_number"`
	Sequence		uint64		`protobuf:"varint,6,opt,name=sequence,proto3" json:"sequence"`
	Status			string		`protobuf:"bytes,7,opt,name=status,proto3" json:"status"`
	AuthPolicy		AuthPolicy	`protobuf:"bytes,8,opt,name=auth_policy,json=authPolicy,proto3" json:"auth_policy"`
	FeatureFlags		[]string	`protobuf:"bytes,9,rep,name=features,proto3" json:"features,omitempty"`
	Metadata		AccountMetadata	`protobuf:"bytes,10,opt,name=metadata,proto3" json:"metadata,omitempty"`
	ReputationID		string		`protobuf:"bytes,11,opt,name=reputation_id,json=reputationID,proto3" json:"reputation_id,omitempty"`
	CreatedHeight		uint64		`protobuf:"varint,12,opt,name=created_height,json=createdHeight,proto3" json:"created_height"`
	LastActiveHeight	uint64		`protobuf:"varint,13,opt,name=last_active_height,json=lastActiveHeight,proto3" json:"last_active_height,omitempty"`
	LastStorageChargeHeight	uint64		`protobuf:"varint,14,opt,name=last_storage_charge_height,json=lastStorageChargeHeight,proto3" json:"last_storage_charge_height,omitempty"`
	StorageRentDebt		uint64		`protobuf:"varint,15,opt,name=storage_rent_debt,json=storageRentDebt,proto3" json:"storage_rent_debt,omitempty"`
}

type AuthPolicy struct {
	Version		uint64		`protobuf:"varint,1,opt,name=version,proto3" json:"version"`
	Mode		string		`protobuf:"bytes,2,opt,name=mode,proto3" json:"mode"`
	Keys		[]AuthKey	`protobuf:"bytes,3,rep,name=keys,proto3" json:"keys,omitempty"`
	Threshold	uint64		`protobuf:"varint,4,opt,name=threshold,proto3" json:"threshold,omitempty"`
	Weights		[]AuthWeight	`protobuf:"bytes,5,rep,name=weights,proto3" json:"weights,omitempty"`
	RecoveryPolicy	RecoveryPolicy	`protobuf:"bytes,6,opt,name=recovery_policy,json=recoveryPolicy,proto3" json:"recovery_policy,omitempty"`
	Timelock	TimelockPolicy	`protobuf:"bytes,7,opt,name=timelock,proto3" json:"timelock,omitempty"`
	SpendingLimits	[]SpendingLimit	`protobuf:"bytes,8,rep,name=spending_limits,json=spendingLimits,proto3" json:"spending_limits,omitempty"`
}

type AccountMetadata struct {
	MetadataHash	string	`protobuf:"bytes,1,opt,name=metadata_hash,json=metadataHash,proto3" json:"metadata_hash,omitempty"`
	DisplayNameHash	string	`protobuf:"bytes,2,opt,name=display_name_hash,json=displayNameHash,proto3" json:"display_name_hash,omitempty"`
	DomainAlias	string	`protobuf:"bytes,3,opt,name=domain_alias,json=domainAlias,proto3" json:"domain_alias,omitempty"`
	CreatedHeight	uint64	`protobuf:"varint,4,opt,name=created_height,json=createdHeight,proto3" json:"created_height,omitempty"`
}
