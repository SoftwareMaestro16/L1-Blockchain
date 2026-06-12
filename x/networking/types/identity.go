package types

import (
	"bytes"
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"fmt"
)

type IdentityTransitionRecord struct {
	TransitionID		string
	FromNodeID		string
	ToNodeID		string
	FromNodePubKey		[]byte
	ToNodePubKey		[]byte
	FromValidatorPubKey	[]byte
	ToValidatorPubKey	[]byte
	FromRoles		[]NodeRole
	ToRoles			[]NodeRole
	EffectiveHeight		uint64
	ExpiresHeight		uint64
	Nonce			[]byte
	OldSignature		[]byte
	NewSignature		[]byte
}

func NormalizeIdentityTransition(record IdentityTransitionRecord) IdentityTransitionRecord {
	record.TransitionID = normalizeHashText(record.TransitionID)
	record.FromNodeID = normalizeHashText(record.FromNodeID)
	record.ToNodeID = normalizeHashText(record.ToNodeID)
	record.FromNodePubKey = cloneBytes(record.FromNodePubKey)
	record.ToNodePubKey = cloneBytes(record.ToNodePubKey)
	record.FromValidatorPubKey = cloneBytes(record.FromValidatorPubKey)
	record.ToValidatorPubKey = cloneBytes(record.ToValidatorPubKey)
	record.FromRoles = normalizeRoles(record.FromRoles)
	record.ToRoles = normalizeRoles(record.ToRoles)
	record.Nonce = cloneBytes(record.Nonce)
	record.OldSignature = cloneBytes(record.OldSignature)
	record.NewSignature = cloneBytes(record.NewSignature)
	return record
}

func (r IdentityTransitionRecord) SigningPayload() ([]byte, error) {
	normalized := NormalizeIdentityTransition(r)
	if normalized.TransitionID == "" {
		normalized.TransitionID = ComputeIdentityTransitionID(normalized)
	}
	normalized.OldSignature = nil
	normalized.NewSignature = nil
	bz, err := json.Marshal(normalized)
	if err != nil {
		return nil, err
	}
	return bz, nil
}

func ComputeIdentityTransitionID(record IdentityTransitionRecord) string {
	normalized := NormalizeIdentityTransition(record)
	normalized.TransitionID = ""
	normalized.OldSignature = nil
	normalized.NewSignature = nil
	bz, _ := json.Marshal(normalized)
	return hashBytes("aetra-identity-transition-v1", bz)
}

func SignIdentityTransition(oldRecord, newRecord NodeRecord, oldPrivateKey, newPrivateKey ed25519.PrivateKey, networkSalt []byte, effectiveHeight, expiresHeight uint64, nonce []byte) (IdentityTransitionRecord, error) {
	if len(oldPrivateKey) != ed25519.PrivateKeySize || len(newPrivateKey) != ed25519.PrivateKeySize {
		return IdentityTransitionRecord{}, errors.New("networking identity transition private keys must be ed25519")
	}
	oldRecord = NormalizeNodeRecord(oldRecord)
	newRecord = NormalizeNodeRecord(newRecord)
	if err := oldRecord.Validate(networkSalt, 0); err != nil {
		return IdentityTransitionRecord{}, err
	}
	if err := newRecord.Validate(networkSalt, 0); err != nil {
		return IdentityTransitionRecord{}, err
	}
	oldPubKey, ok := oldPrivateKey.Public().(ed25519.PublicKey)
	if !ok || !bytes.Equal(oldPubKey, oldRecord.NodePubKey) {
		return IdentityTransitionRecord{}, errors.New("networking identity transition old private key does not match old node record")
	}
	newPubKey, ok := newPrivateKey.Public().(ed25519.PublicKey)
	if !ok || !bytes.Equal(newPubKey, newRecord.NodePubKey) {
		return IdentityTransitionRecord{}, errors.New("networking identity transition new private key does not match new node record")
	}
	record := NormalizeIdentityTransition(IdentityTransitionRecord{
		FromNodeID:		oldRecord.NodeID,
		ToNodeID:		newRecord.NodeID,
		FromNodePubKey:		oldRecord.NodePubKey,
		ToNodePubKey:		newRecord.NodePubKey,
		FromValidatorPubKey:	oldRecord.ValidatorPubKey,
		ToValidatorPubKey:	newRecord.ValidatorPubKey,
		FromRoles:		oldRecord.Roles,
		ToRoles:		newRecord.Roles,
		EffectiveHeight:	effectiveHeight,
		ExpiresHeight:		expiresHeight,
		Nonce:			nonce,
	})
	record.TransitionID = ComputeIdentityTransitionID(record)
	payload, err := record.SigningPayload()
	if err != nil {
		return IdentityTransitionRecord{}, err
	}
	record.OldSignature = ed25519.Sign(oldPrivateKey, payload)
	record.NewSignature = ed25519.Sign(newPrivateKey, payload)
	if err := ValidateIdentityTransition(oldRecord, newRecord, record, networkSalt, effectiveHeight); err != nil {
		return IdentityTransitionRecord{}, err
	}
	return NormalizeIdentityTransition(record), nil
}

func (r IdentityTransitionRecord) ValidateBasic() error {
	record := NormalizeIdentityTransition(r)
	if err := ValidateHash("networking identity transition id", record.TransitionID); err != nil {
		return err
	}
	if err := ValidateHash("networking identity transition from node id", record.FromNodeID); err != nil {
		return err
	}
	if err := ValidateHash("networking identity transition to node id", record.ToNodeID); err != nil {
		return err
	}
	if record.FromNodeID == record.ToNodeID {
		return errors.New("networking identity transition must rotate to a distinct node id")
	}
	if len(record.FromNodePubKey) != ed25519.PublicKeySize || len(record.ToNodePubKey) != ed25519.PublicKeySize {
		return fmt.Errorf("networking identity transition node pub keys must be %d bytes", ed25519.PublicKeySize)
	}
	if len(record.FromValidatorPubKey) > 0 && len(record.FromValidatorPubKey) != ed25519.PublicKeySize {
		return fmt.Errorf("networking identity transition validator pub keys must be %d bytes", ed25519.PublicKeySize)
	}
	if len(record.ToValidatorPubKey) > 0 && len(record.ToValidatorPubKey) != ed25519.PublicKeySize {
		return fmt.Errorf("networking identity transition validator pub keys must be %d bytes", ed25519.PublicKeySize)
	}
	if len(record.FromRoles) == 0 || len(record.ToRoles) == 0 {
		return errors.New("networking identity transition roles are required")
	}
	if len(record.FromRoles) > MaxRolesPerNode || len(record.ToRoles) > MaxRolesPerNode {
		return fmt.Errorf("networking identity transition roles must be <= %d", MaxRolesPerNode)
	}
	for _, role := range append(append([]NodeRole(nil), record.FromRoles...), record.ToRoles...) {
		if !IsNodeRole(role) {
			return fmt.Errorf("unknown networking node role %q", role)
		}
	}
	if record.EffectiveHeight == 0 || record.ExpiresHeight <= record.EffectiveHeight {
		return errors.New("networking identity transition height range is invalid")
	}
	if len(record.Nonce) == 0 || len(record.Nonce) > MaxNonceBytes {
		return fmt.Errorf("networking identity transition nonce must be between 1 and %d bytes", MaxNonceBytes)
	}
	if len(record.OldSignature) != ed25519.SignatureSize || len(record.NewSignature) != ed25519.SignatureSize {
		return fmt.Errorf("networking identity transition signatures must be %d bytes", ed25519.SignatureSize)
	}
	if ComputeIdentityTransitionID(record) != record.TransitionID {
		return errors.New("networking identity transition id mismatch")
	}
	return nil
}

func ValidateIdentityTransition(oldRecord, newRecord NodeRecord, transition IdentityTransitionRecord, networkSalt []byte, currentHeight uint64) error {
	oldRecord = NormalizeNodeRecord(oldRecord)
	newRecord = NormalizeNodeRecord(newRecord)
	transition = NormalizeIdentityTransition(transition)
	if err := transition.ValidateBasic(); err != nil {
		return err
	}
	if err := oldRecord.Validate(networkSalt, currentHeight); err != nil {
		return err
	}
	if err := newRecord.Validate(networkSalt, currentHeight); err != nil {
		return err
	}
	if currentHeight > 0 && (currentHeight < transition.EffectiveHeight || currentHeight > transition.ExpiresHeight) {
		return errors.New("networking identity transition is outside its active height range")
	}
	if transition.ExpiresHeight > oldRecord.ExpiresHeight || transition.ExpiresHeight > newRecord.ExpiresHeight {
		return errors.New("networking identity transition cannot outlive node records")
	}
	if transition.FromNodeID != oldRecord.NodeID || transition.ToNodeID != newRecord.NodeID {
		return errors.New("networking identity transition node id mismatch")
	}
	if !bytes.Equal(transition.FromNodePubKey, oldRecord.NodePubKey) || !bytes.Equal(transition.ToNodePubKey, newRecord.NodePubKey) {
		return errors.New("networking identity transition node pub key mismatch")
	}
	if !bytes.Equal(transition.FromValidatorPubKey, oldRecord.ValidatorPubKey) || !bytes.Equal(transition.ToValidatorPubKey, newRecord.ValidatorPubKey) {
		return errors.New("networking identity transition validator pub key mismatch")
	}
	if transition.FromNodeID != ComputeNodeID(identityPubKeyForRecord(oldRecord), networkSalt) {
		return errors.New("networking identity transition old node id does not match identity key")
	}
	if transition.ToNodeID != ComputeNodeID(identityPubKeyForRecord(newRecord), networkSalt) {
		return errors.New("networking identity transition new node id does not match identity key")
	}
	if !equalNodeRoles(transition.FromRoles, oldRecord.Roles) || !equalNodeRoles(transition.ToRoles, newRecord.Roles) {
		return errors.New("networking identity transition roles must match signed node records")
	}
	payload, err := transition.SigningPayload()
	if err != nil {
		return err
	}
	if !ed25519.Verify(oldRecord.NodePubKey, payload, transition.OldSignature) {
		return errors.New("networking identity transition old signature verification failed")
	}
	if !ed25519.Verify(newRecord.NodePubKey, payload, transition.NewSignature) {
		return errors.New("networking identity transition new signature verification failed")
	}
	return nil
}

func identityPubKeyForRecord(record NodeRecord) []byte {
	if len(record.ValidatorPubKey) > 0 {
		return record.ValidatorPubKey
	}
	return record.NodePubKey
}

func equalNodeRoles(left, right []NodeRole) bool {
	left = normalizeRoles(left)
	right = normalizeRoles(right)
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}
