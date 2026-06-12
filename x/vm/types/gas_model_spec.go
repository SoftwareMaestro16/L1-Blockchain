package types

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"

	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

const (
	AVMGasClassExecution			AVMGasClass	= "execution"
	AVMGasClassStorage			AVMGasClass	= "storage"
	AVMGasClassScheduling			AVMGasClass	= "scheduling"
	AVMGasClassCrossZoneRouting		AVMGasClass	= "cross_zone_routing"
	AVMGasClassProofVerification		AVMGasClass	= "proof_verification"
	AVMGasClassContinuation			AVMGasClass	= "continuation"
	AVMGasClassInterfaceIntrospection	AVMGasClass	= "interface_introspection"

	MaxAVMGasCharges	= 32
)

type AVMGasClass string

type AVMGasClassBudget struct {
	Class	AVMGasClass
	Limit	uint64
}

type AVMGasCharge struct {
	Class	AVMGasClass
	Amount	uint64
}

type AVMGasSchedule struct {
	ClassBudgets			[]AVMGasClassBudget
	SchedulingGas			uint64
	CrossZoneRoutingGas		uint64
	ProofVerificationGas		uint64
	ContinuationGas			uint64
	InterfaceIntrospectionGas	uint64
	RetryGas			uint64
	BounceGas			uint64
	RefundUnused			bool
	MaxRefundGas			uint64
	ScheduleHash			string
}

type AVMGasPolicy struct {
	BaseMessageGas		uint64
	PerBytePayloadGas	uint64
	StorageReadGas		uint64
	StorageWriteGas		uint64
	QueueInsertGas		uint64
	QueuePopGas		uint64
	CrossZoneBaseGas	uint64
	ProofVerifyBaseGas	uint64
	ContinuationStoreGas	uint64
	BounceBaseGas		uint64
	PolicyHash		string
}

type AVMAsyncGasReserve struct {
	MessageID	string
	ZoneID		zonestypes.ZoneID
	ReservedGas	uint64
	EscrowedGas	uint64
	RemainingGas	uint64
	Consumed	[]AVMGasCharge
	RefundGas	uint64
	Refundable	bool
	ReserveHash	string
}

type AVMZoneAsyncGasMeter struct {
	ZoneID		zonestypes.ZoneID
	Budget		zonestypes.ZoneExecutionBudget
	Consumed	[]AVMGasCharge
	MeterHash	string
}

func DefaultAVMGasSchedule() (AVMGasSchedule, error) {
	policy, err := DefaultAVMGasPolicy()
	if err != nil {
		return AVMGasSchedule{}, err
	}
	schedule, err := AVMGasScheduleFromPolicy(policy, true, 1_000_000)
	if err != nil {
		return AVMGasSchedule{}, err
	}
	return schedule, nil
}

func DefaultAVMGasPolicy() (AVMGasPolicy, error) {
	policy := AVMGasPolicy{
		BaseMessageGas:		10,
		PerBytePayloadGas:	1,
		StorageReadGas:		2,
		StorageWriteGas:	4,
		QueueInsertGas:		10,
		QueuePopGas:		10,
		CrossZoneBaseGas:	20,
		ProofVerifyBaseGas:	15,
		ContinuationStoreGas:	30,
		BounceBaseGas:		25,
	}
	policy.PolicyHash = ComputeAVMGasPolicyHash(policy)
	return policy, policy.Validate()
}

func AVMGasScheduleFromPolicy(policy AVMGasPolicy, refundUnused bool, maxRefundGas uint64) (AVMGasSchedule, error) {
	policy = canonicalAVMGasPolicy(policy)
	if err := policy.Validate(); err != nil {
		return AVMGasSchedule{}, err
	}
	schedule := AVMGasSchedule{
		ClassBudgets: []AVMGasClassBudget{
			{Class: AVMGasClassExecution, Limit: 1_000_000},
			{Class: AVMGasClassStorage, Limit: 250_000},
			{Class: AVMGasClassScheduling, Limit: 50_000},
			{Class: AVMGasClassCrossZoneRouting, Limit: 100_000},
			{Class: AVMGasClassProofVerification, Limit: 200_000},
			{Class: AVMGasClassContinuation, Limit: 150_000},
			{Class: AVMGasClassInterfaceIntrospection, Limit: 25_000},
		},
		SchedulingGas:			policy.QueueInsertGas,
		CrossZoneRoutingGas:		policy.CrossZoneBaseGas,
		ProofVerificationGas:		policy.ProofVerifyBaseGas,
		ContinuationGas:		policy.ContinuationStoreGas,
		InterfaceIntrospectionGas:	5,
		RetryGas:			policy.QueuePopGas,
		BounceGas:			policy.BounceBaseGas,
		RefundUnused:			refundUnused,
		MaxRefundGas:			maxRefundGas,
	}
	schedule = canonicalAVMGasSchedule(schedule)
	schedule.ScheduleHash = ComputeAVMGasScheduleHash(schedule)
	return schedule, schedule.Validate()
}

func NewAVMAsyncGasReserve(msg AVMAsyncMessage, schedule AVMGasSchedule) (AVMAsyncGasReserve, error) {
	msg = canonicalAVMAsyncMessage(msg)
	if err := msg.Validate(); err != nil {
		return AVMAsyncGasReserve{}, err
	}
	schedule = canonicalAVMGasSchedule(schedule)
	if err := schedule.Validate(); err != nil {
		return AVMAsyncGasReserve{}, err
	}
	required, charges, err := RequiredAVMAsyncGas(msg, schedule)
	if err != nil {
		return AVMAsyncGasReserve{}, err
	}
	reserve := AVMAsyncGasReserve{
		MessageID:	msg.ID,
		ZoneID:		msg.DestinationZone,
		ReservedGas:	required,
		EscrowedGas:	required,
		RemainingGas:	required,
		Consumed:	charges,
		Refundable:	schedule.RefundUnused,
	}
	reserve = canonicalAVMAsyncGasReserve(reserve)
	reserve.ReserveHash = ComputeAVMAsyncGasReserveHash(reserve)
	return reserve, reserve.Validate()
}

func NewAVMAsyncGasReserveWithPolicy(msg AVMAsyncMessage, policy AVMGasPolicy, refundUnused bool, maxRefundGas uint64) (AVMAsyncGasReserve, error) {
	schedule, err := AVMGasScheduleFromPolicy(policy, refundUnused, maxRefundGas)
	if err != nil {
		return AVMAsyncGasReserve{}, err
	}
	return NewAVMAsyncGasReserve(msg, schedule)
}

func RequiredAVMAsyncGas(msg AVMAsyncMessage, schedule AVMGasSchedule) (uint64, []AVMGasCharge, error) {
	if err := schedule.Validate(); err != nil {
		return 0, nil, err
	}
	charges := []AVMGasCharge{
		{Class: AVMGasClassExecution, Amount: msg.GasLimit},
		{Class: AVMGasClassScheduling, Amount: schedule.SchedulingGas},
	}
	if msg.SourceZone != msg.DestinationZone {
		charges = append(charges, AVMGasCharge{Class: AVMGasClassCrossZoneRouting, Amount: schedule.CrossZoneRoutingGas})
	}
	if msg.AuthProofOptional != "" {
		charges = append(charges, AVMGasCharge{Class: AVMGasClassProofVerification, Amount: schedule.ProofVerificationGas})
	}
	if msg.StateProofOptional != "" {
		charges = append(charges, AVMGasCharge{Class: AVMGasClassProofVerification, Amount: schedule.ProofVerificationGas})
	}
	if msg.DestinationActorOptional != "" {
		charges = append(charges, AVMGasCharge{Class: AVMGasClassContinuation, Amount: schedule.ContinuationGas})
	}
	return sumAVMGasCharges(charges)
}

func AVMMessageAdmissionGas(msg AVMAsyncMessage, policy AVMGasPolicy) (uint64, error) {
	msg = canonicalAVMAsyncMessage(msg)
	if err := msg.Validate(); err != nil {
		return 0, err
	}
	policy = canonicalAVMGasPolicy(policy)
	if err := policy.Validate(); err != nil {
		return 0, err
	}
	payloadGas, err := checkedAVMGasMul(policy.PerBytePayloadGas, uint64(len(msg.Payload)))
	if err != nil {
		return 0, err
	}
	total, err := checkedAVMGasAdd(policy.BaseMessageGas, payloadGas)
	if err != nil {
		return 0, err
	}
	for _, amount := range []uint64{msg.GasLimit, policy.QueueInsertGas} {
		total, err = checkedAVMGasAdd(total, amount)
		if err != nil {
			return 0, err
		}
	}
	if msg.SourceZone != msg.DestinationZone {
		total, err = checkedAVMGasAdd(total, policy.CrossZoneBaseGas)
		if err != nil {
			return 0, err
		}
	}
	proofs := uint32(0)
	if msg.AuthProofOptional != "" {
		proofs++
	}
	if msg.StateProofOptional != "" {
		proofs++
	}
	if proofs > 0 {
		proofGas, err := AVMProofVerificationGas(proofs, policy)
		if err != nil {
			return 0, err
		}
		total, err = checkedAVMGasAdd(total, proofGas)
		if err != nil {
			return 0, err
		}
	}
	if msg.DestinationActorOptional != "" {
		total, err = checkedAVMGasAdd(total, policy.ContinuationStoreGas)
		if err != nil {
			return 0, err
		}
	}
	return total, nil
}

func AVMStorageReadGas(bytes uint64, policy AVMGasPolicy) (uint64, error) {
	return avmStorageByteGas("AVM storage read", bytes, policy, policy.StorageReadGas)
}

func AVMStorageWriteGas(bytes uint64, policy AVMGasPolicy) (uint64, error) {
	return avmStorageByteGas("AVM storage write", bytes, policy, policy.StorageWriteGas)
}

func AVMProofVerificationGas(proofCount uint32, policy AVMGasPolicy) (uint64, error) {
	policy = canonicalAVMGasPolicy(policy)
	if err := policy.Validate(); err != nil {
		return 0, err
	}
	if proofCount == 0 {
		return 0, errors.New("AVM proof verification count must be positive")
	}
	return checkedAVMGasMul(policy.ProofVerifyBaseGas, uint64(proofCount))
}

func NewAVMZoneAsyncGasMeter(zoneID zonestypes.ZoneID, budget zonestypes.ZoneExecutionBudget) (AVMZoneAsyncGasMeter, error) {
	if err := zonestypes.ValidateZoneID(zoneID); err != nil {
		return AVMZoneAsyncGasMeter{}, err
	}
	if budget.MaxGas == 0 && budget.MaxMessages == 0 {
		budget = zonestypes.DefaultZoneExecutionBudget()
	}
	if err := budget.Validate(); err != nil {
		return AVMZoneAsyncGasMeter{}, err
	}
	meter := AVMZoneAsyncGasMeter{
		ZoneID:	zoneID,
		Budget:	budget,
	}
	meter.MeterHash = ComputeAVMZoneAsyncGasMeterHash(meter)
	return meter, meter.Validate()
}

func ConsumeAVMZoneAsyncGas(meter AVMZoneAsyncGasMeter, class AVMGasClass, gas uint64, messages uint32) (AVMZoneAsyncGasMeter, error) {
	meter = canonicalAVMZoneAsyncGasMeter(meter)
	if err := meter.Validate(); err != nil {
		return AVMZoneAsyncGasMeter{}, err
	}
	if !IsAVMGasClass(class) {
		return AVMZoneAsyncGasMeter{}, fmt.Errorf("invalid AVM gas class %q", class)
	}
	if gas == 0 {
		return AVMZoneAsyncGasMeter{}, errors.New("AVM zone async gas consumption must be positive")
	}
	nextBudget, err := meter.Budget.Consume(gas, messages)
	if err != nil {
		return AVMZoneAsyncGasMeter{}, err
	}
	meter.Budget = nextBudget
	meter.Consumed = append(meter.Consumed, AVMGasCharge{Class: class, Amount: gas})
	meter = canonicalAVMZoneAsyncGasMeter(meter)
	meter.MeterHash = ComputeAVMZoneAsyncGasMeterHash(meter)
	return meter, meter.Validate()
}

func ValidateAVMExecutionGasLimit(msg AVMAsyncMessage, gasUsed uint64) error {
	msg = canonicalAVMAsyncMessage(msg)
	if err := msg.Validate(); err != nil {
		return err
	}
	if gasUsed > msg.GasLimit {
		return errors.New("AVM execution gas used exceeds message gas limit")
	}
	return nil
}

func ValidateAVMContractEmissionGasReserve(msg AVMAsyncMessage, reserve AVMAsyncGasReserve) error {
	msg = canonicalAVMAsyncMessage(msg)
	if err := msg.Validate(); err != nil {
		return err
	}
	reserve = canonicalAVMAsyncGasReserve(reserve)
	if err := reserve.Validate(); err != nil {
		return err
	}
	if reserve.MessageID != msg.ID {
		return errors.New("AVM contract emitted message gas reserve id mismatch")
	}
	if reserve.ZoneID != msg.DestinationZone {
		return errors.New("AVM contract emitted message gas reserve zone mismatch")
	}
	if reserveChargeTotalForClass(reserve, AVMGasClassExecution) < msg.GasLimit {
		return errors.New("AVM contract emitted message lacks execution gas reserve")
	}
	if reserveChargeTotalForClass(reserve, AVMGasClassScheduling) == 0 {
		return errors.New("AVM contract emitted message lacks scheduling gas reserve")
	}
	if msg.SourceZone != msg.DestinationZone && reserveChargeTotalForClass(reserve, AVMGasClassCrossZoneRouting) == 0 {
		return errors.New("AVM contract emitted message lacks routing gas reserve")
	}
	if proofCountForMessage(msg) > 0 && (msg.AuthProofOptional != "" || msg.StateProofOptional != "") && reserveChargeTotalForClass(reserve, AVMGasClassProofVerification) == 0 {
		return errors.New("AVM contract emitted message lacks proof verification gas reserve")
	}
	if msg.DestinationActorOptional != "" && reserveChargeTotalForClass(reserve, AVMGasClassContinuation) == 0 {
		return errors.New("AVM contract emitted message lacks continuation gas reserve")
	}
	return nil
}

func ValidateAVMProofGasMetering(proofCount uint32, charged uint64, policy AVMGasPolicy) error {
	required, err := AVMProofVerificationGas(proofCount, policy)
	if err != nil {
		return err
	}
	if charged < required {
		return errors.New("AVM proof verification gas is under-metered")
	}
	return nil
}

func ValidateAVMStorageWriteGas(bytes uint64, charged uint64, policy AVMGasPolicy) error {
	required, err := AVMStorageWriteGas(bytes, policy)
	if err != nil {
		return err
	}
	if charged < required {
		return errors.New("AVM storage write gas is under-metered")
	}
	return nil
}

func ValidateAVMFailedExecutionGas(receipt AVMExecutionReceipt, gasUsedBeforeFailure uint64) error {
	receipt = canonicalAVMExecutionReceipt(receipt)
	if err := receipt.Validate(); err != nil {
		return err
	}
	if receipt.Status != AVMReceiptStatusFailed {
		return errors.New("AVM failed execution gas invariant requires failed receipt")
	}
	if gasUsedBeforeFailure == 0 {
		return errors.New("AVM failed execution must consume gas used before failure")
	}
	if receipt.GasUsed != gasUsedBeforeFailure {
		return errors.New("AVM failed execution receipt gas drift")
	}
	return nil
}

func ConsumeAVMReservedGas(reserve AVMAsyncGasReserve, class AVMGasClass, amount uint64) (AVMAsyncGasReserve, error) {
	reserve = canonicalAVMAsyncGasReserve(reserve)
	if err := reserve.Validate(); err != nil {
		return AVMAsyncGasReserve{}, err
	}
	if !IsAVMGasClass(class) {
		return AVMAsyncGasReserve{}, fmt.Errorf("invalid AVM gas class %q", class)
	}
	if amount == 0 {
		return AVMAsyncGasReserve{}, errors.New("AVM gas consumption amount must be positive")
	}
	if amount > reserve.RemainingGas {
		return AVMAsyncGasReserve{}, errors.New("AVM gas consumption exceeds reserved gas")
	}
	reserve.RemainingGas -= amount
	reserve.Consumed = append(reserve.Consumed, AVMGasCharge{Class: class, Amount: amount})
	reserve = canonicalAVMAsyncGasReserve(reserve)
	reserve.ReserveHash = ComputeAVMAsyncGasReserveHash(reserve)
	return reserve, reserve.Validate()
}

func ChargeAVMRetryGas(reserve AVMAsyncGasReserve, policy AVMRetryPolicy, schedule AVMGasSchedule) (AVMAsyncGasReserve, error) {
	if err := schedule.Validate(); err != nil {
		return AVMAsyncGasReserve{}, err
	}
	if policy.Mode == AVMRetryModeNone || !policy.ChargeRetryGas {
		return reserve, reserve.Validate()
	}
	return ConsumeAVMReservedGas(reserve, AVMGasClassScheduling, schedule.RetryGas)
}

func ChargeAVMBounceGas(reserve AVMAsyncGasReserve, schedule AVMGasSchedule) (AVMAsyncGasReserve, error) {
	if err := schedule.Validate(); err != nil {
		return AVMAsyncGasReserve{}, err
	}
	return ConsumeAVMReservedGas(reserve, AVMGasClassCrossZoneRouting, schedule.BounceGas)
}

func FinalizeAVMAsyncGasRefund(reserve AVMAsyncGasReserve, schedule AVMGasSchedule) (AVMAsyncGasReserve, error) {
	reserve = canonicalAVMAsyncGasReserve(reserve)
	if err := reserve.Validate(); err != nil {
		return AVMAsyncGasReserve{}, err
	}
	if err := schedule.Validate(); err != nil {
		return AVMAsyncGasReserve{}, err
	}
	refund := uint64(0)
	if schedule.RefundUnused && reserve.Refundable {
		refund = reserve.RemainingGas
		if schedule.MaxRefundGas > 0 && refund > schedule.MaxRefundGas {
			refund = schedule.MaxRefundGas
		}
	}
	reserve.RefundGas = refund
	reserve.RemainingGas -= refund
	reserve = canonicalAVMAsyncGasReserve(reserve)
	reserve.ReserveHash = ComputeAVMAsyncGasReserveHash(reserve)
	return reserve, reserve.Validate()
}

func (s AVMGasSchedule) Validate() error {
	s = canonicalAVMGasSchedule(s)
	if len(s.ClassBudgets) != len(allAVMGasClasses()) {
		return errors.New("AVM gas schedule must define every gas class")
	}
	seen := make(map[AVMGasClass]struct{}, len(s.ClassBudgets))
	for i, budget := range s.ClassBudgets {
		if !IsAVMGasClass(budget.Class) {
			return fmt.Errorf("invalid AVM gas class %q", budget.Class)
		}
		if budget.Limit == 0 {
			return fmt.Errorf("AVM gas class %q limit must be positive", budget.Class)
		}
		if _, found := seen[budget.Class]; found {
			return fmt.Errorf("duplicate AVM gas class %q", budget.Class)
		}
		seen[budget.Class] = struct{}{}
		if i > 0 && s.ClassBudgets[i-1].Class >= budget.Class {
			return errors.New("AVM gas class budgets must be sorted canonically")
		}
	}
	for _, class := range allAVMGasClasses() {
		if _, found := seen[class]; !found {
			return fmt.Errorf("AVM gas schedule missing class %q", class)
		}
	}
	for _, item := range []struct {
		name	string
		value	uint64
	}{
		{name: "AVM scheduling gas", value: s.SchedulingGas},
		{name: "AVM cross-zone routing gas", value: s.CrossZoneRoutingGas},
		{name: "AVM proof verification gas", value: s.ProofVerificationGas},
		{name: "AVM continuation gas", value: s.ContinuationGas},
		{name: "AVM interface introspection gas", value: s.InterfaceIntrospectionGas},
		{name: "AVM retry gas", value: s.RetryGas},
		{name: "AVM bounce gas", value: s.BounceGas},
	} {
		if item.value == 0 {
			return fmt.Errorf("%s must be positive", item.name)
		}
	}
	if s.RefundUnused && s.MaxRefundGas == 0 {
		return errors.New("AVM gas refund policy must bound max refund gas")
	}
	if s.ScheduleHash == "" {
		return errors.New("AVM gas schedule hash is required")
	}
	if err := zonestypes.ValidateHash("AVM gas schedule hash", s.ScheduleHash); err != nil {
		return err
	}
	if s.ScheduleHash != ComputeAVMGasScheduleHash(s) {
		return errors.New("AVM gas schedule hash mismatch")
	}
	return nil
}

func (p AVMGasPolicy) Validate() error {
	p = canonicalAVMGasPolicy(p)
	for _, item := range []struct {
		name	string
		value	uint64
	}{
		{name: "AVM base message gas", value: p.BaseMessageGas},
		{name: "AVM per-byte payload gas", value: p.PerBytePayloadGas},
		{name: "AVM storage read gas", value: p.StorageReadGas},
		{name: "AVM storage write gas", value: p.StorageWriteGas},
		{name: "AVM queue insert gas", value: p.QueueInsertGas},
		{name: "AVM queue pop gas", value: p.QueuePopGas},
		{name: "AVM cross-zone base gas", value: p.CrossZoneBaseGas},
		{name: "AVM proof verify base gas", value: p.ProofVerifyBaseGas},
		{name: "AVM continuation store gas", value: p.ContinuationStoreGas},
		{name: "AVM bounce base gas", value: p.BounceBaseGas},
	} {
		if item.value == 0 {
			return fmt.Errorf("%s must be positive", item.name)
		}
	}
	if p.PolicyHash == "" {
		return errors.New("AVM gas policy hash is required")
	}
	if err := zonestypes.ValidateHash("AVM gas policy hash", p.PolicyHash); err != nil {
		return err
	}
	if p.PolicyHash != ComputeAVMGasPolicyHash(p) {
		return errors.New("AVM gas policy hash mismatch")
	}
	return nil
}

func (r AVMAsyncGasReserve) Validate() error {
	r = canonicalAVMAsyncGasReserve(r)
	if err := zonestypes.ValidateHash("AVM async gas reserve message id", r.MessageID); err != nil {
		return err
	}
	if err := zonestypes.ValidateZoneID(r.ZoneID); err != nil {
		return err
	}
	if r.ReservedGas == 0 {
		return errors.New("AVM async reserved gas must be positive")
	}
	if r.EscrowedGas != r.ReservedGas {
		return errors.New("AVM async gas reserve must be fully escrowed upfront")
	}
	if r.RemainingGas > r.EscrowedGas {
		return errors.New("AVM async remaining gas exceeds escrow")
	}
	if r.RefundGas > r.EscrowedGas {
		return errors.New("AVM async refund gas exceeds escrow")
	}
	if len(r.Consumed) > MaxAVMGasCharges {
		return fmt.Errorf("AVM async gas charges must be <= %d", MaxAVMGasCharges)
	}
	for i, charge := range r.Consumed {
		if err := charge.Validate(); err != nil {
			return err
		}
		if i > 0 && compareAVMGasCharges(r.Consumed[i-1], charge) > 0 {
			return errors.New("AVM async gas charges must be sorted canonically")
		}
	}
	if r.ReserveHash == "" {
		return errors.New("AVM async gas reserve hash is required")
	}
	if err := zonestypes.ValidateHash("AVM async gas reserve hash", r.ReserveHash); err != nil {
		return err
	}
	if r.ReserveHash != ComputeAVMAsyncGasReserveHash(r) {
		return errors.New("AVM async gas reserve hash mismatch")
	}
	return nil
}

func (m AVMZoneAsyncGasMeter) Validate() error {
	m = canonicalAVMZoneAsyncGasMeter(m)
	if err := zonestypes.ValidateZoneID(m.ZoneID); err != nil {
		return err
	}
	if err := m.Budget.Validate(); err != nil {
		return err
	}
	if len(m.Consumed) > MaxAVMGasCharges {
		return fmt.Errorf("AVM zone async gas charges must be <= %d", MaxAVMGasCharges)
	}
	for i, charge := range m.Consumed {
		if err := charge.Validate(); err != nil {
			return err
		}
		if i > 0 && compareAVMGasCharges(m.Consumed[i-1], charge) > 0 {
			return errors.New("AVM zone async gas charges must be sorted canonically")
		}
	}
	if m.MeterHash == "" {
		return errors.New("AVM zone async gas meter hash is required")
	}
	if err := zonestypes.ValidateHash("AVM zone async gas meter hash", m.MeterHash); err != nil {
		return err
	}
	if m.MeterHash != ComputeAVMZoneAsyncGasMeterHash(m) {
		return errors.New("AVM zone async gas meter hash mismatch")
	}
	return nil
}

func (c AVMGasCharge) Validate() error {
	if !IsAVMGasClass(c.Class) {
		return fmt.Errorf("invalid AVM gas charge class %q", c.Class)
	}
	if c.Amount == 0 {
		return errors.New("AVM gas charge amount must be positive")
	}
	return nil
}

func IsAVMGasClass(class AVMGasClass) bool {
	for _, candidate := range allAVMGasClasses() {
		if class == candidate {
			return true
		}
	}
	return false
}

func ComputeAVMGasScheduleHash(schedule AVMGasSchedule) string {
	schedule = canonicalAVMGasSchedule(schedule)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-gas-schedule-v1")
	writeEngineUint64(h, uint64(len(schedule.ClassBudgets)))
	for _, budget := range schedule.ClassBudgets {
		writeEnginePart(h, string(budget.Class))
		writeEngineUint64(h, budget.Limit)
	}
	writeEngineUint64(h, schedule.SchedulingGas)
	writeEngineUint64(h, schedule.CrossZoneRoutingGas)
	writeEngineUint64(h, schedule.ProofVerificationGas)
	writeEngineUint64(h, schedule.ContinuationGas)
	writeEngineUint64(h, schedule.InterfaceIntrospectionGas)
	writeEngineUint64(h, schedule.RetryGas)
	writeEngineUint64(h, schedule.BounceGas)
	writeEngineBool(h, schedule.RefundUnused)
	writeEngineUint64(h, schedule.MaxRefundGas)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMGasPolicyHash(policy AVMGasPolicy) string {
	policy = canonicalAVMGasPolicy(policy)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-gas-policy-v1")
	writeEngineUint64(h, policy.BaseMessageGas)
	writeEngineUint64(h, policy.PerBytePayloadGas)
	writeEngineUint64(h, policy.StorageReadGas)
	writeEngineUint64(h, policy.StorageWriteGas)
	writeEngineUint64(h, policy.QueueInsertGas)
	writeEngineUint64(h, policy.QueuePopGas)
	writeEngineUint64(h, policy.CrossZoneBaseGas)
	writeEngineUint64(h, policy.ProofVerifyBaseGas)
	writeEngineUint64(h, policy.ContinuationStoreGas)
	writeEngineUint64(h, policy.BounceBaseGas)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMAsyncGasReserveHash(reserve AVMAsyncGasReserve) string {
	reserve = canonicalAVMAsyncGasReserve(reserve)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-async-gas-reserve-v1")
	writeEnginePart(h, reserve.MessageID)
	writeEnginePart(h, string(reserve.ZoneID))
	writeEngineUint64(h, reserve.ReservedGas)
	writeEngineUint64(h, reserve.EscrowedGas)
	writeEngineUint64(h, reserve.RemainingGas)
	writeEngineUint64(h, reserve.RefundGas)
	writeEngineBool(h, reserve.Refundable)
	writeEngineUint64(h, uint64(len(reserve.Consumed)))
	for _, charge := range reserve.Consumed {
		writeEnginePart(h, string(charge.Class))
		writeEngineUint64(h, charge.Amount)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMZoneAsyncGasMeterHash(meter AVMZoneAsyncGasMeter) string {
	meter = canonicalAVMZoneAsyncGasMeter(meter)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-zone-async-gas-meter-v1")
	writeEnginePart(h, string(meter.ZoneID))
	writeEngineUint64(h, meter.Budget.MaxGas)
	writeEngineUint64(h, meter.Budget.GasUsed)
	writeEngineUint64(h, uint64(meter.Budget.MaxMessages))
	writeEngineUint64(h, uint64(meter.Budget.MessagesUsed))
	writeEngineUint64(h, uint64(len(meter.Consumed)))
	for _, charge := range meter.Consumed {
		writeEnginePart(h, string(charge.Class))
		writeEngineUint64(h, charge.Amount)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func allAVMGasClasses() []AVMGasClass {
	return []AVMGasClass{
		AVMGasClassContinuation,
		AVMGasClassCrossZoneRouting,
		AVMGasClassExecution,
		AVMGasClassInterfaceIntrospection,
		AVMGasClassProofVerification,
		AVMGasClassScheduling,
		AVMGasClassStorage,
	}
}

func sumAVMGasCharges(charges []AVMGasCharge) (uint64, []AVMGasCharge, error) {
	out := append([]AVMGasCharge(nil), charges...)
	var total uint64
	for _, charge := range out {
		if err := charge.Validate(); err != nil {
			return 0, nil, err
		}
		if charge.Amount > ^uint64(0)-total {
			return 0, nil, errors.New("AVM gas charge overflow")
		}
		total += charge.Amount
	}
	sort.SliceStable(out, func(i, j int) bool {
		return compareAVMGasCharges(out[i], out[j]) < 0
	})
	return total, out, nil
}

func canonicalAVMGasSchedule(schedule AVMGasSchedule) AVMGasSchedule {
	schedule.ScheduleHash = strings.TrimSpace(schedule.ScheduleHash)
	schedule.ClassBudgets = append([]AVMGasClassBudget(nil), schedule.ClassBudgets...)
	sort.SliceStable(schedule.ClassBudgets, func(i, j int) bool {
		return schedule.ClassBudgets[i].Class < schedule.ClassBudgets[j].Class
	})
	return schedule
}

func canonicalAVMGasPolicy(policy AVMGasPolicy) AVMGasPolicy {
	policy.PolicyHash = strings.TrimSpace(policy.PolicyHash)
	return policy
}

func canonicalAVMAsyncGasReserve(reserve AVMAsyncGasReserve) AVMAsyncGasReserve {
	reserve.MessageID = strings.TrimSpace(reserve.MessageID)
	reserve.ReserveHash = strings.TrimSpace(reserve.ReserveHash)
	reserve.Consumed = append([]AVMGasCharge(nil), reserve.Consumed...)
	sort.SliceStable(reserve.Consumed, func(i, j int) bool {
		return compareAVMGasCharges(reserve.Consumed[i], reserve.Consumed[j]) < 0
	})
	return reserve
}

func canonicalAVMZoneAsyncGasMeter(meter AVMZoneAsyncGasMeter) AVMZoneAsyncGasMeter {
	meter.MeterHash = strings.TrimSpace(meter.MeterHash)
	meter.Consumed = append([]AVMGasCharge(nil), meter.Consumed...)
	sort.SliceStable(meter.Consumed, func(i, j int) bool {
		return compareAVMGasCharges(meter.Consumed[i], meter.Consumed[j]) < 0
	})
	return meter
}

func compareAVMGasCharges(left, right AVMGasCharge) int {
	if left.Class < right.Class {
		return -1
	}
	if left.Class > right.Class {
		return 1
	}
	if left.Amount < right.Amount {
		return -1
	}
	if left.Amount > right.Amount {
		return 1
	}
	return 0
}

func avmStorageByteGas(name string, bytes uint64, policy AVMGasPolicy, perByte uint64) (uint64, error) {
	policy = canonicalAVMGasPolicy(policy)
	if err := policy.Validate(); err != nil {
		return 0, err
	}
	if bytes == 0 {
		return 0, fmt.Errorf("%s bytes must be positive", name)
	}
	return checkedAVMGasMul(perByte, bytes)
}

func checkedAVMGasAdd(left, right uint64) (uint64, error) {
	if right > ^uint64(0)-left {
		return 0, errors.New("AVM gas addition overflow")
	}
	return left + right, nil
}

func checkedAVMGasMul(left, right uint64) (uint64, error) {
	if left != 0 && right > ^uint64(0)/left {
		return 0, errors.New("AVM gas multiplication overflow")
	}
	return left * right, nil
}

func reserveChargeTotalForClass(reserve AVMAsyncGasReserve, class AVMGasClass) uint64 {
	var total uint64
	for _, charge := range reserve.Consumed {
		if charge.Class == class {
			if charge.Amount > ^uint64(0)-total {
				return ^uint64(0)
			}
			total += charge.Amount
		}
	}
	return total
}

func proofCountForMessage(msg AVMAsyncMessage) uint32 {
	var count uint32
	if msg.AuthProofOptional != "" {
		count++
	}
	if msg.StateProofOptional != "" {
		count++
	}
	if count == 0 {
		return 1
	}
	return count
}
