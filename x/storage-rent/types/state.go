package types

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
)

const (
	ContractStatusActive	= "active"
	ContractStatusFrozen	= "frozen"
	ContractStatusDeleted	= "deleted"
	ContractStatusArchived	= "archived"

	DistributionScale	= uint32(10_000)
	DefaultProofRoot	= "0000000000000000000000000000000000000000000000000000000000000000"
)

type StorageRentParams struct {
	RentRatePerByteBlock	uint64
	RentRatePerByteSecond	uint64
	FreeStorageAllowance	uint64
	RetentionBlocks		uint64
	UnfreezeBufferBlocks	uint64
	MaxContracts		uint32
	MaxExemptions		uint32
	FeeCollectorAccount	string
	TreasuryAccount		string
	FeeCollectorShare	uint32
	TreasuryShare		uint32
	BurnShare		uint32
}

type StorageRentState struct {
	Contracts	[]ContractRentRecord
	Distributions	[]RentDistributionRecord
	Exemptions	[]RentExemption
	SystemReserve	SystemRentReserve
}

type SystemRentReserve struct {
	AvailableFunds				uint64
	ProjectedRentPerBlock			uint64
	WarningRunwayBlocks			uint64
	CriticalRunwayBlocks			uint64
	FeeCollectorBalance			uint64
	TreasuryBalance				uint64
	GovernanceConfiguredPayerBalance	uint64
	RequiredTopUp				uint64
	ProtocolCriticalExecutable		bool
	LastRunwayBlocks			uint64
	LastAlert				string
	LastTopUpAmount				uint64
	LastRemainingDebt			uint64
	LastUpdatedHeight			uint64
}

type ContractRentRecord struct {
	ContractAddress			string
	ActorID				string
	StorageBytes			uint64
	PrepaidRentBalance		uint64
	RentDebt			uint64
	LastChargedHeight		uint64
	Status				string
	FreezeHeight			uint64
	DeletionEligibilityHeight	uint64
	ArchivalProofRoot		string
	Exempt				bool
}

type RentDistributionRecord struct {
	ContractAddress		string
	Height			uint64
	Amount			uint64
	FeeCollectorAmount	uint64
	TreasuryAmount		uint64
	BurnAmount		uint64
}

type RentExemption struct {
	Account	string
	Reason	string
}

type MsgWithdrawExcessRent struct {
	Authority	string
	ContractAddress	string
	Amount		uint64
	Height		uint64
}

type MsgFreezeExpiredContract struct {
	Authority	string
	ContractAddress	string
	Height		uint64
}

type MsgDeleteExpiredContract struct {
	Authority		string
	ContractAddress		string
	Height			uint64
	ArchivalProofRoot	string
}

type MsgUpdateStorageRentParams struct {
	Authority	string
	Params		StorageRentParams
}

func DefaultStorageRentParams() StorageRentParams {
	return StorageRentParams{
		RentRatePerByteBlock:	1,
		RentRatePerByteSecond:	1,
		RetentionBlocks:	10_000,
		UnfreezeBufferBlocks:	100,
		MaxContracts:		100_000,
		MaxExemptions:		1_024,
		FeeCollectorAccount:	"fee_collector",
		TreasuryAccount:	"treasury",
		FeeCollectorShare:	5_000,
		TreasuryShare:		4_000,
		BurnShare:		1_000,
	}
}

func EmptyStorageRentState() StorageRentState {
	return StorageRentState{
		Contracts:	[]ContractRentRecord{},
		Distributions:	[]RentDistributionRecord{},
		Exemptions:	[]RentExemption{},
		SystemReserve:	DefaultSystemRentReserve(),
	}
}

func DefaultSystemRentReserve() SystemRentReserve {
	return SystemRentReserve{
		ProtocolCriticalExecutable: true,
	}
}

func (p StorageRentParams) Validate() error {
	if p.RentRatePerByteBlock == 0 {
		return errors.New("storage rent rate must be positive")
	}
	if p.RentRatePerByteSecond == 0 {
		return errors.New("storage rent per-second rate must be positive")
	}
	if p.RetentionBlocks == 0 {
		return errors.New("storage rent retention blocks must be positive")
	}
	if p.MaxContracts == 0 || p.MaxExemptions == 0 {
		return errors.New("storage rent limits must be positive")
	}
	if p.FeeCollectorShare+p.TreasuryShare+p.BurnShare != DistributionScale {
		return errors.New("storage rent distribution shares must sum to scale")
	}
	if p.FeeCollectorShare > 0 && strings.TrimSpace(p.FeeCollectorAccount) == "" {
		return errors.New("storage rent fee collector account is required")
	}
	if p.TreasuryShare > 0 && strings.TrimSpace(p.TreasuryAccount) == "" {
		return errors.New("storage rent treasury account is required")
	}
	return nil
}

func (s StorageRentState) Export() StorageRentState {
	out := StorageRentState{
		Contracts:	cloneContracts(s.Contracts),
		Distributions:	cloneDistributions(s.Distributions),
		Exemptions:	cloneExemptions(s.Exemptions),
		SystemReserve:	s.SystemReserve,
	}
	SortContracts(out.Contracts)
	SortDistributions(out.Distributions)
	SortExemptions(out.Exemptions)
	return out
}

func (s StorageRentState) Validate(params StorageRentParams) error {
	if err := params.Validate(); err != nil {
		return err
	}
	s = s.Export()
	if uint32(len(s.Contracts)) > params.MaxContracts {
		return errors.New("storage rent contract count exceeds limit")
	}
	if uint32(len(s.Exemptions)) > params.MaxExemptions {
		return errors.New("storage rent exemption count exceeds limit")
	}
	seenContracts := map[string]struct{}{}
	for _, contract := range s.Contracts {
		if err := contract.Validate(); err != nil {
			return err
		}
		if _, found := seenContracts[contract.ContractAddress]; found {
			return fmt.Errorf("duplicate storage rent contract %q", contract.ContractAddress)
		}
		seenContracts[contract.ContractAddress] = struct{}{}
	}
	seenExemptions := map[string]struct{}{}
	for _, exemption := range s.Exemptions {
		if err := exemption.Validate(); err != nil {
			return err
		}
		if _, found := seenExemptions[exemption.Account]; found {
			return fmt.Errorf("duplicate storage rent exemption %q", exemption.Account)
		}
		seenExemptions[exemption.Account] = struct{}{}
	}
	for _, distribution := range s.Distributions {
		if err := distribution.Validate(); err != nil {
			return err
		}
		if _, found := seenContracts[distribution.ContractAddress]; !found {
			return fmt.Errorf("storage rent distribution references unknown contract %q", distribution.ContractAddress)
		}
	}
	if err := s.SystemReserve.Validate(); err != nil {
		return err
	}
	return nil
}

func (r SystemRentReserve) Validate() error {
	if r.CriticalRunwayBlocks > r.WarningRunwayBlocks {
		return errors.New("storage rent system critical runway must not exceed warning runway")
	}
	switch r.LastAlert {
	case "", SystemRentAlertWarning, SystemRentAlertCritical, SystemRentAlertInvariant:
		return nil
	default:
		return fmt.Errorf("unsupported storage rent system alert %q", r.LastAlert)
	}
}

func (r SystemRentReserve) Accounting() SystemRentAccounting {
	return SystemRentAccounting{
		AvailableFunds:				r.AvailableFunds,
		ProjectedRentPerBlock:			r.ProjectedRentPerBlock,
		WarningRunwayBlocks:			r.WarningRunwayBlocks,
		CriticalRunwayBlocks:			r.CriticalRunwayBlocks,
		FeeCollectorBalance:			r.FeeCollectorBalance,
		TreasuryBalance:			r.TreasuryBalance,
		GovernanceConfiguredPayerBalance:	r.GovernanceConfiguredPayerBalance,
		RequiredTopUp:				r.RequiredTopUp,
		ProtocolCriticalExecutable:		r.ProtocolCriticalExecutable,
	}
}

func (r SystemRentReserve) Evaluate() SystemRentResult {
	return ComputeSystemRentAccounting(r.Accounting())
}

func (r SystemRentReserve) WithResult(height uint64, result SystemRentResult) SystemRentReserve {
	r.LastRunwayBlocks = result.RunwayBlocks
	r.LastAlert = result.Alert
	r.LastTopUpAmount = result.TopUpAmount
	r.LastRemainingDebt = result.RemainingDebt
	r.LastUpdatedHeight = height
	return r
}

func (c ContractRentRecord) Normalize() ContractRentRecord {
	c.ContractAddress = strings.TrimSpace(c.ContractAddress)
	c.ActorID = strings.TrimSpace(c.ActorID)
	c.Status = strings.TrimSpace(c.Status)
	if c.Status == "" {
		c.Status = ContractStatusActive
	}
	c.ArchivalProofRoot = strings.TrimSpace(c.ArchivalProofRoot)
	if c.ArchivalProofRoot == "" {
		c.ArchivalProofRoot = DefaultProofRoot
	}
	return c
}

func (c ContractRentRecord) Validate() error {
	c = c.Normalize()
	if c.ContractAddress == "" {
		return errors.New("storage rent contract address is required")
	}
	if c.ActorID == "" {
		return errors.New("storage rent actor id is required")
	}
	if !IsContractStatus(c.Status) {
		return errors.New("storage rent contract status is invalid")
	}
	if c.LastChargedHeight == 0 {
		return errors.New("storage rent last charged height must be positive")
	}
	if (c.Status == ContractStatusFrozen || c.Status == ContractStatusFrozenLimited) && (c.FreezeHeight == 0 || c.DeletionEligibilityHeight == 0) {
		return errors.New("storage rent frozen contract requires freeze and deletion eligibility heights")
	}
	if (c.Status == ContractStatusDeleted || c.Status == ContractStatusArchived) && c.DeletionEligibilityHeight == 0 {
		return errors.New("storage rent deleted contract requires deletion eligibility height")
	}
	if err := ValidateHexRoot("storage rent archival proof root", c.ArchivalProofRoot); err != nil {
		return err
	}
	return nil
}

func (d RentDistributionRecord) Validate() error {
	d.ContractAddress = strings.TrimSpace(d.ContractAddress)
	if d.ContractAddress == "" {
		return errors.New("storage rent distribution contract address is required")
	}
	if d.Height == 0 {
		return errors.New("storage rent distribution height must be positive")
	}
	if d.Amount == 0 {
		return errors.New("storage rent distribution amount must be positive")
	}
	if d.FeeCollectorAmount+d.TreasuryAmount+d.BurnAmount != d.Amount {
		return errors.New("storage rent distribution must conserve coins")
	}
	return nil
}

func (e RentExemption) Validate() error {
	e.Account = strings.TrimSpace(e.Account)
	e.Reason = strings.TrimSpace(e.Reason)
	if e.Account == "" || e.Reason == "" {
		return errors.New("storage rent exemption account and reason are required")
	}
	return nil
}

func AccrueRent(contract ContractRentRecord, params StorageRentParams, height uint64) (ContractRentRecord, uint64, error) {
	if err := params.Validate(); err != nil {
		return ContractRentRecord{}, 0, err
	}
	contract = contract.Normalize()
	if height == 0 || height < contract.LastChargedHeight {
		return ContractRentRecord{}, 0, errors.New("storage rent accrual height must be monotonic")
	}
	if contract.Exempt {
		contract.LastChargedHeight = height
		return contract, 0, nil
	}
	blocks := height - contract.LastChargedHeight
	due, err := RentForBlocks(contract.StorageBytes, params, blocks)
	if err != nil {
		return ContractRentRecord{}, 0, err
	}
	contract.LastChargedHeight = height
	if due <= contract.PrepaidRentBalance {
		contract.PrepaidRentBalance -= due
		return contract, due, nil
	}
	unpaid := due - contract.PrepaidRentBalance
	contract.PrepaidRentBalance = 0
	if contract.RentDebt > math.MaxUint64-unpaid {
		return ContractRentRecord{}, 0, errors.New("storage rent debt overflow")
	}
	contract.RentDebt += unpaid
	return contract, due, nil
}

func RentForBlocks(storageBytes uint64, params StorageRentParams, blocks uint64) (uint64, error) {
	if storageBytes <= params.FreeStorageAllowance || blocks == 0 {
		return 0, nil
	}
	billable := storageBytes - params.FreeStorageAllowance
	if billable > math.MaxUint64/params.RentRatePerByteBlock {
		return 0, errors.New("storage rent byte-rate overflow")
	}
	perBlock := billable * params.RentRatePerByteBlock
	if blocks > math.MaxUint64/perBlock {
		return 0, errors.New("storage rent block overflow")
	}
	return perBlock * blocks, nil
}

func RentForSeconds(storageBytes uint64, params StorageRentParams, elapsedSeconds uint64) (uint64, error) {
	if storageBytes <= params.FreeStorageAllowance || elapsedSeconds == 0 {
		return 0, nil
	}
	billable := storageBytes - params.FreeStorageAllowance
	if billable > math.MaxUint64/params.RentRatePerByteSecond {
		return 0, errors.New("storage rent byte-second rate overflow")
	}
	perSecond := billable * params.RentRatePerByteSecond
	if elapsedSeconds > math.MaxUint64/perSecond {
		return 0, errors.New("storage rent elapsed seconds overflow")
	}
	return perSecond * elapsedSeconds, nil
}

func RequiredUnfreezePayment(contract ContractRentRecord, params StorageRentParams) (uint64, error) {
	buffer, err := RentForBlocks(contract.StorageBytes, params, params.UnfreezeBufferBlocks)
	if err != nil {
		return 0, err
	}
	if contract.RentDebt > math.MaxUint64-buffer {
		return 0, errors.New("storage rent unfreeze payment overflow")
	}
	return contract.RentDebt + buffer, nil
}

func ApplyRentPayment(contract ContractRentRecord, amount uint64) (ContractRentRecord, error) {
	if amount <= contract.RentDebt {
		contract.RentDebt -= amount
		return contract, nil
	}
	excess := amount - contract.RentDebt
	contract.RentDebt = 0
	if contract.PrepaidRentBalance > math.MaxUint64-excess {
		return ContractRentRecord{}, errors.New("storage rent prepaid balance overflow")
	}
	contract.PrepaidRentBalance += excess
	return contract, nil
}

func BuildDistribution(contractAddress string, height, amount uint64, params StorageRentParams) RentDistributionRecord {
	fee := amount * uint64(params.FeeCollectorShare) / uint64(DistributionScale)
	treasury := amount * uint64(params.TreasuryShare) / uint64(DistributionScale)
	burn := amount - fee - treasury
	return RentDistributionRecord{
		ContractAddress:	strings.TrimSpace(contractAddress),
		Height:			height,
		Amount:			amount,
		FeeCollectorAmount:	fee,
		TreasuryAmount:		treasury,
		BurnAmount:		burn,
	}
}

func CanExecuteContract(contract ContractRentRecord) bool {
	return contract.Normalize().Status == ContractStatusActive
}

func IsContractStatus(status string) bool {
	switch status {
	case ContractStatusActive, ContractStatusFrozen, ContractStatusFrozenLimited, ContractStatusDeleted, ContractStatusArchived:
		return true
	default:
		return false
	}
}

func ValidateHexRoot(name, value string) error {
	value = strings.TrimSpace(value)
	if len(value) != 64 {
		return fmt.Errorf("%s must be 32-byte hex", name)
	}
	if _, err := hex.DecodeString(value); err != nil {
		return fmt.Errorf("%s must be hex: %w", name, err)
	}
	return nil
}

func SortContracts(contracts []ContractRentRecord) {
	sort.SliceStable(contracts, func(i, j int) bool {
		return contracts[i].Normalize().ContractAddress < contracts[j].Normalize().ContractAddress
	})
}

func SortDistributions(distributions []RentDistributionRecord) {
	sort.SliceStable(distributions, func(i, j int) bool {
		if distributions[i].Height != distributions[j].Height {
			return distributions[i].Height < distributions[j].Height
		}
		return distributions[i].ContractAddress < distributions[j].ContractAddress
	})
}

func SortExemptions(exemptions []RentExemption) {
	sort.SliceStable(exemptions, func(i, j int) bool { return exemptions[i].Account < exemptions[j].Account })
}

func cloneContracts(contracts []ContractRentRecord) []ContractRentRecord {
	out := make([]ContractRentRecord, len(contracts))
	for i, contract := range contracts {
		out[i] = contract.Normalize()
	}
	return out
}

func cloneDistributions(distributions []RentDistributionRecord) []RentDistributionRecord {
	out := append([]RentDistributionRecord(nil), distributions...)
	for i := range out {
		out[i].ContractAddress = strings.TrimSpace(out[i].ContractAddress)
	}
	return out
}

func cloneExemptions(exemptions []RentExemption) []RentExemption {
	out := append([]RentExemption(nil), exemptions...)
	for i := range out {
		out[i].Account = strings.TrimSpace(out[i].Account)
		out[i].Reason = strings.TrimSpace(out[i].Reason)
	}
	return out
}
