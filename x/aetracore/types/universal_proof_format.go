package types

import (
	"errors"
	"fmt"
)

const UniversalProofVersionV1 = uint64(1)

type UniversalProofType string
type UniversalProofFailureCode string

const (
	ProofTypeAccountState		UniversalProofType	= "AccountStateProof"
	ProofTypeBalance		UniversalProofType	= "BalanceProof"
	ProofTypeZoneRoot		UniversalProofType	= "ZoneRootProof"
	ProofTypeShardRoot		UniversalProofType	= "ShardRootProof"
	ProofTypeMessageInclusion	UniversalProofType	= "MessageInclusionProof"
	ProofTypeMessageReceipt		UniversalProofType	= "MessageReceiptProof"
	ProofTypeDomainOwnership	UniversalProofType	= "DomainOwnershipProof"
	ProofTypeResolverRecord		UniversalProofType	= "ResolverRecordProof"
	ProofTypeContractState		UniversalProofType	= "ContractStateProof"
	ProofTypePaymentSettlement	UniversalProofType	= "PaymentSettlementProof"
	ProofTypeNonExistence		UniversalProofType	= "NonExistenceProof"
)

const (
	ProofFailureNone			UniversalProofFailureCode	= ""
	ProofFailureUntrustedHeader		UniversalProofFailureCode	= "ERR_UNTRUSTED_HEADER"
	ProofFailureChainIDMismatch		UniversalProofFailureCode	= "ERR_CHAIN_ID_MISMATCH"
	ProofFailureHeightUnavailable		UniversalProofFailureCode	= "ERR_HEIGHT_UNAVAILABLE"
	ProofFailureRootMismatch		UniversalProofFailureCode	= "ERR_ROOT_MISMATCH"
	ProofFailureZoneNotFound		UniversalProofFailureCode	= "ERR_ZONE_NOT_FOUND"
	ProofFailureShardNotFound		UniversalProofFailureCode	= "ERR_SHARD_NOT_FOUND"
	ProofFailureStoreProofInvalid		UniversalProofFailureCode	= "ERR_STORE_PROOF_INVALID"
	ProofFailureMessageNotIncluded		UniversalProofFailureCode	= "ERR_MESSAGE_NOT_INCLUDED"
	ProofFailureReceiptNotFound		UniversalProofFailureCode	= "ERR_RECEIPT_NOT_FOUND"
	ProofFailureObjectExpired		UniversalProofFailureCode	= "ERR_OBJECT_EXPIRED"
	ProofFailureNonExistenceProofInvalid	UniversalProofFailureCode	= "ERR_NON_EXISTENCE_PROOF_INVALID"
)

type UniversalTrustedHeader struct {
	ChainID		string
	Height		uint64
	AppHash		string
	HeaderHash	string
	Trusted		bool
}

type UniversalStoreProof struct {
	ProofVersion		uint64
	Key			[]byte
	Value			[]byte
	NonExistenceMarker	[]byte
	StoreRoot		string
	ProofOps		[]string
	ProofHash		string
}

type UniversalRootStep struct {
	Index		uint32
	FromRootType	RootType
	FromRoot	string
	ToRootType	RootType
	ToRoot		string
	Scope		string
	StepHash	string
}

type UniversalShardCommitment struct {
	Height		uint64
	ZoneID		ZoneID
	ShardID		ShardID
	ShardRoot	string
	ShardRootsRoot	string
	CommitmentHash	string
}

type UniversalMessageCommitment struct {
	Height			uint64
	MessageID		string
	MessageRoot		string
	SourceOutboxRoot	string
	DestinationInboxRoot	string
	ReceiptRoot		string
	ReceiptHash		string
	MessageCommitmentHash	string
	ReceiptCommitmentHash	string
	DeliveryCommitmentHash	string
}

type UniversalProofEnvelope struct {
	ProofType		UniversalProofType
	ProofVersion		uint64
	ChainID			string
	Height			uint64
	AppHash			string
	RootType		RootType
	ZoneID			ZoneID
	ShardID			ShardID
	Key			[]byte
	Value			[]byte
	AbsenceMarker		[]byte
	ObjectExpiryHeight	uint64
	StoreProof		UniversalStoreProof
	ZoneCommitment		ZoneCommitment
	HasZoneCommit		bool
	ShardCommitment		UniversalShardCommitment
	HasShardCommit		bool
	MessageCommit		UniversalMessageCommitment
	HasMessageCommit	bool
	VerificationPath	[]UniversalRootStep
	ProofHash		string
}

type UniversalProofVerificationResult struct {
	Verified	bool
	VerifiedAbsent	bool
	Value		[]byte
	FailureCode	UniversalProofFailureCode
	FailureMessage	string
	ProofHash	string
}

type universalProofRequirement struct {
	RootType	RootType
	ZoneID		ZoneID
	RequiresZone	bool
	RequiresShard	bool
	RequiresMessage	bool
	RequiresReceipt	bool
	RequiresAbsence	bool
}

func NewUniversalStoreProof(proof UniversalStoreProof) (UniversalStoreProof, error) {
	if proof.ProofVersion == 0 {
		proof.ProofVersion = UniversalProofVersionV1
	}
	proof.ProofOps = append([]string(nil), proof.ProofOps...)
	proof.ProofHash = ComputeUniversalStoreProofHash(proof)
	return proof, proof.Validate()
}

func NewUniversalRootStep(step UniversalRootStep) (UniversalRootStep, error) {
	step.StepHash = ComputeUniversalRootStepHash(step)
	return step, step.Validate()
}

func NewUniversalShardCommitment(commitment UniversalShardCommitment) (UniversalShardCommitment, error) {
	commitment.CommitmentHash = ComputeUniversalShardCommitmentHash(commitment)
	return commitment, commitment.Validate()
}

func NewUniversalMessageCommitment(commitment UniversalMessageCommitment) (UniversalMessageCommitment, error) {
	commitment.MessageCommitmentHash = ComputeUniversalMessageCommitmentHash(commitment)
	commitment.ReceiptCommitmentHash = ComputeUniversalReceiptCommitmentHash(commitment)
	commitment.DeliveryCommitmentHash = ComputeUniversalDeliveryCommitmentHash(commitment)
	return commitment, commitment.Validate()
}

func NewUniversalProofEnvelope(proof UniversalProofEnvelope) (UniversalProofEnvelope, error) {
	if proof.ProofVersion == 0 {
		proof.ProofVersion = UniversalProofVersionV1
	}
	proof.VerificationPath = append([]UniversalRootStep(nil), proof.VerificationPath...)
	proof.Value = append([]byte(nil), proof.Value...)
	proof.AbsenceMarker = append([]byte(nil), proof.AbsenceMarker...)
	proof.Key = append([]byte(nil), proof.Key...)
	proof.ProofHash = ComputeUniversalProofEnvelopeHash(proof)
	return proof, proof.ValidateFormat()
}

func SupportedUniversalProofTypes() []UniversalProofType {
	return []UniversalProofType{
		ProofTypeAccountState,
		ProofTypeBalance,
		ProofTypeZoneRoot,
		ProofTypeShardRoot,
		ProofTypeMessageInclusion,
		ProofTypeMessageReceipt,
		ProofTypeDomainOwnership,
		ProofTypeResolverRecord,
		ProofTypeContractState,
		ProofTypePaymentSettlement,
		ProofTypeNonExistence,
	}
}

func UniversalProofRequirementForType(proofType UniversalProofType) (universalProofRequirement, bool) {
	switch proofType {
	case ProofTypeAccountState:
		return universalProofRequirement{RootType: AccountProofRootType}, true
	case ProofTypeBalance:
		return universalProofRequirement{RootType: BalanceProofRootType, ZoneID: ZoneIDFinancial, RequiresZone: true, RequiresShard: true}, true
	case ProofTypeZoneRoot:
		return universalProofRequirement{RootType: ZoneStateProofRootType, RequiresZone: true}, true
	case ProofTypeShardRoot:
		return universalProofRequirement{RootType: ShardStateProofRootType, RequiresZone: true, RequiresShard: true}, true
	case ProofTypeMessageInclusion:
		return universalProofRequirement{RootType: MessageProofRootType, RequiresMessage: true}, true
	case ProofTypeMessageReceipt:
		return universalProofRequirement{RootType: ReceiptProofRootType, RequiresMessage: true, RequiresReceipt: true}, true
	case ProofTypeDomainOwnership:
		return universalProofRequirement{RootType: DomainOwnershipProofRootType, ZoneID: ZoneIDIdentity, RequiresZone: true, RequiresShard: true}, true
	case ProofTypeResolverRecord:
		return universalProofRequirement{RootType: ResolverProofRootType, ZoneID: ZoneIDIdentity, RequiresZone: true, RequiresShard: true}, true
	case ProofTypeContractState:
		return universalProofRequirement{RootType: ContractStateProofRootType, ZoneID: ZoneIDContract, RequiresZone: true, RequiresShard: true}, true
	case ProofTypePaymentSettlement:
		return universalProofRequirement{RootType: PaymentSettlementProofRootType, ZoneID: ZoneIDFinancial, RequiresZone: true, RequiresShard: true}, true
	case ProofTypeNonExistence:
		return universalProofRequirement{RequiresAbsence: true}, true
	default:
		return universalProofRequirement{}, false
	}
}

func VerifyUniversalProof(proof UniversalProofEnvelope, trusted UniversalTrustedHeader) UniversalProofVerificationResult {
	if !trusted.Trusted {
		return universalProofFailure(proof, ProofFailureUntrustedHeader, "trusted header is not marked trusted")
	}
	if trusted.Height == 0 || proof.Height == 0 || trusted.Height != proof.Height {
		return universalProofFailure(proof, ProofFailureHeightUnavailable, "proof height is not available in trusted header")
	}
	if trusted.ChainID != proof.ChainID {
		return universalProofFailure(proof, ProofFailureChainIDMismatch, "proof chain_id does not match trusted header")
	}
	if err := trusted.Validate(); err != nil {
		return universalProofFailure(proof, ProofFailureUntrustedHeader, err.Error())
	}
	if proof.AppHash != trusted.AppHash {
		return universalProofFailure(proof, ProofFailureRootMismatch, "proof app_hash does not match trusted header")
	}
	if err := proof.ValidateFormat(); err != nil {
		return universalProofFailure(proof, failureCodeForProofError(proof, err), err.Error())
	}
	if err := proof.validateCompatibility(); err != nil {
		return universalProofFailure(proof, ProofFailureRootMismatch, err.Error())
	}
	if err := verifyUniversalRootPath(proof); err != nil {
		return universalProofFailure(proof, ProofFailureRootMismatch, err.Error())
	}
	if err := verifyUniversalZoneScope(proof); err != nil {
		return universalProofFailure(proof, ProofFailureZoneNotFound, err.Error())
	}
	if err := verifyUniversalShardScope(proof); err != nil {
		return universalProofFailure(proof, ProofFailureShardNotFound, err.Error())
	}
	if err := verifyUniversalStoreProof(proof); err != nil {
		if proof.ProofType == ProofTypeNonExistence {
			return universalProofFailure(proof, ProofFailureNonExistenceProofInvalid, err.Error())
		}
		return universalProofFailure(proof, ProofFailureStoreProofInvalid, err.Error())
	}
	if err := verifyUniversalMessageScope(proof); err != nil {
		if proof.ProofType == ProofTypeMessageReceipt {
			return universalProofFailure(proof, ProofFailureReceiptNotFound, err.Error())
		}
		return universalProofFailure(proof, ProofFailureMessageNotIncluded, err.Error())
	}
	if err := verifyUniversalObjectRules(proof); err != nil {
		return universalProofFailure(proof, failureCodeForProofError(proof, err), err.Error())
	}
	return UniversalProofVerificationResult{
		Verified:	true,
		VerifiedAbsent:	proof.ProofType == ProofTypeNonExistence,
		Value:		append([]byte(nil), proof.Value...),
		FailureCode:	ProofFailureNone,
		ProofHash:	proof.ProofHash,
		FailureMessage:	"",
	}
}

func (h UniversalTrustedHeader) Validate() error {
	if h.ChainID == "" {
		return errors.New("aetracore universal proof trusted header chain_id is required")
	}
	if h.Height == 0 {
		return errors.New("aetracore universal proof trusted header height must be positive")
	}
	if err := ValidateHash("aetracore universal proof trusted header app hash", h.AppHash); err != nil {
		return err
	}
	if h.HeaderHash != "" {
		return ValidateHash("aetracore universal proof trusted header hash", h.HeaderHash)
	}
	return nil
}

func (p UniversalStoreProof) Validate() error {
	if p.ProofVersion != UniversalProofVersionV1 {
		return fmt.Errorf("aetracore universal store proof version %d is not supported", p.ProofVersion)
	}
	if len(p.Key) == 0 {
		return errors.New("aetracore universal store proof key is required")
	}
	if err := ValidateHash("aetracore universal store proof root", p.StoreRoot); err != nil {
		return err
	}
	for _, op := range p.ProofOps {
		if err := ValidateHash("aetracore universal store proof op", op); err != nil {
			return err
		}
	}
	if err := ValidateHash("aetracore universal store proof hash", p.ProofHash); err != nil {
		return err
	}
	if p.ProofHash != ComputeUniversalStoreProofHash(p) {
		return errors.New("aetracore universal store proof hash mismatch")
	}
	return nil
}

func (s UniversalRootStep) Validate() error {
	if err := ValidateHash("aetracore universal proof path parent root", s.FromRoot); err != nil {
		return err
	}
	if err := ValidateHash("aetracore universal proof path child root", s.ToRoot); err != nil {
		return err
	}
	if err := validateToken("aetracore universal proof path parent root type", string(s.FromRootType), MaxScopeLength); err != nil {
		return err
	}
	if err := validateToken("aetracore universal proof path child root type", string(s.ToRootType), MaxScopeLength); err != nil {
		return err
	}
	if s.Scope != "" {
		if err := validateToken("aetracore universal proof path scope", s.Scope, MaxScopeLength); err != nil {
			return err
		}
	}
	if err := ValidateHash("aetracore universal proof path step hash", s.StepHash); err != nil {
		return err
	}
	if s.StepHash != ComputeUniversalRootStepHash(s) {
		return errors.New("aetracore universal proof path step hash mismatch")
	}
	return nil
}

func (c UniversalShardCommitment) Validate() error {
	if c.Height == 0 {
		return errors.New("aetracore universal shard commitment height must be positive")
	}
	if err := ValidateZoneID(c.ZoneID); err != nil {
		return err
	}
	if err := ValidateShardID(c.ShardID); err != nil {
		return err
	}
	if err := ValidateHash("aetracore universal shard commitment shard root", c.ShardRoot); err != nil {
		return err
	}
	if err := ValidateHash("aetracore universal shard commitment aggregate root", c.ShardRootsRoot); err != nil {
		return err
	}
	if err := ValidateHash("aetracore universal shard commitment hash", c.CommitmentHash); err != nil {
		return err
	}
	if c.CommitmentHash != ComputeUniversalShardCommitmentHash(c) {
		return errors.New("aetracore universal shard commitment hash mismatch")
	}
	return nil
}

func (c UniversalMessageCommitment) Validate() error {
	if c.Height == 0 {
		return errors.New("aetracore universal message commitment height must be positive")
	}
	for _, field := range []struct {
		name	string
		value	string
	}{
		{"aetracore universal message commitment message id", c.MessageID},
		{"aetracore universal message root", c.MessageRoot},
		{"aetracore universal source outbox root", c.SourceOutboxRoot},
		{"aetracore universal destination inbox root", c.DestinationInboxRoot},
		{"aetracore universal receipt root", c.ReceiptRoot},
		{"aetracore universal receipt hash", c.ReceiptHash},
		{"aetracore universal message commitment hash", c.MessageCommitmentHash},
		{"aetracore universal receipt commitment hash", c.ReceiptCommitmentHash},
		{"aetracore universal delivery commitment hash", c.DeliveryCommitmentHash},
	} {
		if err := ValidateHash(field.name, field.value); err != nil {
			return err
		}
	}
	if c.MessageCommitmentHash != ComputeUniversalMessageCommitmentHash(c) {
		return errors.New("aetracore universal message commitment hash mismatch")
	}
	if c.ReceiptCommitmentHash != ComputeUniversalReceiptCommitmentHash(c) {
		return errors.New("aetracore universal receipt commitment hash mismatch")
	}
	if c.DeliveryCommitmentHash != ComputeUniversalDeliveryCommitmentHash(c) {
		return errors.New("aetracore universal delivery commitment hash mismatch")
	}
	return nil
}

func (p UniversalProofEnvelope) ValidateFormat() error {
	if p.ProofVersion != UniversalProofVersionV1 {
		return fmt.Errorf("aetracore universal proof version %d is not supported", p.ProofVersion)
	}
	if p.ChainID == "" {
		return errors.New("aetracore universal proof chain_id is required")
	}
	if p.Height == 0 {
		return errors.New("aetracore universal proof height must be positive")
	}
	if err := ValidateHash("aetracore universal proof app hash", p.AppHash); err != nil {
		return err
	}
	if err := validateToken("aetracore universal proof type", string(p.ProofType), MaxScopeLength); err != nil {
		return err
	}
	if err := validateToken("aetracore universal proof root type", string(p.RootType), MaxScopeLength); err != nil {
		return err
	}
	if len(p.Key) == 0 {
		return errors.New("aetracore universal proof key is required")
	}
	if len(p.VerificationPath) == 0 {
		return errors.New("aetracore universal proof verification path is required")
	}
	if p.ZoneID != "" {
		if err := ValidateZoneID(p.ZoneID); err != nil {
			return err
		}
	}
	if p.ShardID != "" {
		if err := ValidateShardID(p.ShardID); err != nil {
			return err
		}
	}
	if err := p.StoreProof.Validate(); err != nil {
		return err
	}
	for _, step := range p.VerificationPath {
		if err := step.Validate(); err != nil {
			return err
		}
	}
	if p.HasZoneCommit {
		if err := p.ZoneCommitment.ValidateHash(); err != nil {
			return err
		}
	}
	if p.HasShardCommit {
		if err := p.ShardCommitment.Validate(); err != nil {
			return err
		}
	}
	if p.HasMessageCommit {
		if err := p.MessageCommit.Validate(); err != nil {
			return err
		}
	}
	if err := ValidateHash("aetracore universal proof envelope hash", p.ProofHash); err != nil {
		return err
	}
	if p.ProofHash != ComputeUniversalProofEnvelopeHash(p) {
		return errors.New("aetracore universal proof envelope hash mismatch")
	}
	return nil
}

func (p UniversalProofEnvelope) validateCompatibility() error {
	req, found := UniversalProofRequirementForType(p.ProofType)
	if !found {
		return fmt.Errorf("aetracore universal proof type %q is not supported", p.ProofType)
	}
	if req.RootType != "" && p.RootType != req.RootType {
		return fmt.Errorf("aetracore proof type %s requires root type %s", p.ProofType, req.RootType)
	}
	if req.ZoneID != "" && p.ZoneID != req.ZoneID {
		return fmt.Errorf("aetracore proof type %s requires zone %s", p.ProofType, req.ZoneID)
	}
	if req.RequiresZone && (!p.HasZoneCommit || p.ZoneID == "") {
		return fmt.Errorf("aetracore proof type %s requires zone commitment", p.ProofType)
	}
	if req.RequiresShard && (!p.HasShardCommit || p.ShardID == "") {
		return fmt.Errorf("aetracore proof type %s requires shard commitment", p.ProofType)
	}
	if req.RequiresMessage && !p.HasMessageCommit {
		return fmt.Errorf("aetracore proof type %s requires message commitment", p.ProofType)
	}
	if req.RequiresAbsence && len(p.AbsenceMarker) == 0 {
		return fmt.Errorf("aetracore proof type %s requires non-existence marker", p.ProofType)
	}
	return nil
}

func ComputeUniversalStoreProofHash(proof UniversalStoreProof) string {
	return hashRoot("aetra-next-universal-store-proof-v1", func(w byteWriter) {
		writeUint64(w, proof.ProofVersion)
		writeBytes(w, proof.Key)
		writeBytes(w, proof.Value)
		writeBytes(w, proof.NonExistenceMarker)
		writePart(w, proof.StoreRoot)
		writeUint64(w, uint64(len(proof.ProofOps)))
		for _, op := range proof.ProofOps {
			writePart(w, op)
		}
	})
}

func ComputeUniversalRootStepHash(step UniversalRootStep) string {
	return hashRoot("aetra-next-universal-root-step-v1", func(w byteWriter) {
		writeUint64(w, uint64(step.Index))
		writePart(w, string(step.FromRootType))
		writePart(w, step.FromRoot)
		writePart(w, string(step.ToRootType))
		writePart(w, step.ToRoot)
		writePart(w, step.Scope)
	})
}

func ComputeUniversalShardCommitmentHash(commitment UniversalShardCommitment) string {
	return hashRoot("aetra-next-universal-shard-commitment-v1", func(w byteWriter) {
		writeUint64(w, commitment.Height)
		writePart(w, string(commitment.ZoneID))
		writePart(w, string(commitment.ShardID))
		writePart(w, commitment.ShardRoot)
		writePart(w, commitment.ShardRootsRoot)
	})
}

func ComputeUniversalMessageCommitmentHash(commitment UniversalMessageCommitment) string {
	return hashRoot("aetra-next-universal-message-commitment-v1", func(w byteWriter) {
		writeUint64(w, commitment.Height)
		writePart(w, commitment.MessageID)
		writePart(w, commitment.MessageRoot)
		writePart(w, commitment.SourceOutboxRoot)
		writePart(w, commitment.DestinationInboxRoot)
	})
}

func ComputeUniversalReceiptCommitmentHash(commitment UniversalMessageCommitment) string {
	return hashRoot("aetra-next-universal-receipt-commitment-v1", func(w byteWriter) {
		writeUint64(w, commitment.Height)
		writePart(w, commitment.MessageID)
		writePart(w, commitment.ReceiptRoot)
		writePart(w, commitment.ReceiptHash)
	})
}

func ComputeUniversalDeliveryCommitmentHash(commitment UniversalMessageCommitment) string {
	return hashRoot("aetra-next-universal-delivery-commitment-v1", func(w byteWriter) {
		writeUint64(w, commitment.Height)
		writePart(w, commitment.MessageID)
		writePart(w, commitment.MessageCommitmentHash)
		writePart(w, commitment.ReceiptCommitmentHash)
	})
}

func ComputeUniversalProofEnvelopeHash(proof UniversalProofEnvelope) string {
	return hashRoot("aetra-next-universal-proof-envelope-v1", func(w byteWriter) {
		writePart(w, string(proof.ProofType))
		writeUint64(w, proof.ProofVersion)
		writePart(w, proof.ChainID)
		writeUint64(w, proof.Height)
		writePart(w, proof.AppHash)
		writePart(w, string(proof.RootType))
		writePart(w, string(proof.ZoneID))
		writePart(w, string(proof.ShardID))
		writeBytes(w, proof.Key)
		writeBytes(w, proof.Value)
		writeBytes(w, proof.AbsenceMarker)
		writeUint64(w, proof.ObjectExpiryHeight)
		writePart(w, proof.StoreProof.ProofHash)
		writeBool(w, proof.HasZoneCommit)
		if proof.HasZoneCommit {
			writePart(w, proof.ZoneCommitment.CommitmentHash)
		}
		writeBool(w, proof.HasShardCommit)
		if proof.HasShardCommit {
			writePart(w, proof.ShardCommitment.CommitmentHash)
		}
		writeBool(w, proof.HasMessageCommit)
		if proof.HasMessageCommit {
			writePart(w, proof.MessageCommit.DeliveryCommitmentHash)
		}
		writeUint64(w, uint64(len(proof.VerificationPath)))
		for _, step := range proof.VerificationPath {
			writePart(w, step.StepHash)
		}
	})
}

func verifyUniversalRootPath(proof UniversalProofEnvelope) error {
	current := proof.AppHash
	for i, step := range proof.VerificationPath {
		if uint32(i) != step.Index {
			return errors.New("aetracore universal proof path index mismatch")
		}
		if step.FromRoot != current {
			return errors.New("aetracore universal proof path is not contiguous")
		}
		if err := step.Validate(); err != nil {
			return err
		}
		current = step.ToRoot
	}
	if current != proof.StoreProof.StoreRoot {
		return errors.New("aetracore universal proof path does not end at store root")
	}
	last := proof.VerificationPath[len(proof.VerificationPath)-1]
	if last.ToRootType != proof.RootType {
		return fmt.Errorf("aetracore universal proof path ends at root type %s, expected %s", last.ToRootType, proof.RootType)
	}
	return nil
}

func verifyUniversalZoneScope(proof UniversalProofEnvelope) error {
	if proof.ZoneID == "" && !proof.HasZoneCommit {
		return nil
	}
	if !proof.HasZoneCommit {
		return errors.New("aetracore universal proof missing zone commitment")
	}
	if proof.ZoneCommitment.Height != proof.Height {
		return errors.New("aetracore universal proof zone commitment height mismatch")
	}
	if proof.ZoneCommitment.ZoneID != proof.ZoneID {
		return errors.New("aetracore universal proof zone commitment zone mismatch")
	}
	if proof.ProofType == ProofTypeZoneRoot && proof.ZoneCommitment.StateRoot != proof.StoreProof.StoreRoot {
		return errors.New("aetracore universal proof zone state root mismatch")
	}
	return nil
}

func verifyUniversalShardScope(proof UniversalProofEnvelope) error {
	if proof.ShardID == "" && !proof.HasShardCommit {
		return nil
	}
	if !proof.HasShardCommit {
		return errors.New("aetracore universal proof missing shard commitment")
	}
	if proof.ShardCommitment.Height != proof.Height {
		return errors.New("aetracore universal proof shard commitment height mismatch")
	}
	if proof.ShardCommitment.ZoneID != proof.ZoneID || proof.ShardCommitment.ShardID != proof.ShardID {
		return errors.New("aetracore universal proof shard commitment scope mismatch")
	}
	if proof.HasZoneCommit && proof.ShardCommitment.ShardRootsRoot != proof.ZoneCommitment.ShardRootsRoot {
		return errors.New("aetracore universal proof shard aggregate root mismatch")
	}
	if proof.ProofType == ProofTypeShardRoot && proof.ShardCommitment.ShardRoot != proof.StoreProof.StoreRoot {
		return errors.New("aetracore universal proof shard state root mismatch")
	}
	return nil
}

func verifyUniversalStoreProof(proof UniversalProofEnvelope) error {
	if string(proof.Key) != string(proof.StoreProof.Key) {
		return errors.New("aetracore universal proof store key mismatch")
	}
	if proof.ProofType == ProofTypeNonExistence {
		if len(proof.StoreProof.NonExistenceMarker) == 0 || len(proof.AbsenceMarker) == 0 {
			return errors.New("aetracore universal non-existence marker is required")
		}
		if string(proof.StoreProof.NonExistenceMarker) != string(proof.AbsenceMarker) {
			return errors.New("aetracore universal non-existence marker mismatch")
		}
		if len(proof.StoreProof.Value) != 0 || len(proof.Value) != 0 {
			return errors.New("aetracore universal non-existence proof must not carry value")
		}
		return nil
	}
	if len(proof.Value) == 0 || len(proof.StoreProof.Value) == 0 {
		return errors.New("aetracore universal existence proof value is required")
	}
	if string(proof.Value) != string(proof.StoreProof.Value) {
		return errors.New("aetracore universal proof value mismatch")
	}
	return nil
}

func verifyUniversalMessageScope(proof UniversalProofEnvelope) error {
	if !proof.HasMessageCommit {
		return nil
	}
	if proof.MessageCommit.Height != proof.Height {
		return errors.New("aetracore universal proof message commitment height mismatch")
	}
	switch proof.ProofType {
	case ProofTypeMessageInclusion:
		if proof.MessageCommit.MessageRoot != proof.StoreProof.StoreRoot {
			return errors.New("aetracore universal proof message root mismatch")
		}
	case ProofTypeMessageReceipt:
		if proof.MessageCommit.ReceiptRoot != proof.StoreProof.StoreRoot {
			return errors.New("aetracore universal proof receipt root mismatch")
		}
		if proof.MessageCommit.ReceiptHash == EmptyRootHash {
			return errors.New("aetracore universal proof receipt not found")
		}
	}
	return nil
}

func verifyUniversalObjectRules(proof UniversalProofEnvelope) error {
	req, _ := UniversalProofRequirementForType(proof.ProofType)
	if req.RequiresAbsence {
		return nil
	}
	if proof.Height == 0 {
		return errors.New("aetracore universal proof object height is unavailable")
	}
	if proof.ObjectExpiryHeight != 0 && proof.Height > proof.ObjectExpiryHeight {
		return errors.New("aetracore universal proof object expired at proof height")
	}
	return nil
}

func failureCodeForProofError(proof UniversalProofEnvelope, err error) UniversalProofFailureCode {
	if err == nil {
		return ProofFailureNone
	}
	if proof.ProofType == ProofTypeNonExistence {
		return ProofFailureNonExistenceProofInvalid
	}
	if proof.ProofType == ProofTypeMessageReceipt {
		return ProofFailureReceiptNotFound
	}
	if proof.ProofType == ProofTypeMessageInclusion {
		return ProofFailureMessageNotIncluded
	}
	if proof.ObjectExpiryHeight != 0 && proof.Height > proof.ObjectExpiryHeight {
		return ProofFailureObjectExpired
	}
	return ProofFailureRootMismatch
}

func universalProofFailure(proof UniversalProofEnvelope, code UniversalProofFailureCode, message string) UniversalProofVerificationResult {
	return UniversalProofVerificationResult{
		Verified:	false,
		VerifiedAbsent:	false,
		FailureCode:	code,
		FailureMessage:	message,
		ProofHash:	proof.ProofHash,
	}
}

func writeBytes(w byteWriter, value []byte) {
	writeUint64(w, uint64(len(value)))
	_, _ = w.Write(value)
}

func writeBool(w byteWriter, value bool) {
	if value {
		writeUint64(w, 1)
		return
	}
	writeUint64(w, 0)
}
