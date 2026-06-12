package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	IdentityCrossZoneBindingV2Prefix	= IdentityZonePrefix + "/cross_zone/bindings"
	IdentityCrossZoneInvalidationV2Prefix	= IdentityZonePrefix + "/cross_zone/invalidations"
)

type IdentityCrossZoneBindingTargetType string
type IdentityCrossZoneBindingConfirmationType string
type IdentityCrossZoneBindingInvalidationReason string

const (
	IdentityBindingTargetAccount		IdentityCrossZoneBindingTargetType	= "account"
	IdentityBindingTargetService		IdentityCrossZoneBindingTargetType	= "service"
	IdentityBindingTargetContract		IdentityCrossZoneBindingTargetType	= "contract"
	IdentityBindingTargetZoneEndpoint	IdentityCrossZoneBindingTargetType	= "zone_endpoint"
	IdentityBindingTargetComposite		IdentityCrossZoneBindingTargetType	= "composite_identity_object"

	IdentityBindingConfirmationProof	IdentityCrossZoneBindingConfirmationType	= "proof_verifiable"
	IdentityBindingConfirmationMessage	IdentityCrossZoneBindingConfirmationType	= "message_confirmed"

	IdentityBindingInvalidatedCreated	IdentityCrossZoneBindingInvalidationReason	= "created"
	IdentityBindingInvalidatedUpdated	IdentityCrossZoneBindingInvalidationReason	= "updated"
	IdentityBindingInvalidatedRevoked	IdentityCrossZoneBindingInvalidationReason	= "revoked"
	IdentityBindingInvalidatedExpired	IdentityCrossZoneBindingInvalidationReason	= "expired"
)

type CrossZoneIdentityBindingV2 struct {
	IdentityID	string
	TargetZone	string
	TargetType	IdentityCrossZoneBindingTargetType
	TargetKey	string
	ProofRequired	bool
	ExpiresHeight	uint64
	BindingVersion	uint64
	BindingHash	string
}

type CrossZoneBindingConfirmationV2 struct {
	ConfirmationType	IdentityCrossZoneBindingConfirmationType
	ProofRoot		string
	ProofHash		string
	MessageID		string
	ReceiptHash		string
	ConfirmedHeight		uint64
	ConfirmationHash	string
}

type CrossZoneBindingInvalidationEventV2 struct {
	EventID		string
	BindingHash	string
	IdentityID	string
	TargetZone	string
	TargetType	IdentityCrossZoneBindingTargetType
	TargetKey	string
	Reason		IdentityCrossZoneBindingInvalidationReason
	Height		uint64
	EventHash	string
}

type CrossZoneIdentityBindingStateV2 struct {
	Bindings	[]CrossZoneIdentityBindingV2
	Confirmations	[]CrossZoneBindingConfirmationV2
	Invalidations	[]CrossZoneBindingInvalidationEventV2
	Height		uint64
	RootHash	string
}

type MsgBindCrossZoneIdentityV2 struct {
	Authority	sdk.AccAddress
	Binding		CrossZoneIdentityBindingV2
	Confirmation	CrossZoneBindingConfirmationV2
	Height		uint64
	MessageHash	string
}

type MsgUpdateCrossZoneIdentityBindingV2 struct {
	Authority	sdk.AccAddress
	Binding		CrossZoneIdentityBindingV2
	Confirmation	CrossZoneBindingConfirmationV2
	Height		uint64
	MessageHash	string
}

type MsgRevokeCrossZoneIdentityBindingV2 struct {
	Authority	sdk.AccAddress
	IdentityID	string
	TargetZone	string
	TargetType	IdentityCrossZoneBindingTargetType
	TargetKey	string
	Height		uint64
	MessageHash	string
}

func CrossZoneIdentityBindingV2Key(identityID, targetZone string, targetType IdentityCrossZoneBindingTargetType, targetKey string) (string, error) {
	if err := validateIdentityGraphID("cross-zone binding identity id", identityID); err != nil {
		return "", err
	}
	if err := validateCrossZoneTargetZoneV2(targetZone); err != nil {
		return "", err
	}
	if !IsCrossZoneBindingTargetTypeV2(targetType) {
		return "", fmt.Errorf("unknown cross-zone binding target type %q", targetType)
	}
	if err := validateIdentityGraphID("cross-zone binding target key", targetKey); err != nil {
		return "", err
	}
	return IdentityCrossZoneBindingV2Prefix + "/" + identityID + "/" + targetZone + "/" + string(targetType) + "/" + targetKey, nil
}

func CrossZoneBindingInvalidationV2Key(height uint64, eventID string) (string, error) {
	if height == 0 {
		return "", errors.New("cross-zone binding invalidation height must be positive")
	}
	if err := validateIdentityGraphID("cross-zone binding invalidation event id", eventID); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%020d/%s", IdentityCrossZoneInvalidationV2Prefix, height, eventID), nil
}

func NewCrossZoneIdentityBindingV2(binding CrossZoneIdentityBindingV2) (CrossZoneIdentityBindingV2, error) {
	if binding.BindingHash != "" {
		return CrossZoneIdentityBindingV2{}, errors.New("cross-zone identity binding hash must be empty before construction")
	}
	if err := binding.ValidateFormat(); err != nil {
		return CrossZoneIdentityBindingV2{}, err
	}
	binding.BindingHash = ComputeCrossZoneIdentityBindingV2Hash(binding)
	return binding, binding.Validate()
}

func NewCrossZoneBindingConfirmationV2(confirmation CrossZoneBindingConfirmationV2) (CrossZoneBindingConfirmationV2, error) {
	if confirmation.ConfirmationHash != "" {
		return CrossZoneBindingConfirmationV2{}, errors.New("cross-zone binding confirmation hash must be empty before construction")
	}
	if err := confirmation.ValidateFormat(); err != nil {
		return CrossZoneBindingConfirmationV2{}, err
	}
	confirmation.ConfirmationHash = ComputeCrossZoneBindingConfirmationV2Hash(confirmation)
	return confirmation, confirmation.Validate()
}

func NewCrossZoneBindingInvalidationEventV2(event CrossZoneBindingInvalidationEventV2) (CrossZoneBindingInvalidationEventV2, error) {
	if event.EventHash != "" {
		return CrossZoneBindingInvalidationEventV2{}, errors.New("cross-zone binding invalidation event hash must be empty before construction")
	}
	if event.EventID == "" {
		event.EventID = ComputeCrossZoneBindingInvalidationEventIDV2(event)
	}
	if err := event.ValidateFormat(); err != nil {
		return CrossZoneBindingInvalidationEventV2{}, err
	}
	event.EventHash = ComputeCrossZoneBindingInvalidationEventV2Hash(event)
	return event, event.Validate()
}

func BuildCrossZoneIdentityBindingStateV2(bindings []CrossZoneIdentityBindingV2, confirmations []CrossZoneBindingConfirmationV2, invalidations []CrossZoneBindingInvalidationEventV2, height uint64) (CrossZoneIdentityBindingStateV2, error) {
	state := CrossZoneIdentityBindingStateV2{
		Bindings:	normalizeCrossZoneIdentityBindingsV2(bindings),
		Confirmations:	normalizeCrossZoneBindingConfirmationsV2(confirmations),
		Invalidations:	normalizeCrossZoneBindingInvalidationsV2(invalidations),
		Height:		height,
	}
	if err := state.ValidateFormat(); err != nil {
		return CrossZoneIdentityBindingStateV2{}, err
	}
	state.RootHash = ComputeCrossZoneIdentityBindingStateV2Root(state)
	return state, state.Validate()
}

func ApplyCrossZoneIdentityBindV2(state CrossZoneIdentityBindingStateV2, graph IdentityGraphStateV2, msg MsgBindCrossZoneIdentityV2) (CrossZoneIdentityBindingStateV2, CrossZoneBindingInvalidationEventV2, error) {
	if err := msg.Validate(graph); err != nil {
		return CrossZoneIdentityBindingStateV2{}, CrossZoneBindingInvalidationEventV2{}, err
	}
	if _, found := findCrossZoneIdentityBindingV2(state.Bindings, msg.Binding); found {
		return CrossZoneIdentityBindingStateV2{}, CrossZoneBindingInvalidationEventV2{}, errors.New("cross-zone identity binding already exists")
	}
	event, err := NewCrossZoneBindingInvalidationEventV2(CrossZoneBindingInvalidationEventV2{
		BindingHash:	msg.Binding.BindingHash,
		IdentityID:	msg.Binding.IdentityID,
		TargetZone:	msg.Binding.TargetZone,
		TargetType:	msg.Binding.TargetType,
		TargetKey:	msg.Binding.TargetKey,
		Reason:		IdentityBindingInvalidatedCreated,
		Height:		msg.Height,
	})
	if err != nil {
		return CrossZoneIdentityBindingStateV2{}, CrossZoneBindingInvalidationEventV2{}, err
	}
	nextBindings := append([]CrossZoneIdentityBindingV2(nil), state.Bindings...)
	nextBindings = append(nextBindings, msg.Binding)
	nextConfirmations := append([]CrossZoneBindingConfirmationV2(nil), state.Confirmations...)
	nextConfirmations = append(nextConfirmations, msg.Confirmation)
	nextInvalidations := append([]CrossZoneBindingInvalidationEventV2(nil), state.Invalidations...)
	nextInvalidations = append(nextInvalidations, event)
	next, err := BuildCrossZoneIdentityBindingStateV2(nextBindings, nextConfirmations, nextInvalidations, msg.Height)
	return next, event, err
}

func ApplyCrossZoneIdentityBindingUpdateV2(state CrossZoneIdentityBindingStateV2, graph IdentityGraphStateV2, msg MsgUpdateCrossZoneIdentityBindingV2) (CrossZoneIdentityBindingStateV2, CrossZoneBindingInvalidationEventV2, error) {
	if err := msg.Validate(graph); err != nil {
		return CrossZoneIdentityBindingStateV2{}, CrossZoneBindingInvalidationEventV2{}, err
	}
	idx, found := findCrossZoneIdentityBindingV2(state.Bindings, msg.Binding)
	if !found {
		return CrossZoneIdentityBindingStateV2{}, CrossZoneBindingInvalidationEventV2{}, errors.New("cross-zone identity binding not found")
	}
	if msg.Binding.BindingVersion <= state.Bindings[idx].BindingVersion {
		return CrossZoneIdentityBindingStateV2{}, CrossZoneBindingInvalidationEventV2{}, errors.New("cross-zone identity binding update must increase version")
	}
	event, err := NewCrossZoneBindingInvalidationEventV2(CrossZoneBindingInvalidationEventV2{
		BindingHash:	state.Bindings[idx].BindingHash,
		IdentityID:	msg.Binding.IdentityID,
		TargetZone:	msg.Binding.TargetZone,
		TargetType:	msg.Binding.TargetType,
		TargetKey:	msg.Binding.TargetKey,
		Reason:		IdentityBindingInvalidatedUpdated,
		Height:		msg.Height,
	})
	if err != nil {
		return CrossZoneIdentityBindingStateV2{}, CrossZoneBindingInvalidationEventV2{}, err
	}
	nextBindings := append([]CrossZoneIdentityBindingV2(nil), state.Bindings...)
	nextBindings[idx] = msg.Binding
	nextConfirmations := append([]CrossZoneBindingConfirmationV2(nil), state.Confirmations...)
	nextConfirmations = append(nextConfirmations, msg.Confirmation)
	nextInvalidations := append([]CrossZoneBindingInvalidationEventV2(nil), state.Invalidations...)
	nextInvalidations = append(nextInvalidations, event)
	next, err := BuildCrossZoneIdentityBindingStateV2(nextBindings, nextConfirmations, nextInvalidations, msg.Height)
	return next, event, err
}

func ApplyCrossZoneIdentityBindingRevokeV2(state CrossZoneIdentityBindingStateV2, graph IdentityGraphStateV2, msg MsgRevokeCrossZoneIdentityBindingV2) (CrossZoneIdentityBindingStateV2, CrossZoneBindingInvalidationEventV2, error) {
	if err := msg.Validate(graph); err != nil {
		return CrossZoneIdentityBindingStateV2{}, CrossZoneBindingInvalidationEventV2{}, err
	}
	probe, err := NewCrossZoneIdentityBindingV2(CrossZoneIdentityBindingV2{
		IdentityID:	msg.IdentityID,
		TargetZone:	msg.TargetZone,
		TargetType:	msg.TargetType,
		TargetKey:	msg.TargetKey,
		ProofRequired:	false,
		ExpiresHeight:	msg.Height + 1,
		BindingVersion:	1,
	})
	if err != nil {
		return CrossZoneIdentityBindingStateV2{}, CrossZoneBindingInvalidationEventV2{}, err
	}
	idx, found := findCrossZoneIdentityBindingV2(state.Bindings, probe)
	if !found {
		return CrossZoneIdentityBindingStateV2{}, CrossZoneBindingInvalidationEventV2{}, errors.New("cross-zone identity binding not found")
	}
	event, err := NewCrossZoneBindingInvalidationEventV2(CrossZoneBindingInvalidationEventV2{
		BindingHash:	state.Bindings[idx].BindingHash,
		IdentityID:	msg.IdentityID,
		TargetZone:	msg.TargetZone,
		TargetType:	msg.TargetType,
		TargetKey:	msg.TargetKey,
		Reason:		IdentityBindingInvalidatedRevoked,
		Height:		msg.Height,
	})
	if err != nil {
		return CrossZoneIdentityBindingStateV2{}, CrossZoneBindingInvalidationEventV2{}, err
	}
	nextBindings := append([]CrossZoneIdentityBindingV2(nil), state.Bindings...)
	nextBindings = append(nextBindings[:idx], nextBindings[idx+1:]...)
	nextInvalidations := append([]CrossZoneBindingInvalidationEventV2(nil), state.Invalidations...)
	nextInvalidations = append(nextInvalidations, event)
	next, err := BuildCrossZoneIdentityBindingStateV2(nextBindings, state.Confirmations, nextInvalidations, msg.Height)
	return next, event, err
}

func ValidateCrossZoneIdentityBindingAuthorizationV2(graph IdentityGraphStateV2, identityID string, authority sdk.AccAddress) error {
	if err := graph.Validate(); err != nil {
		return err
	}
	if err := validateSpecAddress("cross-zone identity binding authority", authority); err != nil {
		return err
	}
	for _, node := range graph.Nodes {
		if node.IdentityID == identityID {
			if len(node.Owner) == 0 {
				return errors.New("cross-zone identity binding identity node has no owner")
			}
			if !addressesEqual(node.Owner, authority) {
				return errors.New("cross-zone identity binding must be authorized by identity owner")
			}
			return nil
		}
	}
	return errors.New("cross-zone identity binding identity node not found")
}

func ValidateCrossZoneIdentityBindingForRoutingV2(binding CrossZoneIdentityBindingV2, confirmation CrossZoneBindingConfirmationV2, height uint64) error {
	if height == 0 {
		return errors.New("cross-zone identity binding routing height must be positive")
	}
	if err := binding.Validate(); err != nil {
		return err
	}
	if err := ValidateCrossZoneBindingConfirmationForBindingV2(binding, confirmation); err != nil {
		return err
	}
	if binding.ExpiresHeight <= height {
		return errors.New("expired cross-zone identity binding cannot be used for routing")
	}
	return nil
}

func ValidateCrossZoneBindingConfirmationForBindingV2(binding CrossZoneIdentityBindingV2, confirmation CrossZoneBindingConfirmationV2) error {
	if err := binding.Validate(); err != nil {
		return err
	}
	if err := confirmation.Validate(); err != nil {
		return err
	}
	if binding.ProofRequired && confirmation.ConfirmationType != IdentityBindingConfirmationProof {
		return errors.New("cross-zone identity binding target must be proof-verifiable")
	}
	if !binding.ProofRequired && confirmation.ConfirmationType != IdentityBindingConfirmationProof && confirmation.ConfirmationType != IdentityBindingConfirmationMessage {
		return errors.New("cross-zone identity binding target must be proof-verifiable or message-confirmed")
	}
	return nil
}

func (binding CrossZoneIdentityBindingV2) ValidateFormat() error {
	if _, err := CrossZoneIdentityBindingV2Key(binding.IdentityID, binding.TargetZone, binding.TargetType, binding.TargetKey); err != nil {
		return err
	}
	if binding.ExpiresHeight == 0 {
		return errors.New("cross-zone identity binding expires_height must be positive")
	}
	if binding.BindingVersion == 0 {
		return errors.New("cross-zone identity binding version must be positive")
	}
	if binding.BindingHash != "" {
		return validateHexHash("cross-zone identity binding hash", binding.BindingHash)
	}
	return nil
}

func (binding CrossZoneIdentityBindingV2) Validate() error {
	if err := binding.ValidateFormat(); err != nil {
		return err
	}
	if binding.BindingHash == "" {
		return errors.New("cross-zone identity binding hash is required")
	}
	if binding.BindingHash != ComputeCrossZoneIdentityBindingV2Hash(binding) {
		return errors.New("cross-zone identity binding hash mismatch")
	}
	return nil
}

func (confirmation CrossZoneBindingConfirmationV2) ValidateFormat() error {
	if !IsCrossZoneBindingConfirmationTypeV2(confirmation.ConfirmationType) {
		return fmt.Errorf("unknown cross-zone binding confirmation type %q", confirmation.ConfirmationType)
	}
	if confirmation.ConfirmedHeight == 0 {
		return errors.New("cross-zone binding confirmation height must be positive")
	}
	switch confirmation.ConfirmationType {
	case IdentityBindingConfirmationProof:
		if err := validateHexHash("cross-zone binding confirmation proof root", confirmation.ProofRoot); err != nil {
			return err
		}
		if err := validateHexHash("cross-zone binding confirmation proof hash", confirmation.ProofHash); err != nil {
			return err
		}
	case IdentityBindingConfirmationMessage:
		if err := validateHexHash("cross-zone binding confirmation message id", confirmation.MessageID); err != nil {
			return err
		}
		if err := validateHexHash("cross-zone binding confirmation receipt hash", confirmation.ReceiptHash); err != nil {
			return err
		}
	}
	if confirmation.ConfirmationHash != "" {
		return validateHexHash("cross-zone binding confirmation hash", confirmation.ConfirmationHash)
	}
	return nil
}

func (confirmation CrossZoneBindingConfirmationV2) Validate() error {
	if err := confirmation.ValidateFormat(); err != nil {
		return err
	}
	if confirmation.ConfirmationHash == "" {
		return errors.New("cross-zone binding confirmation hash is required")
	}
	if confirmation.ConfirmationHash != ComputeCrossZoneBindingConfirmationV2Hash(confirmation) {
		return errors.New("cross-zone binding confirmation hash mismatch")
	}
	return nil
}

func (event CrossZoneBindingInvalidationEventV2) ValidateFormat() error {
	if _, err := CrossZoneBindingInvalidationV2Key(event.Height, event.EventID); err != nil {
		return err
	}
	if err := validateHexHash("cross-zone binding invalidation binding hash", event.BindingHash); err != nil {
		return err
	}
	if _, err := CrossZoneIdentityBindingV2Key(event.IdentityID, event.TargetZone, event.TargetType, event.TargetKey); err != nil {
		return err
	}
	if !IsCrossZoneBindingInvalidationReasonV2(event.Reason) {
		return fmt.Errorf("unknown cross-zone binding invalidation reason %q", event.Reason)
	}
	if event.EventHash != "" {
		return validateHexHash("cross-zone binding invalidation event hash", event.EventHash)
	}
	return nil
}

func (event CrossZoneBindingInvalidationEventV2) Validate() error {
	if err := event.ValidateFormat(); err != nil {
		return err
	}
	if event.EventHash == "" {
		return errors.New("cross-zone binding invalidation event hash is required")
	}
	if event.EventHash != ComputeCrossZoneBindingInvalidationEventV2Hash(event) {
		return errors.New("cross-zone binding invalidation event hash mismatch")
	}
	return nil
}

func (state CrossZoneIdentityBindingStateV2) ValidateFormat() error {
	if state.Height == 0 {
		return errors.New("cross-zone identity binding state height must be positive")
	}
	if err := validateCrossZoneIdentityBindingsV2(state.Bindings); err != nil {
		return err
	}
	if err := validateCrossZoneBindingConfirmationsV2(state.Confirmations); err != nil {
		return err
	}
	if err := validateCrossZoneBindingInvalidationsV2(state.Invalidations); err != nil {
		return err
	}
	if state.RootHash != "" {
		return validateHexHash("cross-zone identity binding state root", state.RootHash)
	}
	return nil
}

func (state CrossZoneIdentityBindingStateV2) Validate() error {
	if err := state.ValidateFormat(); err != nil {
		return err
	}
	if state.RootHash == "" {
		return errors.New("cross-zone identity binding state root is required")
	}
	if state.RootHash != ComputeCrossZoneIdentityBindingStateV2Root(state) {
		return errors.New("cross-zone identity binding state root mismatch")
	}
	return nil
}

func (msg MsgBindCrossZoneIdentityV2) Validate(graph IdentityGraphStateV2) error {
	if err := ValidateCrossZoneIdentityBindingAuthorizationV2(graph, msg.Binding.IdentityID, msg.Authority); err != nil {
		return err
	}
	if msg.Height == 0 {
		return errors.New("cross-zone identity binding message height must be positive")
	}
	if err := ValidateCrossZoneBindingConfirmationForBindingV2(msg.Binding, msg.Confirmation); err != nil {
		return err
	}
	if msg.Binding.ExpiresHeight <= msg.Height {
		return errors.New("cross-zone identity binding cannot expire before message height")
	}
	if msg.MessageHash != "" && msg.MessageHash != ComputeMsgBindCrossZoneIdentityV2Hash(msg) {
		return errors.New("cross-zone identity binding message hash mismatch")
	}
	return nil
}

func (msg MsgUpdateCrossZoneIdentityBindingV2) Validate(graph IdentityGraphStateV2) error {
	if err := ValidateCrossZoneIdentityBindingAuthorizationV2(graph, msg.Binding.IdentityID, msg.Authority); err != nil {
		return err
	}
	if msg.Height == 0 {
		return errors.New("cross-zone identity binding update height must be positive")
	}
	if err := ValidateCrossZoneBindingConfirmationForBindingV2(msg.Binding, msg.Confirmation); err != nil {
		return err
	}
	if msg.Binding.ExpiresHeight <= msg.Height {
		return errors.New("cross-zone identity binding cannot expire before update height")
	}
	if msg.MessageHash != "" && msg.MessageHash != ComputeMsgUpdateCrossZoneIdentityBindingV2Hash(msg) {
		return errors.New("cross-zone identity binding update message hash mismatch")
	}
	return nil
}

func (msg MsgRevokeCrossZoneIdentityBindingV2) Validate(graph IdentityGraphStateV2) error {
	if err := ValidateCrossZoneIdentityBindingAuthorizationV2(graph, msg.IdentityID, msg.Authority); err != nil {
		return err
	}
	if msg.Height == 0 {
		return errors.New("cross-zone identity binding revoke height must be positive")
	}
	if _, err := CrossZoneIdentityBindingV2Key(msg.IdentityID, msg.TargetZone, msg.TargetType, msg.TargetKey); err != nil {
		return err
	}
	if msg.MessageHash != "" && msg.MessageHash != ComputeMsgRevokeCrossZoneIdentityBindingV2Hash(msg) {
		return errors.New("cross-zone identity binding revoke message hash mismatch")
	}
	return nil
}

func ComputeCrossZoneIdentityBindingV2Hash(binding CrossZoneIdentityBindingV2) string {
	return identityHash(
		"cross-zone-identity-binding-v1",
		binding.IdentityID,
		binding.TargetZone,
		string(binding.TargetType),
		binding.TargetKey,
		fmt.Sprintf("%t", binding.ProofRequired),
		fmt.Sprintf("%020d", binding.ExpiresHeight),
		fmt.Sprintf("%020d", binding.BindingVersion),
	)
}

func ComputeCrossZoneBindingConfirmationV2Hash(confirmation CrossZoneBindingConfirmationV2) string {
	return identityHash(
		"cross-zone-binding-confirmation-v1",
		string(confirmation.ConfirmationType),
		confirmation.ProofRoot,
		confirmation.ProofHash,
		confirmation.MessageID,
		confirmation.ReceiptHash,
		fmt.Sprintf("%020d", confirmation.ConfirmedHeight),
	)
}

func ComputeCrossZoneBindingInvalidationEventIDV2(event CrossZoneBindingInvalidationEventV2) string {
	return identityHash("cross-zone-binding-invalidation-id-v1", event.BindingHash, event.IdentityID, event.TargetZone, string(event.TargetType), event.TargetKey, string(event.Reason), fmt.Sprintf("%020d", event.Height))
}

func ComputeCrossZoneBindingInvalidationEventV2Hash(event CrossZoneBindingInvalidationEventV2) string {
	return identityHash("cross-zone-binding-invalidation-v1", event.EventID, event.BindingHash, event.IdentityID, event.TargetZone, string(event.TargetType), event.TargetKey, string(event.Reason), fmt.Sprintf("%020d", event.Height))
}

func ComputeCrossZoneIdentityBindingV2Root(bindings []CrossZoneIdentityBindingV2) string {
	ordered := normalizeCrossZoneIdentityBindingsV2(bindings)
	parts := []string{"cross-zone-identity-binding-root-v1", fmt.Sprintf("%020d", len(ordered))}
	for _, binding := range ordered {
		parts = append(parts, binding.BindingHash)
	}
	return identityHash(parts...)
}

func ComputeCrossZoneBindingConfirmationV2Root(confirmations []CrossZoneBindingConfirmationV2) string {
	ordered := normalizeCrossZoneBindingConfirmationsV2(confirmations)
	parts := []string{"cross-zone-binding-confirmation-root-v1", fmt.Sprintf("%020d", len(ordered))}
	for _, confirmation := range ordered {
		parts = append(parts, confirmation.ConfirmationHash)
	}
	return identityHash(parts...)
}

func ComputeCrossZoneBindingInvalidationV2Root(events []CrossZoneBindingInvalidationEventV2) string {
	ordered := normalizeCrossZoneBindingInvalidationsV2(events)
	parts := []string{"cross-zone-binding-invalidation-root-v1", fmt.Sprintf("%020d", len(ordered))}
	for _, event := range ordered {
		parts = append(parts, event.EventHash)
	}
	return identityHash(parts...)
}

func ComputeCrossZoneIdentityBindingStateV2Root(state CrossZoneIdentityBindingStateV2) string {
	return identityHash(
		"cross-zone-identity-binding-state-root-v1",
		fmt.Sprintf("%020d", state.Height),
		ComputeCrossZoneIdentityBindingV2Root(state.Bindings),
		ComputeCrossZoneBindingConfirmationV2Root(state.Confirmations),
		ComputeCrossZoneBindingInvalidationV2Root(state.Invalidations),
	)
}

func ComputeMsgBindCrossZoneIdentityV2Hash(msg MsgBindCrossZoneIdentityV2) string {
	return identityHash("msg-bind-cross-zone-identity-v1", fmt.Sprintf("%x", []byte(msg.Authority)), msg.Binding.BindingHash, msg.Confirmation.ConfirmationHash, fmt.Sprintf("%020d", msg.Height))
}

func ComputeMsgUpdateCrossZoneIdentityBindingV2Hash(msg MsgUpdateCrossZoneIdentityBindingV2) string {
	return identityHash("msg-update-cross-zone-identity-binding-v1", fmt.Sprintf("%x", []byte(msg.Authority)), msg.Binding.BindingHash, msg.Confirmation.ConfirmationHash, fmt.Sprintf("%020d", msg.Height))
}

func ComputeMsgRevokeCrossZoneIdentityBindingV2Hash(msg MsgRevokeCrossZoneIdentityBindingV2) string {
	return identityHash("msg-revoke-cross-zone-identity-binding-v1", fmt.Sprintf("%x", []byte(msg.Authority)), msg.IdentityID, msg.TargetZone, string(msg.TargetType), msg.TargetKey, fmt.Sprintf("%020d", msg.Height))
}

func IsCrossZoneBindingTargetTypeV2(targetType IdentityCrossZoneBindingTargetType) bool {
	switch targetType {
	case IdentityBindingTargetAccount, IdentityBindingTargetService, IdentityBindingTargetContract, IdentityBindingTargetZoneEndpoint, IdentityBindingTargetComposite:
		return true
	default:
		return false
	}
}

func IsCrossZoneBindingConfirmationTypeV2(confirmationType IdentityCrossZoneBindingConfirmationType) bool {
	switch confirmationType {
	case IdentityBindingConfirmationProof, IdentityBindingConfirmationMessage:
		return true
	default:
		return false
	}
}

func IsCrossZoneBindingInvalidationReasonV2(reason IdentityCrossZoneBindingInvalidationReason) bool {
	switch reason {
	case IdentityBindingInvalidatedCreated, IdentityBindingInvalidatedUpdated, IdentityBindingInvalidatedRevoked, IdentityBindingInvalidatedExpired:
		return true
	default:
		return false
	}
}

func normalizeCrossZoneIdentityBindingsV2(bindings []CrossZoneIdentityBindingV2) []CrossZoneIdentityBindingV2 {
	out := append([]CrossZoneIdentityBindingV2(nil), bindings...)
	sort.SliceStable(out, func(i, j int) bool {
		left, _ := CrossZoneIdentityBindingV2Key(out[i].IdentityID, out[i].TargetZone, out[i].TargetType, out[i].TargetKey)
		right, _ := CrossZoneIdentityBindingV2Key(out[j].IdentityID, out[j].TargetZone, out[j].TargetType, out[j].TargetKey)
		return left < right
	})
	return out
}

func normalizeCrossZoneBindingConfirmationsV2(confirmations []CrossZoneBindingConfirmationV2) []CrossZoneBindingConfirmationV2 {
	out := append([]CrossZoneBindingConfirmationV2(nil), confirmations...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].ConfirmationHash < out[j].ConfirmationHash })
	return out
}

func normalizeCrossZoneBindingInvalidationsV2(events []CrossZoneBindingInvalidationEventV2) []CrossZoneBindingInvalidationEventV2 {
	out := append([]CrossZoneBindingInvalidationEventV2(nil), events...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Height != out[j].Height {
			return out[i].Height < out[j].Height
		}
		return out[i].EventID < out[j].EventID
	})
	return out
}

func validateCrossZoneIdentityBindingsV2(bindings []CrossZoneIdentityBindingV2) error {
	seen := map[string]struct{}{}
	previous := ""
	for _, binding := range bindings {
		if err := binding.Validate(); err != nil {
			return err
		}
		key, _ := CrossZoneIdentityBindingV2Key(binding.IdentityID, binding.TargetZone, binding.TargetType, binding.TargetKey)
		if _, found := seen[key]; found {
			return fmt.Errorf("duplicate cross-zone identity binding %s", key)
		}
		seen[key] = struct{}{}
		if previous != "" && previous >= key {
			return errors.New("cross-zone identity bindings must be sorted canonically")
		}
		previous = key
	}
	return nil
}

func validateCrossZoneBindingConfirmationsV2(confirmations []CrossZoneBindingConfirmationV2) error {
	seen := map[string]struct{}{}
	previous := ""
	for _, confirmation := range confirmations {
		if err := confirmation.Validate(); err != nil {
			return err
		}
		if _, found := seen[confirmation.ConfirmationHash]; found {
			return fmt.Errorf("duplicate cross-zone binding confirmation %s", confirmation.ConfirmationHash)
		}
		seen[confirmation.ConfirmationHash] = struct{}{}
		if previous != "" && previous >= confirmation.ConfirmationHash {
			return errors.New("cross-zone binding confirmations must be sorted canonically")
		}
		previous = confirmation.ConfirmationHash
	}
	return nil
}

func validateCrossZoneBindingInvalidationsV2(events []CrossZoneBindingInvalidationEventV2) error {
	seen := map[string]struct{}{}
	previous := ""
	for _, event := range events {
		if err := event.Validate(); err != nil {
			return err
		}
		key, _ := CrossZoneBindingInvalidationV2Key(event.Height, event.EventID)
		if _, found := seen[key]; found {
			return fmt.Errorf("duplicate cross-zone binding invalidation %s", key)
		}
		seen[key] = struct{}{}
		if previous != "" && previous >= key {
			return errors.New("cross-zone binding invalidations must be sorted canonically")
		}
		previous = key
	}
	return nil
}

func findCrossZoneIdentityBindingV2(bindings []CrossZoneIdentityBindingV2, probe CrossZoneIdentityBindingV2) (int, bool) {
	for i, binding := range bindings {
		if binding.IdentityID == probe.IdentityID &&
			binding.TargetZone == probe.TargetZone &&
			binding.TargetType == probe.TargetType &&
			binding.TargetKey == probe.TargetKey {
			return i, true
		}
	}
	return -1, false
}

func validateCrossZoneTargetZoneV2(zone string) error {
	if strings.TrimSpace(zone) != zone || zone == "" {
		return errors.New("cross-zone binding target zone is required and must not have surrounding whitespace")
	}
	if len(zone) > 64 {
		return errors.New("cross-zone binding target zone must not exceed 64 bytes")
	}
	for _, r := range zone {
		if r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_' || r == '-' {
			continue
		}
		return errors.New("cross-zone binding target zone contains invalid character")
	}
	return nil
}
