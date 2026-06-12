package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/sovereign-l1/l1/app/addressing"
)

type ServiceRegistryMessage interface {
	ServiceRegistryMessageName() string
	ServiceRegistrySigner() string
	ValidateBasic() error
}

type MsgRegisterService struct {
	Authority		string
	Descriptor		ServiceDescriptor
	OwnerAuthorizationHash	string
	MessageHash		string
}

type MsgUpdateService struct {
	Authority	string
	Descriptor	ServiceDescriptor
	ExpectedVersion	uint64
	MessageHash	string
}

type MsgRenewService struct {
	Authority	string
	ServiceID	string
	ExpiryHeight	uint64
	ExpectedVersion	uint64
	MessageHash	string
}

type MsgDisableService struct {
	Authority	string
	ServiceID	string
	Reason		string
	ExpectedVersion	uint64
	MessageHash	string
}

type MsgTransferService struct {
	Authority	string
	ServiceID	string
	NewOwner	string
	ExpectedVersion	uint64
	MessageHash	string
}

type MsgBindServiceIdentity struct {
	Authority	string
	ServiceID	string
	IdentityName	string
	ExpectedVersion	uint64
	MessageHash	string
}

type MsgUnbindServiceIdentity struct {
	Authority	string
	ServiceID	string
	IdentityName	string
	ExpectedVersion	uint64
	MessageHash	string
}

type MsgRegisterProvider struct {
	Authority	string
	ServiceID	string
	Provider	FogProviderRecord
	MessageHash	string
}

type MsgUpdateProvider struct {
	Authority	string
	ServiceID	string
	Provider	FogProviderRecord
	MessageHash	string
}

type MsgStakeProviderCollateral struct {
	Authority	string
	ServiceID	string
	ProviderID	string
	Denom		string
	Amount		string
	Height		uint64
	MessageHash	string
}

type MsgUnstakeProviderCollateral struct {
	Authority	string
	ServiceID	string
	ProviderID	string
	Denom		string
	Amount		string
	Height		uint64
	MessageHash	string
}

type MsgAnchorServiceReceipt struct {
	Authority	string
	Receipt		ServiceReceipt
	AnchorHash	string
	MessageHash	string
}

type MsgSubmitServiceDispute struct {
	Authority	string
	ServiceID	string
	CallID		string
	ProviderID	string
	EvidenceHash	string
	Reason		string
	OpenedHeight	uint64
	MessageHash	string
}

type QueryPagination struct {
	Offset	uint64
	Limit	uint64
}

type QueryService struct {
	ServiceID	string
	IncludeAnchor	bool
	IncludeProof	bool
}

type QueryServiceResponse struct {
	Descriptor	ServiceDescriptor
	Anchor		ServiceAnchor
	Proof		ServiceRegistryProof
	Found		bool
}

type QueryServiceByName struct {
	ServiceName	string
	IncludeProof	bool
}

type QueryServicesByOwner struct {
	Owner		string
	Pagination	QueryPagination
}

type QueryServicesByIdentity struct {
	IdentityName	string
	Pagination	QueryPagination
}

type QueryProvidersByService struct {
	ServiceID	string
	Pagination	QueryPagination
}

type QueryServiceInterface struct {
	InterfaceHash string
}

type QueryServicePaymentModel struct {
	ServiceID string
}

type QueryServiceVerificationModel struct {
	ServiceID string
}

type QueryServiceReceipt struct {
	ServiceID	string
	CallID		string
}

type QueryServiceProof struct {
	ServiceID string
}

type QueryServiceParams struct{}

type QueryServicesResponse struct {
	Services	[]ServiceDescriptor
	Total		uint64
}

type QueryProvidersResponse struct {
	Providers	[]ProviderRecord
	Total		uint64
}

type QueryServiceInterfaceResponse struct {
	Interface	ServiceInterface
	Found		bool
}

type QueryServicePaymentModelResponse struct {
	ServiceID	string
	PaymentModel	string
	SettlementMode	ServicePaymentSettlementMode
	Denom		string
	Amount		string
	PricingUnit	ServicePricingUnit
	Found		bool
}

type QueryServiceVerificationModelResponse struct {
	ServiceID		string
	TrustModel		ServiceTrustModel
	VerificationModel	ServiceVerificationModel
	ProofFormat		string
	ChallengeWindow		uint64
	FallbackServiceID	string
	ProviderCollateral	string
	Found			bool
}

type QueryServiceReceiptResponse struct {
	Receipt	ServiceReceipt
	Found	bool
}

type QueryServiceProofResponse struct {
	Proof	ServiceRegistryProof
	Found	bool
}

type QueryServiceParamsResponse struct {
	Params AetraCoreParams
}

func ValidateServiceRegistryMessage(msg ServiceRegistryMessage) error {
	if msg == nil {
		return errors.New("aetracore service registry message is required")
	}
	return msg.ValidateBasic()
}

func NewMsgRegisterService(authority string, descriptor ServiceDescriptor, ownerAuthorizationHash string) (MsgRegisterService, error) {
	msg := MsgRegisterService{Authority: strings.TrimSpace(authority), Descriptor: CanonicalServiceDescriptor(descriptor), OwnerAuthorizationHash: strings.ToLower(strings.TrimSpace(ownerAuthorizationHash))}
	msg.MessageHash = ComputeServiceRegistryMessageHash(msg)
	return msg, msg.ValidateBasic()
}

func NewMsgUpdateService(authority string, descriptor ServiceDescriptor, expectedVersion uint64) (MsgUpdateService, error) {
	msg := MsgUpdateService{Authority: strings.TrimSpace(authority), Descriptor: CanonicalServiceDescriptor(descriptor), ExpectedVersion: expectedVersion}
	msg.MessageHash = ComputeServiceRegistryMessageHash(msg)
	return msg, msg.ValidateBasic()
}

func NewMsgRenewService(authority, serviceID string, expiryHeight, expectedVersion uint64) (MsgRenewService, error) {
	msg := MsgRenewService{Authority: strings.TrimSpace(authority), ServiceID: strings.TrimSpace(serviceID), ExpiryHeight: expiryHeight, ExpectedVersion: expectedVersion}
	msg.MessageHash = ComputeServiceRegistryMessageHash(msg)
	return msg, msg.ValidateBasic()
}

func NewMsgDisableService(authority, serviceID, reason string, expectedVersion uint64) (MsgDisableService, error) {
	msg := MsgDisableService{Authority: strings.TrimSpace(authority), ServiceID: strings.TrimSpace(serviceID), Reason: strings.TrimSpace(reason), ExpectedVersion: expectedVersion}
	msg.MessageHash = ComputeServiceRegistryMessageHash(msg)
	return msg, msg.ValidateBasic()
}

func NewMsgTransferService(authority, serviceID, newOwner string, expectedVersion uint64) (MsgTransferService, error) {
	msg := MsgTransferService{Authority: strings.TrimSpace(authority), ServiceID: strings.TrimSpace(serviceID), NewOwner: strings.TrimSpace(newOwner), ExpectedVersion: expectedVersion}
	msg.MessageHash = ComputeServiceRegistryMessageHash(msg)
	return msg, msg.ValidateBasic()
}

func NewMsgBindServiceIdentity(authority, serviceID, identityName string, expectedVersion uint64) (MsgBindServiceIdentity, error) {
	msg := MsgBindServiceIdentity{Authority: strings.TrimSpace(authority), ServiceID: strings.TrimSpace(serviceID), IdentityName: strings.TrimSpace(identityName), ExpectedVersion: expectedVersion}
	msg.MessageHash = ComputeServiceRegistryMessageHash(msg)
	return msg, msg.ValidateBasic()
}

func NewMsgUnbindServiceIdentity(authority, serviceID, identityName string, expectedVersion uint64) (MsgUnbindServiceIdentity, error) {
	msg := MsgUnbindServiceIdentity{Authority: strings.TrimSpace(authority), ServiceID: strings.TrimSpace(serviceID), IdentityName: strings.TrimSpace(identityName), ExpectedVersion: expectedVersion}
	msg.MessageHash = ComputeServiceRegistryMessageHash(msg)
	return msg, msg.ValidateBasic()
}

func NewMsgRegisterProvider(authority, serviceID string, provider FogProviderRecord) (MsgRegisterProvider, error) {
	msg := MsgRegisterProvider{Authority: strings.TrimSpace(authority), ServiceID: strings.TrimSpace(serviceID), Provider: CanonicalFogProviderRecord(provider)}
	msg.MessageHash = ComputeServiceRegistryMessageHash(msg)
	return msg, msg.ValidateBasic()
}

func NewMsgUpdateProvider(authority, serviceID string, provider FogProviderRecord) (MsgUpdateProvider, error) {
	msg := MsgUpdateProvider{Authority: strings.TrimSpace(authority), ServiceID: strings.TrimSpace(serviceID), Provider: CanonicalFogProviderRecord(provider)}
	msg.MessageHash = ComputeServiceRegistryMessageHash(msg)
	return msg, msg.ValidateBasic()
}

func NewMsgStakeProviderCollateral(authority, serviceID, providerID, denom, amount string, height uint64) (MsgStakeProviderCollateral, error) {
	msg := MsgStakeProviderCollateral{Authority: strings.TrimSpace(authority), ServiceID: strings.TrimSpace(serviceID), ProviderID: strings.TrimSpace(providerID), Denom: strings.TrimSpace(denom), Amount: strings.TrimSpace(amount), Height: height}
	msg.MessageHash = ComputeServiceRegistryMessageHash(msg)
	return msg, msg.ValidateBasic()
}

func NewMsgUnstakeProviderCollateral(authority, serviceID, providerID, denom, amount string, height uint64) (MsgUnstakeProviderCollateral, error) {
	msg := MsgUnstakeProviderCollateral{Authority: strings.TrimSpace(authority), ServiceID: strings.TrimSpace(serviceID), ProviderID: strings.TrimSpace(providerID), Denom: strings.TrimSpace(denom), Amount: strings.TrimSpace(amount), Height: height}
	msg.MessageHash = ComputeServiceRegistryMessageHash(msg)
	return msg, msg.ValidateBasic()
}

func NewMsgAnchorServiceReceipt(authority string, receipt ServiceReceipt, anchorHash string) (MsgAnchorServiceReceipt, error) {
	msg := MsgAnchorServiceReceipt{Authority: strings.TrimSpace(authority), Receipt: receipt, AnchorHash: strings.ToLower(strings.TrimSpace(anchorHash))}
	msg.MessageHash = ComputeServiceRegistryMessageHash(msg)
	return msg, msg.ValidateBasic()
}

func NewMsgSubmitServiceDispute(authority, serviceID, callID, providerID, evidenceHash, reason string, openedHeight uint64) (MsgSubmitServiceDispute, error) {
	msg := MsgSubmitServiceDispute{
		Authority:	strings.TrimSpace(authority),
		ServiceID:	strings.TrimSpace(serviceID),
		CallID:		strings.ToLower(strings.TrimSpace(callID)),
		ProviderID:	strings.TrimSpace(providerID),
		EvidenceHash:	strings.ToLower(strings.TrimSpace(evidenceHash)),
		Reason:		strings.TrimSpace(reason),
		OpenedHeight:	openedHeight,
	}
	msg.MessageHash = ComputeServiceRegistryMessageHash(msg)
	return msg, msg.ValidateBasic()
}

func (m MsgRegisterService) ServiceRegistryMessageName() string	{ return "MsgRegisterService" }
func (m MsgRegisterService) ServiceRegistrySigner() string	{ return m.Authority }
func (m MsgRegisterService) ValidateBasic() error {
	if err := validateRegistryAuthority(m.Authority); err != nil {
		return err
	}
	descriptor := CanonicalServiceDescriptor(m.Descriptor)
	if err := descriptor.Validate(); err != nil {
		return err
	}
	if descriptor.Owner != m.Authority {
		return errors.New("aetracore service registration authority must match descriptor owner")
	}
	if err := ValidateHash("aetracore service registration owner authorization", m.OwnerAuthorizationHash); err != nil {
		return err
	}
	return validateRegistryMessageHash(m, m.MessageHash)
}

func (m MsgUpdateService) ServiceRegistryMessageName() string	{ return "MsgUpdateService" }
func (m MsgUpdateService) ServiceRegistrySigner() string	{ return m.Authority }
func (m MsgUpdateService) ValidateBasic() error {
	if err := validateRegistryAuthority(m.Authority); err != nil {
		return err
	}
	descriptor := CanonicalServiceDescriptor(m.Descriptor)
	if err := descriptor.Validate(); err != nil {
		return err
	}
	if descriptor.Owner != m.Authority {
		return errors.New("aetracore service update authority must match descriptor owner")
	}
	if m.ExpectedVersion == 0 || descriptor.Version < m.ExpectedVersion {
		return errors.New("aetracore service update expected version is invalid")
	}
	return validateRegistryMessageHash(m, m.MessageHash)
}

func (m MsgRenewService) ServiceRegistryMessageName() string	{ return "MsgRenewService" }
func (m MsgRenewService) ServiceRegistrySigner() string		{ return m.Authority }
func (m MsgRenewService) ValidateBasic() error {
	if err := validateRegistryAuthority(m.Authority); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore service renewal service id", m.ServiceID); err != nil {
		return err
	}
	if m.ExpiryHeight == 0 || m.ExpectedVersion == 0 {
		return errors.New("aetracore service renewal requires expiry height and expected version")
	}
	return validateRegistryMessageHash(m, m.MessageHash)
}

func (m MsgDisableService) ServiceRegistryMessageName() string	{ return "MsgDisableService" }
func (m MsgDisableService) ServiceRegistrySigner() string	{ return m.Authority }
func (m MsgDisableService) ValidateBasic() error {
	if err := validateRegistryAuthority(m.Authority); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore service disable service id", m.ServiceID); err != nil {
		return err
	}
	if m.ExpectedVersion == 0 {
		return errors.New("aetracore service disable requires expected version")
	}
	if m.Reason != "" {
		if err := validatePolicyID("aetracore service disable reason", m.Reason); err != nil {
			return err
		}
	}
	return validateRegistryMessageHash(m, m.MessageHash)
}

func (m MsgTransferService) ServiceRegistryMessageName() string	{ return "MsgTransferService" }
func (m MsgTransferService) ServiceRegistrySigner() string	{ return m.Authority }
func (m MsgTransferService) ValidateBasic() error {
	if err := validateRegistryAuthority(m.Authority); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore service transfer service id", m.ServiceID); err != nil {
		return err
	}
	if err := addressing.ValidateAuthorityAddress("aetracore service transfer new owner", m.NewOwner); err != nil {
		return err
	}
	if m.NewOwner == m.Authority {
		return errors.New("aetracore service transfer new owner must differ")
	}
	if m.ExpectedVersion == 0 {
		return errors.New("aetracore service transfer requires expected version")
	}
	return validateRegistryMessageHash(m, m.MessageHash)
}

func (m MsgBindServiceIdentity) ServiceRegistryMessageName() string	{ return "MsgBindServiceIdentity" }
func (m MsgBindServiceIdentity) ServiceRegistrySigner() string		{ return m.Authority }
func (m MsgBindServiceIdentity) ValidateBasic() error {
	return validateServiceIdentityBindingMessage(m.Authority, m.ServiceID, m.IdentityName, m.ExpectedVersion, m, m.MessageHash)
}

func (m MsgUnbindServiceIdentity) ServiceRegistryMessageName() string {
	return "MsgUnbindServiceIdentity"
}
func (m MsgUnbindServiceIdentity) ServiceRegistrySigner() string	{ return m.Authority }
func (m MsgUnbindServiceIdentity) ValidateBasic() error {
	return validateServiceIdentityBindingMessage(m.Authority, m.ServiceID, m.IdentityName, m.ExpectedVersion, m, m.MessageHash)
}

func (m MsgRegisterProvider) ServiceRegistryMessageName() string	{ return "MsgRegisterProvider" }
func (m MsgRegisterProvider) ServiceRegistrySigner() string		{ return m.Authority }
func (m MsgRegisterProvider) ValidateBasic() error {
	return validateProviderMessage(m.Authority, m.ServiceID, m.Provider, m, m.MessageHash)
}

func (m MsgUpdateProvider) ServiceRegistryMessageName() string	{ return "MsgUpdateProvider" }
func (m MsgUpdateProvider) ServiceRegistrySigner() string	{ return m.Authority }
func (m MsgUpdateProvider) ValidateBasic() error {
	return validateProviderMessage(m.Authority, m.ServiceID, m.Provider, m, m.MessageHash)
}

func (m MsgStakeProviderCollateral) ServiceRegistryMessageName() string {
	return "MsgStakeProviderCollateral"
}
func (m MsgStakeProviderCollateral) ServiceRegistrySigner() string	{ return m.Authority }
func (m MsgStakeProviderCollateral) ValidateBasic() error {
	return validateProviderCollateralMessage(m.Authority, m.ServiceID, m.ProviderID, m.Denom, m.Amount, m.Height, m, m.MessageHash)
}

func (m MsgUnstakeProviderCollateral) ServiceRegistryMessageName() string {
	return "MsgUnstakeProviderCollateral"
}
func (m MsgUnstakeProviderCollateral) ServiceRegistrySigner() string	{ return m.Authority }
func (m MsgUnstakeProviderCollateral) ValidateBasic() error {
	return validateProviderCollateralMessage(m.Authority, m.ServiceID, m.ProviderID, m.Denom, m.Amount, m.Height, m, m.MessageHash)
}

func (m MsgAnchorServiceReceipt) ServiceRegistryMessageName() string {
	return "MsgAnchorServiceReceipt"
}
func (m MsgAnchorServiceReceipt) ServiceRegistrySigner() string	{ return m.Authority }
func (m MsgAnchorServiceReceipt) ValidateBasic() error {
	if err := validateRegistryAuthority(m.Authority); err != nil {
		return err
	}
	if err := m.Receipt.Validate(); err != nil {
		return err
	}
	if m.AnchorHash != "" {
		if err := ValidateHash("aetracore service receipt anchor hash", m.AnchorHash); err != nil {
			return err
		}
	}
	return validateRegistryMessageHash(m, m.MessageHash)
}

func (m MsgSubmitServiceDispute) ServiceRegistryMessageName() string {
	return "MsgSubmitServiceDispute"
}
func (m MsgSubmitServiceDispute) ServiceRegistrySigner() string	{ return m.Authority }
func (m MsgSubmitServiceDispute) ValidateBasic() error {
	if err := validateRegistryAuthority(m.Authority); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore service dispute service id", m.ServiceID); err != nil {
		return err
	}
	if err := ValidateHash("aetracore service dispute call id", m.CallID); err != nil {
		return err
	}
	if m.ProviderID != "" {
		if err := validatePolicyID("aetracore service dispute provider id", m.ProviderID); err != nil {
			return err
		}
	}
	if err := ValidateHash("aetracore service dispute evidence hash", m.EvidenceHash); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore service dispute reason", m.Reason); err != nil {
		return err
	}
	if m.OpenedHeight == 0 {
		return errors.New("aetracore service dispute opened height must be positive")
	}
	return validateRegistryMessageHash(m, m.MessageHash)
}

func (q QueryPagination) Normalize(params AetraCoreParams) QueryPagination {
	limit := q.Limit
	if limit == 0 {
		limit = params.DefaultQueryLimit
	}
	if limit > params.MaxQueryLimit {
		limit = params.MaxQueryLimit
	}
	return QueryPagination{Offset: q.Offset, Limit: limit}
}

func (q QueryService) Validate() error {
	return validatePolicyID("aetracore query service id", q.ServiceID)
}

func (q QueryServiceByName) Validate() error {
	return validatePolicyID("aetracore query service name", q.ServiceName)
}

func (q QueryServicesByOwner) Validate(params AetraCoreParams) error {
	if err := params.Validate(); err != nil {
		return err
	}
	if err := addressing.ValidateAuthorityAddress("aetracore query services owner", q.Owner); err != nil {
		return err
	}
	_ = q.Pagination.Normalize(params)
	return nil
}

func (q QueryServicesByIdentity) Validate(params AetraCoreParams) error {
	if err := params.Validate(); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore query identity name", q.IdentityName); err != nil {
		return err
	}
	_ = q.Pagination.Normalize(params)
	return nil
}

func (q QueryProvidersByService) Validate(params AetraCoreParams) error {
	if err := params.Validate(); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore query providers service id", q.ServiceID); err != nil {
		return err
	}
	_ = q.Pagination.Normalize(params)
	return nil
}

func (q QueryServiceInterface) Validate() error {
	return ValidateHash("aetracore query service interface hash", q.InterfaceHash)
}

func (q QueryServicePaymentModel) Validate() error {
	return validatePolicyID("aetracore query service payment model service id", q.ServiceID)
}

func (q QueryServiceVerificationModel) Validate() error {
	return validatePolicyID("aetracore query service verification model service id", q.ServiceID)
}

func (q QueryServiceReceipt) Validate() error {
	if err := validatePolicyID("aetracore query service receipt service id", q.ServiceID); err != nil {
		return err
	}
	return ValidateHash("aetracore query service receipt call id", q.CallID)
}

func (q QueryServiceProof) Validate() error {
	return validatePolicyID("aetracore query service proof service id", q.ServiceID)
}

func (q QueryServiceParams) Validate() error {
	return nil
}

func (state ServiceRegistryState) QueryService(q QueryService) (QueryServiceResponse, error) {
	if err := q.Validate(); err != nil {
		return QueryServiceResponse{}, err
	}
	descriptor, found := state.ServiceDescriptorByID(q.ServiceID)
	if !found {
		return QueryServiceResponse{Found: false}, nil
	}
	response := QueryServiceResponse{Descriptor: descriptor, Found: true}
	if q.IncludeAnchor {
		response.Anchor, _ = state.ServiceAnchorByID(q.ServiceID)
	}
	if q.IncludeProof {
		response.Proof = state.serviceProof(q.ServiceID, ComputeServiceDescriptorHash(descriptor), descriptor.Interface.InterfaceHash)
	}
	return response, nil
}

func (state ServiceRegistryState) QueryServiceByName(q QueryServiceByName) (QueryServiceResponse, error) {
	if err := q.Validate(); err != nil {
		return QueryServiceResponse{}, err
	}
	serviceID, found := state.ServiceIDByName(q.ServiceName)
	if !found {
		return QueryServiceResponse{Found: false}, nil
	}
	return state.QueryService(QueryService{ServiceID: serviceID, IncludeAnchor: true, IncludeProof: q.IncludeProof})
}

func (state ServiceRegistryState) QueryServicesByOwner(q QueryServicesByOwner, params AetraCoreParams) (QueryServicesResponse, error) {
	if err := q.Validate(params); err != nil {
		return QueryServicesResponse{}, err
	}
	ids := state.ServiceIDsByOwner(q.Owner)
	return state.queryServicesByIDs(ids, q.Pagination.Normalize(params)), nil
}

func (state ServiceRegistryState) QueryServicesByIdentity(q QueryServicesByIdentity, params AetraCoreParams) (QueryServicesResponse, error) {
	if err := q.Validate(params); err != nil {
		return QueryServicesResponse{}, err
	}
	ids := state.ServiceIDsByIdentity(q.IdentityName)
	return state.queryServicesByIDs(ids, q.Pagination.Normalize(params)), nil
}

func (state ServiceRegistryState) QueryProvidersByService(q QueryProvidersByService, params AetraCoreParams) (QueryProvidersResponse, error) {
	if err := q.Validate(params); err != nil {
		return QueryProvidersResponse{}, err
	}
	providers := state.ProviderRecordsByService(q.ServiceID)
	total := uint64(len(providers))
	window := applyProviderPagination(providers, q.Pagination.Normalize(params))
	return QueryProvidersResponse{Providers: window, Total: total}, nil
}

func (state ServiceRegistryState) QueryServiceInterface(q QueryServiceInterface) (QueryServiceInterfaceResponse, error) {
	if err := q.Validate(); err != nil {
		return QueryServiceInterfaceResponse{}, err
	}
	iface, found := state.ServiceInterfaceByHash(q.InterfaceHash)
	return QueryServiceInterfaceResponse{Interface: iface, Found: found}, nil
}

func (state ServiceRegistryState) QueryServicePaymentModel(q QueryServicePaymentModel) (QueryServicePaymentModelResponse, error) {
	if err := q.Validate(); err != nil {
		return QueryServicePaymentModelResponse{}, err
	}
	descriptor, found := state.ServiceDescriptorByID(q.ServiceID)
	if !found {
		return QueryServicePaymentModelResponse{ServiceID: q.ServiceID, Found: false}, nil
	}
	return QueryServicePaymentModelResponse{
		ServiceID:	descriptor.ServiceID,
		PaymentModel:	registryPaymentModel(descriptor),
		SettlementMode:	descriptor.Payment.SettlementMode,
		Denom:		descriptor.Payment.Denom,
		Amount:		descriptor.Payment.Amount,
		PricingUnit:	descriptor.Payment.PricingUnit,
		Found:		true,
	}, nil
}

func (state ServiceRegistryState) QueryServiceVerificationModel(q QueryServiceVerificationModel) (QueryServiceVerificationModelResponse, error) {
	if err := q.Validate(); err != nil {
		return QueryServiceVerificationModelResponse{}, err
	}
	descriptor, found := state.ServiceDescriptorByID(q.ServiceID)
	if !found {
		return QueryServiceVerificationModelResponse{ServiceID: q.ServiceID, Found: false}, nil
	}
	return QueryServiceVerificationModelResponse{
		ServiceID:		descriptor.ServiceID,
		TrustModel:		descriptor.Verification.TrustModel,
		VerificationModel:	descriptor.Verification.Model,
		ProofFormat:		descriptor.Verification.ProofFormat,
		ChallengeWindow:	descriptor.Verification.ChallengeWindow,
		FallbackServiceID:	descriptor.Verification.FallbackServiceID,
		ProviderCollateral:	descriptor.Verification.ProviderCollateralAmount,
		Found:			true,
	}, nil
}

func (state ServiceRegistryState) QueryServiceReceipt(q QueryServiceReceipt) (QueryServiceReceiptResponse, error) {
	if err := q.Validate(); err != nil {
		return QueryServiceReceiptResponse{}, err
	}
	receipt, found := state.ServiceReceiptByID(q.ServiceID, q.CallID)
	return QueryServiceReceiptResponse{Receipt: receipt, Found: found}, nil
}

func (state ServiceRegistryState) QueryServiceProof(q QueryServiceProof) (QueryServiceProofResponse, error) {
	if err := q.Validate(); err != nil {
		return QueryServiceProofResponse{}, err
	}
	descriptor, found := state.ServiceDescriptorByID(q.ServiceID)
	if !found {
		return QueryServiceProofResponse{Found: false}, nil
	}
	return QueryServiceProofResponse{Proof: state.serviceProof(q.ServiceID, ComputeServiceDescriptorHash(descriptor), descriptor.Interface.InterfaceHash), Found: true}, nil
}

func QueryServiceParamsResponseFor(params AetraCoreParams) (QueryServiceParamsResponse, error) {
	if err := params.Validate(); err != nil {
		return QueryServiceParamsResponse{}, err
	}
	return QueryServiceParamsResponse{Params: params}, nil
}

func (state ServiceRegistryState) ServiceDescriptorByID(serviceID string) (ServiceDescriptor, bool) {
	for _, descriptor := range state.Descriptors {
		if descriptor.ServiceID == serviceID {
			return descriptor, true
		}
	}
	return ServiceDescriptor{}, false
}

func (state ServiceRegistryState) ServiceAnchorByID(serviceID string) (ServiceAnchor, bool) {
	for _, anchor := range state.Anchors {
		if anchor.ServiceID == serviceID {
			return anchor, true
		}
	}
	return ServiceAnchor{}, false
}

func (state ServiceRegistryState) ServiceIDByName(serviceName string) (string, bool) {
	key, err := ServiceNameStateKey(serviceName)
	if err != nil {
		return "", false
	}
	for _, entry := range state.NameIndex {
		if entry.Key == key {
			return entry.Value, true
		}
	}
	return "", false
}

func (state ServiceRegistryState) ServiceIDsByOwner(owner string) []string {
	prefix := ServiceRegistryOwnersPrefix + "/" + strings.TrimSpace(owner) + "/"
	ids := []string{}
	for _, entry := range state.OwnerIndex {
		if strings.HasPrefix(entry.Key, prefix) {
			ids = append(ids, entry.Value)
		}
	}
	sort.Strings(ids)
	return ids
}

func (state ServiceRegistryState) ServiceIDsByIdentity(identityName string) []string {
	prefix := ServiceRegistryIdentityBindingsPrefix + "/" + strings.TrimSpace(identityName) + "/"
	ids := []string{}
	for _, binding := range state.IdentityBindings {
		if binding.IdentityName == identityName {
			ids = append(ids, binding.ServiceID)
		}
	}
	if len(ids) == 0 {
		for _, entry := range state.Entries {
			if entry.EntryType == ServiceRegistryStateIdentityBinding && strings.HasPrefix(entry.Key, prefix) {
				ids = append(ids, entry.Value)
			}
		}
	}
	sort.Strings(ids)
	return ids
}

func (state ServiceRegistryState) ProviderRecordsByService(serviceID string) []ProviderRecord {
	out := []ProviderRecord{}
	for _, provider := range state.Providers {
		if provider.ServiceID == serviceID {
			out = append(out, provider)
		}
	}
	sortProviderRecords(out)
	return out
}

func (state ServiceRegistryState) ServiceInterfaceByHash(interfaceHash string) (ServiceInterface, bool) {
	interfaceHash = strings.ToLower(strings.TrimSpace(interfaceHash))
	for _, iface := range state.Interfaces {
		if iface.InterfaceHash == interfaceHash {
			return iface, true
		}
	}
	return ServiceInterface{}, false
}

func (state ServiceRegistryState) ServiceReceiptByID(serviceID, callID string) (ServiceReceipt, bool) {
	callID = strings.ToLower(strings.TrimSpace(callID))
	for _, receipt := range state.Receipts {
		if receipt.ServiceID == serviceID && receipt.CallID == callID {
			return receipt, true
		}
	}
	return ServiceReceipt{}, false
}

func ComputeServiceRegistryMessageHash(msg ServiceRegistryMessage) string {
	if msg == nil {
		return EmptyRootHash
	}
	switch m := msg.(type) {
	case MsgRegisterService:
		return hashParts("aetra-aek-msg-register-service-v1", m.Authority, ComputeServiceDescriptorHash(m.Descriptor), m.OwnerAuthorizationHash)
	case MsgUpdateService:
		return hashParts("aetra-aek-msg-update-service-v1", m.Authority, ComputeServiceDescriptorHash(m.Descriptor), fmt.Sprint(m.ExpectedVersion))
	case MsgRenewService:
		return hashParts("aetra-aek-msg-renew-service-v1", m.Authority, m.ServiceID, fmt.Sprint(m.ExpiryHeight), fmt.Sprint(m.ExpectedVersion))
	case MsgDisableService:
		return hashParts("aetra-aek-msg-disable-service-v1", m.Authority, m.ServiceID, m.Reason, fmt.Sprint(m.ExpectedVersion))
	case MsgTransferService:
		return hashParts("aetra-aek-msg-transfer-service-v1", m.Authority, m.ServiceID, m.NewOwner, fmt.Sprint(m.ExpectedVersion))
	case MsgBindServiceIdentity:
		return hashParts("aetra-aek-msg-bind-service-identity-v1", m.Authority, m.ServiceID, m.IdentityName, fmt.Sprint(m.ExpectedVersion))
	case MsgUnbindServiceIdentity:
		return hashParts("aetra-aek-msg-unbind-service-identity-v1", m.Authority, m.ServiceID, m.IdentityName, fmt.Sprint(m.ExpectedVersion))
	case MsgRegisterProvider:
		return hashParts("aetra-aek-msg-register-provider-v1", m.Authority, m.ServiceID, m.Provider.ProviderID, m.Provider.ProviderHash)
	case MsgUpdateProvider:
		return hashParts("aetra-aek-msg-update-provider-v1", m.Authority, m.ServiceID, m.Provider.ProviderID, m.Provider.ProviderHash)
	case MsgStakeProviderCollateral:
		return hashParts("aetra-aek-msg-stake-provider-collateral-v1", m.Authority, m.ServiceID, m.ProviderID, m.Denom, m.Amount, fmt.Sprint(m.Height))
	case MsgUnstakeProviderCollateral:
		return hashParts("aetra-aek-msg-unstake-provider-collateral-v1", m.Authority, m.ServiceID, m.ProviderID, m.Denom, m.Amount, fmt.Sprint(m.Height))
	case MsgAnchorServiceReceipt:
		return hashParts("aetra-aek-msg-anchor-service-receipt-v1", m.Authority, m.Receipt.ServiceID, m.Receipt.CallID, m.Receipt.ReceiptHash, m.AnchorHash)
	case MsgSubmitServiceDispute:
		return hashParts("aetra-aek-msg-submit-service-dispute-v1", m.Authority, m.ServiceID, m.CallID, m.ProviderID, m.EvidenceHash, m.Reason, fmt.Sprint(m.OpenedHeight))
	default:
		return EmptyRootHash
	}
}

func validateRegistryAuthority(authority string) error {
	return addressing.ValidateAuthorityAddress("aetracore service registry message authority", strings.TrimSpace(authority))
}

func validateRegistryMessageHash(msg ServiceRegistryMessage, messageHash string) error {
	if err := ValidateHash("aetracore service registry message hash", messageHash); err != nil {
		return err
	}
	if expected := ComputeServiceRegistryMessageHash(msg); messageHash != expected {
		return fmt.Errorf("aetracore service registry message hash mismatch: expected %s", expected)
	}
	return nil
}

func validateServiceIdentityBindingMessage(authority, serviceID, identityName string, expectedVersion uint64, msg ServiceRegistryMessage, messageHash string) error {
	if err := validateRegistryAuthority(authority); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore service identity binding service id", serviceID); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore service identity binding identity name", identityName); err != nil {
		return err
	}
	if expectedVersion == 0 {
		return errors.New("aetracore service identity binding requires expected version")
	}
	return validateRegistryMessageHash(msg, messageHash)
}

func validateProviderMessage(authority, serviceID string, provider FogProviderRecord, msg ServiceRegistryMessage, messageHash string) error {
	if err := validateRegistryAuthority(authority); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore service provider message service id", serviceID); err != nil {
		return err
	}
	if err := provider.Validate(); err != nil {
		return err
	}
	return validateRegistryMessageHash(msg, messageHash)
}

func validateProviderCollateralMessage(authority, serviceID, providerID, denom, amount string, height uint64, msg ServiceRegistryMessage, messageHash string) error {
	if err := validateRegistryAuthority(authority); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore provider collateral service id", serviceID); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore provider collateral provider id", providerID); err != nil {
		return err
	}
	if err := validatePolicyID("aetracore provider collateral denom", denom); err != nil {
		return err
	}
	if err := validateAmountString("aetracore provider collateral amount", amount); err != nil {
		return err
	}
	if amount == "0" {
		return errors.New("aetracore provider collateral amount must be positive")
	}
	if height == 0 {
		return errors.New("aetracore provider collateral height must be positive")
	}
	return validateRegistryMessageHash(msg, messageHash)
}

func (state ServiceRegistryState) serviceProof(serviceID, descriptorHash, interfaceHash string) ServiceRegistryProof {
	key, _ := ServiceDescriptorStateKey(serviceID)
	recordHash := hashParts("aetra-aek-service-registry-state-proof-record-v1", key, descriptorHash)
	proof := ServiceRegistryProof{
		ServiceID:	serviceID,
		RegistryMode:	ServiceRegistryOnChain,
		RegistryRoot:	state.StateRoot,
		RecordHash:	recordHash,
		DescriptorHash:	descriptorHash,
		InterfaceHash:	interfaceHash,
		ProofHeight:	state.UpdatedHeight,
	}
	proof.ProofHash = ComputeServiceRegistryProofHash(proof)
	return proof
}

func (state ServiceRegistryState) queryServicesByIDs(ids []string, pagination QueryPagination) QueryServicesResponse {
	services := []ServiceDescriptor{}
	for _, id := range ids {
		if descriptor, found := state.ServiceDescriptorByID(id); found {
			services = append(services, descriptor)
		}
	}
	sortServiceDescriptors(services)
	total := uint64(len(services))
	return QueryServicesResponse{Services: applyServicePagination(services, pagination), Total: total}
}

func applyServicePagination(services []ServiceDescriptor, pagination QueryPagination) []ServiceDescriptor {
	if pagination.Offset >= uint64(len(services)) {
		return []ServiceDescriptor{}
	}
	end := pagination.Offset + pagination.Limit
	if end > uint64(len(services)) {
		end = uint64(len(services))
	}
	return append([]ServiceDescriptor(nil), services[int(pagination.Offset):int(end)]...)
}

func applyProviderPagination(providers []ProviderRecord, pagination QueryPagination) []ProviderRecord {
	if pagination.Offset >= uint64(len(providers)) {
		return []ProviderRecord{}
	}
	end := pagination.Offset + pagination.Limit
	if end > uint64(len(providers)) {
		end = uint64(len(providers))
	}
	return append([]ProviderRecord(nil), providers[int(pagination.Offset):int(end)]...)
}
