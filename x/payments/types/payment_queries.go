package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type PaymentProofObjectType string

const (
	PaymentProofObjectIntent	PaymentProofObjectType	= "intent"
	PaymentProofObjectChannel	PaymentProofObjectType	= "channel"
	PaymentProofObjectCondition	PaymentProofObjectType	= "condition"
	PaymentProofObjectRoute		PaymentProofObjectType	= "route"
	PaymentProofObjectSettlement	PaymentProofObjectType	= "settlement"
	PaymentProofObjectDispute	PaymentProofObjectType	= "dispute"
	PaymentProofObjectReceipt	PaymentProofObjectType	= "receipt"
)

type PaymentProofEnvelope struct {
	ObjectType	PaymentProofObjectType
	ObjectID	string
	StateKey	string
	ObjectHash	string
	PaymentRoot	string
	RootAtHeight	string
	Height		uint64
	ProofMetadata	[]string
	ProofHash	string
}

type QueryPaymentIntent struct {
	PaymentID	string
	Height		uint64
}

type QueryPaymentChannel struct {
	ChannelID	string
	Height		uint64
}

type QueryConditionalPayment struct {
	ConditionID	string
	Height		uint64
}

type QueryPaymentRoute struct {
	RouteID	string
	Height	uint64
}

type QueryPaymentSettlement struct {
	PaymentID	string
	Height		uint64
}

type QueryPaymentDispute struct {
	DisputeID	string
	Height		uint64
}

type QueryPaymentProof struct {
	ObjectType	PaymentProofObjectType
	ObjectID	string
	Height		uint64
}

type QueryPaymentIntentResponse struct {
	Intent		PaymentIntent
	Found		bool
	ExpiryHeight	uint64
	RoutePolicy	string
	ProofMetadata	PaymentProofEnvelope
}

type QueryPaymentChannelResponse struct {
	Channel		PaymentChannel
	Found		bool
	ProofMetadata	PaymentProofEnvelope
}

type QueryConditionalPaymentResponse struct {
	Condition	NativeConditionalPayment
	Found		bool
	ProofMetadata	PaymentProofEnvelope
}

type QueryPaymentRouteResponse struct {
	Route		PaymentRouteCommitment
	Found		bool
	Metadata	PaymentRouteQueryMetadata
	ProofMetadata	PaymentProofEnvelope
}

type QueryPaymentSettlementResponse struct {
	Settlement	PaymentSettlement
	Found		bool
	ProofMetadata	PaymentProofEnvelope
}

type QueryPaymentDisputeResponse struct {
	Dispute		PaymentDispute
	Found		bool
	ProofMetadata	PaymentProofEnvelope
}

type QueryPaymentProofResponse struct {
	Proof	PaymentProofEnvelope
	Found	bool
}

type PaymentRouteQueryMetadata struct {
	RouteID			string
	Committer		string
	CommitmentHash		string
	Signed			bool
	Reserved		bool
	ExpiresHeight		uint64
	RouteRoot		string
	ReservationStatus	string
}

func QueryPaymentIntentFromState(state FinancialZonePaymentState, query QueryPaymentIntent) (QueryPaymentIntentResponse, error) {
	state, err := validatePaymentQueryState(state, query.Height)
	if err != nil {
		return QueryPaymentIntentResponse{}, err
	}
	paymentID := normalizeHash(query.PaymentID)
	if err := ValidateHash("payments query intent id", paymentID); err != nil {
		return QueryPaymentIntentResponse{}, err
	}
	for _, intent := range state.Intents {
		if intent.PaymentID != paymentID {
			continue
		}
		proof, err := BuildPaymentProofEnvelope(state, PaymentProofObjectIntent, paymentID)
		if err != nil {
			return QueryPaymentIntentResponse{}, err
		}
		return QueryPaymentIntentResponse{
			Intent:		intent,
			Found:		true,
			ExpiryHeight:	intent.ExpiryHeight,
			RoutePolicy:	"committed-or-reserved-route",
			ProofMetadata:	proof,
		}, nil
	}
	return QueryPaymentIntentResponse{}, nil
}

func QueryPaymentChannelFromState(state FinancialZonePaymentState, query QueryPaymentChannel) (QueryPaymentChannelResponse, error) {
	state, err := validatePaymentQueryState(state, query.Height)
	if err != nil {
		return QueryPaymentChannelResponse{}, err
	}
	channelID := normalizeHash(query.ChannelID)
	if err := ValidateHash("payments query channel id", channelID); err != nil {
		return QueryPaymentChannelResponse{}, err
	}
	for _, channel := range state.Channels {
		if channel.ChannelID != channelID {
			continue
		}
		proof, err := BuildPaymentProofEnvelope(state, PaymentProofObjectChannel, channelID)
		if err != nil {
			return QueryPaymentChannelResponse{}, err
		}
		return QueryPaymentChannelResponse{Channel: channel, Found: true, ProofMetadata: proof}, nil
	}
	return QueryPaymentChannelResponse{}, nil
}

func QueryConditionalPaymentFromState(state FinancialZonePaymentState, query QueryConditionalPayment) (QueryConditionalPaymentResponse, error) {
	state, err := validatePaymentQueryState(state, query.Height)
	if err != nil {
		return QueryConditionalPaymentResponse{}, err
	}
	conditionID := normalizeHash(query.ConditionID)
	if err := ValidateHash("payments query condition id", conditionID); err != nil {
		return QueryConditionalPaymentResponse{}, err
	}
	for _, condition := range state.Conditions {
		if condition.ConditionID != conditionID {
			continue
		}
		proof, err := BuildPaymentProofEnvelope(state, PaymentProofObjectCondition, conditionID)
		if err != nil {
			return QueryConditionalPaymentResponse{}, err
		}
		return QueryConditionalPaymentResponse{Condition: condition, Found: true, ProofMetadata: proof}, nil
	}
	return QueryConditionalPaymentResponse{}, nil
}

func QueryPaymentRouteFromState(state FinancialZonePaymentState, query QueryPaymentRoute) (QueryPaymentRouteResponse, error) {
	state, err := validatePaymentQueryState(state, query.Height)
	if err != nil {
		return QueryPaymentRouteResponse{}, err
	}
	routeID := normalizeHash(query.RouteID)
	if err := ValidateHash("payments query route id", routeID); err != nil {
		return QueryPaymentRouteResponse{}, err
	}
	for _, route := range state.Routes {
		if route.RouteID != routeID {
			continue
		}
		proof, err := BuildPaymentProofEnvelope(state, PaymentProofObjectRoute, routeID)
		if err != nil {
			return QueryPaymentRouteResponse{}, err
		}
		return QueryPaymentRouteResponse{Route: route, Found: true, Metadata: PaymentRouteQueryMetadataFromCommitment(route, state.RouteRoot), ProofMetadata: proof}, nil
	}
	return QueryPaymentRouteResponse{}, nil
}

func PaymentRouteQueryMetadataFromCommitment(route PaymentRouteCommitment, routeRoot string) PaymentRouteQueryMetadata {
	route = route.Normalize()
	status := "unreserved"
	if route.Reserved {
		status = "reserved"
	}
	return PaymentRouteQueryMetadata{
		RouteID:		route.RouteID,
		Committer:		route.Committer,
		CommitmentHash:		route.CommitmentHash,
		Signed:			route.Signed,
		Reserved:		route.Reserved,
		ExpiresHeight:		route.ExpiresHeight,
		RouteRoot:		normalizeHash(routeRoot),
		ReservationStatus:	status,
	}
}

func QueryPaymentSettlementFromState(state FinancialZonePaymentState, query QueryPaymentSettlement) (QueryPaymentSettlementResponse, error) {
	state, err := validatePaymentQueryState(state, query.Height)
	if err != nil {
		return QueryPaymentSettlementResponse{}, err
	}
	paymentID := normalizeHash(query.PaymentID)
	if err := ValidateHash("payments query settlement id", paymentID); err != nil {
		return QueryPaymentSettlementResponse{}, err
	}
	for _, settlement := range state.Settlements {
		if settlement.PaymentID != paymentID {
			continue
		}
		proof, err := BuildPaymentProofEnvelope(state, PaymentProofObjectSettlement, paymentID)
		if err != nil {
			return QueryPaymentSettlementResponse{}, err
		}
		return QueryPaymentSettlementResponse{Settlement: settlement, Found: true, ProofMetadata: proof}, nil
	}
	return QueryPaymentSettlementResponse{}, nil
}

func QueryPaymentDisputeFromState(state FinancialZonePaymentState, query QueryPaymentDispute) (QueryPaymentDisputeResponse, error) {
	state, err := validatePaymentQueryState(state, query.Height)
	if err != nil {
		return QueryPaymentDisputeResponse{}, err
	}
	disputeID := normalizeHash(query.DisputeID)
	if err := ValidateHash("payments query dispute id", disputeID); err != nil {
		return QueryPaymentDisputeResponse{}, err
	}
	for _, dispute := range state.Disputes {
		if dispute.DisputeID != disputeID {
			continue
		}
		proof, err := BuildPaymentProofEnvelope(state, PaymentProofObjectDispute, disputeID)
		if err != nil {
			return QueryPaymentDisputeResponse{}, err
		}
		return QueryPaymentDisputeResponse{Dispute: dispute, Found: true, ProofMetadata: proof}, nil
	}
	return QueryPaymentDisputeResponse{}, nil
}

func QueryPaymentProofFromState(state FinancialZonePaymentState, query QueryPaymentProof) (QueryPaymentProofResponse, error) {
	state, err := validatePaymentQueryState(state, query.Height)
	if err != nil {
		return QueryPaymentProofResponse{}, err
	}
	objectID := normalizeHash(query.ObjectID)
	if err := ValidateHash("payments query proof object id", objectID); err != nil {
		return QueryPaymentProofResponse{}, err
	}
	if _, _, _, found := paymentProofObjectForState(state, query.ObjectType, objectID); !found {
		return QueryPaymentProofResponse{}, nil
	}
	proof, err := BuildPaymentProofEnvelope(state, query.ObjectType, objectID)
	if err != nil {
		return QueryPaymentProofResponse{}, err
	}
	return QueryPaymentProofResponse{Proof: proof, Found: true}, nil
}

func BuildPaymentProofEnvelope(state FinancialZonePaymentState, objectType PaymentProofObjectType, objectID string) (PaymentProofEnvelope, error) {
	state = state.Normalize()
	objectID = normalizeHash(objectID)
	if err := state.Validate(); err != nil {
		return PaymentProofEnvelope{}, err
	}
	stateKey, objectHash, metadata, found := paymentProofObjectForState(state, objectType, objectID)
	if !found {
		return PaymentProofEnvelope{}, errors.New("payments proof object not found")
	}
	proof := PaymentProofEnvelope{
		ObjectType:	objectType,
		ObjectID:	objectID,
		StateKey:	stateKey,
		ObjectHash:	objectHash,
		PaymentRoot:	state.PaymentRoot,
		RootAtHeight:	state.PaymentRoot,
		Height:		state.Height,
		ProofMetadata:	metadata,
	}
	proof.ProofHash = ComputePaymentProofEnvelopeHash(proof)
	return proof, proof.Validate()
}

func (proof PaymentProofEnvelope) Normalize() PaymentProofEnvelope {
	proof.ObjectID = normalizeHash(proof.ObjectID)
	proof.StateKey = strings.TrimSpace(proof.StateKey)
	proof.ObjectHash = normalizeHash(proof.ObjectHash)
	proof.PaymentRoot = normalizeHash(proof.PaymentRoot)
	proof.RootAtHeight = normalizeHash(proof.RootAtHeight)
	proof.ProofMetadata = normalizePaymentQueryMetadata(proof.ProofMetadata)
	proof.ProofHash = normalizeOptionalHash(proof.ProofHash)
	return proof
}

func (proof PaymentProofEnvelope) Validate() error {
	proof = proof.Normalize()
	if !IsPaymentProofObjectType(proof.ObjectType) {
		return fmt.Errorf("unknown payments proof object type %q", proof.ObjectType)
	}
	if err := ValidateHash("payments proof object id", proof.ObjectID); err != nil {
		return err
	}
	if !strings.HasPrefix(proof.StateKey, FinancialPaymentsPrefix+"/") {
		return errors.New("payments proof state key must be under financial payments prefix")
	}
	if err := ValidateHash("payments proof object hash", proof.ObjectHash); err != nil {
		return err
	}
	if err := ValidateHash("payments proof payment root", proof.PaymentRoot); err != nil {
		return err
	}
	if proof.RootAtHeight != proof.PaymentRoot {
		return errors.New("payments proof root-at-height mismatch")
	}
	if proof.Height == 0 {
		return errors.New("payments proof height must be positive")
	}
	if err := ValidateHash("payments proof hash", proof.ProofHash); err != nil {
		return err
	}
	if expected := ComputePaymentProofEnvelopeHash(proof); proof.ProofHash != expected {
		return fmt.Errorf("payments proof hash mismatch: expected %s", expected)
	}
	return nil
}

func ComputePaymentProofEnvelopeHash(proof PaymentProofEnvelope) string {
	proof = proof.Normalize()
	parts := []string{
		"aetra-financial-payment-proof-envelope-v1",
		string(proof.ObjectType),
		proof.ObjectID,
		proof.StateKey,
		proof.ObjectHash,
		proof.PaymentRoot,
		proof.RootAtHeight,
		fmt.Sprintf("%020d", proof.Height),
	}
	parts = append(parts, proof.ProofMetadata...)
	return HashParts(parts...)
}

func IsPaymentProofObjectType(objectType PaymentProofObjectType) bool {
	switch objectType {
	case PaymentProofObjectIntent,
		PaymentProofObjectChannel,
		PaymentProofObjectCondition,
		PaymentProofObjectRoute,
		PaymentProofObjectSettlement,
		PaymentProofObjectDispute,
		PaymentProofObjectReceipt:
		return true
	default:
		return false
	}
}

func validatePaymentQueryState(state FinancialZonePaymentState, height uint64) (FinancialZonePaymentState, error) {
	state = state.Normalize()
	if err := state.Validate(); err != nil {
		return FinancialZonePaymentState{}, err
	}
	if height == 0 {
		return FinancialZonePaymentState{}, errors.New("payments query height must be positive")
	}
	if height != state.Height {
		return FinancialZonePaymentState{}, errors.New("payments query height is outside available root history")
	}
	return state, nil
}

func paymentProofObjectForState(state FinancialZonePaymentState, objectType PaymentProofObjectType, objectID string) (string, string, []string, bool) {
	state = state.Normalize()
	objectID = normalizeHash(objectID)
	switch objectType {
	case PaymentProofObjectIntent:
		for _, intent := range state.Intents {
			if intent.PaymentID == objectID {
				key, _ := FinancialPaymentIntentStateKey(objectID)
				return key, intent.IntentHash, []string{"expiry:" + fmt.Sprintf("%020d", intent.ExpiryHeight), "route:" + intent.RouteIDOptional}, true
			}
		}
	case PaymentProofObjectChannel:
		for _, channel := range state.Channels {
			if channel.ChannelID == objectID {
				key, _ := FinancialPaymentChannelStateKey(objectID)
				return key, channel.ChannelRoot, []string{"nonce:" + fmt.Sprintf("%020d", channel.LatestNonce), "status:" + string(channel.SettlementStatus)}, true
			}
		}
	case PaymentProofObjectCondition:
		for _, condition := range state.Conditions {
			if condition.ConditionID == objectID {
				key, _ := FinancialPaymentConditionStateKey(objectID)
				return key, condition.ConditionRoot, []string{"timeout:" + fmt.Sprintf("%020d", condition.TimeoutHeight), "status:" + string(condition.Status), "route:" + condition.RouteID}, true
			}
		}
	case PaymentProofObjectRoute:
		for _, route := range state.Routes {
			if route.RouteID == objectID {
				key, _ := FinancialPaymentRouteStateKey(objectID)
				return key, route.CommitmentHash, []string{"signed:" + fmt.Sprintf("%t", route.Signed), "reserved:" + fmt.Sprintf("%t", route.Reserved), "expires:" + fmt.Sprintf("%020d", route.ExpiresHeight)}, true
			}
		}
	case PaymentProofObjectSettlement:
		for _, settlement := range state.Settlements {
			if settlement.PaymentID == objectID {
				key, _ := FinancialPaymentSettlementStateKey(objectID)
				return key, settlement.SettlementHash, []string{"close:" + string(settlement.CloseStatus), "receipt:" + settlement.ReceiptHash}, true
			}
		}
	case PaymentProofObjectDispute:
		for _, dispute := range state.Disputes {
			if dispute.DisputeID == objectID {
				key, _ := FinancialPaymentDisputeStateKey(objectID)
				return key, dispute.DisputeRoot, []string{"status:" + string(dispute.Status), "challenge_end:" + fmt.Sprintf("%020d", dispute.ChallengeEnd)}, true
			}
		}
	case PaymentProofObjectReceipt:
		for _, receipt := range state.Receipts {
			if receipt.PaymentID == objectID {
				key := FinancialPaymentReceiptsPrefix + "/" + objectID
				return key, receipt.ReceiptHash, []string{"status:" + string(receipt.Status), "height:" + fmt.Sprintf("%020d", receipt.Height)}, true
			}
		}
	}
	return "", "", nil, false
}

func normalizePaymentQueryMetadata(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			out = append(out, value)
		}
	}
	sort.Strings(out)
	return out
}
