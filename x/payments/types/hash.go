package types

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
)

const HashHexLength = 64

const (
	domainHashParts			= "aetra-payments-hash-parts-v1"
	domainOpeningCommitment		= "aetra-payment-opening-commitment"
	domainBalanceCommitment		= "aetra-payment-balance-state-commitment"
	domainChannelState		= "aetra-payment-channel-state"
	domainAsyncDelta		= "aetra-payment-async-delta"
	domainAsyncDeltaRoot		= "aetra-payment-async-delta-root"
	domainConditionalPromise	= "aetra-payment-conditional-promise"
	domainConditionRootCommitment	= "aetra-payment-condition-root-commitment"
	domainConditionsRoot		= "aetra-payment-conditions-root"
	domainCooperativeClose		= "aetra-payment-cooperative-close"
	domainDisputeProof		= "aetra-payment-dispute-proof"
	domainSettlementResult		= "aetra-payment-settlement-result"
	domainVirtualChannelState	= "aetra-virtual-payment-channel-state"
	domainVirtualChannelAnchor	= "aetra-virtual-payment-channel-anchor"
	domainStateSignaturePreimage	= "aetra-payment-state-signature-preimage-hash"
	domainSignatureEnvelope		= "aetra-payment-signature-envelope"
	domainSignedNonceWAL		= "aetra-payment-signed-nonce-wal"
	domainValidatorPaymentService	= "aetra-payment-validator-service-metadata"
)

func HashParts(parts ...string) string {
	h := sha256.New()
	writeString(h, domainHashParts)
	for _, part := range parts {
		writeString(h, part)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeValidatorPaymentServiceMetadataHash(metadata ValidatorPaymentServiceMetadata) string {
	metadata = metadata.Normalize()
	h := sha256.New()
	writeString(h, domainValidatorPaymentService)
	writeString(h, metadata.ValidatorAddress)
	writeString(h, metadata.ServiceAddress)
	writeString(h, metadata.WatchEndpoint)
	writeString(h, metadata.RoutingEndpoint)
	writeString(h, metadata.PublicKey)
	writeString(h, metadata.MinDelegation)
	writeUint64(h, uint64(metadata.CommissionBps))
	writeString(h, fmt.Sprintf("%t", metadata.Active))
	writeUint64(h, metadata.UpdatedHeight)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeStateHash(state ChannelState) string {
	hash, err := ComputeStateHashForEncodingVersion(state, CanonicalEncodingVersion)
	if err != nil {
		panic(err)
	}
	return hash
}

func ComputeStateHashForEncodingVersion(state ChannelState, version byte) (string, error) {
	if version != CanonicalEncodingVersion {
		return "", fmt.Errorf("payments unsupported channel state encoding version %d", version)
	}
	state = state.Normalize()
	h := sha256.New()
	writeString(h, domainChannelState)
	writeByte(h, version)
	writeString(h, state.ChainID)
	writeUint64(h, uint64(state.AppVersion))
	writeString(h, state.ModuleName)
	for _, field := range state.RequiredFields {
		writeString(h, field)
	}
	writeString(h, state.ChannelID)
	writeString(h, string(state.ChannelType))
	writeString(h, state.ParticipantSetHash)
	writeString(h, state.Denom)
	writeUint64(h, uint64(state.Version))
	writeString(h, state.ParticipantA)
	writeString(h, state.ParticipantB)
	writeString(h, state.BalanceA)
	writeString(h, state.BalanceB)
	writeString(h, state.ReserveA)
	writeString(h, state.ReserveB)
	writeString(h, state.AccruedFees)
	writeUint64(h, state.Epoch)
	writeUint64(h, state.Nonce)
	writeString(h, state.PendingConditionsRoot)
	writeString(h, state.ConditionRoot)
	writeUint64(h, uint64(state.ConditionCount))
	writeString(h, state.PreviousStateHash)
	writeUint64(h, state.TimeoutHeight)
	writeInt64(h, state.TimeoutTimestamp)
	writeUint64(h, state.ChallengePeriod)
	writeUint64(h, state.CloseDelay)
	writeString(h, state.FeePolicyID)
	writeString(h, state.RequiredSignerBitmap)
	writeString(h, state.SignatureScheme)
	writeString(h, state.SignaturePreimageHash)
	writeUint64(h, state.CheckpointNonce)
	writeString(h, state.AsyncUpdateRoot)
	writeString(h, state.AcceptedUpdateRoot)
	writeUint64(h, state.SendWindow)
	writeUint64(h, state.ReceiveWindow)
	writeString(h, state.MaxUnackedAmount)
	writeUint64(h, state.ExpiryHeight)
	for _, balance := range state.Balances {
		writeString(h, balance.Participant)
		writeString(h, balance.Amount)
	}
	for _, balance := range state.CheckpointBalances {
		writeString(h, balance.Participant)
		writeString(h, balance.Amount)
	}
	for _, condition := range state.Conditions {
		writeString(h, condition.ConditionID)
		writeString(h, string(condition.ConditionType))
		writeString(h, condition.Payer)
		writeString(h, condition.Payee)
		writeString(h, condition.Amount)
		writeString(h, condition.HashLock)
		writeUint64(h, condition.TimeoutHeight)
		writeUint64(h, condition.NonceStart)
		writeUint64(h, condition.NonceEnd)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func ComputeOpeningCommitment(channel ChannelRecord) string {
	channel = channel.Normalize()
	h := sha256.New()
	writeString(h, domainOpeningCommitment)
	writeByte(h, CanonicalEncodingVersion)
	writeString(h, channel.ChainID)
	writeString(h, channel.ChannelID)
	writeString(h, string(channel.ChannelType))
	writeString(h, channel.Denom)
	writeString(h, channel.Collateral)
	writeUint64(h, channel.OpenHeight)
	writeUint64(h, channel.CloseDelay)
	writeUint64(h, channel.DisputePeriod)
	for _, participant := range channel.Participants {
		writeString(h, participant)
	}
	for _, signer := range channel.RequiredSigners {
		writeString(h, signer)
	}
	writeString(h, channel.OpeningStateHash)
	writeString(h, channel.LatestState.StateHash)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeBalanceStateCommitment(channel ChannelRecord, state ChannelState) string {
	channel = channel.Normalize()
	state = state.Normalize()
	h := sha256.New()
	writeString(h, domainBalanceCommitment)
	writeByte(h, CanonicalEncodingVersion)
	writeString(h, channel.ChainID)
	writeString(h, channel.ChannelID)
	writeString(h, string(channel.ChannelType))
	writeString(h, state.StateHash)
	writeUint64(h, state.Nonce)
	writeUint64(h, state.Epoch)
	writeString(h, state.BalanceA)
	writeString(h, state.BalanceB)
	writeString(h, state.ReserveA)
	writeString(h, state.ReserveB)
	for _, balance := range state.Balances {
		writeString(h, balance.Participant)
		writeString(h, balance.Amount)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeParticipantSetHash(participants []string) string {
	h := sha256.New()
	writeString(h, "aetra-payment-participant-set-v1")
	for _, participant := range normalizeAddressSet(participants) {
		writeString(h, participant)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeRequiredSignerBitmap(participants, required []string) string {
	participants = normalizeAddressSet(participants)
	required = normalizeAddressSet(required)
	requiredSet := make(map[string]struct{}, len(required))
	for _, signer := range required {
		requiredSet[signer] = struct{}{}
	}
	var buf bytes.Buffer
	for _, participant := range participants {
		if _, found := requiredSet[participant]; found {
			buf.WriteByte('1')
			continue
		}
		buf.WriteByte('0')
	}
	return buf.String()
}

func ComputeStateSignaturePreimageHash(state ChannelState) string {
	appVersion := state.AppVersion
	if appVersion == 0 {
		appVersion = CurrentAppVersion
	}
	moduleName := strings.TrimSpace(state.ModuleName)
	if moduleName == "" {
		moduleName = ModuleName
	}
	requiredFields := normalizeRequiredFields(state.RequiredFields)
	if len(requiredFields) == 0 {
		requiredFields = CanonicalStateRequiredFields()
	}
	version := state.Version
	if version == 0 {
		version = CurrentStateVersion
	}
	balances := normalizeBalances(state.Balances)
	conditions := normalizeConditions(state.Conditions)
	conditionRoot := normalizeOptionalHash(state.ConditionRoot)
	pendingRoot := normalizeOptionalHash(state.PendingConditionsRoot)
	if conditionRoot == "" && pendingRoot != "" {
		conditionRoot = pendingRoot
	}
	if conditionRoot == "" {
		conditionRoot = ComputeConditionsRoot(conditions)
	}
	if pendingRoot == "" {
		pendingRoot = conditionRoot
	}
	conditionCount := state.ConditionCount
	if conditionCount == 0 && len(conditions) > 0 {
		conditionCount = uint32(len(conditions))
	}
	participantSetHash := normalizeOptionalHash(state.ParticipantSetHash)
	if participantSetHash == "" {
		participantSetHash = ComputeParticipantSetHash(participantsFromBalances(balances))
	}
	accruedFees := strings.TrimSpace(state.AccruedFees)
	if accruedFees == "" {
		accruedFees = "0"
	}
	challengePeriod := state.ChallengePeriod
	if challengePeriod == 0 {
		challengePeriod = state.CloseDelay
	}
	feePolicyID := strings.TrimSpace(state.FeePolicyID)
	if feePolicyID == "" {
		feePolicyID = NativeDenom
	}
	requiredSignerBitmap := strings.TrimSpace(state.RequiredSignerBitmap)
	if requiredSignerBitmap == "" {
		participants := participantsFromBalances(balances)
		requiredSignerBitmap = ComputeRequiredSignerBitmap(participants, participants)
	}
	signatureScheme := strings.TrimSpace(state.SignatureScheme)
	if signatureScheme == "" {
		signatureScheme = SignatureSchemeEd25519
	}
	checkpointBalances := normalizeBalances(state.CheckpointBalances)

	h := sha256.New()
	writeString(h, domainStateSignaturePreimage)
	writeByte(h, CanonicalEncodingVersion)
	writeString(h, strings.TrimSpace(state.ChainID))
	writeUint64(h, uint64(appVersion))
	writeString(h, moduleName)
	for _, field := range requiredFields {
		writeString(h, field)
	}
	writeString(h, normalizeHash(state.ChannelID))
	writeString(h, string(state.ChannelType))
	writeString(h, participantSetHash)
	writeString(h, strings.TrimSpace(state.Denom))
	writeUint64(h, uint64(version))
	writeString(h, strings.TrimSpace(state.ParticipantA))
	writeString(h, strings.TrimSpace(state.ParticipantB))
	writeString(h, strings.TrimSpace(state.BalanceA))
	writeString(h, strings.TrimSpace(state.BalanceB))
	writeString(h, strings.TrimSpace(state.ReserveA))
	writeString(h, strings.TrimSpace(state.ReserveB))
	writeString(h, accruedFees)
	writeUint64(h, state.Epoch)
	writeUint64(h, state.Nonce)
	writeString(h, pendingRoot)
	writeString(h, conditionRoot)
	writeUint64(h, uint64(conditionCount))
	writeString(h, normalizeOptionalHash(state.PreviousStateHash))
	writeUint64(h, state.TimeoutHeight)
	writeInt64(h, state.TimeoutTimestamp)
	writeUint64(h, challengePeriod)
	writeUint64(h, state.CloseDelay)
	writeString(h, feePolicyID)
	writeString(h, requiredSignerBitmap)
	writeString(h, signatureScheme)
	writeUint64(h, state.CheckpointNonce)
	writeString(h, normalizeOptionalHash(state.AsyncUpdateRoot))
	writeString(h, normalizeOptionalHash(state.AcceptedUpdateRoot))
	writeUint64(h, state.SendWindow)
	writeUint64(h, state.ReceiveWindow)
	writeString(h, strings.TrimSpace(state.MaxUnackedAmount))
	writeUint64(h, state.ExpiryHeight)
	for _, balance := range balances {
		writeString(h, balance.Participant)
		writeString(h, balance.Amount)
	}
	for _, balance := range checkpointBalances {
		writeString(h, balance.Participant)
		writeString(h, balance.Amount)
	}
	for _, condition := range conditions {
		writeString(h, condition.ConditionID)
		writeString(h, string(condition.ConditionType))
		writeString(h, condition.Payer)
		writeString(h, condition.Payee)
		writeString(h, condition.Amount)
		writeString(h, condition.HashLock)
		writeUint64(h, condition.TimeoutHeight)
		writeUint64(h, condition.NonceStart)
		writeUint64(h, condition.NonceEnd)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeConditionsRoot(conditions []ConditionalPayment) string {
	h := sha256.New()
	writeString(h, domainConditionsRoot)
	writeByte(h, CanonicalEncodingVersion)
	for _, condition := range normalizeConditions(conditions) {
		writeString(h, ComputeConditionalPromiseHash(condition))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeConditionRootCommitment(channel ChannelRecord, state ChannelState) string {
	channel = channel.Normalize()
	state = state.Normalize()
	h := sha256.New()
	writeString(h, domainConditionRootCommitment)
	writeByte(h, CanonicalEncodingVersion)
	writeString(h, channel.ChainID)
	writeString(h, channel.ChannelID)
	writeString(h, string(channel.ChannelType))
	writeUint64(h, state.Nonce)
	writeString(h, state.ConditionRoot)
	writeUint64(h, uint64(state.ConditionCount))
	for _, condition := range state.Conditions {
		writeString(h, ComputeConditionalPromiseHash(condition))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeConditionalPromiseHash(condition ConditionalPayment) string {
	condition = condition.Normalize()
	h := sha256.New()
	writeString(h, domainConditionalPromise)
	writeByte(h, CanonicalEncodingVersion)
	writeString(h, condition.ConditionID)
	writeString(h, string(condition.ConditionType))
	writeString(h, condition.Payer)
	writeString(h, condition.Payee)
	writeString(h, condition.Amount)
	writeString(h, condition.HashLock)
	writeUint64(h, condition.TimeoutHeight)
	writeUint64(h, condition.NonceStart)
	writeUint64(h, condition.NonceEnd)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeConditionalTransferPromiseHash(promise ConditionalPromise) string {
	promise = promise.Normalize()
	h := sha256.New()
	writeString(h, domainConditionalPromise)
	writeByte(h, CanonicalEncodingVersion)
	writeString(h, promise.PromiseID)
	writeString(h, promise.ChannelID)
	writeString(h, promise.Source)
	writeString(h, promise.Destination)
	writeString(h, promise.Amount)
	writeString(h, promise.Fee)
	writeString(h, promise.HashLock)
	writeUint64(h, promise.TimeoutHeight)
	writeInt64(h, promise.TimeoutTimestamp)
	writeString(h, string(promise.ConditionType))
	writeString(h, promise.RouteIDOptional)
	writeString(h, promise.PreviousPromiseIDOptional)
	writeString(h, promise.NextPromiseIDOptional)
	writeUint64(h, promise.Nonce)
	return hex.EncodeToString(h.Sum(nil))
}

func SignaturePreimage(signer, stateHash string) []byte {
	var buf bytes.Buffer
	writeString(&buf, "aetra-payment-state-signature-preimage-v1")
	writeString(&buf, signer)
	writeString(&buf, stateHash)
	return buf.Bytes()
}

func ComputeSignatureEnvelopeHash(signer, chainID, channelID, objectType string, version uint32, nonce uint64, objectID string, expirationHeight uint64, commitmentHash string) string {
	h := sha256.New()
	writeString(h, domainSignatureEnvelope)
	writeByte(h, CanonicalEncodingVersion)
	writeString(h, strings.TrimSpace(signer))
	writeString(h, strings.TrimSpace(chainID))
	writeString(h, normalizeHash(channelID))
	writeString(h, strings.TrimSpace(objectType))
	writeUint64(h, uint64(version))
	writeUint64(h, nonce)
	writeString(h, strings.TrimSpace(objectID))
	writeUint64(h, expirationHeight)
	writeString(h, normalizeHash(commitmentHash))
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeSignedNonceWALHash(record SignedNonceRecord) string {
	record = record.Normalize()
	h := sha256.New()
	writeString(h, domainSignedNonceWAL)
	writeByte(h, CanonicalEncodingVersion)
	writeString(h, record.Signer)
	writeString(h, record.ChainID)
	writeString(h, record.ChannelID)
	writeUint64(h, record.Epoch)
	writeUint64(h, record.Nonce)
	writeString(h, record.StateHash)
	writeString(h, record.IsolationMode)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeSignatureHash(signer, stateHash string) string {
	h := sha256.New()
	writeString(h, "aetra-payment-state-signature-v1")
	_, _ = h.Write(SignaturePreimage(signer, stateHash))
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeUnidirectionalClaimHash(claim UnidirectionalClaim) string {
	claim = claim.Normalize()
	h := sha256.New()
	writeString(h, "aetra-payment-unidirectional-claim-v1")
	writeByte(h, CanonicalEncodingVersion)
	writeString(h, claim.ChainID)
	writeString(h, claim.ChannelID)
	writeString(h, claim.Payer)
	writeString(h, claim.Receiver)
	writeString(h, claim.LockedAmount)
	writeString(h, claim.ClaimedAmount)
	writeUint64(h, claim.Nonce)
	writeUint64(h, claim.ExpirationHeight)
	writeInt64(h, claim.ExpirationTimestamp)
	return hex.EncodeToString(h.Sum(nil))
}

func ClaimSignaturePreimage(signer, claimHash string) []byte {
	var buf bytes.Buffer
	writeString(&buf, "aetra-payment-unidirectional-claim-signature-preimage-v1")
	writeString(&buf, signer)
	writeString(&buf, claimHash)
	return buf.Bytes()
}

func ComputeClaimSignatureHash(signer, claimHash string) string {
	h := sha256.New()
	writeString(h, "aetra-payment-unidirectional-claim-signature-v1")
	_, _ = h.Write(ClaimSignaturePreimage(signer, claimHash))
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAsyncDeltaHash(delta AsyncPaymentDelta) string {
	delta = delta.Normalize()
	h := sha256.New()
	writeString(h, domainAsyncDelta)
	writeByte(h, CanonicalEncodingVersion)
	writeString(h, delta.UpdateID)
	writeString(h, delta.ChainID)
	writeString(h, delta.ChannelID)
	writeString(h, delta.From)
	writeString(h, delta.To)
	writeString(h, delta.Direction)
	writeString(h, delta.Amount)
	writeUint64(h, delta.NonceStart)
	writeUint64(h, delta.NonceEnd)
	writeUint64(h, delta.ExpiryHeight)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAsyncDeltaRoot(deltas []AsyncPaymentDelta) string {
	h := sha256.New()
	writeString(h, domainAsyncDeltaRoot)
	writeByte(h, CanonicalEncodingVersion)
	for _, delta := range normalizeAsyncDeltas(deltas) {
		writeString(h, delta.UpdateID)
		writeString(h, delta.DeltaHash)
		writeString(h, delta.Signature.Signer)
		writeString(h, delta.Signature.SignatureHash)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAsyncDeltaRootForChannel(channel ChannelRecord, deltas []AsyncPaymentDelta) string {
	channel = channel.Normalize()
	h := sha256.New()
	writeString(h, domainAsyncDeltaRoot)
	writeByte(h, CanonicalEncodingVersion)
	writeString(h, channel.ChainID)
	writeString(h, channel.ChannelID)
	writeString(h, string(channel.ChannelType))
	for _, delta := range normalizeAsyncDeltas(deltas) {
		writeString(h, delta.UpdateID)
		writeString(h, delta.ChainID)
		writeString(h, delta.ChannelID)
		writeString(h, delta.DeltaHash)
		writeString(h, delta.Signature.Signer)
		writeString(h, delta.Signature.SignatureHash)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func DeltaSignaturePreimage(signer, deltaHash string) []byte {
	var buf bytes.Buffer
	writeString(&buf, "aetra-payment-async-delta-signature-preimage-v1")
	writeString(&buf, signer)
	writeString(&buf, deltaHash)
	return buf.Bytes()
}

func ComputeDeltaSignatureHash(signer, deltaHash string) string {
	h := sha256.New()
	writeString(h, "aetra-payment-async-delta-signature-v1")
	_, _ = h.Write(DeltaSignaturePreimage(signer, deltaHash))
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeVirtualChannelAnchor(vc VirtualChannel) string {
	vc = vc.Normalize()
	h := sha256.New()
	writeString(h, domainVirtualChannelAnchor)
	writeByte(h, CanonicalEncodingVersion)
	writeString(h, vc.ChainID)
	writeString(h, vc.VirtualChannelID)
	writeUint64(h, vc.Nonce)
	writeString(h, vc.ParentRouteID)
	for _, id := range vc.ParentChannelIDs {
		writeString(h, id)
	}
	for _, endpoint := range vc.Endpoints {
		writeString(h, endpoint)
	}
	writeString(h, vc.EndpointA)
	writeString(h, vc.EndpointB)
	writeString(h, vc.IntermediarySetHash)
	writeString(h, vc.Capacity)
	writeString(h, vc.BalanceA)
	writeString(h, vc.BalanceB)
	writeString(h, vc.RoutingFeeAmount)
	writeString(h, vc.AnchorFeePaid)
	writeUint64(h, vc.ExpiresHeight)
	writeString(h, vc.ConditionRoot)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeVirtualChannelStateHash(vc VirtualChannel) string {
	vc = vc.Normalize()
	h := sha256.New()
	writeString(h, domainVirtualChannelState)
	writeByte(h, CanonicalEncodingVersion)
	writeString(h, vc.ChainID)
	writeString(h, vc.VirtualChannelID)
	writeUint64(h, vc.Nonce)
	writeString(h, vc.ParentRouteID)
	for _, id := range vc.ParentChannelIDs {
		writeString(h, id)
	}
	for _, endpoint := range vc.Endpoints {
		writeString(h, endpoint)
	}
	writeString(h, vc.EndpointA)
	writeString(h, vc.EndpointB)
	for _, intermediary := range vc.Intermediaries {
		writeString(h, intermediary)
	}
	writeString(h, vc.IntermediarySetHash)
	writeString(h, vc.Capacity)
	writeString(h, vc.BalanceA)
	writeString(h, vc.BalanceB)
	writeString(h, vc.RoutingFeeAmount)
	writeString(h, vc.AnchorFeePaid)
	writeUint64(h, vc.ExpiresHeight)
	writeString(h, string(vc.Status))
	writeString(h, vc.AnchorCommitment)
	writeString(h, vc.ConditionRoot)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeCooperativeCloseHash(chainID, channelID, stateHash string, nonce uint64) string {
	h := sha256.New()
	writeString(h, domainCooperativeClose)
	writeByte(h, CanonicalEncodingVersion)
	writeString(h, strings.TrimSpace(chainID))
	writeString(h, normalizeHash(channelID))
	writeString(h, normalizeHash(stateHash))
	writeUint64(h, nonce)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeDisputeProofHash(proof FraudProof) string {
	proof = proof.Normalize()
	h := sha256.New()
	writeString(h, domainDisputeProof)
	writeByte(h, CanonicalEncodingVersion)
	writeString(h, proof.StateA.ChainID)
	writeString(h, proof.ChannelID())
	writeString(h, proof.ProofID)
	writeString(h, string(proof.ProofType))
	writeString(h, proof.SubmittedBy)
	writeString(h, proof.OffendingSigner)
	writeString(h, proof.StateA.StateHash)
	writeUint64(h, proof.StateA.Nonce)
	writeString(h, proof.StateB.StateHash)
	writeUint64(h, proof.StateB.Nonce)
	writeString(h, proof.PenaltyDenom)
	writeString(h, proof.PenaltyAmount)
	writeString(h, proof.EvidenceHash)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeSettlementHash(settlement SettlementRecord) string {
	settlement = settlement.Normalize()
	h := sha256.New()
	writeString(h, domainSettlementResult)
	writeByte(h, CanonicalEncodingVersion)
	writeString(h, settlement.ChainID)
	writeString(h, settlement.ChannelID)
	writeString(h, settlement.StateHash)
	writeUint64(h, settlement.Nonce)
	writeUint64(h, settlement.SettledHeight)
	writeString(h, settlement.SettlementFeeDenom)
	writeString(h, settlement.SettlementFee)
	for _, balance := range settlement.FinalBalances {
		writeString(h, balance.Participant)
		writeString(h, balance.Amount)
	}
	for _, penalty := range settlement.Penalties {
		writeString(h, penalty.Offender)
		writeString(h, penalty.Recipient)
		writeString(h, penalty.Denom)
		writeString(h, penalty.Amount)
	}
	for _, allocation := range settlement.PenaltyAllocations {
		writeString(h, allocation.Offender)
		writeString(h, string(allocation.Route))
		writeString(h, allocation.Denom)
		writeString(h, allocation.Amount)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeSettlementResultCommitment(channel ChannelRecord, settlement SettlementRecord) string {
	channel = channel.Normalize()
	settlement = settlement.Normalize()
	if settlement.ChainID == "" {
		settlement.ChainID = channel.ChainID
	}
	return ComputeSettlementHash(settlement)
}

func ComputeBatchRoot(operations []SettlementOperation) string {
	ordered := SortSettlementOperations(operations)
	h := sha256.New()
	writeString(h, "aetra-payment-settlement-batch-v1")
	for _, op := range ordered {
		writeString(h, op.OperationID)
		writeString(h, string(op.OperationType))
		writeString(h, op.ChannelID)
		writeUint64(h, op.Nonce)
		writeString(h, op.StateHash)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ValidateHash(fieldName, value string) error {
	if len(value) != HashHexLength {
		return fmt.Errorf("%s must be %d lowercase hex chars", fieldName, HashHexLength)
	}
	for _, r := range value {
		if r >= '0' && r <= '9' || r >= 'a' && r <= 'f' {
			continue
		}
		return fmt.Errorf("%s must be %d lowercase hex chars", fieldName, HashHexLength)
	}
	return nil
}

func writeString(w interface{ Write([]byte) (int, error) }, value string) {
	writeUint64(w, uint64(len(value)))
	_, _ = w.Write([]byte(value))
}

func writeByte(w interface{ Write([]byte) (int, error) }, value byte) {
	_, _ = w.Write([]byte{value})
}

func writeUint64(w interface{ Write([]byte) (int, error) }, value uint64) {
	var bz [8]byte
	binary.BigEndian.PutUint64(bz[:], value)
	_, _ = w.Write(bz[:])
}

func writeInt64(w interface{ Write([]byte) (int, error) }, value int64) {
	writeUint64(w, uint64(value))
}

func compareString(left, right string) int {
	return bytes.Compare([]byte(left), []byte(right))
}

func sortStrings(values []string) {
	sort.SliceStable(values, func(i, j int) bool {
		return values[i] < values[j]
	})
}
