package types

import (
	"errors"
	"sort"
	"strings"
)

const (
	AuthModeMultisig	= "multisig"
	AuthModeThreshold	= "threshold"
	AuthModeWeighted	= "weighted"
	AuthModeTwoDevice	= "two_device"

	AuthKeyRolePrimary	= "primary"
	AuthKeyRoleDevice	= "device"
	AuthKeyRoleRecovery	= "recovery"

	AuthOperationTransfer		= "transfer"
	AuthOperationStakingChange	= "staking_change"
	AuthOperationAuthPolicyUpdate	= "auth_policy_update"
	AuthOperationRecoverAccount	= "recover_account"
	AuthOperationFreezeAccount	= "freeze_account"
	AuthOperationPayStorageDebt	= "pay_storage_debt"
	AuthOperationUnfreezeAccount	= "unfreeze_account"
	AuthOperationMetadataUpdate	= "metadata_update"
	AuthOperationParamsUpdate	= "params_update"
)

type AuthKey struct {
	ID		string	`protobuf:"bytes,1,opt,name=id,proto3" json:"id"`
	PublicKey	string	`protobuf:"bytes,2,opt,name=public_key,json=publicKey,proto3" json:"public_key"`
	Role		string	`protobuf:"bytes,3,opt,name=role,proto3" json:"role,omitempty"`
}

type AuthWeight struct {
	KeyID	string	`protobuf:"bytes,1,opt,name=key_id,json=keyID,proto3" json:"key_id"`
	Weight	uint64	`protobuf:"varint,2,opt,name=weight,proto3" json:"weight"`
}

type RecoveryPolicy struct {
	Keys			[]string	`protobuf:"bytes,1,rep,name=keys,proto3" json:"keys,omitempty"`
	Threshold		uint64		`protobuf:"varint,2,opt,name=threshold,proto3" json:"threshold,omitempty"`
	TimelockEndHeight	uint64		`protobuf:"varint,3,opt,name=timelock_end_height,json=timelockEndHeight,proto3" json:"timelock_end_height,omitempty"`
}

type TimelockPolicy struct {
	AuthPolicyUpdateEndHeight	uint64	`protobuf:"varint,1,opt,name=auth_policy_update_end_height,json=authPolicyUpdateEndHeight,proto3" json:"auth_policy_update_end_height,omitempty"`
	RecoveryEndHeight		uint64	`protobuf:"varint,2,opt,name=recovery_end_height,json=recoveryEndHeight,proto3" json:"recovery_end_height,omitempty"`
}

type SpendingLimit struct {
	Operation	string	`protobuf:"bytes,1,opt,name=operation,proto3" json:"operation"`
	MaxAmount	uint64	`protobuf:"varint,2,opt,name=max_amount,json=maxAmount,proto3" json:"max_amount"`
}

type AuthzResult struct {
	Authorized	bool
	Mode		string
	Signers		[]string
	Weight		uint64
}

func (p AuthPolicy) Normalize() AuthPolicy {
	p.Mode = strings.TrimSpace(p.Mode)
	p.Keys = append([]AuthKey(nil), p.Keys...)
	for i := range p.Keys {
		p.Keys[i] = p.Keys[i].Normalize()
	}
	sort.SliceStable(p.Keys, func(i, j int) bool { return p.Keys[i].ID < p.Keys[j].ID })
	p.Weights = append([]AuthWeight(nil), p.Weights...)
	for i := range p.Weights {
		p.Weights[i].KeyID = strings.TrimSpace(p.Weights[i].KeyID)
	}
	sort.SliceStable(p.Weights, func(i, j int) bool { return p.Weights[i].KeyID < p.Weights[j].KeyID })
	p.RecoveryPolicy = p.RecoveryPolicy.Normalize()
	p.SpendingLimits = append([]SpendingLimit(nil), p.SpendingLimits...)
	for i := range p.SpendingLimits {
		p.SpendingLimits[i].Operation = strings.TrimSpace(p.SpendingLimits[i].Operation)
	}
	sort.SliceStable(p.SpendingLimits, func(i, j int) bool {
		if p.SpendingLimits[i].Operation != p.SpendingLimits[j].Operation {
			return p.SpendingLimits[i].Operation < p.SpendingLimits[j].Operation
		}
		return p.SpendingLimits[i].MaxAmount < p.SpendingLimits[j].MaxAmount
	})
	return p
}

func (k AuthKey) Normalize() AuthKey {
	k.ID = strings.TrimSpace(k.ID)
	k.PublicKey = strings.TrimSpace(k.PublicKey)
	k.Role = strings.TrimSpace(k.Role)
	return k
}

func (p RecoveryPolicy) Normalize() RecoveryPolicy {
	p.Keys = append([]string(nil), p.Keys...)
	for i := range p.Keys {
		p.Keys[i] = strings.TrimSpace(p.Keys[i])
	}
	sort.Strings(p.Keys)
	return p
}

func (p RecoveryPolicy) Validate() error {
	p = p.Normalize()
	if len(p.Keys) == 0 && p.Threshold == 0 && p.TimelockEndHeight == 0 {
		return nil
	}
	if len(p.Keys) == 0 {
		return errors.New("native account recovery policy keys are required")
	}
	if p.Threshold == 0 || p.Threshold > uint64(len(p.Keys)) {
		return errors.New("native account recovery policy threshold is invalid")
	}
	previous := ""
	for _, key := range p.Keys {
		if key == "" {
			return errors.New("native account recovery key is required")
		}
		if containsSecretLikeText(key) {
			return errors.New("native account recovery policy must not contain private keys or seed phrases")
		}
		if key <= previous {
			return errors.New("native account recovery keys must be sorted and unique")
		}
		previous = key
	}
	return nil
}

func (p TimelockPolicy) Validate() error {
	return nil
}
