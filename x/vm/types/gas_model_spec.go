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
	AVMGasClassExecution              AVMGasClass = "execution"
	AVMGasClassStorage                AVMGasClass = "storage"
	AVMGasClassScheduling             AVMGasClass = "scheduling"
	AVMGasClassCrossZoneRouting       AVMGasClass = "cross_zone_routing"
	AVMGasClassProofVerification      AVMGasClass = "proof_verification"
	AVMGasClassContinuation           AVMGasClass = "continuation"
	AVMGasClassInterfaceIntrospection AVMGasClass = "interface_introspection"

	MaxAVMGasCharges = 32
)

type AVMGasClass string

type AVMGasClassBudget struct {
	Class AVMGasClass
	Limit uint64
}

type AVMGasCharge struct {
	Class  AVMGasClass
	Amount uint64
}

type AVMGasSchedule struct {
	ClassBudgets              []AVMGasClassBudget
	SchedulingGas             uint64
	CrossZoneRoutingGas       uint64
	ProofVerificationGas      uint64
	ContinuationGas           uint64
	InterfaceIntrospectionGas uint64
	RetryGas                  uint64
	BounceGas                 uint64
	RefundUnused              bool
	MaxRefundGas              uint64
	ScheduleHash              string
}

type AVMAsyncGasReserve struct {
	MessageID    string
	ZoneID       zonestypes.ZoneID
	ReservedGas  uint64
	EscrowedGas  uint64
	RemainingGas uint64
	Consumed     []AVMGasCharge
	RefundGas    uint64
	Refundable   bool
	ReserveHash  string
}

func DefaultAVMGasSchedule() (AVMGasSchedule, error) {
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
		SchedulingGas:             10,
		CrossZoneRoutingGas:       20,
		ProofVerificationGas:      15,
		ContinuationGas:           30,
		InterfaceIntrospectionGas: 5,
		RetryGas:                  10,
		BounceGas:                 25,
		RefundUnused:              true,
		MaxRefundGas:              1_000_000,
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
		MessageID:    msg.ID,
		ZoneID:       msg.DestinationZone,
		ReservedGas:  required,
		EscrowedGas:  required,
		RemainingGas: required,
		Consumed:     charges,
		Refundable:   schedule.RefundUnused,
	}
	reserve = canonicalAVMAsyncGasReserve(reserve)
	reserve.ReserveHash = ComputeAVMAsyncGasReserveHash(reserve)
	return reserve, reserve.Validate()
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
		name  string
		value uint64
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
	writeEnginePart(h, "aetheris-avm-gas-schedule-v1")
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

func ComputeAVMAsyncGasReserveHash(reserve AVMAsyncGasReserve) string {
	reserve = canonicalAVMAsyncGasReserve(reserve)
	h := sha256.New()
	writeEnginePart(h, "aetheris-avm-async-gas-reserve-v1")
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

func canonicalAVMAsyncGasReserve(reserve AVMAsyncGasReserve) AVMAsyncGasReserve {
	reserve.MessageID = strings.TrimSpace(reserve.MessageID)
	reserve.ReserveHash = strings.TrimSpace(reserve.ReserveHash)
	reserve.Consumed = append([]AVMGasCharge(nil), reserve.Consumed...)
	sort.SliceStable(reserve.Consumed, func(i, j int) bool {
		return compareAVMGasCharges(reserve.Consumed[i], reserve.Consumed[j]) < 0
	})
	return reserve
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
