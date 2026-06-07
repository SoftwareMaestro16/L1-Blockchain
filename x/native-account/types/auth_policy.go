package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

const (
	AuthModeMultisig  = "multisig"
	AuthModeThreshold = "threshold"
	AuthModeWeighted  = "weighted"
	AuthModeTwoDevice = "two_device"

	AuthKeyRolePrimary  = "primary"
	AuthKeyRoleDevice   = "device"
	AuthKeyRoleRecovery = "recovery"

	AuthOperationTransfer         = "transfer"
	AuthOperationStakingChange    = "staking_change"
	AuthOperationAuthPolicyUpdate = "auth_policy_update"
	AuthOperationRecoverAccount   = "recover_account"
	AuthOperationFreezeAccount    = "freeze_account"
	AuthOperationUnfreezeAccount  = "unfreeze_account"
	AuthOperationMetadataUpdate   = "metadata_update"
	AuthOperationParamsUpdate     = "params_update"
)

type AuthKey struct {
	ID        string `json:"id"`
	PublicKey string `json:"public_key"`
	Role      string `json:"role,omitempty"`
}

type AuthWeight struct {
	KeyID  string `json:"key_id"`
	Weight uint64 `json:"weight"`
}

type RecoveryPolicy struct {
	Keys              []string `json:"keys,omitempty"`
	Threshold         uint64   `json:"threshold,omitempty"`
	TimelockEndHeight uint64   `json:"timelock_end_height,omitempty"`
}

type TimelockPolicy struct {
	AuthPolicyUpdateEndHeight uint64 `json:"auth_policy_update_end_height,omitempty"`
	RecoveryEndHeight         uint64 `json:"recovery_end_height,omitempty"`
}

type SpendingLimit struct {
	Operation string `json:"operation"`
	MaxAmount uint64 `json:"max_amount"`
}

type AuthzResult struct {
	Authorized bool
	Mode       string
	Signers    []string
	Weight     uint64
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

func AuthorizeAuthPolicy(account Account, msg ExternalMessage) (AuthzResult, error) {
	policy := account.AuthPolicy.Normalize()
	if err := policy.Validate(); err != nil {
		return AuthzResult{}, err
	}
	keys := effectiveAuthKeys(account)
	signers := canonicalSigners(msg.Signers)
	switch policy.Mode {
	case AuthModeSingleKey:
		if operationWithinSpendingLimit(policy, msg.Operation, msg.Amount) || len(policy.SpendingLimits) == 0 {
			if signedByAnyKey(keys, signers) {
				return AuthzResult{Authorized: true, Mode: policy.Mode, Signers: signers}, nil
			}
			return AuthzResult{}, errors.New("external message missing authorized single-key signer")
		}
		return AuthzResult{}, errors.New("external message amount exceeds single-key spending limit")
	case AuthModeMultisig, AuthModeThreshold:
		threshold := policy.Threshold
		if policy.Mode == AuthModeMultisig && threshold == 0 {
			threshold = uint64(len(keys))
		}
		count := countSignedKeys(keys, signers)
		if count < threshold {
			return AuthzResult{}, fmt.Errorf("external message signatures %d below threshold %d", count, threshold)
		}
		return AuthzResult{Authorized: true, Mode: policy.Mode, Signers: signers}, nil
	case AuthModeWeighted:
		weight := signedWeight(policy.Weights, signers)
		if weight < policy.Threshold {
			return AuthzResult{}, fmt.Errorf("external message signer weight %d below threshold %d", weight, policy.Threshold)
		}
		return AuthzResult{Authorized: true, Mode: policy.Mode, Signers: signers, Weight: weight}, nil
	case AuthModeTwoDevice:
		if operationWithinSpendingLimit(policy, msg.Operation, msg.Amount) && signedByRole(keys, signers, AuthKeyRolePrimary) {
			return AuthzResult{Authorized: true, Mode: policy.Mode, Signers: signers}, nil
		}
		if signedByRole(keys, signers, AuthKeyRolePrimary) && signedByRole(keys, signers, AuthKeyRoleDevice) {
			return AuthzResult{Authorized: true, Mode: policy.Mode, Signers: signers}, nil
		}
		return AuthzResult{}, errors.New("two-device auth requires primary and device signatures")
	default:
		return AuthzResult{}, fmt.Errorf("unsupported external auth policy mode %q", policy.Mode)
	}
}

func AuthorizeRecoveryPolicy(account Account, msg MsgRecoverAccount) error {
	policy := account.AuthPolicy.Normalize()
	if err := policy.RecoveryPolicy.Validate(); err != nil {
		return err
	}
	if len(policy.RecoveryPolicy.Keys) == 0 {
		return errors.New("native account recovery policy is not configured")
	}
	if msg.CurrentHeight < policy.RecoveryPolicy.TimelockEndHeight || msg.CurrentHeight < policy.Timelock.RecoveryEndHeight {
		return errors.New("native account recovery timelock has not expired")
	}
	signers := canonicalSigners(msg.Signers)
	count := countSignedStrings(policy.RecoveryPolicy.Keys, signers)
	if count < policy.RecoveryPolicy.Threshold {
		return fmt.Errorf("recovery signatures %d below threshold %d", count, policy.RecoveryPolicy.Threshold)
	}
	return nil
}

func effectiveAuthKeys(account Account) []AuthKey {
	policy := account.AuthPolicy.Normalize()
	if len(policy.Keys) > 0 {
		return policy.Keys
	}
	keys := make([]AuthKey, 0, len(account.PubKeys))
	for idx, pubKey := range account.PubKeys {
		keys = append(keys, AuthKey{ID: fmt.Sprintf("legacy-%020d", idx), PublicKey: pubKey, Role: AuthKeyRolePrimary})
	}
	return keys
}

func validateAuthKeys(keys []AuthKey, allowEmpty bool) error {
	if len(keys) == 0 {
		if allowEmpty {
			return nil
		}
		return errors.New("native account auth policy keys are required")
	}
	previous := ""
	for _, key := range keys {
		key = key.Normalize()
		if key.ID == "" || key.PublicKey == "" {
			return errors.New("native account auth key id and public key are required")
		}
		if containsSecretLikeText(key.ID) || containsSecretLikeText(key.PublicKey) || containsSecretLikeText(key.Role) {
			return errors.New("native account auth policy must not contain private keys or seed phrases")
		}
		if key.ID <= previous {
			return errors.New("native account auth keys must be sorted and unique")
		}
		previous = key.ID
	}
	return nil
}

func validateAuthWeights(keys []AuthKey, weights []AuthWeight, threshold uint64) error {
	if len(weights) == 0 {
		return errors.New("native account weighted auth weights are required")
	}
	keyIDs := map[string]struct{}{}
	for _, key := range keys {
		keyIDs[key.ID] = struct{}{}
	}
	total := uint64(0)
	previous := ""
	for _, weight := range weights {
		weight.KeyID = strings.TrimSpace(weight.KeyID)
		if weight.KeyID == "" || weight.Weight == 0 {
			return errors.New("native account weighted auth key id and weight are required")
		}
		if containsSecretLikeText(weight.KeyID) {
			return errors.New("native account auth policy must not contain private keys or seed phrases")
		}
		if _, found := keyIDs[weight.KeyID]; !found {
			return fmt.Errorf("native account weighted auth references unknown key %q", weight.KeyID)
		}
		if weight.KeyID <= previous {
			return errors.New("native account weighted auth weights must be sorted and unique")
		}
		previous = weight.KeyID
		total += weight.Weight
	}
	if total < threshold {
		return errors.New("native account weighted auth total weight below threshold")
	}
	return nil
}

func validateSpendingLimits(limits []SpendingLimit) error {
	previous := ""
	for _, limit := range limits {
		operation := strings.TrimSpace(limit.Operation)
		if operation == "" {
			return errors.New("native account spending limit operation is required")
		}
		if containsSecretLikeText(operation) {
			return errors.New("native account spending limits must not contain private keys or seed phrases")
		}
		key := fmt.Sprintf("%s/%020d", operation, limit.MaxAmount)
		if key <= previous {
			return errors.New("native account spending limits must be sorted and unique")
		}
		previous = key
	}
	return nil
}

func hasAuthKeyRole(keys []AuthKey, role string) bool {
	for _, key := range keys {
		if key.Role == role {
			return true
		}
	}
	return false
}

func operationWithinSpendingLimit(policy AuthPolicy, operation string, amount uint64) bool {
	operation = strings.TrimSpace(operation)
	for _, limit := range policy.SpendingLimits {
		if limit.Operation == operation && amount <= limit.MaxAmount {
			return true
		}
	}
	return false
}

func canonicalSigners(signers []string) []string {
	out := make([]string, 0, len(signers))
	seen := map[string]struct{}{}
	for _, signer := range signers {
		signer = strings.TrimSpace(signer)
		if signer == "" {
			continue
		}
		if _, found := seen[signer]; found {
			continue
		}
		seen[signer] = struct{}{}
		out = append(out, signer)
	}
	sort.Strings(out)
	return out
}

func signedByAnyKey(keys []AuthKey, signers []string) bool {
	for _, key := range keys {
		for _, signer := range signers {
			if signer == key.PublicKey || signer == key.ID {
				return true
			}
		}
	}
	return false
}

func signedByRole(keys []AuthKey, signers []string, role string) bool {
	for _, key := range keys {
		if key.Role != role {
			continue
		}
		for _, signer := range signers {
			if signer == key.PublicKey || signer == key.ID {
				return true
			}
		}
	}
	return false
}

func countSignedKeys(keys []AuthKey, signers []string) uint64 {
	count := uint64(0)
	for _, key := range keys {
		for _, signer := range signers {
			if signer == key.PublicKey || signer == key.ID {
				count++
				break
			}
		}
	}
	return count
}

func countSignedStrings(keys []string, signers []string) uint64 {
	count := uint64(0)
	for _, key := range keys {
		for _, signer := range signers {
			if signer == key {
				count++
				break
			}
		}
	}
	return count
}

func signedWeight(weights []AuthWeight, signers []string) uint64 {
	total := uint64(0)
	signed := map[string]struct{}{}
	for _, signer := range signers {
		signed[signer] = struct{}{}
	}
	for _, weight := range weights {
		if _, found := signed[weight.KeyID]; found {
			total += weight.Weight
		}
	}
	return total
}
