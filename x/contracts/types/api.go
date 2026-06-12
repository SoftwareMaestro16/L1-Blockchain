package types

import (
	"errors"
	"fmt"
	"strings"
)

const (
	MaxContractMetadataBytes	= 1024
	MaxContractPayloadBytes		= 64 * 1024
	MaxContractQueryLimit		= 100
)

type MsgDeployContract struct {
	Creator		string
	CodeID		string
	ChainID		string
	Namespace	string
	Salt		string
	StateInit	*StateInit
	InitPayload	[]byte
	InitialBalance	uint64
	Admin		string
	Upgradeable	bool
	SystemOwned	bool
	SchemaVersion	uint64
	Metadata	[]byte
	Height		uint64
}

type MsgExecuteExternal struct {
	Sender		string
	ContractAddress	string
	ChainID		string
	Namespace	string
	StateInit	*StateInit
	Payload		[]byte
	Funds		uint64
	GasLimit	uint64
	Metadata	[]byte
	Height		uint64
}

type MsgExecuteInternal struct {
	Message	InternalMessage
	Height	uint64
}

type MsgSendInternalMessage struct {
	Message	InternalMessage
	Height	uint64
}

type MsgUpdateContractParams struct {
	Authority	string
	Params		Params
}

type MsgUpdateContractParamsResponse struct {
	StateRoot string
}

type PageRequest struct {
	Limit uint32
}

type QueryParamsRequest struct{}

type QueryParamsResponse struct {
	Params Params
}

type QueryCodeRequest struct {
	CodeID string
}

type QueryCodeResponse struct {
	Code	CodeRecord
	Found	bool
}

type QueryCodesRequest struct {
	Pagination PageRequest
}

type QueryCodesResponse struct {
	Codes []CodeRecord
}

type QueryContractsRequest struct {
	Pagination PageRequest
}

type QueryContractsResponse struct {
	Contracts []Contract
}

type QueryContractStorageRequest struct {
	ContractAddress	string
	KeyPrefix	[]byte
	Pagination	PageRequest
}

type QueryContractStorageResponse struct {
	Entries []ContractStorageEntry
}

type QueryContractReceiptsRequest struct {
	ContractAddress	string
	Pagination	PageRequest
}

type QueryContractReceiptsResponse struct {
	Receipts []ContractReceipt
}

type QueryContractQueueRequest struct {
	ContractAddress	string
	Pagination	PageRequest
}

type QueryContractQueueResponse struct {
	Messages []InternalMessage
}

type QueryContractEventsRequest struct {
	ContractAddress	string
	Pagination	PageRequest
}

type QueryContractEventsResponse struct{}

type QueryContractStateRootRequest struct {
	ContractAddress string
}

type QueryContractStateRootResponse struct {
	StateRoot string
}

func (m MsgStoreCode) ValidateBasic(params Params) error {
	if err := ValidateUserFacingAEAddress("store code authority", m.Authority); err != nil {
		return err
	}
	if len(m.Bytecode) > 0 {
		return ValidateAVMBytecode(params, m.Bytecode)
	}
	if m.CodeBytes == 0 || m.CodeBytes > params.MaxCodeBytes {
		return errors.New(ErrInvalidBytecode + ": code size out of bounds")
	}
	return validateHashText("store code hash", m.CodeHash)
}

func (m MsgDeployContract) ValidateBasic(params Params) error {
	if err := ValidateUserFacingAEAddress("deploy creator", m.Creator); err != nil {
		return err
	}
	if m.CodeID == "" {
		return errors.New("deploy code id is required")
	}
	if m.StateInit != nil {
		if err := m.StateInit.Validate(params); err != nil {
			return err
		}
	}
	if len(m.InitPayload) > MaxContractPayloadBytes {
		return errors.New("deploy payload exceeds maximum size")
	}
	if len(m.Metadata) > MaxContractMetadataBytes {
		return errors.New("deploy metadata exceeds maximum size")
	}
	if m.Admin != "" {
		if err := ValidateUserFacingAEAddress("deploy admin", m.Admin); err != nil {
			return err
		}
	}
	if m.Height == 0 {
		return errors.New("deploy height must be positive")
	}
	return nil
}

func (m MsgExecuteExternal) ValidateBasic(params Params) error {
	if err := ValidateUserFacingAEAddress("external execute sender", m.Sender); err != nil {
		return err
	}
	if err := ValidateContractAddress(m.ContractAddress); err != nil {
		return err
	}
	if m.StateInit != nil {
		if err := m.StateInit.Validate(params); err != nil {
			return err
		}
	}
	if len(m.Payload) > MaxContractPayloadBytes {
		return errors.New("external execute payload exceeds maximum size")
	}
	if len(m.Metadata) > MaxContractMetadataBytes {
		return errors.New("external execute metadata exceeds maximum size")
	}
	if m.GasLimit == 0 || m.GasLimit > params.MaxGasPerExecution {
		return errors.New("external execute gas limit out of bounds")
	}
	if m.Height == 0 {
		return errors.New("external execute height must be positive")
	}
	return nil
}

func (m MsgExecuteInternal) ValidateBasic(_ Params) error {
	if m.Height == 0 {
		return errors.New("internal execute height must be positive")
	}
	msg := m.Message
	if msg.Height == 0 {
		msg.Height = m.Height
	}
	return msg.Validate()
}

func (m MsgSendInternalMessage) ValidateBasic(_ Params) error {
	if m.Height == 0 {
		return errors.New("send internal height must be positive")
	}
	msg := m.Message
	if msg.Height == 0 {
		msg.Height = m.Height
	}
	return msg.Validate()
}

func (m MsgUpgradeContractCode) ValidateBasic(params Params) error {
	if strings.TrimSpace(m.Actor) == "" {
		return errors.New("contract upgrade actor is required")
	}
	if err := ValidateContractAddress(m.ContractAddress); err != nil {
		return err
	}
	if strings.TrimSpace(m.NewCodeID) == "" {
		return errors.New("contract upgrade code id is required")
	}
	if len(m.MigrationHandler) > MaxContractMetadataBytes {
		return errors.New("contract migration handler exceeds maximum size")
	}
	if m.Height == 0 {
		return errors.New("contract upgrade height must be positive")
	}
	_ = params
	return nil
}

func (m MsgMigrateContractState) ValidateBasic(_ Params) error {
	if strings.TrimSpace(m.Actor) == "" {
		return errors.New("contract migration actor is required")
	}
	if err := ValidateContractAddress(m.ContractAddress); err != nil {
		return err
	}
	if m.FromSchemaVersion == 0 || m.ToSchemaVersion == 0 || m.ToSchemaVersion <= m.FromSchemaVersion {
		return errors.New("contract migration schema versions are invalid")
	}
	if strings.TrimSpace(m.MigrationHandler) == "" {
		return errors.New("contract migration handler is required")
	}
	if len(m.MigrationHandler) > MaxContractMetadataBytes {
		return errors.New("contract migration handler exceeds maximum size")
	}
	if len(m.Payload) > MaxContractPayloadBytes {
		return errors.New("contract migration payload exceeds maximum size")
	}
	if m.Height == 0 {
		return errors.New("contract migration height must be positive")
	}
	return nil
}

func (m MsgSetContractAdmin) ValidateBasic(_ Params) error {
	if strings.TrimSpace(m.Actor) == "" {
		return errors.New("contract admin actor is required")
	}
	if err := ValidateContractAddress(m.ContractAddress); err != nil {
		return err
	}
	if err := ValidateUserFacingAEAddress("new contract admin", m.NewAdmin); err != nil {
		return err
	}
	if m.Height == 0 {
		return errors.New("contract admin height must be positive")
	}
	return nil
}

func (m MsgDisableContractUpgrades) ValidateBasic(_ Params) error {
	if strings.TrimSpace(m.Actor) == "" {
		return errors.New("contract upgrade disable actor is required")
	}
	if err := ValidateContractAddress(m.ContractAddress); err != nil {
		return err
	}
	if m.Height == 0 {
		return errors.New("contract upgrade disable height must be positive")
	}
	return nil
}

func (m MsgUpdateContractParams) ValidateBasic() error {
	if err := m.Params.Authorize(m.Authority); err != nil {
		return err
	}
	return m.Params.Validate()
}

func ValidateQueryPagination(req PageRequest) error {
	if req.Limit == 0 || req.Limit > MaxContractQueryLimit {
		return fmt.Errorf("query limit must be within 1..%d", MaxContractQueryLimit)
	}
	return nil
}
