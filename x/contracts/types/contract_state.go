package types

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/sovereign-l1/l1/app/addressing"
)

const (
	ContractStatusActive        = "active"
	ContractStatusFrozen        = "frozen"
	ContractStatusFrozenLimited = "frozen_limited"
	ContractStatusArchived      = "archived"
	ContractStatusDeleted       = "deleted"

	AssetTypeToken = "token"
	AssetTypeNFT   = "nft"
	AssetTypeDEX   = "dex"

	NativeStakingCapability = "native_staking_injection"

	ErrUnauthorized    = "contracts_unauthorized"
	ErrAccountInactive = "contracts_account_inactive"
	ErrAccountFrozen   = "contracts_account_frozen"
	ErrStorageRent     = "contracts_storage_rent"
)

type State struct {
	Codes                []CodeRecord
	Contracts            []Contract
	InternalMessages     []InternalMessage
	AssetOwnership       []AssetOwnershipRecord
	StakingCapabilities  []ContractCapability
	NativeStakingInjects []NativeStakingInjectionRecord
}

type CodeRecord struct {
	CodeID    string
	CodeHash  string
	CodeBytes uint64
	Bytecode  []byte
	Owner     string
}

type Contract struct {
	AddressUser             string
	AddressRaw              string
	CodeID                  string
	CodeHash                string
	Creator                 string
	Owner                   string
	Admin                   string
	InitMsg                 []byte
	Data                    []byte
	Balance                 uint64
	StateRoot               string
	Status                  string
	StorageBytes            uint64
	LastStorageChargeHeight uint64
	StorageRentDebt         uint64
	LogicalTime             uint64
	CreatedHeight           uint64
	UpdatedHeight           uint64
}

type ContractCapability struct {
	ContractAddressUser string
	ContractAddressRaw  string
	Capability          string
	PoolID              string
	GrantedHeight       uint64
}

type InternalMessage struct {
	SourceContractUser string
	DestinationAccount string
	Funds              uint64
	Opcode             uint32
	QueryID            uint64
	Body               []byte
	Bounce             bool
	Deadline           uint64
	GasLimit           uint64
	LogicalTime        uint64
	MessageID          string
	Refunded           bool
	Height             uint64
}

type AssetOwnershipRecord struct {
	AssetType           string
	ContractAddressUser string
	Owner               string
	AssetID             string
}

type NativeStakingInjectionRecord struct {
	ContractAddressUser string
	ContractAddressRaw  string
	PoolID              string
	Amount              uint64
	Height              uint64
}

type MsgInstantiateContract struct {
	Creator      string
	CodeID       string
	InitMsg      []byte
	Funds        uint64
	Admin        string
	Salt         string
	StorageBytes uint64
	Height       uint64
}

type InstantiateContractResponse struct {
	ContractAddressUser string
	ContractAddressRaw  string
	Owner               string
	Admin               string
	Balance             uint64
	Events              []ContractEvent
}

type MsgExecuteContract struct {
	Sender          string
	ContractAddress string
	Msg             []byte
	Funds           uint64
	Height          uint64
}

type ExecuteContractResponse struct {
	ContractAddressUser string
	Owner               string
	Balance             uint64
	Events              []ContractEvent
}

type MsgTopUpContract struct {
	Sender          string
	ContractAddress string
	Amount          uint64
	Height          uint64
}

type MsgPayContractStorageDebt struct {
	Sender          string
	ContractAddress string
	Amount          uint64
	Height          uint64
}

type MsgUnfreezeContract struct {
	Sender          string
	ContractAddress string
	Height          uint64
}

type MsgGrantNativeStakingCapability struct {
	Authority           string
	ContractAddressUser string
	ContractAddressRaw  string
	PoolID              string
	Height              uint64
}

type MsgInjectNativeStaking struct {
	CallerContractUser string
	CallerContractRaw  string
	PoolID             string
	Amount             uint64
	Height             uint64
}

type MsgReceiveInternalMessage struct {
	SourceContractUser string
	DestinationAccount string
	Funds              uint64
	Opcode             uint32
	QueryID            uint64
	Body               []byte
	Bounce             bool
	Deadline           uint64
	GasLimit           uint64
	LogicalTime        uint64
	MessageID          string
	Height             uint64
}

type QueryAssetOwnerRequest struct {
	AssetType           string
	ContractAddressUser string
	AssetID             string
}

type QueryAssetOwnerResponse struct {
	Owner string
	Found bool
}

type ContractEvent struct {
	Type        string
	Actor       string
	Contract    string
	Amount      uint64
	InternalRaw string
}

func (s State) Normalize() State {
	out := cloneState(s)
	sort.SliceStable(out.Codes, func(i, j int) bool { return out.Codes[i].CodeID < out.Codes[j].CodeID })
	sort.SliceStable(out.Contracts, func(i, j int) bool { return out.Contracts[i].AddressUser < out.Contracts[j].AddressUser })
	sort.SliceStable(out.InternalMessages, func(i, j int) bool {
		if out.InternalMessages[i].Height != out.InternalMessages[j].Height {
			return out.InternalMessages[i].Height < out.InternalMessages[j].Height
		}
		if out.InternalMessages[i].LogicalTime != out.InternalMessages[j].LogicalTime {
			return out.InternalMessages[i].LogicalTime < out.InternalMessages[j].LogicalTime
		}
		if out.InternalMessages[i].MessageID != out.InternalMessages[j].MessageID {
			return out.InternalMessages[i].MessageID < out.InternalMessages[j].MessageID
		}
		if out.InternalMessages[i].SourceContractUser != out.InternalMessages[j].SourceContractUser {
			return out.InternalMessages[i].SourceContractUser < out.InternalMessages[j].SourceContractUser
		}
		return out.InternalMessages[i].DestinationAccount < out.InternalMessages[j].DestinationAccount
	})
	sort.SliceStable(out.AssetOwnership, func(i, j int) bool {
		return assetKey(out.AssetOwnership[i]) < assetKey(out.AssetOwnership[j])
	})
	sort.SliceStable(out.StakingCapabilities, func(i, j int) bool {
		return capabilityKey(out.StakingCapabilities[i]) < capabilityKey(out.StakingCapabilities[j])
	})
	sort.SliceStable(out.NativeStakingInjects, func(i, j int) bool {
		if out.NativeStakingInjects[i].Height != out.NativeStakingInjects[j].Height {
			return out.NativeStakingInjects[i].Height < out.NativeStakingInjects[j].Height
		}
		if out.NativeStakingInjects[i].PoolID != out.NativeStakingInjects[j].PoolID {
			return out.NativeStakingInjects[i].PoolID < out.NativeStakingInjects[j].PoolID
		}
		return out.NativeStakingInjects[i].ContractAddressUser < out.NativeStakingInjects[j].ContractAddressUser
	})
	return out
}

func (s State) Validate(params Params) error {
	s = s.Normalize()
	seenCodes := map[string]struct{}{}
	for _, code := range s.Codes {
		if err := code.Validate(params); err != nil {
			return err
		}
		if _, found := seenCodes[code.CodeID]; found {
			return errors.New("duplicate contract code")
		}
		seenCodes[code.CodeID] = struct{}{}
	}
	seenContracts := map[string]struct{}{}
	for _, contract := range s.Contracts {
		if err := contract.Validate(params); err != nil {
			return err
		}
		if _, found := seenCodes[contract.CodeID]; !found {
			return errors.New("contract references unknown code")
		}
		if _, found := seenContracts[contract.AddressUser]; found {
			return errors.New("duplicate contract address")
		}
		seenContracts[contract.AddressUser] = struct{}{}
	}
	for _, msg := range s.InternalMessages {
		if err := msg.Validate(); err != nil {
			return err
		}
	}
	seenAssets := map[string]struct{}{}
	for _, asset := range s.AssetOwnership {
		if err := asset.Validate(); err != nil {
			return err
		}
		key := assetKey(asset)
		if _, found := seenAssets[key]; found {
			return errors.New("duplicate contract asset ownership")
		}
		seenAssets[key] = struct{}{}
	}
	seenCaps := map[string]struct{}{}
	for _, cap := range s.StakingCapabilities {
		if err := cap.Validate(); err != nil {
			return err
		}
		if _, found := seenContracts[cap.ContractAddressUser]; !found {
			return errors.New("staking capability references unknown contract")
		}
		key := capabilityKey(cap)
		if _, found := seenCaps[key]; found {
			return errors.New("duplicate contract capability")
		}
		seenCaps[key] = struct{}{}
	}
	for _, inject := range s.NativeStakingInjects {
		if err := inject.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (c CodeRecord) Validate(params Params) error {
	if strings.TrimSpace(c.CodeID) == "" {
		return errors.New("contract code id is required")
	}
	if err := ValidateUserFacingAEAddress("contract code owner", c.Owner); err != nil {
		return err
	}
	if err := validateHashText("contract code hash", c.CodeHash); err != nil {
		return err
	}
	if c.CodeBytes == 0 || c.CodeBytes > params.MaxCodeBytes {
		return errors.New(ErrInvalidBytecode + ": code size out of bounds")
	}
	if len(c.Bytecode) > 0 {
		if err := ValidateAVMBytecode(params, c.Bytecode); err != nil {
			return err
		}
		if c.CodeBytes != uint64(len(c.Bytecode)) {
			return errors.New(ErrInvalidBytecode + ": code bytes must match bytecode length")
		}
		if c.CodeHash != CanonicalCodeHash(c.Bytecode) {
			return errors.New(ErrInvalidBytecode + ": code hash must match canonical bytecode hash")
		}
	}
	return nil
}

func (c Contract) Validate(params Params) error {
	if err := ValidateUserFacingAEAddress("contract address", c.AddressUser); err != nil {
		return err
	}
	if err := ValidateRawAddress("contract raw address", c.AddressRaw); err != nil {
		return err
	}
	if err := ValidateAddressPair("contract address pair", c.AddressUser, c.AddressRaw); err != nil {
		return err
	}
	if err := ValidateUserFacingAEAddress("contract creator", c.Creator); err != nil {
		return err
	}
	if err := ValidateUserFacingAEAddress("contract owner", c.Owner); err != nil {
		return err
	}
	if err := ValidateUserFacingAEAddress("contract admin", c.Admin); err != nil {
		return err
	}
	if strings.TrimSpace(c.CodeID) == "" {
		return errors.New("contract code id is required")
	}
	if c.CodeHash != "" {
		if err := validateHashText("contract code hash", c.CodeHash); err != nil {
			return err
		}
	}
	if c.StateRoot != "" {
		if err := validateHashText("contract state root", c.StateRoot); err != nil {
			return err
		}
	}
	if c.Status != ContractStatusActive && c.Status != ContractStatusFrozen && c.Status != ContractStatusFrozenLimited && c.Status != ContractStatusArchived && c.Status != ContractStatusDeleted {
		return fmt.Errorf("unsupported contract status %q", c.Status)
	}
	if c.StorageBytes > params.MaxContractStorageBytes {
		return errors.New(ErrStorageRent + ": contract storage exceeds configured limit")
	}
	if c.CreatedHeight == 0 || c.UpdatedHeight < c.CreatedHeight {
		return errors.New("contract heights are invalid")
	}
	if c.LogicalTime == 0 && c.Status != ContractStatusDeleted {
		return errors.New("contract logical time must be positive")
	}
	return nil
}

func (m InternalMessage) Validate() error {
	if err := ValidateUserFacingAEAddress("internal message source contract", m.SourceContractUser); err != nil {
		return err
	}
	if err := ValidateUserFacingAEAddress("internal message destination account", m.DestinationAccount); err != nil {
		return err
	}
	if m.Height == 0 {
		return errors.New("internal message height must be positive")
	}
	if len(m.Body) > MaxContractPayloadBytes {
		return errors.New("internal message body exceeds maximum size")
	}
	if m.Deadline != 0 && m.Deadline < m.Height {
		return errors.New("internal message is expired")
	}
	if m.MessageID != "" {
		if err := validateHashText("internal message id", m.MessageID); err != nil {
			return err
		}
	}
	return nil
}

func (a AssetOwnershipRecord) Validate() error {
	if a.AssetType != AssetTypeToken && a.AssetType != AssetTypeNFT && a.AssetType != AssetTypeDEX {
		return fmt.Errorf("unsupported contract asset type %q", a.AssetType)
	}
	if err := ValidateUserFacingAEAddress("asset contract address", a.ContractAddressUser); err != nil {
		return err
	}
	if err := ValidateUserFacingAEAddress("asset owner", a.Owner); err != nil {
		return err
	}
	if strings.TrimSpace(a.AssetID) == "" {
		return errors.New("asset id is required")
	}
	assetID := strings.ToLower(strings.TrimSpace(a.AssetID))
	if strings.HasPrefix(assetID, "stake_reputation") || strings.HasPrefix(assetID, "reputation/stake") {
		return errors.New("stake reputation is account-owned and cannot be transferred as token or NFT")
	}
	return nil
}

func (c ContractCapability) Validate() error {
	if err := ValidateUserFacingAEAddress("contract capability user address", c.ContractAddressUser); err != nil {
		return err
	}
	if err := ValidateRawAddress("contract capability raw address", c.ContractAddressRaw); err != nil {
		return err
	}
	if err := ValidateAddressPair("contract capability address pair", c.ContractAddressUser, c.ContractAddressRaw); err != nil {
		return err
	}
	if c.Capability != NativeStakingCapability {
		return fmt.Errorf("unsupported contract capability %q", c.Capability)
	}
	if strings.TrimSpace(c.PoolID) == "" {
		return errors.New("contract capability pool id is required")
	}
	if c.GrantedHeight == 0 {
		return errors.New("contract capability height must be positive")
	}
	return nil
}

func (n NativeStakingInjectionRecord) Validate() error {
	if err := ValidateUserFacingAEAddress("native staking injection contract", n.ContractAddressUser); err != nil {
		return err
	}
	if err := ValidateRawAddress("native staking injection raw contract", n.ContractAddressRaw); err != nil {
		return err
	}
	if err := ValidateAddressPair("native staking injection contract pair", n.ContractAddressUser, n.ContractAddressRaw); err != nil {
		return err
	}
	if strings.TrimSpace(n.PoolID) == "" {
		return errors.New("native staking injection pool id is required")
	}
	if n.Amount == 0 || n.Height == 0 {
		return errors.New("native staking injection amount and height must be positive")
	}
	return nil
}

func ValidateUserFacingAEAddress(field, text string) error {
	text = strings.TrimSpace(text)
	if !strings.HasPrefix(text, addressing.UserFriendlyPrefix) {
		return fmt.Errorf("%s must use AE user-facing address format", field)
	}
	return addressing.ValidateUserAddress(field, text)
}

func ValidateRawAddress(field, text string) error {
	text = strings.TrimSpace(text)
	if !strings.HasPrefix(text, addressing.RawPrefix) {
		return fmt.Errorf("%s must use 4: raw address format", field)
	}
	if _, err := addressing.Parse(text); err != nil {
		return fmt.Errorf("invalid %s: %w", field, err)
	}
	return nil
}

func ValidateAddressPair(field, userAddress, rawAddress string) error {
	userPair, err := addressing.PairFromUserAddress(addressing.AddressRoleAccount, userAddress)
	if err != nil {
		return err
	}
	rawPair, err := addressing.PairFromRawAddress(addressing.AddressRoleAccount, rawAddress)
	if err != nil {
		return err
	}
	if userPair.User != rawPair.User || userPair.Raw != rawPair.Raw {
		return fmt.Errorf("%s AE and raw addresses must represent the same account", field)
	}
	return nil
}

func RawAddressForUserAddress(userAddress string) (string, error) {
	pair, err := addressing.PairFromUserAddress(addressing.AddressRoleAccount, userAddress)
	if err != nil {
		return "", err
	}
	return pair.Raw, nil
}

func UserAddressForRawAddress(rawAddress string) (string, error) {
	pair, err := addressing.PairFromRawAddress(addressing.AddressRoleAccount, rawAddress)
	if err != nil {
		return "", err
	}
	return pair.User, nil
}

func DeriveContractAddress(creator string, codeID string, salt string) (string, string, error) {
	if err := ValidateUserFacingAEAddress("contract creator", creator); err != nil {
		return "", "", err
	}
	if strings.TrimSpace(codeID) == "" {
		return "", "", errors.New("contract code id is required")
	}
	sum := sha256.Sum256([]byte("aetra-contract-v1/" + creator + "/" + codeID + "/" + salt))
	user, err := addressing.FormatUserFriendly(sum[:])
	if err != nil {
		return "", "", err
	}
	return user, addressing.Format(sum[:]), nil
}

func ComputeContractStateRoot(contract Contract) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf(
		"aetra-contract-state-v1/%s/%s/%s/%020d/%x",
		contract.AddressUser,
		contract.CodeID,
		contract.CodeHash,
		contract.LogicalTime,
		contract.Data,
	)))
	return hex.EncodeToString(sum[:])
}

func ComputeInternalMessageID(msg InternalMessage) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf(
		"aetra-internal-message-v1/%s/%s/%020d/%010d/%020d/%020d/%t/%020d/%020d/%x",
		msg.SourceContractUser,
		msg.DestinationAccount,
		msg.Funds,
		msg.Opcode,
		msg.QueryID,
		msg.Height,
		msg.Bounce,
		msg.Deadline,
		msg.LogicalTime,
		msg.Body,
	)))
	return hex.EncodeToString(sum[:])
}

func RefreshStateRoot(gs GenesisState) GenesisState {
	gs.State = gs.State.Normalize()
	gs.StateRoot = ComputeContractsStateRoot(gs)
	return gs
}

func cloneState(s State) State {
	return State{
		Codes:                cloneCodes(s.Codes),
		Contracts:            cloneContracts(s.Contracts),
		InternalMessages:     cloneInternalMessages(s.InternalMessages),
		AssetOwnership:       append([]AssetOwnershipRecord(nil), s.AssetOwnership...),
		StakingCapabilities:  append([]ContractCapability(nil), s.StakingCapabilities...),
		NativeStakingInjects: append([]NativeStakingInjectionRecord(nil), s.NativeStakingInjects...),
	}
}

func cloneCodes(values []CodeRecord) []CodeRecord {
	out := append([]CodeRecord(nil), values...)
	for i := range out {
		out[i].Bytecode = append([]byte(nil), out[i].Bytecode...)
	}
	return out
}

func cloneContracts(values []Contract) []Contract {
	out := append([]Contract(nil), values...)
	for i := range out {
		out[i].InitMsg = append([]byte(nil), out[i].InitMsg...)
		out[i].Data = append([]byte(nil), out[i].Data...)
	}
	return out
}

func cloneInternalMessages(values []InternalMessage) []InternalMessage {
	out := append([]InternalMessage(nil), values...)
	for i := range out {
		out[i].Body = append([]byte(nil), out[i].Body...)
	}
	return out
}

func assetKey(a AssetOwnershipRecord) string {
	return a.AssetType + "/" + a.ContractAddressUser + "/" + a.AssetID
}

func capabilityKey(c ContractCapability) string {
	return c.ContractAddressUser + "/" + c.Capability + "/" + c.PoolID
}

func validateHashText(field string, text string) error {
	text = strings.TrimSpace(text)
	if len(text) == 64 {
		if _, err := hex.DecodeString(text); err == nil {
			return nil
		}
	}
	return fmt.Errorf("%s must be 32-byte lowercase hex", field)
}
