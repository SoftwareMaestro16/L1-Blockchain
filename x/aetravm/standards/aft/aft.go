package aft

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	aetraaddress "github.com/sovereign-l1/l1/app/addressing"
	appparams "github.com/sovereign-l1/l1/app/params"
	feestypes "github.com/sovereign-l1/l1/x/fees/types"
)

const (
	StandardName             = "AFT-44"
	MaxTokenNameLength       = 64
	MaxTokenSymbolLength     = 16
	MaxMetadataRefLength     = 256
	MaxDecimals              = uint32(18)
	WalletCodeHashLength     = 32
	DefaultWalletCodeVersion = uint32(1)
	walletDerivationDomain   = "aetra/AFT-44/wallet/v1"
)

type TokenMetadata struct {
	Name        string
	Symbol      string
	Decimals    uint32
	ContentRef  string
	DisplayName string
}

type MasterState struct {
	Address               sdk.AccAddress
	Admin                 sdk.AccAddress
	TotalSupply           sdkmath.Int
	Mintable              bool
	Burnable              bool
	Metadata              TokenMetadata
	WalletCodeHash        []byte
	WalletCodeUpgradeable bool
	AdminTransferPending  sdk.AccAddress
	AdminRenounced        bool
}

type WalletState struct {
	Address           sdk.AccAddress
	MasterAddress     sdk.AccAddress
	Owner             sdk.AccAddress
	Balance           sdkmath.Int
	WalletCodeVersion uint32
	PendingQueryIDs   map[uint64]struct{}
	ProcessedQueryIDs map[uint64]struct{}
}

type State struct {
	Master  MasterState
	Wallets map[string]WalletState
}

func NewState(master MasterState) (*State, error) {
	if err := master.Validate(); err != nil {
		return nil, err
	}
	return &State{
		Master:  master,
		Wallets: make(map[string]WalletState),
	}, nil
}

func DeriveWalletAddress(master, owner sdk.AccAddress, walletCodeHash []byte) (sdk.AccAddress, error) {
	if err := aetraaddress.RejectZeroAddress("token master", master); err != nil {
		return nil, err
	}
	if err := aetraaddress.RejectZeroAddress("wallet owner", owner); err != nil {
		return nil, err
	}
	if len(walletCodeHash) != WalletCodeHashLength {
		return nil, fmt.Errorf("wallet code hash must be %d bytes", WalletCodeHashLength)
	}
	h := sha256.New()
	h.Write([]byte(walletDerivationDomain))
	writePart(h.Write, master)
	writePart(h.Write, owner)
	writePart(h.Write, walletCodeHash)
	return sdk.AccAddress(h.Sum(nil)), nil
}

func (m MasterState) Validate() error {
	if err := aetraaddress.RejectZeroAddress("token master", m.Address); err != nil {
		return err
	}
	if !m.AdminRenounced {
		if err := aetraaddress.RejectZeroAddress("token admin", m.Admin); err != nil {
			return err
		}
	}
	if m.AdminRenounced && len(m.AdminTransferPending) != 0 {
		return errors.New("renounced admin cannot have pending admin transfer")
	}
	if !m.TotalSupply.IsNil() && m.TotalSupply.IsNegative() {
		return errors.New("total supply must not be negative")
	}
	if len(m.WalletCodeHash) != WalletCodeHashLength {
		return fmt.Errorf("wallet code hash must be %d bytes", WalletCodeHashLength)
	}
	return ValidateMetadata(m.Metadata)
}

func (w WalletState) Validate() error {
	if err := aetraaddress.RejectZeroAddress("token wallet", w.Address); err != nil {
		return err
	}
	if err := aetraaddress.RejectZeroAddress("token master", w.MasterAddress); err != nil {
		return err
	}
	if err := aetraaddress.RejectZeroAddress("wallet owner", w.Owner); err != nil {
		return err
	}
	if !w.Balance.IsNil() && w.Balance.IsNegative() {
		return errors.New("wallet balance must not be negative")
	}
	if w.WalletCodeVersion == 0 {
		return errors.New("wallet code version must be positive")
	}
	return nil
}

func ValidateMetadata(metadata TokenMetadata) error {
	name := strings.TrimSpace(metadata.Name)
	symbol := strings.TrimSpace(metadata.Symbol)
	contentRef := strings.TrimSpace(metadata.ContentRef)
	if name == "" {
		return errors.New("token name must be set")
	}
	if symbol == "" {
		return errors.New("token symbol must be set")
	}
	if len(name) > MaxTokenNameLength {
		return fmt.Errorf("token name length must be <= %d", MaxTokenNameLength)
	}
	if len(symbol) > MaxTokenSymbolLength {
		return fmt.Errorf("token symbol length must be <= %d", MaxTokenSymbolLength)
	}
	if len(contentRef) > MaxMetadataRefLength {
		return fmt.Errorf("metadata content reference length must be <= %d", MaxMetadataRefLength)
	}
	if metadata.Decimals > MaxDecimals {
		return fmt.Errorf("token decimals must be <= %d", MaxDecimals)
	}
	for _, value := range []string{name, symbol, strings.TrimSpace(metadata.DisplayName)} {
		if spoofsNative(value) {
			return errors.New("AFT-44 metadata must not spoof native AET/naet metadata")
		}
	}
	return nil
}

func ValidateOperationFees(fees sdk.Coins) error {
	return feestypes.ValidateFeeCoins(feestypes.DefaultParams(), fees, true)
}

func (s *State) WalletAddress(owner sdk.AccAddress) (sdk.AccAddress, error) {
	return DeriveWalletAddress(s.Master.Address, owner, s.Master.WalletCodeHash)
}

func (s *State) Wallet(owner sdk.AccAddress) (WalletState, bool, error) {
	addr, err := s.WalletAddress(owner)
	if err != nil {
		return WalletState{}, false, err
	}
	wallet, ok := s.Wallets[string(addr)]
	return wallet, ok, nil
}

func (s *State) Mint(caller, recipient sdk.AccAddress, amount sdkmath.Int) error {
	if err := s.requireAdmin(caller); err != nil {
		return err
	}
	if !s.Master.Mintable {
		return errors.New("token minting is closed")
	}
	if !amount.IsPositive() {
		return errors.New("mint amount must be positive")
	}
	wallet, err := s.ensureWallet(recipient)
	if err != nil {
		return err
	}
	wallet.Balance = wallet.Balance.Add(amount)
	s.Wallets[string(wallet.Address)] = wallet
	s.Master.TotalSupply = s.Master.TotalSupply.Add(amount)
	return s.ValidateAccounting()
}

func (s *State) Transfer(owner, recipient sdk.AccAddress, amount sdkmath.Int, queryID uint64) error {
	if !amount.IsPositive() {
		return errors.New("transfer amount must be positive")
	}
	if queryID == 0 {
		return errors.New("query id must be non-zero")
	}
	from, err := s.requireWallet(owner)
	if err != nil {
		return err
	}
	if from.hasSeenQuery(queryID) {
		return errors.New("replayed wallet message")
	}
	if from.Balance.LT(amount) {
		return errors.New("insufficient token wallet balance")
	}
	to, err := s.ensureWallet(recipient)
	if err != nil {
		return err
	}
	from.Balance = from.Balance.Sub(amount)
	from.PendingQueryIDs[queryID] = struct{}{}
	to.Balance = to.Balance.Add(amount)
	s.Wallets[string(from.Address)] = from
	s.Wallets[string(to.Address)] = to
	return s.ValidateAccounting()
}

func (s *State) Burn(owner sdk.AccAddress, amount sdkmath.Int, queryID uint64) error {
	if !s.Master.Burnable {
		return errors.New("token burning is closed")
	}
	if !amount.IsPositive() {
		return errors.New("burn amount must be positive")
	}
	if queryID == 0 {
		return errors.New("query id must be non-zero")
	}
	wallet, err := s.requireWallet(owner)
	if err != nil {
		return err
	}
	if wallet.hasSeenQuery(queryID) {
		return errors.New("replayed wallet message")
	}
	if wallet.Balance.LT(amount) {
		return errors.New("insufficient token wallet balance")
	}
	wallet.Balance = wallet.Balance.Sub(amount)
	wallet.PendingQueryIDs[queryID] = struct{}{}
	s.Wallets[string(wallet.Address)] = wallet
	s.Master.TotalSupply = s.Master.TotalSupply.Sub(amount)
	return s.ValidateAccounting()
}

func (s *State) BounceTransfer(owner sdk.AccAddress, amount sdkmath.Int, queryID uint64) error {
	if !amount.IsPositive() {
		return errors.New("bounce amount must be positive")
	}
	wallet, err := s.requireWallet(owner)
	if err != nil {
		return err
	}
	if _, exists := wallet.PendingQueryIDs[queryID]; !exists {
		return errors.New("unknown pending query id")
	}
	delete(wallet.PendingQueryIDs, queryID)
	wallet.ProcessedQueryIDs[queryID] = struct{}{}
	s.Wallets[string(wallet.Address)] = wallet
	return s.ValidateAccounting()
}

func (s *State) ChangeAdmin(caller, nextAdmin sdk.AccAddress) error {
	if err := s.requireAdmin(caller); err != nil {
		return err
	}
	if err := aetraaddress.RejectZeroAddress("new token admin", nextAdmin); err != nil {
		return err
	}
	s.Master.AdminTransferPending = append(sdk.AccAddress(nil), nextAdmin...)
	return s.Master.Validate()
}

func (s *State) AcceptAdmin(caller sdk.AccAddress) error {
	if len(s.Master.AdminTransferPending) == 0 {
		return errors.New("no pending admin transfer")
	}
	if !s.Master.AdminTransferPending.Equals(caller) {
		return errors.New("only pending admin can accept admin transfer")
	}
	s.Master.Admin = append(sdk.AccAddress(nil), caller...)
	s.Master.AdminTransferPending = nil
	return s.Master.Validate()
}

func (s *State) RenounceAdmin(caller sdk.AccAddress) error {
	if err := s.requireAdmin(caller); err != nil {
		return err
	}
	s.Master.Admin = nil
	s.Master.AdminTransferPending = nil
	s.Master.AdminRenounced = true
	s.Master.Mintable = false
	return s.Master.Validate()
}

func (s *State) ChangeMetadata(caller sdk.AccAddress, metadata TokenMetadata) error {
	if err := s.requireAdmin(caller); err != nil {
		return err
	}
	if err := ValidateMetadata(metadata); err != nil {
		return err
	}
	s.Master.Metadata = metadata
	return s.Master.Validate()
}

func (s *State) CloseMinting(caller sdk.AccAddress) error {
	if err := s.requireAdmin(caller); err != nil {
		return err
	}
	s.Master.Mintable = false
	return nil
}

func (s *State) ValidateAccounting() error {
	if err := s.Master.Validate(); err != nil {
		return err
	}
	total := sdkmath.ZeroInt()
	for _, wallet := range s.Wallets {
		if err := wallet.Validate(); err != nil {
			return err
		}
		if !wallet.MasterAddress.Equals(s.Master.Address) {
			return errors.New("wallet master address mismatch")
		}
		expected, err := s.WalletAddress(wallet.Owner)
		if err != nil {
			return err
		}
		if !expected.Equals(wallet.Address) {
			return errors.New("wallet address derivation mismatch")
		}
		total = total.Add(wallet.Balance)
	}
	if !total.Equal(s.Master.TotalSupply) {
		return fmt.Errorf("AFT-44 supply mismatch: master %s != wallets %s", s.Master.TotalSupply, total)
	}
	return nil
}

func (s *State) ensureWallet(owner sdk.AccAddress) (WalletState, error) {
	addr, err := s.WalletAddress(owner)
	if err != nil {
		return WalletState{}, err
	}
	if wallet, ok := s.Wallets[string(addr)]; ok {
		return wallet, nil
	}
	wallet := WalletState{
		Address:           addr,
		MasterAddress:     append(sdk.AccAddress(nil), s.Master.Address...),
		Owner:             append(sdk.AccAddress(nil), owner...),
		Balance:           sdkmath.ZeroInt(),
		WalletCodeVersion: DefaultWalletCodeVersion,
		PendingQueryIDs:   make(map[uint64]struct{}),
		ProcessedQueryIDs: make(map[uint64]struct{}),
	}
	return wallet, wallet.Validate()
}

func (s *State) requireWallet(owner sdk.AccAddress) (WalletState, error) {
	wallet, ok, err := s.Wallet(owner)
	if err != nil {
		return WalletState{}, err
	}
	if !ok {
		return WalletState{}, errors.New("token wallet does not exist")
	}
	if wallet.PendingQueryIDs == nil {
		wallet.PendingQueryIDs = make(map[uint64]struct{})
	}
	if wallet.ProcessedQueryIDs == nil {
		wallet.ProcessedQueryIDs = make(map[uint64]struct{})
	}
	return wallet, nil
}

func (s *State) requireAdmin(caller sdk.AccAddress) error {
	if s.Master.AdminRenounced {
		return errors.New("token admin is renounced")
	}
	if !s.Master.Admin.Equals(caller) {
		return errors.New("only token admin can perform this action")
	}
	return nil
}

func spoofsNative(value string) bool {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "" {
		return false
	}
	return normalized == strings.ToLower(appparams.TokenName) ||
		normalized == strings.ToLower(appparams.TokenSymbol) ||
		normalized == strings.ToLower(appparams.DisplayDenom) ||
		normalized == strings.ToLower(appparams.BaseDenom)
}

func writePart(write func([]byte) (int, error), bz []byte) {
	var length [8]byte
	binary.BigEndian.PutUint64(length[:], uint64(len(bz)))
	_, _ = write(length[:])
	_, _ = write(bz)
}

func (w WalletState) hasSeenQuery(queryID uint64) bool {
	if _, exists := w.PendingQueryIDs[queryID]; exists {
		return true
	}
	_, exists := w.ProcessedQueryIDs[queryID]
	return exists
}
