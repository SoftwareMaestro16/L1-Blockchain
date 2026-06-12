package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/sovereign-l1/l1/app/addressing"
)

type SignerKeyRole string

const (
	SignerKeyRoleParticipant	SignerKeyRole	= "PARTICIPANT"
	SignerKeyRoleRoutingGossip	SignerKeyRole	= "ROUTING_GOSSIP"
)

type ChannelSigningLimit struct {
	ChannelID		string
	MaxNonce		uint64
	MaxAmount		string
	MaxSignatures		uint32
	ValidUntilHeight	uint64
}

type SignerEmergencyPause struct {
	Signer		string
	ChannelID	string
	Reason		string
	PausedHeight	uint64
}

type SignedStateAuditLog struct {
	LogID		string
	Signer		string
	KeyRole		SignerKeyRole
	ChainID		string
	ChannelID	string
	Epoch		uint64
	Nonce		uint64
	StateHash	string
	CommitmentHash	string
	WALHash		string
	SignatureHash	string
	Height		uint64
	AuditHash	string
}

type PaymentSignerConfig struct {
	Signer			string
	KeyRole			SignerKeyRole
	FundsKey		string
	GossipKey		string
	IsolationMode		string
	ChannelLimits		[]ChannelSigningLimit
	EmergencyPauses		[]SignerEmergencyPause
	AutomatedSigning	bool
	MaxAutomatedAmount	string
	MaxAutomatedPerBlock	uint32
}

type PaymentSignerAPI struct {
	Config		PaymentSignerConfig
	NonceStore	SignerPersistence
	AuditLogs	[]SignedStateAuditLog
	BlockHeight	uint64
	BlockSigned	uint32
}

type SignStateRequest struct {
	State		ChannelState
	Signer		string
	KeyRole		SignerKeyRole
	Amount		string
	Automated	bool
	CurrentHeight	uint64
}

type SignStateResponse struct {
	Signature	StateSignature
	AuditLog	SignedStateAuditLog
	WALRecord	SignedNonceRecord
}

type KeyCompromiseCloseRequest struct {
	ChannelID	string
	CompromisedKey	string
	SafeSubmitter	string
	LatestState	ChannelState
	CurrentHeight	uint64
	SettlementFee	string
	EvidenceHash	string
}

func NewPaymentSignerAPI(config PaymentSignerConfig, store SignerPersistence, logs []SignedStateAuditLog) (PaymentSignerAPI, error) {
	config = config.Normalize()
	if err := config.Validate(); err != nil {
		return PaymentSignerAPI{}, err
	}
	store = store.Normalize()
	if config.IsolationMode != "" {
		store.IsolationMode = config.IsolationMode
	}
	api := PaymentSignerAPI{
		Config:		config,
		NonceStore:	store.Normalize(),
		AuditLogs:	normalizeSignedStateAuditLogs(logs),
	}
	return api, nil
}

func (api PaymentSignerAPI) SignState(req SignStateRequest) (PaymentSignerAPI, SignStateResponse, error) {
	api = api.Normalize()
	req = req.Normalize()
	if err := req.Validate(api.Config); err != nil {
		return PaymentSignerAPI{}, SignStateResponse{}, err
	}
	if err := api.Config.ValidateSigningRequest(req); err != nil {
		return PaymentSignerAPI{}, SignStateResponse{}, err
	}
	if err := api.validateSignatureQuota(req); err != nil {
		return PaymentSignerAPI{}, SignStateResponse{}, err
	}
	if req.CurrentHeight != api.BlockHeight {
		api.BlockHeight = req.CurrentHeight
		api.BlockSigned = 0
	}
	if req.Automated && api.Config.MaxAutomatedPerBlock > 0 && api.BlockSigned >= api.Config.MaxAutomatedPerBlock {
		return PaymentSignerAPI{}, SignStateResponse{}, errors.New("payments signer automated per-block limit exceeded")
	}
	nextStore, sig, err := SignStateWithWriteAhead(api.NonceStore.Records, req.State, req.Signer, api.Config.IsolationMode)
	if err != nil {
		return PaymentSignerAPI{}, SignStateResponse{}, err
	}
	api.NonceStore.Records = nextStore
	wal, found := signedNonceRecordForSignature(api.NonceStore.Records, sig)
	if !found {
		return PaymentSignerAPI{}, SignStateResponse{}, errors.New("payments signer wal record missing after signature")
	}
	log := BuildSignedStateAuditLog(req, sig, wal)
	if err := log.Validate(); err != nil {
		return PaymentSignerAPI{}, SignStateResponse{}, err
	}
	api.AuditLogs = append(api.AuditLogs, log)
	api.AuditLogs = normalizeSignedStateAuditLogs(api.AuditLogs)
	if req.Automated {
		api.BlockSigned++
	}
	return api.Normalize(), SignStateResponse{Signature: sig, AuditLog: log, WALRecord: wal}, nil
}

func (api PaymentSignerAPI) SignGossip(message GossipMessage, signer string, currentHeight uint64) (PaymentSignerAPI, SignedGossipEnvelope, error) {
	api = api.Normalize()
	signer = strings.TrimSpace(signer)
	if err := api.Config.ValidateGossipSigningRequest(signer, currentHeight); err != nil {
		return PaymentSignerAPI{}, SignedGossipEnvelope{}, err
	}
	message = message.Normalize()
	if currentHeight < message.ValidAfterHeight {
		return PaymentSignerAPI{}, SignedGossipEnvelope{}, errors.New("payments gossip signer message is not yet valid")
	}
	if currentHeight > message.ValidUntilHeight {
		return PaymentSignerAPI{}, SignedGossipEnvelope{}, errors.New("payments gossip signer message is expired")
	}
	built, err := BuildGossipMessage(message)
	if err != nil {
		return PaymentSignerAPI{}, SignedGossipEnvelope{}, err
	}
	if built.NodeID != signer {
		return PaymentSignerAPI{}, SignedGossipEnvelope{}, errors.New("payments gossip signer must match routing node")
	}
	sig, err := SignatureForGossip(built, signer)
	if err != nil {
		return PaymentSignerAPI{}, SignedGossipEnvelope{}, err
	}
	envelope := SignedGossipEnvelope{
		Message:	built,
		MessageHash:	built.MessageID,
		Signature:	sig,
		ReceivedFrom:	signer,
		ReceivedAt:	currentHeight,
	}.Normalize()
	return api, envelope, nil
}

func DurableNonceStoreCommit(store SignerPersistence, record SignedNonceRecord) (SignerPersistence, error) {
	store = store.Normalize()
	record = record.Normalize()
	if err := record.Validate(); err != nil {
		return SignerPersistence{}, err
	}
	for _, existing := range store.Records {
		if existing.Signer == record.Signer && existing.ChainID == record.ChainID && existing.ChannelID == record.ChannelID && existing.Epoch == record.Epoch && existing.Nonce == record.Nonce && existing.StateHash != record.StateHash {
			return SignerPersistence{}, errors.New("payments durable nonce store rejects same nonce divergent state")
		}
	}
	store.Records = appendOrReplaceSignedNonceRecord(store.Records, record)
	return store.Normalize(), nil
}

func BuildSignedStateAuditLog(req SignStateRequest, sig StateSignature, wal SignedNonceRecord) SignedStateAuditLog {
	req = req.Normalize()
	sig = sig.Normalize()
	wal = wal.Normalize()
	log := SignedStateAuditLog{
		Signer:		sig.Signer,
		KeyRole:	req.KeyRole,
		ChainID:	sig.ChainID,
		ChannelID:	sig.ChannelID,
		Epoch:		wal.Epoch,
		Nonce:		sig.Nonce,
		StateHash:	sig.StateHash,
		CommitmentHash:	sig.CommitmentHash,
		WALHash:	wal.WALHash,
		SignatureHash:	sig.SignatureHash,
		Height:		req.CurrentHeight,
	}
	log.LogID = HashParts("signed-state-audit-log", log.Signer, log.ChannelID, fmt.Sprintf("%d", log.Epoch), fmt.Sprintf("%d", log.Nonce), log.StateHash, log.SignatureHash)
	log.AuditHash = ComputeSignedStateAuditHash(log)
	return log.Normalize()
}

func ComputeSignedStateAuditHash(log SignedStateAuditLog) string {
	log = log.Normalize()
	return HashParts(
		"signed-state-audit",
		log.LogID,
		log.Signer,
		string(log.KeyRole),
		log.ChainID,
		log.ChannelID,
		fmt.Sprintf("%d", log.Epoch),
		fmt.Sprintf("%d", log.Nonce),
		log.StateHash,
		log.CommitmentHash,
		log.WALHash,
		log.SignatureHash,
		fmt.Sprintf("%d", log.Height),
	)
}

func EmergencyPauseSigner(config PaymentSignerConfig, signer, channelID, reason string, currentHeight uint64) (PaymentSignerConfig, error) {
	config = config.Normalize()
	if currentHeight == 0 {
		return PaymentSignerConfig{}, errors.New("payments signer pause height must be positive")
	}
	pause := SignerEmergencyPause{
		Signer:		strings.TrimSpace(signer),
		ChannelID:	normalizeHash(channelID),
		Reason:		strings.TrimSpace(reason),
		PausedHeight:	currentHeight,
	}
	if err := pause.Validate(); err != nil {
		return PaymentSignerConfig{}, err
	}
	replaced := false
	for i, existing := range config.EmergencyPauses {
		existing = existing.Normalize()
		if existing.Signer == pause.Signer && existing.ChannelID == pause.ChannelID {
			config.EmergencyPauses[i] = pause
			replaced = true
			break
		}
	}
	if !replaced {
		config.EmergencyPauses = append(config.EmergencyPauses, pause)
	}
	config.EmergencyPauses = normalizeSignerEmergencyPauses(config.EmergencyPauses)
	return config, config.Validate()
}

func SubmitKeyCompromiseClose(state PaymentsState, req KeyCompromiseCloseRequest) (PaymentsState, error) {
	state = state.Export()
	req = req.Normalize()
	channel, found := state.ChannelByID(req.ChannelID)
	if !found {
		return PaymentsState{}, errors.New("payments key compromise close channel not found")
	}
	if err := req.ValidateForChannel(channel); err != nil {
		return PaymentsState{}, err
	}
	return SubmitCloseWithRequest(state, ChannelCloseRequest{
		ChannelID:	req.ChannelID,
		ClosingState:	req.LatestState,
		CloseReason:	CloseReasonFraud,
		Submitter:	req.SafeSubmitter,
		CurrentHeight:	req.CurrentHeight,
		SettlementFee:	req.SettlementFee,
	})
}

func (api PaymentSignerAPI) Normalize() PaymentSignerAPI {
	api.Config = api.Config.Normalize()
	api.NonceStore = api.NonceStore.Normalize()
	api.AuditLogs = normalizeSignedStateAuditLogs(api.AuditLogs)
	return api
}

func (c PaymentSignerConfig) Normalize() PaymentSignerConfig {
	c.Signer = strings.TrimSpace(c.Signer)
	c.FundsKey = strings.TrimSpace(c.FundsKey)
	c.GossipKey = strings.TrimSpace(c.GossipKey)
	c.IsolationMode = strings.TrimSpace(c.IsolationMode)
	if c.IsolationMode == "" {
		c.IsolationMode = SignerIsolationProcess
	}
	if c.KeyRole == "" {
		c.KeyRole = SignerKeyRoleParticipant
	}
	c.MaxAutomatedAmount = strings.TrimSpace(c.MaxAutomatedAmount)
	if c.MaxAutomatedAmount == "" {
		c.MaxAutomatedAmount = "0"
	}
	c.ChannelLimits = normalizeChannelSigningLimits(c.ChannelLimits)
	c.EmergencyPauses = normalizeSignerEmergencyPauses(c.EmergencyPauses)
	return c
}

func (c PaymentSignerConfig) Validate() error {
	c = c.Normalize()
	if err := addressing.ValidateUserAddress("payments signer api signer", c.Signer); err != nil {
		return err
	}
	if !IsSignerKeyRole(c.KeyRole) {
		return fmt.Errorf("unknown payments signer key role %q", c.KeyRole)
	}
	if c.IsolationMode != SignerIsolationProcess && c.IsolationMode != SignerIsolationHardware {
		return errors.New("payments signer api isolation mode is unsupported")
	}
	if c.FundsKey != "" {
		if err := addressing.ValidateUserAddress("payments signer funds key", c.FundsKey); err != nil {
			return err
		}
	}
	if c.GossipKey != "" {
		if err := addressing.ValidateUserAddress("payments signer gossip key", c.GossipKey); err != nil {
			return err
		}
	}
	if c.FundsKey != "" && c.GossipKey != "" && c.FundsKey == c.GossipKey {
		return errors.New("payments routing signer gossip key must be separate from funds key")
	}
	if err := validateNonNegativeInt("payments signer automated amount limit", c.MaxAutomatedAmount); err != nil {
		return err
	}
	for _, limit := range c.ChannelLimits {
		if err := limit.Validate(); err != nil {
			return err
		}
	}
	for _, pause := range c.EmergencyPauses {
		if err := pause.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (c PaymentSignerConfig) ValidateSigningRequest(req SignStateRequest) error {
	c = c.Normalize()
	req = req.Normalize()
	if req.Signer != c.Signer {
		return errors.New("payments signer request signer mismatch")
	}
	if c.FundsKey != "" && req.Signer != c.FundsKey {
		return errors.New("payments signer request funds key mismatch")
	}
	if c.GossipKey != "" && req.Signer == c.GossipKey {
		return errors.New("payments routing gossip key cannot sign channel state")
	}
	if req.KeyRole != c.KeyRole {
		return errors.New("payments signer request key role mismatch")
	}
	for _, pause := range c.EmergencyPauses {
		pause = pause.Normalize()
		if pause.Signer == req.Signer && (pause.ChannelID == "" || pause.ChannelID == req.State.ChannelID) {
			return errors.New("payments signer is emergency paused")
		}
	}
	for _, limit := range c.ChannelLimits {
		limit = limit.Normalize()
		if limit.ChannelID != req.State.ChannelID {
			continue
		}
		if limit.ValidUntilHeight > 0 && req.CurrentHeight > limit.ValidUntilHeight {
			return errors.New("payments signer channel limit expired")
		}
		if limit.MaxNonce > 0 && req.State.Nonce > limit.MaxNonce {
			return errors.New("payments signer channel nonce limit exceeded")
		}
		if limit.MaxAmount != "" && limit.MaxAmount != "0" {
			amount, err := parseNonNegativeInt("payments signer request amount", req.Amount)
			if err != nil {
				return err
			}
			maxAmount, err := parseNonNegativeInt("payments signer channel max amount", limit.MaxAmount)
			if err != nil {
				return err
			}
			if amount.GT(maxAmount) {
				return errors.New("payments signer channel amount limit exceeded")
			}
		}
	}
	if req.Automated {
		if !c.AutomatedSigning {
			return errors.New("payments signer automated signing disabled")
		}
		if c.MaxAutomatedAmount != "" && c.MaxAutomatedAmount != "0" {
			amount, err := parseNonNegativeInt("payments signer automated amount", req.Amount)
			if err != nil {
				return err
			}
			maxAmount, err := parseNonNegativeInt("payments signer max automated amount", c.MaxAutomatedAmount)
			if err != nil {
				return err
			}
			if amount.GT(maxAmount) {
				return errors.New("payments signer automated amount limit exceeded")
			}
		}
	}
	return nil
}

func (c PaymentSignerConfig) ValidateGossipSigningRequest(signer string, currentHeight uint64) error {
	c = c.Normalize()
	signer = strings.TrimSpace(signer)
	if currentHeight == 0 {
		return errors.New("payments gossip signing height must be positive")
	}
	if c.KeyRole != SignerKeyRoleRoutingGossip {
		return errors.New("payments gossip signing requires routing gossip key role")
	}
	if signer != c.Signer {
		return errors.New("payments gossip signer mismatch")
	}
	if c.GossipKey != "" && signer != c.GossipKey {
		return errors.New("payments gossip signer key mismatch")
	}
	if c.FundsKey != "" && signer == c.FundsKey {
		return errors.New("payments funds key cannot sign routing gossip")
	}
	if err := c.Validate(); err != nil {
		return err
	}
	return nil
}

func (r SignStateRequest) Normalize() SignStateRequest {
	r.State = r.State.Normalize()
	r.Signer = strings.TrimSpace(r.Signer)
	if r.Signer == "" {
		r.Signer = firstNonEmptySigner(r.State.Signatures)
	}
	if r.KeyRole == "" {
		r.KeyRole = SignerKeyRoleParticipant
	}
	r.Amount = strings.TrimSpace(r.Amount)
	if r.Amount == "" {
		r.Amount = "0"
	}
	return r
}

func (r SignStateRequest) Validate(config PaymentSignerConfig) error {
	r = r.Normalize()
	if r.CurrentHeight == 0 {
		return errors.New("payments signer request height must be positive")
	}
	if err := addressing.ValidateUserAddress("payments signer request signer", r.Signer); err != nil {
		return err
	}
	if !IsSignerKeyRole(r.KeyRole) {
		return fmt.Errorf("unknown payments signer request role %q", r.KeyRole)
	}
	if err := validateNonNegativeInt("payments signer request amount", r.Amount); err != nil {
		return err
	}
	if r.State.StateHash == "" {
		return errors.New("payments signer request state hash is required")
	}
	if config.KeyRole == SignerKeyRoleRoutingGossip {
		return errors.New("payments routing gossip key cannot sign channel state")
	}
	return nil
}

func (r ChannelSigningLimit) Normalize() ChannelSigningLimit {
	r.ChannelID = normalizeHash(r.ChannelID)
	r.MaxAmount = strings.TrimSpace(r.MaxAmount)
	if r.MaxAmount == "" {
		r.MaxAmount = "0"
	}
	return r
}

func (r ChannelSigningLimit) Validate() error {
	limit := r.Normalize()
	if err := ValidateHash("payments signer channel limit id", limit.ChannelID); err != nil {
		return err
	}
	if err := validateNonNegativeInt("payments signer channel amount limit", limit.MaxAmount); err != nil {
		return err
	}
	if limit.MaxNonce == 0 && limit.MaxAmount == "0" && limit.MaxSignatures == 0 {
		return errors.New("payments signer channel limit must constrain nonce amount or signatures")
	}
	return nil
}

func (p SignerEmergencyPause) Normalize() SignerEmergencyPause {
	p.Signer = strings.TrimSpace(p.Signer)
	p.ChannelID = normalizeOptionalHash(p.ChannelID)
	p.Reason = strings.TrimSpace(p.Reason)
	return p
}

func (p SignerEmergencyPause) Validate() error {
	pause := p.Normalize()
	if err := addressing.ValidateUserAddress("payments signer pause signer", pause.Signer); err != nil {
		return err
	}
	if pause.ChannelID != "" {
		if err := ValidateHash("payments signer pause channel id", pause.ChannelID); err != nil {
			return err
		}
	}
	if pause.Reason == "" {
		return errors.New("payments signer pause reason is required")
	}
	if pause.PausedHeight == 0 {
		return errors.New("payments signer pause height must be positive")
	}
	return nil
}

func (l SignedStateAuditLog) Normalize() SignedStateAuditLog {
	l.LogID = normalizeOptionalHash(l.LogID)
	l.Signer = strings.TrimSpace(l.Signer)
	l.ChainID = strings.TrimSpace(l.ChainID)
	l.ChannelID = normalizeHash(l.ChannelID)
	l.StateHash = normalizeHash(l.StateHash)
	l.CommitmentHash = normalizeHash(l.CommitmentHash)
	l.WALHash = normalizeHash(l.WALHash)
	l.SignatureHash = normalizeHash(l.SignatureHash)
	l.AuditHash = normalizeOptionalHash(l.AuditHash)
	if l.KeyRole == "" {
		l.KeyRole = SignerKeyRoleParticipant
	}
	return l
}

func (l SignedStateAuditLog) Validate() error {
	log := l.Normalize()
	if err := ValidateHash("payments signed state audit log id", log.LogID); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments signed state audit signer", log.Signer); err != nil {
		return err
	}
	if !IsSignerKeyRole(log.KeyRole) {
		return fmt.Errorf("unknown payments signed state audit key role %q", log.KeyRole)
	}
	if log.ChainID == "" {
		return errors.New("payments signed state audit chain id is required")
	}
	if log.Epoch == 0 || log.Nonce == 0 || log.Height == 0 {
		return errors.New("payments signed state audit heights and nonces must be positive")
	}
	for _, item := range []struct {
		name	string
		value	string
	}{
		{"payments signed state audit channel id", log.ChannelID},
		{"payments signed state audit state hash", log.StateHash},
		{"payments signed state audit commitment hash", log.CommitmentHash},
		{"payments signed state audit wal hash", log.WALHash},
		{"payments signed state audit signature hash", log.SignatureHash},
		{"payments signed state audit hash", log.AuditHash},
	} {
		if err := ValidateHash(item.name, item.value); err != nil {
			return err
		}
	}
	if expected := ComputeSignedStateAuditHash(log); log.AuditHash != expected {
		return errors.New("payments signed state audit hash mismatch")
	}
	return nil
}

func (r KeyCompromiseCloseRequest) Normalize() KeyCompromiseCloseRequest {
	r.ChannelID = normalizeHash(r.ChannelID)
	r.CompromisedKey = strings.TrimSpace(r.CompromisedKey)
	r.SafeSubmitter = strings.TrimSpace(r.SafeSubmitter)
	r.LatestState = r.LatestState.Normalize()
	r.SettlementFee = strings.TrimSpace(r.SettlementFee)
	if r.SettlementFee == "" {
		r.SettlementFee = "0"
	}
	r.EvidenceHash = normalizeOptionalHash(r.EvidenceHash)
	return r
}

func (r KeyCompromiseCloseRequest) ValidateForChannel(channel ChannelRecord) error {
	req := r.Normalize()
	channel = channel.Normalize()
	if req.ChannelID != channel.ChannelID {
		return errors.New("payments key compromise close channel mismatch")
	}
	if err := addressing.ValidateUserAddress("payments compromised key", req.CompromisedKey); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments key compromise safe submitter", req.SafeSubmitter); err != nil {
		return err
	}
	if req.CompromisedKey == req.SafeSubmitter {
		return errors.New("payments key compromise close requires separate safe submitter")
	}
	if !containsString(channel.Participants, req.CompromisedKey) || !containsString(channel.Participants, req.SafeSubmitter) {
		return errors.New("payments key compromise close keys must be channel participants")
	}
	if err := req.LatestState.ValidateForChannel(channel, true); err != nil {
		return err
	}
	if req.LatestState.Nonce < channel.LatestState.Nonce {
		return errors.New("payments key compromise close state nonce below latest")
	}
	if err := validateNonNegativeInt("payments key compromise settlement fee", req.SettlementFee); err != nil {
		return err
	}
	if req.EvidenceHash != "" {
		if err := ValidateHash("payments key compromise evidence hash", req.EvidenceHash); err != nil {
			return err
		}
	}
	return nil
}

func (r SignedNonceRecord) Validate() error {
	record := r.Normalize()
	if err := addressing.ValidateUserAddress("payments signed nonce signer", record.Signer); err != nil {
		return err
	}
	if record.ChainID == "" {
		return errors.New("payments signed nonce chain id is required")
	}
	if err := ValidateHash("payments signed nonce channel id", record.ChannelID); err != nil {
		return err
	}
	if record.Epoch == 0 || record.Nonce == 0 {
		return errors.New("payments signed nonce epoch and nonce must be positive")
	}
	if err := ValidateHash("payments signed nonce state hash", record.StateHash); err != nil {
		return err
	}
	if record.IsolationMode != SignerIsolationProcess && record.IsolationMode != SignerIsolationHardware {
		return errors.New("payments signed nonce isolation mode is unsupported")
	}
	if err := ValidateHash("payments signed nonce wal hash", record.WALHash); err != nil {
		return err
	}
	if expected := ComputeSignedNonceWALHash(record); record.WALHash != expected {
		return errors.New("payments signed nonce wal hash mismatch")
	}
	return nil
}

func IsSignerKeyRole(value SignerKeyRole) bool {
	switch value {
	case SignerKeyRoleParticipant, SignerKeyRoleRoutingGossip:
		return true
	default:
		return false
	}
}

func signedNonceRecordForSignature(records []SignedNonceRecord, sig StateSignature) (SignedNonceRecord, bool) {
	sig = sig.Normalize()
	for _, record := range normalizeSignedNonceRecords(records) {
		if record.Signer == sig.Signer && record.ChainID == sig.ChainID && record.ChannelID == sig.ChannelID && record.Nonce == sig.Nonce && record.StateHash == sig.StateHash {
			return record, true
		}
	}
	return SignedNonceRecord{}, false
}

func (api PaymentSignerAPI) validateSignatureQuota(req SignStateRequest) error {
	for _, limit := range api.Config.ChannelLimits {
		limit = limit.Normalize()
		if limit.ChannelID != req.State.ChannelID || limit.MaxSignatures == 0 {
			continue
		}
		var count uint32
		alreadyReleased := false
		for _, record := range normalizeSignedNonceRecords(api.NonceStore.Records) {
			if record.Signer != req.Signer || record.ChannelID != req.State.ChannelID || !record.Released {
				continue
			}
			count++
			if record.ChainID == req.State.ChainID && record.Epoch == req.State.Epoch && record.Nonce == req.State.Nonce && record.StateHash == req.State.StateHash {
				alreadyReleased = true
			}
		}
		if count >= limit.MaxSignatures && !alreadyReleased {
			return errors.New("payments signer channel signature limit exceeded")
		}
	}
	return nil
}

func appendOrReplaceSignedNonceRecord(records []SignedNonceRecord, next SignedNonceRecord) []SignedNonceRecord {
	next = next.Normalize()
	out := make([]SignedNonceRecord, 0, len(records)+1)
	replaced := false
	for _, record := range normalizeSignedNonceRecords(records) {
		if record.Signer == next.Signer && record.ChainID == next.ChainID && record.ChannelID == next.ChannelID && record.Epoch == next.Epoch && record.Nonce == next.Nonce {
			out = append(out, next)
			replaced = true
			continue
		}
		out = append(out, record)
	}
	if !replaced {
		out = append(out, next)
	}
	return normalizeSignedNonceRecords(out)
}

func firstNonEmptySigner(signatures []StateSignature) string {
	for _, sig := range normalizeSignatures(signatures) {
		if sig.Signer != "" {
			return sig.Signer
		}
	}
	return ""
}

func normalizeChannelSigningLimits(limits []ChannelSigningLimit) []ChannelSigningLimit {
	out := make([]ChannelSigningLimit, len(limits))
	for i, limit := range limits {
		out[i] = limit.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].ChannelID < out[j].ChannelID })
	return out
}

func normalizeSignerEmergencyPauses(pauses []SignerEmergencyPause) []SignerEmergencyPause {
	out := make([]SignerEmergencyPause, len(pauses))
	for i, pause := range pauses {
		out[i] = pause.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Signer == out[j].Signer {
			return out[i].ChannelID < out[j].ChannelID
		}
		return out[i].Signer < out[j].Signer
	})
	return out
}

func normalizeSignedStateAuditLogs(logs []SignedStateAuditLog) []SignedStateAuditLog {
	out := make([]SignedStateAuditLog, len(logs))
	for i, log := range logs {
		out[i] = log.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Height == out[j].Height {
			return out[i].LogID < out[j].LogID
		}
		return out[i].Height < out[j].Height
	})
	return out
}
