package types

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"

	coretypes "github.com/sovereign-l1/l1/x/aetracore/types"
)

const CurrentGenesisVersion = 1

type ServiceDescriptor = coretypes.ServiceDescriptor
type ServiceAnchor = coretypes.ServiceAnchor
type ServiceInterface = coretypes.ServiceInterface
type IdentityServiceBinding = coretypes.IdentityServiceBinding
type ProviderRecord = coretypes.ProviderRecord
type ReputationRecord = coretypes.ReputationRecord
type ServiceReceipt = coretypes.ServiceReceipt
type ServiceRegistryState = coretypes.ServiceRegistryState
type ServiceRegistryProof = coretypes.ServiceRegistryProof
type ServiceConsensusContext = coretypes.ServiceConsensusContext
type ServiceCallEnvelope = coretypes.ServiceCallEnvelope
type ServiceCallKind = coretypes.ServiceCallKind
type ServiceCallStatus = coretypes.ServiceCallStatus
type ServicePaymentStatus = coretypes.ServicePaymentStatus
type ServiceExecutionOutcome = coretypes.ServiceExecutionOutcome
type ServiceCallReceipt = coretypes.ServiceCallReceipt

type MsgRegisterService = coretypes.MsgRegisterService
type MsgUpdateService = coretypes.MsgUpdateService
type MsgRenewService = coretypes.MsgRenewService
type MsgDisableService = coretypes.MsgDisableService
type MsgTransferService = coretypes.MsgTransferService
type MsgBindServiceIdentity = coretypes.MsgBindServiceIdentity
type MsgUnbindServiceIdentity = coretypes.MsgUnbindServiceIdentity
type MsgRegisterProvider = coretypes.MsgRegisterProvider
type MsgUpdateProvider = coretypes.MsgUpdateProvider
type MsgStakeProviderCollateral = coretypes.MsgStakeProviderCollateral
type MsgUnstakeProviderCollateral = coretypes.MsgUnstakeProviderCollateral
type MsgAnchorServiceReceipt = coretypes.MsgAnchorServiceReceipt
type MsgSubmitServiceDispute = coretypes.MsgSubmitServiceDispute

type QueryService = coretypes.QueryService
type QueryServiceByName = coretypes.QueryServiceByName
type QueryServicesByOwner = coretypes.QueryServicesByOwner
type QueryServicesByIdentity = coretypes.QueryServicesByIdentity
type QueryProvidersByService = coretypes.QueryProvidersByService
type QueryServiceInterface = coretypes.QueryServiceInterface
type QueryServicePaymentModel = coretypes.QueryServicePaymentModel
type QueryServiceVerificationModel = coretypes.QueryServiceVerificationModel
type QueryServiceReceipt = coretypes.QueryServiceReceipt
type QueryServiceProof = coretypes.QueryServiceProof
type QueryServiceParams = coretypes.QueryServiceParams
type QueryServiceResponse = coretypes.QueryServiceResponse
type QueryServicesResponse = coretypes.QueryServicesResponse
type QueryProvidersResponse = coretypes.QueryProvidersResponse
type QueryServiceInterfaceResponse = coretypes.QueryServiceInterfaceResponse
type QueryServicePaymentModelResponse = coretypes.QueryServicePaymentModelResponse
type QueryServiceVerificationModelResponse = coretypes.QueryServiceVerificationModelResponse
type QueryServiceReceiptResponse = coretypes.QueryServiceReceiptResponse
type QueryServiceProofResponse = coretypes.QueryServiceProofResponse
type QueryServiceParamsResponse = coretypes.QueryServiceParamsResponse

type MsgRegisterServiceResponse struct{}
type MsgUpdateServiceResponse struct{}
type MsgRenewServiceResponse struct{}
type MsgDisableServiceResponse struct{}
type MsgTransferServiceResponse struct{}
type MsgBindServiceIdentityResponse struct{}
type MsgUnbindServiceIdentityResponse struct{}
type MsgRegisterProviderResponse struct{}
type MsgUpdateProviderResponse struct{}
type MsgStakeProviderCollateralResponse struct{}
type MsgUnstakeProviderCollateralResponse struct{}
type MsgAnchorServiceReceiptResponse struct{}
type MsgSubmitServiceDisputeResponse struct{}
type MsgRegisterInterfaceResponse struct{}
type MsgUpdateInterfaceResponse struct{}

type MsgServer interface {
	RegisterService(context.Context, *MsgRegisterService) (*MsgRegisterServiceResponse, error)
	UpdateService(context.Context, *MsgUpdateService) (*MsgUpdateServiceResponse, error)
	RegisterInterface(context.Context, *MsgRegisterInterface) (*MsgRegisterInterfaceResponse, error)
	UpdateInterface(context.Context, *MsgUpdateInterface) (*MsgUpdateInterfaceResponse, error)
	RenewService(context.Context, *MsgRenewService) (*MsgRenewServiceResponse, error)
	DisableService(context.Context, *MsgDisableService) (*MsgDisableServiceResponse, error)
	TransferService(context.Context, *MsgTransferService) (*MsgTransferServiceResponse, error)
	BindServiceIdentity(context.Context, *MsgBindServiceIdentity) (*MsgBindServiceIdentityResponse, error)
	UnbindServiceIdentity(context.Context, *MsgUnbindServiceIdentity) (*MsgUnbindServiceIdentityResponse, error)
	RegisterProvider(context.Context, *MsgRegisterProvider) (*MsgRegisterProviderResponse, error)
	UpdateProvider(context.Context, *MsgUpdateProvider) (*MsgUpdateProviderResponse, error)
	StakeProviderCollateral(context.Context, *MsgStakeProviderCollateral) (*MsgStakeProviderCollateralResponse, error)
	UnstakeProviderCollateral(context.Context, *MsgUnstakeProviderCollateral) (*MsgUnstakeProviderCollateralResponse, error)
	AnchorServiceReceipt(context.Context, *MsgAnchorServiceReceipt) (*MsgAnchorServiceReceiptResponse, error)
	SubmitServiceDispute(context.Context, *MsgSubmitServiceDispute) (*MsgSubmitServiceDisputeResponse, error)
}

type QueryServer interface {
	Service(context.Context, *QueryService) (*QueryServiceResponse, error)
	ServiceByName(context.Context, *QueryServiceByName) (*QueryServiceResponse, error)
	ServicesByOwner(context.Context, *QueryServicesByOwner) (*QueryServicesResponse, error)
	ServicesByIdentity(context.Context, *QueryServicesByIdentity) (*QueryServicesResponse, error)
	ProvidersByService(context.Context, *QueryProvidersByService) (*QueryProvidersResponse, error)
	ServiceInterface(context.Context, *QueryServiceInterface) (*QueryServiceInterfaceResponse, error)
	ServicePaymentModel(context.Context, *QueryServicePaymentModel) (*QueryServicePaymentModelResponse, error)
	ServiceVerificationModel(context.Context, *QueryServiceVerificationModel) (*QueryServiceVerificationModelResponse, error)
	ServiceReceipt(context.Context, *QueryServiceReceipt) (*QueryServiceReceiptResponse, error)
	ServiceProof(context.Context, *QueryServiceProof) (*QueryServiceProofResponse, error)
	ServiceParams(context.Context, *QueryServiceParams) (*QueryServiceParamsResponse, error)
}

type ServiceDisputeRecord struct {
	DisputeID	string
	ServiceID	string
	CallID		string
	ProviderID	string
	EvidenceHash	string
	Reason		string
	OpenedHeight	uint64
	Submitter	string
	DisputeHash	string
}

type GenesisState struct {
	Version		uint64
	Params		coretypes.AetraCoreParams
	Registry	ServiceRegistryState
	Disputes	[]ServiceDisputeRecord
}

func DefaultGenesis() GenesisState {
	registry, _ := coretypes.NewServiceRegistryState(nil, nil, nil, nil, nil, nil, 1)
	return GenesisState{
		Version:	CurrentGenesisVersion,
		Params:		coretypes.TestnetParams(),
		Registry:	registry,
		Disputes:	[]ServiceDisputeRecord{},
	}
}

func (gs GenesisState) Validate() error {
	if gs.Version != CurrentGenesisVersion {
		return fmt.Errorf("services unsupported genesis version %d", gs.Version)
	}
	if err := gs.Params.Validate(); err != nil {
		return err
	}
	if err := gs.Registry.Validate(); err != nil {
		return err
	}
	return ValidateRegistryInvariants(gs.Registry)
}

func NewServiceDisputeRecord(msg MsgSubmitServiceDispute) (ServiceDisputeRecord, error) {
	if err := msg.ValidateBasic(); err != nil {
		return ServiceDisputeRecord{}, err
	}
	record := ServiceDisputeRecord{
		DisputeID:	servicesHashParts("aetra-services-dispute-id-v1", msg.ServiceID, msg.CallID, msg.ProviderID, msg.EvidenceHash),
		ServiceID:	msg.ServiceID,
		CallID:		msg.CallID,
		ProviderID:	msg.ProviderID,
		EvidenceHash:	msg.EvidenceHash,
		Reason:		msg.Reason,
		OpenedHeight:	msg.OpenedHeight,
		Submitter:	msg.Authority,
	}
	record.DisputeHash = ComputeServiceDisputeHash(record)
	return record, record.Validate()
}

func (record ServiceDisputeRecord) Validate() error {
	if record.DisputeID == "" || record.ServiceID == "" || record.CallID == "" || record.EvidenceHash == "" {
		return errors.New("services dispute requires id, service, call, and evidence")
	}
	if record.OpenedHeight == 0 {
		return errors.New("services dispute opened height must be positive")
	}
	if record.DisputeHash != ComputeServiceDisputeHash(record) {
		return errors.New("services dispute hash mismatch")
	}
	return nil
}

func ComputeServiceDisputeHash(record ServiceDisputeRecord) string {
	return servicesHashParts(
		"aetra-services-dispute-v1",
		record.DisputeID,
		record.ServiceID,
		record.CallID,
		record.ProviderID,
		record.EvidenceHash,
		record.Reason,
		fmt.Sprint(record.OpenedHeight),
		record.Submitter,
	)
}

func ValidateRegistryInvariants(state ServiceRegistryState) error {
	if err := state.Validate(); err != nil {
		return err
	}
	interfaces := map[string]struct{}{}
	for _, iface := range state.Interfaces {
		if iface.InterfaceHash != coretypes.ComputeServiceInterfaceHash(iface) {
			return fmt.Errorf("services interface hash mismatch %s", iface.InterfaceHash)
		}
		interfaces[iface.InterfaceHash] = struct{}{}
	}
	descriptors := map[string]ServiceDescriptor{}
	for _, descriptor := range state.Descriptors {
		if descriptor.Interface.InterfaceHash != coretypes.ComputeServiceInterfaceHash(descriptor.Interface) {
			return fmt.Errorf("services descriptor interface hash mismatch %s", descriptor.ServiceID)
		}
		if _, found := interfaces[descriptor.Interface.InterfaceHash]; !found {
			return fmt.Errorf("services descriptor %s missing interface index", descriptor.ServiceID)
		}
		descriptors[descriptor.ServiceID] = descriptor
	}
	for _, anchor := range state.Anchors {
		descriptor, found := descriptors[anchor.ServiceID]
		if !found {
			continue
		}
		if anchor.DescriptorHash != coretypes.ComputeServiceDescriptorHash(descriptor) {
			return fmt.Errorf("services anchor descriptor hash mismatch %s", anchor.ServiceID)
		}
		if anchor.InterfaceHash != descriptor.Interface.InterfaceHash {
			return fmt.Errorf("services anchor interface hash mismatch %s", anchor.ServiceID)
		}
	}
	return nil
}

func SortDisputes(disputes []ServiceDisputeRecord) {
	sort.SliceStable(disputes, func(i, j int) bool { return disputes[i].DisputeID < disputes[j].DisputeID })
}

func servicesHashParts(parts ...string) string {
	sum := sha256.Sum256([]byte(strings.Join(parts, "\x00")))
	return hex.EncodeToString(sum[:])
}
