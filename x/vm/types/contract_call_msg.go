package types

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"strings"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sovereign-l1/l1/x/aetravm/async"
	"github.com/sovereign-l1/l1/x/aetravm/avm"
)

const (
	MaxContractMethodLength = 96
)

type MsgContractCall struct {
	Caller		sdk.AccAddress
	ContractAddr	sdk.AccAddress
	Method		string
	Args		[]byte
	Funds		sdkmath.Int
	GasLimit	uint64
	ReplyToOptional	sdk.AccAddress
	ExpiryHeight	uint64
}

type ContractMethodAdmission struct {
	ContractAddr	sdk.AccAddress
	Method		string
	Entrypoint	avm.Entrypoint
	Enabled		bool
}

type ContractFundsEscrow struct {
	Caller		sdk.AccAddress
	ContractAddr	sdk.AccAddress
	Amount		sdkmath.Int
	ExpiryHeight	uint64
	Escrowed	bool
}

type ContractCallAdmission struct {
	CreatedHeight	uint64
	Methods		[]ContractMethodAdmission
	Escrows		[]ContractFundsEscrow
	MaxArgsBytes	uint64
}

func BuildContractCallFromMsg(msg MsgContractCall, state ContractZoneState, admission ContractCallAdmission) (ContractCall, error) {
	msg = msg.Normalize()
	state = normalizeContractState(state)
	admission = admission.Normalize()
	if err := msg.Validate(state, admission); err != nil {
		return ContractCall{}, err
	}
	method, _ := admission.MethodFor(msg.ContractAddr, msg.Method)
	return ContractCall{
		Actor:			cloneAddress(msg.Caller),
		Contract:		cloneAddress(msg.ContractAddr),
		Entrypoint:		method.Entrypoint,
		GasLimit:		msg.GasLimit,
		Body:			append([]byte(nil), msg.Args...),
		EmitDestination:	cloneAddress(msg.ReplyToOptional),
	}, nil
}

func (m MsgContractCall) Normalize() MsgContractCall {
	m.Caller = cloneAddress(m.Caller)
	m.ContractAddr = cloneAddress(m.ContractAddr)
	m.Method = strings.TrimSpace(m.Method)
	m.Args = append([]byte(nil), m.Args...)
	if m.Funds.IsNil() {
		m.Funds = sdkmath.ZeroInt()
	}
	m.ReplyToOptional = cloneAddress(m.ReplyToOptional)
	return m
}

func (m MsgContractCall) Validate(state ContractZoneState, admission ContractCallAdmission) error {
	m = m.Normalize()
	state = normalizeContractState(state)
	admission = admission.Normalize()
	if err := state.Validate(); err != nil {
		return err
	}
	if err := admission.Validate(state); err != nil {
		return err
	}
	if err := validateContractAddress("contract call caller", m.Caller); err != nil {
		return err
	}
	if err := validateContractAddress("contract call address", m.ContractAddr); err != nil {
		return err
	}
	contract, found := findContract(state, m.ContractAddr)
	if !found {
		return errors.New("contract call target does not exist")
	}
	if _, found := findCode(state, contract.CodeID); !found {
		return errors.New("contract call target code does not exist")
	}
	method, found := admission.MethodFor(m.ContractAddr, m.Method)
	if !found || !method.Enabled {
		return errors.New("contract call method is not enabled")
	}
	if !avm.IsValidEntrypoint(method.Entrypoint) || method.Entrypoint == avm.EntryDeploy || method.Entrypoint == avm.EntryMigrate || method.Entrypoint == avm.EntryQuery {
		return errors.New("contract call method selector is invalid")
	}
	maxArgs := admission.MaxArgsBytes
	if maxArgs == 0 {
		maxArgs = uint64(async.DefaultParams().MaxBodySize)
	}
	if uint64(len(m.Args)) > maxArgs {
		return fmt.Errorf("contract call args size must be <= %d", maxArgs)
	}
	if m.Funds.IsNegative() {
		return errors.New("contract call funds must be non-negative")
	}
	if !admission.HasEscrow(m) {
		return errors.New("contract call funds must be escrowed")
	}
	if m.GasLimit == 0 || m.GasLimit > state.Policy.GasModel.ExecuteGas {
		return fmt.Errorf("contract call gas limit must be in 1..%d", state.Policy.GasModel.ExecuteGas)
	}
	if m.ExpiryHeight == 0 || m.ExpiryHeight < admission.CreatedHeight {
		return errors.New("contract call expiry height must be at or after created height")
	}
	if len(m.ReplyToOptional) != 0 {
		if err := validateContractAddress("contract call reply target", m.ReplyToOptional); err != nil {
			return err
		}
	}
	return nil
}

func (a ContractCallAdmission) Normalize() ContractCallAdmission {
	a.Methods = normalizeContractMethodAdmissions(a.Methods)
	a.Escrows = normalizeContractFundsEscrows(a.Escrows)
	return a
}

func (a ContractCallAdmission) Validate(state ContractZoneState) error {
	if a.CreatedHeight == 0 {
		return errors.New("contract call created height must be positive")
	}
	for _, method := range a.Methods {
		if err := method.Validate(state); err != nil {
			return err
		}
	}
	for _, escrow := range a.Escrows {
		if err := escrow.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (a ContractCallAdmission) MethodFor(contractAddr sdk.AccAddress, methodID string) (ContractMethodAdmission, bool) {
	methodID = strings.TrimSpace(methodID)
	for _, method := range a.Methods {
		if bytes.Equal(method.ContractAddr, contractAddr) && method.Method == methodID {
			return method, true
		}
	}
	return ContractMethodAdmission{}, false
}

func (a ContractCallAdmission) HasEscrow(msg MsgContractCall) bool {
	msg = msg.Normalize()
	for _, escrow := range a.Escrows {
		escrow = escrow.Normalize()
		if !escrow.Escrowed {
			continue
		}
		if !bytes.Equal(escrow.Caller, msg.Caller) || !bytes.Equal(escrow.ContractAddr, msg.ContractAddr) {
			continue
		}
		if escrow.Amount.LT(msg.Funds) {
			continue
		}
		if escrow.ExpiryHeight < msg.ExpiryHeight {
			continue
		}
		return true
	}
	return false
}

func (m ContractMethodAdmission) Normalize() ContractMethodAdmission {
	m.ContractAddr = cloneAddress(m.ContractAddr)
	m.Method = strings.TrimSpace(m.Method)
	return m
}

func (m ContractMethodAdmission) Validate(state ContractZoneState) error {
	m = m.Normalize()
	if err := validateContractAddress("contract call method contract", m.ContractAddr); err != nil {
		return err
	}
	if _, found := findContract(state, m.ContractAddr); !found {
		return errors.New("contract call method references missing contract")
	}
	if err := validateContractToken("contract call method", m.Method, MaxContractMethodLength); err != nil {
		return err
	}
	if !avm.IsValidEntrypoint(m.Entrypoint) {
		return errors.New("contract call method entrypoint is invalid")
	}
	return nil
}

func (e ContractFundsEscrow) Normalize() ContractFundsEscrow {
	e.Caller = cloneAddress(e.Caller)
	e.ContractAddr = cloneAddress(e.ContractAddr)
	if e.Amount.IsNil() {
		e.Amount = sdkmath.ZeroInt()
	}
	return e
}

func (e ContractFundsEscrow) Validate() error {
	e = e.Normalize()
	if err := validateContractAddress("contract call escrow caller", e.Caller); err != nil {
		return err
	}
	if err := validateContractAddress("contract call escrow contract", e.ContractAddr); err != nil {
		return err
	}
	if e.Amount.IsNegative() {
		return errors.New("contract call escrow amount must be non-negative")
	}
	if e.ExpiryHeight == 0 {
		return errors.New("contract call escrow expiry height must be positive")
	}
	return nil
}

func normalizeContractMethodAdmissions(values []ContractMethodAdmission) []ContractMethodAdmission {
	out := make([]ContractMethodAdmission, len(values))
	for i, value := range values {
		out[i] = value.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool {
		if cmp := bytes.Compare(out[i].ContractAddr, out[j].ContractAddr); cmp != 0 {
			return cmp < 0
		}
		return out[i].Method < out[j].Method
	})
	return out
}

func normalizeContractFundsEscrows(values []ContractFundsEscrow) []ContractFundsEscrow {
	out := make([]ContractFundsEscrow, len(values))
	for i, value := range values {
		out[i] = value.Normalize()
	}
	sort.SliceStable(out, func(i, j int) bool {
		if cmp := bytes.Compare(out[i].Caller, out[j].Caller); cmp != 0 {
			return cmp < 0
		}
		return bytes.Compare(out[i].ContractAddr, out[j].ContractAddr) < 0
	})
	return out
}

func validateContractToken(field string, value string, max int) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("%s is required", field)
	}
	if len(value) > max {
		return fmt.Errorf("%s must be <= %d bytes", field, max)
	}
	for _, ch := range value {
		if ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' || ch >= '0' && ch <= '9' || ch == '_' || ch == '-' || ch == '.' || ch == ':' {
			continue
		}
		return fmt.Errorf("%s contains unsupported character %q", field, ch)
	}
	return nil
}
