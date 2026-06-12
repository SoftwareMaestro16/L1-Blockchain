package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	IdentityNodeAPIProofPassthroughVersionV2	uint64	= 1
	IdentityIndexerEventSchemaVersionV2		uint64	= 1
)

type IdentityNodeAPIEndpointV2 string
type IdentityWalletSDKHelperNameV2 string
type IdentityIndexerEventTypeV2 string

const (
	IdentityNodeAPIResolvePrimary		IdentityNodeAPIEndpointV2	= "ResolvePrimaryAddress"
	IdentityNodeAPIResolveContractTarget	IdentityNodeAPIEndpointV2	= "ResolveContractTarget"
	IdentityNodeAPIResolveServiceEndpoint	IdentityNodeAPIEndpointV2	= "ResolveServiceEndpoint"
	IdentityNodeAPIResolveInterface		IdentityNodeAPIEndpointV2	= "ResolveInterfaceDescriptor"
	IdentityNodeAPIResolveRoutingMetadata	IdentityNodeAPIEndpointV2	= "ResolveRoutingMetadata"
	IdentityNodeAPIResolveReverse		IdentityNodeAPIEndpointV2	= "ResolveReverseRecord"
	IdentityNodeAPIQueryDomainLifecycle	IdentityNodeAPIEndpointV2	= "QueryDomainLifecycleState"
	IdentityNodeAPIQueryRegistrationPrice	IdentityNodeAPIEndpointV2	= "QueryRegistrationPrice"
	IdentityNodeAPIQueryRenewalPrice	IdentityNodeAPIEndpointV2	= "QueryRenewalPrice"
	IdentityNodeAPIQueryDelegationAuth	IdentityNodeAPIEndpointV2	= "QueryDelegationAuthorization"
)

const (
	IdentityWalletSDKNormalizeName			IdentityWalletSDKHelperNameV2	= "NormalizeName"
	IdentityWalletSDKValidateName			IdentityWalletSDKHelperNameV2	= "ValidateName"
	IdentityWalletSDKResolvePrimaryVerified		IdentityWalletSDKHelperNameV2	= "ResolvePrimaryVerified"
	IdentityWalletSDKResolveContractTargetVerified	IdentityWalletSDKHelperNameV2	= "ResolveContractTargetVerified"
	IdentityWalletSDKResolveServiceVerified		IdentityWalletSDKHelperNameV2	= "ResolveServiceVerified"
	IdentityWalletSDKResolveInterfaceVerified	IdentityWalletSDKHelperNameV2	= "ResolveInterfaceVerified"
	IdentityWalletSDKVerifyResolutionProof		IdentityWalletSDKHelperNameV2	= "VerifyResolutionProof"
	IdentityWalletSDKBuildSendByNameTx		IdentityWalletSDKHelperNameV2	= "BuildSendByNameTx"
	IdentityWalletSDKBuildInvokeByNameTx		IdentityWalletSDKHelperNameV2	= "BuildInvokeByNameTx"
	IdentityWalletSDKRenderVerifiedInterface	IdentityWalletSDKHelperNameV2	= "RenderVerifiedInterface"
	IdentityWalletSDKCheckReverseResolution		IdentityWalletSDKHelperNameV2	= "CheckReverseResolution"
)

const (
	IdentityIndexerEventDomainV2		IdentityIndexerEventTypeV2	= "identity.domain"
	IdentityIndexerEventOwnerV2		IdentityIndexerEventTypeV2	= "identity.owner"
	IdentityIndexerEventResolverV2		IdentityIndexerEventTypeV2	= "identity.resolver"
	IdentityIndexerEventExpiryV2		IdentityIndexerEventTypeV2	= "identity.expiry"
	IdentityIndexerEventDelegationV2	IdentityIndexerEventTypeV2	= "identity.delegation"
	IdentityIndexerEventReverseV2		IdentityIndexerEventTypeV2	= "identity.reverse"
)

const (
	IdentityIndexerAttrName			= "name"
	IdentityIndexerAttrNameHash		= "name_hash"
	IdentityIndexerAttrOwner		= "owner"
	IdentityIndexerAttrResolver		= "resolver"
	IdentityIndexerAttrExpiryHeight		= "expiry_height"
	IdentityIndexerAttrRecordVersion	= "record_version"
	IdentityIndexerAttrDelegate		= "delegate"
	IdentityIndexerAttrScope		= "scope"
	IdentityIndexerAttrAddress		= "address"
	IdentityIndexerAttrProofRef		= "proof_ref"
)

type IdentityAPIAndSDKRequirementsV2 struct {
	NodeAPIEndpoints	[]IdentityNodeAPIEndpointV2
	WalletSDKHelpers	[]IdentityWalletSDKHelperNameV2
	IndexerSchemas		[]IdentityIndexerEventSchemaV2
	RequirementsHash	string
}

type IdentityNodeAPIRequestV2 struct {
	Name			string
	TargetKey		string
	InterfaceID		string
	ExpectedHash		string
	Method			string
	PayloadHash		string
	Address			sdk.AccAddress
	IncludeProof		bool
	ChainID			string
	AppHash			string
	CurrentHeight		uint64
	FreshnessThreshold	uint64
	DurationBlocks		uint64
	RenewalPeriods		uint32
	DemandClass		IdentityDemandClassV2
	Auction			bool
	ResolverPayloadBytes	uint64
	SubdomainMode		IdentitySubdomainModeV2
	Delegation		*DelegationRecordV2
	DelegationAuth		PartialDelegationAuthorizationV2
	AuthorizedAliasKeys	[]string
}

type IdentityNodeAPIResponseV2 struct {
	Endpoint		IdentityNodeAPIEndpointV2
	QueryCode		IdentityQueryCodeV2
	FailureCode		IdentityLightClientFailureCodeV2
	Error			string
	Height			uint64
	RecordVersion		uint64
	Address			sdk.AccAddress
	ContractTarget		*ContractTargetV2
	ServiceEndpoint		*ServiceEndpointV2
	InterfaceDescriptor	*InterfaceDescriptorV2
	RoutingMetadata		*RoutingMetadataV2
	ReverseRecord		*ReverseResolutionRecordV2
	Lifecycle		DomainLifecycleStatus
	RegistrationPrice	*IdentityDomainPriceQuoteV2
	RenewalPrice		*IdentityDomainPriceQuoteV2
	DelegationAuthorized	bool
	Delegation		*DelegationRecordV2
	ProofPassthrough	*IdentityProofPassthroughFormatV2
	ResponseHash		string
}

type IdentityNodeAPIV2 struct {
	query		IdentityQueryServiceV2
	chainID		string
	appHash		string
	defaultTTL	uint64
}

type IdentityProofPassthroughFormatV2 struct {
	FormatVersion		uint64
	ChainID			string
	Height			uint64
	AppHash			string
	QueryType		IdentityProofQueryTypeV2
	Name			string
	TargetType		IdentityResolutionTargetTypeV2
	TargetKey		string
	RecordVersion		uint64
	ProofCommitmentHash	string
	ProofReference		string
	TrustlessMode		bool
	Proof			*IdentityResolutionProofFormatV2
	FormatHash		string
}

type IdentityWalletSendByNameTxV2 struct {
	Name		string
	ToAddress	sdk.AccAddress
	AmountDenom	string
	Amount		string
	Memo		string
	ProofHeight	uint64
	RecordVersion	uint64
	ProofVerified	bool
	BuildHash	string
}

type IdentityWalletInvokeByNameTxV2 struct {
	Name		string
	ContractAddress	sdk.AccAddress
	TargetID	string
	Entrypoint	string
	Method		string
	PayloadHash	string
	InterfaceID	string
	InterfaceHash	string
	ProofHeight	uint64
	RecordVersion	uint64
	ProofVerified	bool
	BuildHash	string
}

type IdentityIndexerEventSchemaV2 struct {
	SchemaVersion		uint64
	EventType		IdentityIndexerEventTypeV2
	RequiredKeys		[]string
	MaintainsIndexes	[]string
	RequiresProofReference	bool
	ProofPassthroughFormat	uint64
	SchemaHash		string
}

type IdentityIndexerEventV2 struct {
	EventType		IdentityIndexerEventTypeV2
	Height			uint64
	Attributes		map[string]string
	ProofPassthrough	*IdentityProofPassthroughFormatV2
	EventHash		string
}

type IdentityIndexerReplayResultV2 struct {
	DomainIndex	map[string]string
	OwnerIndex	map[string][]string
	ResolverIndex	map[string][]string
	ExpiryIndex	map[string][]string
	DelegationIndex	map[string][]string
	ReverseIndex	map[string]string
	ProofReferences	map[string]string
	EventsReplayed	uint64
	ReplayHash	string
}

func DefaultIdentityAPIAndSDKRequirementsV2() (IdentityAPIAndSDKRequirementsV2, error) {
	requirements := IdentityAPIAndSDKRequirementsV2{
		NodeAPIEndpoints:	requiredIdentityNodeAPIEndpointsV2(),
		WalletSDKHelpers:	requiredIdentityWalletSDKHelpersV2(),
		IndexerSchemas:		DefaultIdentityIndexerEventSchemasV2(),
	}
	requirements.RequirementsHash = ComputeIdentityAPIAndSDKRequirementsHashV2(requirements)
	return requirements, ValidateIdentityAPIAndSDKRequirementsV2(requirements)
}

func ValidateIdentityAPIAndSDKRequirementsV2(requirements IdentityAPIAndSDKRequirementsV2) error {
	if err := validateNodeAPIEndpointSetV2(requirements.NodeAPIEndpoints); err != nil {
		return err
	}
	if err := validateWalletSDKHelperSetV2(requirements.WalletSDKHelpers); err != nil {
		return err
	}
	if len(requirements.IndexerSchemas) == 0 {
		return errors.New("identity api/sdk indexer event schemas are required")
	}
	for _, schema := range requirements.IndexerSchemas {
		if err := ValidateIdentityIndexerEventSchemaV2(schema); err != nil {
			return err
		}
	}
	if requirements.RequirementsHash == "" || requirements.RequirementsHash != ComputeIdentityAPIAndSDKRequirementsHashV2(requirements) {
		return errors.New("identity api/sdk requirements hash mismatch")
	}
	return nil
}

func NewIdentityNodeAPIV2(ctx IdentityQueryContextV2, chainID string, appHash string) (IdentityNodeAPIV2, error) {
	if strings.TrimSpace(chainID) == "" {
		return IdentityNodeAPIV2{}, errors.New("identity node api chain_id is required")
	}
	if err := validateHexHash("identity node api app_hash", appHash); err != nil {
		return IdentityNodeAPIV2{}, err
	}
	query := NewIdentityQueryServiceV2(ctx)
	return IdentityNodeAPIV2{query: query, chainID: chainID, appHash: appHash, defaultTTL: query.ctx.DefaultTTL}, nil
}

func (api IdentityNodeAPIV2) ResolvePrimaryAddress(request IdentityNodeAPIRequestV2) IdentityNodeAPIResponseV2 {
	queryResp := api.query.QueryResolvePrimary(request.Name)
	resp := identityNodeAPIResponseFromQueryV2(IdentityNodeAPIResolvePrimary, queryResp)
	resp.Address = cloneSpecAddress(queryResp.Address)
	api.attachProof(&resp, request, IdentityProofQueryResolvePrimary, IdentityResolutionTargetPrimary, ResolverKeyPrimary)
	return resp.finalize()
}

func (api IdentityNodeAPIV2) ResolveContractTarget(request IdentityNodeAPIRequestV2) IdentityNodeAPIResponseV2 {
	key := request.TargetKey
	if key == "" {
		key = ResolverKeyContract
	}
	queryResp := api.query.QueryResolveContractTarget(request.Name, key)
	resp := identityNodeAPIResponseFromQueryV2(IdentityNodeAPIResolveContractTarget, queryResp)
	resp.ContractTarget = queryResp.ContractTarget
	resp.Address = cloneSpecAddress(queryResp.Address)
	api.attachProof(&resp, request, IdentityProofQueryResolveRecord, IdentityResolutionTargetContract, key)
	return resp.finalize()
}

func (api IdentityNodeAPIV2) ResolveServiceEndpoint(request IdentityNodeAPIRequestV2) IdentityNodeAPIResponseV2 {
	queryResp := api.query.QueryResolveServiceRecord(request.Name, request.TargetKey, true)
	resp := identityNodeAPIResponseFromQueryV2(IdentityNodeAPIResolveServiceEndpoint, queryResp)
	resp.ServiceEndpoint = queryResp.Service
	api.attachProof(&resp, request, IdentityProofQueryResolveRecord, IdentityResolutionTargetService, request.TargetKey)
	return resp.finalize()
}

func (api IdentityNodeAPIV2) ResolveInterfaceDescriptor(request IdentityNodeAPIRequestV2) IdentityNodeAPIResponseV2 {
	queryResp := api.query.QueryResolveInterface(request.Name, request.TargetKey)
	resp := identityNodeAPIResponseFromQueryV2(IdentityNodeAPIResolveInterface, queryResp)
	resp.InterfaceDescriptor = queryResp.Interface
	api.attachProof(&resp, request, IdentityProofQueryResolveRecord, IdentityResolutionTargetInterface, request.TargetKey)
	return resp.finalize()
}

func (api IdentityNodeAPIV2) ResolveRoutingMetadata(request IdentityNodeAPIRequestV2) IdentityNodeAPIResponseV2 {
	queryResp := api.query.QueryResolveRoute(request.Name)
	resp := identityNodeAPIResponseFromQueryV2(IdentityNodeAPIResolveRoutingMetadata, queryResp)
	resp.RoutingMetadata = queryResp.Route
	api.attachProof(&resp, request, IdentityProofQueryResolveRecord, IdentityResolutionTargetRoute, "route")
	return resp.finalize()
}

func (api IdentityNodeAPIV2) ResolveReverseRecord(request IdentityNodeAPIRequestV2) IdentityNodeAPIResponseV2 {
	queryResp := api.query.QueryVerifiedReverse(request.Address, request.AuthorizedAliasKeys)
	resp := identityNodeAPIResponseFromQueryV2(IdentityNodeAPIResolveReverse, queryResp)
	resp.ReverseRecord = queryResp.Reverse
	if request.IncludeProof && queryResp.Code == IdentityQueryOK && queryResp.Reverse != nil {
		proof, err := BuildProofModuleReverseResolutionProofV2(api.query.ctx.State, api.chainID, api.appHash, queryResp.Reverse.Name, api.query.ctx.Height, api.defaultTTL, request.Address)
		if err != nil {
			resp.QueryCode = IdentityQueryVerificationFailed
			resp.FailureCode = IdentityLightClientErrProofInvalid
			resp.Error = err.Error()
		} else {
			passthrough, err := BuildIdentityProofPassthroughFormatV2(proof, IdentityResolutionTargetPrimary, ResolverKeyPrimary, true, "")
			if err != nil {
				resp.QueryCode = IdentityQueryVerificationFailed
				resp.FailureCode = IdentityLightClientErrProofInvalid
				resp.Error = err.Error()
			} else {
				resp.ProofPassthrough = &passthrough
			}
		}
	}
	return resp.finalize()
}

func (api IdentityNodeAPIV2) QueryDomainLifecycleState(request IdentityNodeAPIRequestV2) IdentityNodeAPIResponseV2 {
	queryResp := api.query.QueryDomainLifecycle(request.Name)
	resp := identityNodeAPIResponseFromQueryV2(IdentityNodeAPIQueryDomainLifecycle, queryResp)
	resp.Lifecycle = queryResp.Lifecycle
	return resp.finalize()
}

func (api IdentityNodeAPIV2) QueryRegistrationPrice(request IdentityNodeAPIRequestV2) IdentityNodeAPIResponseV2 {
	queryResp := api.query.QueryRegistrationPrice(request.Name, request.DurationBlocks, request.DemandClass, request.Auction, request.ResolverPayloadBytes, request.SubdomainMode)
	resp := identityNodeAPIResponseFromQueryV2(IdentityNodeAPIQueryRegistrationPrice, queryResp)
	resp.RegistrationPrice = queryResp.RegistrationPrice
	return resp.finalize()
}

func (api IdentityNodeAPIV2) QueryRenewalPrice(request IdentityNodeAPIRequestV2) IdentityNodeAPIResponseV2 {
	queryResp := api.query.QueryRenewalPrice(request.Name, request.RenewalPeriods, request.ResolverPayloadBytes)
	resp := identityNodeAPIResponseFromQueryV2(IdentityNodeAPIQueryRenewalPrice, queryResp)
	resp.RenewalPrice = queryResp.RenewalPrice
	return resp.finalize()
}

func (api IdentityNodeAPIV2) QueryDelegationAuthorization(request IdentityNodeAPIRequestV2) IdentityNodeAPIResponseV2 {
	resp := IdentityNodeAPIResponseV2{Endpoint: IdentityNodeAPIQueryDelegationAuth, QueryCode: IdentityQueryOK, Height: api.query.ctx.Height}
	if request.Delegation == nil {
		resp.QueryCode = IdentityQueryInvalidRequest
		resp.Error = "identity node api delegation record is required"
		return resp.finalize()
	}
	auth := request.DelegationAuth
	if auth.Height == 0 {
		auth.Height = api.query.ctx.Height
	}
	if auth.ExpectedDelegationVersion == 0 {
		auth.ExpectedDelegationVersion = request.Delegation.DelegationVersion
	}
	if err := ValidatePartialDelegationAuthorizationV2(*request.Delegation, auth); err != nil {
		resp.QueryCode = IdentityQueryVerificationFailed
		resp.FailureCode = IdentityLightClientErrDelegationMissing
		resp.Error = err.Error()
		return resp.finalize()
	}
	resp.DelegationAuthorized = true
	delegation := cloneDelegationRecordV2(*request.Delegation)
	resp.Delegation = &delegation
	resp.RecordVersion = delegation.DelegationVersion
	return resp.finalize()
}

func BuildIdentityProofPassthroughFormatV2(proof IdentityResolutionProofFormatV2, targetType IdentityResolutionTargetTypeV2, targetKey string, trustless bool, proofReference string) (IdentityProofPassthroughFormatV2, error) {
	out := IdentityProofPassthroughFormatV2{
		FormatVersion:		IdentityNodeAPIProofPassthroughVersionV2,
		ChainID:		proof.ChainID,
		Height:			proof.Height,
		AppHash:		proof.AppHash,
		QueryType:		proof.QueryType,
		Name:			proof.Name,
		TargetType:		targetType,
		TargetKey:		targetKey,
		RecordVersion:		proof.RecordVersion,
		ProofCommitmentHash:	proof.ProofCommitmentHash,
		ProofReference:		proofReference,
		TrustlessMode:		trustless,
		Proof:			&proof,
	}
	out.FormatHash = ComputeIdentityProofPassthroughHashV2(out)
	return out, ValidateIdentityProofPassthroughFormatV2(out)
}

func ValidateIdentityProofPassthroughFormatV2(format IdentityProofPassthroughFormatV2) error {
	if format.FormatVersion != IdentityNodeAPIProofPassthroughVersionV2 {
		return errors.New("identity proof passthrough format version mismatch")
	}
	if strings.TrimSpace(format.ChainID) == "" {
		return errors.New("identity proof passthrough chain_id is required")
	}
	if format.Height == 0 {
		return errors.New("identity proof passthrough height is required")
	}
	if err := validateHexHash("identity proof passthrough app_hash", format.AppHash); err != nil {
		return err
	}
	if err := validateIdentityProofQueryTypeV2(format.QueryType); err != nil {
		return err
	}
	if _, err := NormalizeAETDomain(format.Name); err != nil {
		return err
	}
	if err := validateRoutingTargetTypeV2(string(format.TargetType)); err != nil {
		return err
	}
	if format.TargetKey != "" {
		if err := validateUnifiedRecordKey("identity proof passthrough target key", format.TargetKey); err != nil {
			return err
		}
	}
	if format.RecordVersion == 0 {
		return errors.New("identity proof passthrough record_version is required")
	}
	if err := validateHexHash("identity proof passthrough commitment", format.ProofCommitmentHash); err != nil {
		return err
	}
	if format.TrustlessMode && format.Proof == nil && format.ProofReference == "" {
		return errors.New("identity proof passthrough trustless mode requires proof or proof reference")
	}
	if format.Proof != nil {
		if format.Proof.ChainID != format.ChainID || format.Proof.Height != format.Height || format.Proof.AppHash != format.AppHash || format.Proof.ProofCommitmentHash != format.ProofCommitmentHash {
			return errors.New("identity proof passthrough proof header mismatch")
		}
		if err := ValidateIdentityResolutionProofFormatV2(*format.Proof); err != nil {
			return err
		}
	}
	if format.FormatHash == "" || format.FormatHash != ComputeIdentityProofPassthroughHashV2(format) {
		return errors.New("identity proof passthrough format hash mismatch")
	}
	return nil
}

func IdentityWalletSDKNormalizeNameV2(name string) (string, error) {
	return NormalizeAETDomain(name)
}

func IdentityWalletSDKValidateNameV2(name string) error {
	_, err := NormalizeAETDomain(name)
	return err
}

func IdentityWalletSDKResolvePrimaryVerifiedV2(request IdentitySendByNameRequestV2) (IdentitySendByNameResultV2, error) {
	if request.Proof == nil {
		return IdentitySendByNameResultV2{}, errors.New("wallet sdk resolve primary verified requires proof")
	}
	result, err := BuildIdentitySendByNameV2(request)
	if err != nil {
		return IdentitySendByNameResultV2{}, err
	}
	if !result.ProofVerified {
		return IdentitySendByNameResultV2{}, errors.New("wallet sdk resolve primary did not verify proof")
	}
	return result, nil
}

func IdentityWalletSDKResolveContractTargetVerifiedV2(request IdentityInvokeByNameRequestV2) (IdentityInvokeByNameResultV2, error) {
	if request.Proof == nil {
		return IdentityInvokeByNameResultV2{}, errors.New("wallet sdk resolve contract verified requires proof")
	}
	result, err := BuildIdentityInvokeByNameV2(request)
	if err != nil {
		return IdentityInvokeByNameResultV2{}, err
	}
	if !result.ProofVerified {
		return IdentityInvokeByNameResultV2{}, errors.New("wallet sdk resolve contract did not verify proof")
	}
	return result, nil
}

func IdentityWalletSDKResolveServiceVerifiedV2(request IdentityServiceDiscoveryRequestV2) (IdentityServiceDiscoveryResultV2, error) {
	if request.Proof == nil {
		return IdentityServiceDiscoveryResultV2{}, errors.New("wallet sdk resolve service verified requires proof")
	}
	result, err := BuildIdentityServiceDiscoveryV2(request)
	if err != nil {
		return IdentityServiceDiscoveryResultV2{}, err
	}
	if !result.ProofVerified {
		return IdentityServiceDiscoveryResultV2{}, errors.New("wallet sdk resolve service did not verify proof")
	}
	return result, nil
}

func IdentityWalletSDKResolveInterfaceVerifiedV2(request IdentityInterfaceSchemaRequestV2) (IdentityInterfaceSchemaResultV2, error) {
	if request.Proof == nil {
		return IdentityInterfaceSchemaResultV2{}, errors.New("wallet sdk resolve interface verified requires proof")
	}
	result, err := BuildIdentityInterfaceSchemaMappingV2(request)
	if err != nil {
		return IdentityInterfaceSchemaResultV2{}, err
	}
	if !result.ProofVerified || !result.SchemaHashVerified {
		return IdentityInterfaceSchemaResultV2{}, errors.New("wallet sdk resolve interface did not verify proof and schema hash")
	}
	return result, nil
}

func IdentityWalletSDKVerifyResolutionProofV2(request IdentityLightClientVerificationRequestV2) (IdentityLightClientVerifiedTargetV2, error) {
	return VerifyIdentityResolutionProofLightClientV2(request)
}

func IdentityWalletSDKBuildSendByNameTxV2(result IdentitySendByNameResultV2, amountDenom string, amount string) (IdentityWalletSendByNameTxV2, error) {
	if !result.ProofVerified {
		return IdentityWalletSendByNameTxV2{}, errors.New("wallet sdk send-by-name tx requires verified resolution")
	}
	if strings.TrimSpace(amountDenom) == "" || strings.TrimSpace(amount) == "" {
		return IdentityWalletSendByNameTxV2{}, errors.New("wallet sdk send-by-name tx amount is required")
	}
	tx := IdentityWalletSendByNameTxV2{
		Name:		result.NormalizedName,
		ToAddress:	cloneSpecAddress(result.Address),
		AmountDenom:	amountDenom,
		Amount:		amount,
		Memo:		result.AuditMemo,
		ProofHeight:	result.ProofHeight,
		RecordVersion:	result.RecordVersion,
		ProofVerified:	result.ProofVerified,
	}
	tx.BuildHash = ComputeIdentityWalletSendByNameTxHashV2(tx)
	return tx, ValidateIdentityWalletSendByNameTxV2(tx)
}

func IdentityWalletSDKBuildInvokeByNameTxV2(result IdentityInvokeByNameResultV2) (IdentityWalletInvokeByNameTxV2, error) {
	if !result.ProofVerified {
		return IdentityWalletInvokeByNameTxV2{}, errors.New("wallet sdk invoke-by-name tx requires verified resolution")
	}
	if result.InterfaceID != "" && !result.InterfaceDescriptorVerified {
		return IdentityWalletInvokeByNameTxV2{}, errors.New("wallet sdk invoke-by-name tx requires verified interface descriptor")
	}
	tx := IdentityWalletInvokeByNameTxV2{
		Name:			result.NormalizedName,
		ContractAddress:	cloneSpecAddress(result.ContractAddress),
		TargetID:		result.TargetID,
		Entrypoint:		result.Entrypoint,
		Method:			result.Method,
		PayloadHash:		result.PayloadHash,
		InterfaceID:		result.InterfaceID,
		InterfaceHash:		result.InterfaceHash,
		ProofHeight:		result.ProofHeight,
		RecordVersion:		result.RecordVersion,
		ProofVerified:		result.ProofVerified,
	}
	tx.BuildHash = ComputeIdentityWalletInvokeByNameTxHashV2(tx)
	return tx, ValidateIdentityWalletInvokeByNameTxV2(tx)
}

func IdentityWalletSDKRenderVerifiedInterfaceV2(request IdentityInterfaceSchemaRequestV2) (IdentityInterfaceSchemaResultV2, error) {
	return IdentityWalletSDKResolveInterfaceVerifiedV2(request)
}

func IdentityWalletSDKCheckReverseResolutionV2(request IdentityLightClientVerificationRequestV2) (IdentityLightClientVerifiedTargetV2, error) {
	request.RequireReverseResolution = true
	if request.TargetType == "" {
		request.TargetType = IdentityResolutionTargetPrimary
	}
	return VerifyIdentityResolutionProofLightClientV2(request)
}

func ValidateIdentityWalletSendByNameTxV2(tx IdentityWalletSendByNameTxV2) error {
	if _, err := NormalizeAETDomain(tx.Name); err != nil {
		return err
	}
	if err := validateSpecAddress("wallet sdk send-by-name address", tx.ToAddress); err != nil {
		return err
	}
	if strings.TrimSpace(tx.AmountDenom) == "" || strings.TrimSpace(tx.Amount) == "" {
		return errors.New("wallet sdk send-by-name amount is required")
	}
	if tx.ProofHeight == 0 || tx.RecordVersion == 0 || !tx.ProofVerified {
		return errors.New("wallet sdk send-by-name proof metadata is required")
	}
	if tx.BuildHash == "" || tx.BuildHash != ComputeIdentityWalletSendByNameTxHashV2(tx) {
		return errors.New("wallet sdk send-by-name build hash mismatch")
	}
	return nil
}

func ValidateIdentityWalletInvokeByNameTxV2(tx IdentityWalletInvokeByNameTxV2) error {
	if _, err := NormalizeAETDomain(tx.Name); err != nil {
		return err
	}
	if err := validateSpecAddress("wallet sdk invoke-by-name contract address", tx.ContractAddress); err != nil {
		return err
	}
	if err := validateUnifiedRecordKey("wallet sdk invoke target_id", tx.TargetID); err != nil {
		return err
	}
	if err := validateContractEntrypointV2(tx.Entrypoint); err != nil {
		return err
	}
	if tx.PayloadHash != "" {
		if err := validateHexHash("wallet sdk invoke payload hash", tx.PayloadHash); err != nil {
			return err
		}
	}
	if tx.InterfaceHash != "" {
		if err := ValidateInterfaceDescriptorHashFormatV2(tx.InterfaceHash); err != nil {
			return err
		}
	}
	if tx.ProofHeight == 0 || tx.RecordVersion == 0 || !tx.ProofVerified {
		return errors.New("wallet sdk invoke-by-name proof metadata is required")
	}
	if tx.BuildHash == "" || tx.BuildHash != ComputeIdentityWalletInvokeByNameTxHashV2(tx) {
		return errors.New("wallet sdk invoke-by-name build hash mismatch")
	}
	return nil
}

func DefaultIdentityIndexerEventSchemasV2() []IdentityIndexerEventSchemaV2 {
	schemas := []IdentityIndexerEventSchemaV2{
		newIdentityIndexerEventSchemaV2(IdentityIndexerEventDomainV2, []string{IdentityIndexerAttrName, IdentityIndexerAttrNameHash, IdentityIndexerAttrOwner, IdentityIndexerAttrExpiryHeight, IdentityIndexerAttrRecordVersion}, []string{"domain", "owner", "expiry"}),
		newIdentityIndexerEventSchemaV2(IdentityIndexerEventOwnerV2, []string{IdentityIndexerAttrNameHash, IdentityIndexerAttrOwner, IdentityIndexerAttrRecordVersion}, []string{"owner"}),
		newIdentityIndexerEventSchemaV2(IdentityIndexerEventResolverV2, []string{IdentityIndexerAttrName, IdentityIndexerAttrNameHash, IdentityIndexerAttrResolver, IdentityIndexerAttrRecordVersion}, []string{"resolver"}),
		newIdentityIndexerEventSchemaV2(IdentityIndexerEventExpiryV2, []string{IdentityIndexerAttrNameHash, IdentityIndexerAttrExpiryHeight, IdentityIndexerAttrRecordVersion}, []string{"expiry"}),
		newIdentityIndexerEventSchemaV2(IdentityIndexerEventDelegationV2, []string{IdentityIndexerAttrNameHash, IdentityIndexerAttrDelegate, IdentityIndexerAttrScope, IdentityIndexerAttrRecordVersion}, []string{"delegation"}),
		newIdentityIndexerEventSchemaV2(IdentityIndexerEventReverseV2, []string{IdentityIndexerAttrAddress, IdentityIndexerAttrNameHash, IdentityIndexerAttrRecordVersion}, []string{"reverse"}),
	}
	return schemas
}

func BuildIdentityIndexerEventV2(eventType IdentityIndexerEventTypeV2, height uint64, attributes map[string]string, proof *IdentityProofPassthroughFormatV2) (IdentityIndexerEventV2, error) {
	event := IdentityIndexerEventV2{
		EventType:		eventType,
		Height:			height,
		Attributes:		cloneStringMapV2(attributes),
		ProofPassthrough:	proof,
	}
	event.EventHash = ComputeIdentityIndexerEventHashV2(event)
	return event, ValidateIdentityIndexerEventV2(event, DefaultIdentityIndexerEventSchemasV2())
}

func ValidateIdentityIndexerEventSchemaV2(schema IdentityIndexerEventSchemaV2) error {
	if schema.SchemaVersion != IdentityIndexerEventSchemaVersionV2 {
		return errors.New("identity indexer event schema version mismatch")
	}
	if !IsIdentityIndexerEventTypeV2(schema.EventType) {
		return fmt.Errorf("unsupported identity indexer event type %q", schema.EventType)
	}
	if err := validateBreakdownTokenSetV2("identity indexer required key", schema.RequiredKeys, nil); err != nil {
		return err
	}
	if err := validateBreakdownTokenSetV2("identity indexer maintained index", schema.MaintainsIndexes, nil); err != nil {
		return err
	}
	if schema.RequiresProofReference && schema.ProofPassthroughFormat != IdentityNodeAPIProofPassthroughVersionV2 {
		return errors.New("identity indexer event schema proof passthrough version mismatch")
	}
	if schema.SchemaHash == "" || schema.SchemaHash != ComputeIdentityIndexerEventSchemaHashV2(schema) {
		return errors.New("identity indexer event schema hash mismatch")
	}
	return nil
}

func ValidateIdentityIndexerEventV2(event IdentityIndexerEventV2, schemas []IdentityIndexerEventSchemaV2) error {
	if event.Height == 0 {
		return errors.New("identity indexer event height is required")
	}
	schema, found := findIdentityIndexerSchemaV2(event.EventType, schemas)
	if !found {
		return fmt.Errorf("identity indexer schema missing for event %q", event.EventType)
	}
	if err := ValidateIdentityIndexerEventSchemaV2(schema); err != nil {
		return err
	}
	for _, key := range schema.RequiredKeys {
		if strings.TrimSpace(event.Attributes[key]) == "" {
			return fmt.Errorf("identity indexer event missing required key %q", key)
		}
	}
	if schema.RequiresProofReference {
		if event.ProofPassthrough == nil && event.Attributes[IdentityIndexerAttrProofRef] == "" {
			return errors.New("identity indexer trustless event requires proof passthrough or proof reference")
		}
		if event.ProofPassthrough != nil {
			if err := ValidateIdentityProofPassthroughFormatV2(*event.ProofPassthrough); err != nil {
				return err
			}
		}
	}
	if event.EventHash == "" || event.EventHash != ComputeIdentityIndexerEventHashV2(event) {
		return errors.New("identity indexer event hash mismatch")
	}
	return nil
}

func ReplayIdentityIndexerEventsV2(events []IdentityIndexerEventV2, schemas []IdentityIndexerEventSchemaV2) (IdentityIndexerReplayResultV2, error) {
	if len(schemas) == 0 {
		schemas = DefaultIdentityIndexerEventSchemasV2()
	}
	result := IdentityIndexerReplayResultV2{
		DomainIndex:		map[string]string{},
		OwnerIndex:		map[string][]string{},
		ResolverIndex:		map[string][]string{},
		ExpiryIndex:		map[string][]string{},
		DelegationIndex:	map[string][]string{},
		ReverseIndex:		map[string]string{},
		ProofReferences:	map[string]string{},
	}
	var lastHeight uint64
	for _, event := range events {
		if err := ValidateIdentityIndexerEventV2(event, schemas); err != nil {
			return IdentityIndexerReplayResultV2{}, err
		}
		if event.Height < lastHeight {
			return IdentityIndexerReplayResultV2{}, errors.New("identity indexer replay events must be height ordered")
		}
		lastHeight = event.Height
		attrs := event.Attributes
		nameHash := attrs[IdentityIndexerAttrNameHash]
		switch event.EventType {
		case IdentityIndexerEventDomainV2:
			result.DomainIndex[nameHash] = attrs[IdentityIndexerAttrName]
			appendUniqueIndexValueV2(result.OwnerIndex, attrs[IdentityIndexerAttrOwner], nameHash)
			appendUniqueIndexValueV2(result.ExpiryIndex, attrs[IdentityIndexerAttrExpiryHeight], nameHash)
		case IdentityIndexerEventOwnerV2:
			appendUniqueIndexValueV2(result.OwnerIndex, attrs[IdentityIndexerAttrOwner], nameHash)
		case IdentityIndexerEventResolverV2:
			appendUniqueIndexValueV2(result.ResolverIndex, attrs[IdentityIndexerAttrResolver], nameHash)
		case IdentityIndexerEventExpiryV2:
			appendUniqueIndexValueV2(result.ExpiryIndex, attrs[IdentityIndexerAttrExpiryHeight], nameHash)
		case IdentityIndexerEventDelegationV2:
			appendUniqueIndexValueV2(result.DelegationIndex, attrs[IdentityIndexerAttrDelegate]+"/"+attrs[IdentityIndexerAttrScope], nameHash)
		case IdentityIndexerEventReverseV2:
			result.ReverseIndex[attrs[IdentityIndexerAttrAddress]] = nameHash
		}
		if event.ProofPassthrough != nil {
			result.ProofReferences[event.ProofPassthrough.ProofCommitmentHash] = event.ProofPassthrough.FormatHash
		}
		if ref := attrs[IdentityIndexerAttrProofRef]; ref != "" {
			result.ProofReferences[ref] = ref
		}
		result.EventsReplayed++
	}
	sortIndexerReplayIndexesV2(&result)
	result.ReplayHash = ComputeIdentityIndexerReplayHashV2(result)
	return result, nil
}

func ComputeIdentityAPIAndSDKRequirementsHashV2(requirements IdentityAPIAndSDKRequirementsV2) string {
	parts := []string{"identity-api-sdk-requirements-v2"}
	for _, endpoint := range sortedNodeAPIEndpointsV2(requirements.NodeAPIEndpoints) {
		parts = append(parts, "endpoint", string(endpoint))
	}
	for _, helper := range sortedWalletSDKHelpersV2(requirements.WalletSDKHelpers) {
		parts = append(parts, "wallet", string(helper))
	}
	for _, schema := range sortedIdentityIndexerEventSchemasV2(requirements.IndexerSchemas) {
		parts = append(parts, "schema", schema.SchemaHash)
	}
	return identityHash(parts...)
}

func ComputeIdentityProofPassthroughHashV2(format IdentityProofPassthroughFormatV2) string {
	proofHash := ""
	if format.Proof != nil {
		proofHash = format.Proof.ProofCommitmentHash
	}
	return identityHash("identity-proof-passthrough-v2", fmt.Sprint(format.FormatVersion), format.ChainID, fmt.Sprint(format.Height), format.AppHash, string(format.QueryType), format.Name, string(format.TargetType), format.TargetKey, fmt.Sprint(format.RecordVersion), format.ProofCommitmentHash, format.ProofReference, fmt.Sprint(format.TrustlessMode), proofHash)
}

func ComputeIdentityWalletSendByNameTxHashV2(tx IdentityWalletSendByNameTxV2) string {
	return identityHash("identity-wallet-send-by-name-tx-v2", tx.Name, string(tx.ToAddress), tx.AmountDenom, tx.Amount, tx.Memo, fmt.Sprint(tx.ProofHeight), fmt.Sprint(tx.RecordVersion), fmt.Sprint(tx.ProofVerified))
}

func ComputeIdentityWalletInvokeByNameTxHashV2(tx IdentityWalletInvokeByNameTxV2) string {
	return identityHash("identity-wallet-invoke-by-name-tx-v2", tx.Name, string(tx.ContractAddress), tx.TargetID, tx.Entrypoint, tx.Method, tx.PayloadHash, tx.InterfaceID, tx.InterfaceHash, fmt.Sprint(tx.ProofHeight), fmt.Sprint(tx.RecordVersion), fmt.Sprint(tx.ProofVerified))
}

func ComputeIdentityIndexerEventSchemaHashV2(schema IdentityIndexerEventSchemaV2) string {
	parts := []string{"identity-indexer-event-schema-v2", fmt.Sprint(schema.SchemaVersion), string(schema.EventType), fmt.Sprint(schema.RequiresProofReference), fmt.Sprint(schema.ProofPassthroughFormat)}
	parts = append(parts, sortedBreakdownStringsV2(schema.RequiredKeys)...)
	parts = append(parts, sortedBreakdownStringsV2(schema.MaintainsIndexes)...)
	return identityHash(parts...)
}

func ComputeIdentityIndexerEventHashV2(event IdentityIndexerEventV2) string {
	parts := []string{"identity-indexer-event-v2", string(event.EventType), fmt.Sprint(event.Height)}
	for _, key := range sortedStringMapKeysV2(event.Attributes) {
		parts = append(parts, key, event.Attributes[key])
	}
	if event.ProofPassthrough != nil {
		parts = append(parts, event.ProofPassthrough.FormatHash)
	}
	return identityHash(parts...)
}

func ComputeIdentityIndexerReplayHashV2(result IdentityIndexerReplayResultV2) string {
	parts := []string{"identity-indexer-replay-v2", fmt.Sprint(result.EventsReplayed)}
	appendMap := func(prefix string, values map[string][]string) {
		for _, key := range sortedStringMapKeysFromSliceMapV2(values) {
			parts = append(parts, prefix, key)
			parts = append(parts, values[key]...)
		}
	}
	for _, key := range sortedStringMapKeysV2(result.DomainIndex) {
		parts = append(parts, "domain", key, result.DomainIndex[key])
	}
	appendMap("owner", result.OwnerIndex)
	appendMap("resolver", result.ResolverIndex)
	appendMap("expiry", result.ExpiryIndex)
	appendMap("delegation", result.DelegationIndex)
	for _, key := range sortedStringMapKeysV2(result.ReverseIndex) {
		parts = append(parts, "reverse", key, result.ReverseIndex[key])
	}
	for _, key := range sortedStringMapKeysV2(result.ProofReferences) {
		parts = append(parts, "proof", key, result.ProofReferences[key])
	}
	return identityHash(parts...)
}

func (api IdentityNodeAPIV2) attachProof(resp *IdentityNodeAPIResponseV2, request IdentityNodeAPIRequestV2, queryType IdentityProofQueryTypeV2, targetType IdentityResolutionTargetTypeV2, targetKey string) {
	if !request.IncludeProof || resp.QueryCode != IdentityQueryOK {
		return
	}
	proof, err := BuildIdentityResolutionProofFormatV2(api.query.ctx.State, api.chainID, api.appHash, request.Name, queryType, api.query.ctx.Height, api.defaultTTL, request.Address)
	if err != nil {
		resp.QueryCode = IdentityQueryVerificationFailed
		resp.FailureCode = IdentityLightClientErrProofInvalid
		resp.Error = err.Error()
		return
	}
	passthrough, err := BuildIdentityProofPassthroughFormatV2(proof, targetType, targetKey, true, "")
	if err != nil {
		resp.QueryCode = IdentityQueryVerificationFailed
		resp.FailureCode = IdentityLightClientErrProofInvalid
		resp.Error = err.Error()
		return
	}
	resp.ProofPassthrough = &passthrough
}

func identityNodeAPIResponseFromQueryV2(endpoint IdentityNodeAPIEndpointV2, queryResp IdentityQueryResponseV2) IdentityNodeAPIResponseV2 {
	return IdentityNodeAPIResponseV2{
		Endpoint:	endpoint,
		QueryCode:	queryResp.Code,
		FailureCode:	queryResp.FailureCode,
		Error:		queryResp.Error,
		Height:		queryResp.Height,
		RecordVersion:	queryResp.RecordVersion,
	}
}

func (resp IdentityNodeAPIResponseV2) finalize() IdentityNodeAPIResponseV2 {
	resp.ResponseHash = ComputeIdentityNodeAPIResponseHashV2(resp)
	return resp
}

func ComputeIdentityNodeAPIResponseHashV2(resp IdentityNodeAPIResponseV2) string {
	parts := []string{"identity-node-api-response-v2", string(resp.Endpoint), string(resp.QueryCode), string(resp.FailureCode), resp.Error, fmt.Sprint(resp.Height), fmt.Sprint(resp.RecordVersion), string(resp.Address), fmt.Sprint(resp.DelegationAuthorized)}
	if resp.ContractTarget != nil {
		parts = append(parts, contractTargetIDV2(*resp.ContractTarget), string(contractTargetAddressV2(*resp.ContractTarget)))
	}
	if resp.ServiceEndpoint != nil {
		parts = append(parts, serviceEndpointIDV2(*resp.ServiceEndpoint), resp.ServiceEndpoint.Endpoint)
	}
	if resp.InterfaceDescriptor != nil {
		parts = append(parts, resp.InterfaceDescriptor.InterfaceID, interfaceDescriptorSchemaHashV2(*resp.InterfaceDescriptor))
	}
	if resp.RoutingMetadata != nil {
		parts = append(parts, resp.RoutingMetadata.RouteID, resp.RoutingMetadata.TargetType, resp.RoutingMetadata.PreferredTarget)
	}
	if resp.ReverseRecord != nil {
		parts = append(parts, string(resp.ReverseRecord.Address), resp.ReverseRecord.NameHash, resp.ReverseRecord.Name, fmt.Sprint(resp.ReverseRecord.Verified))
	}
	if resp.ProofPassthrough != nil {
		parts = append(parts, resp.ProofPassthrough.FormatHash)
	}
	return identityHash(parts...)
}

func newIdentityIndexerEventSchemaV2(eventType IdentityIndexerEventTypeV2, required []string, indexes []string) IdentityIndexerEventSchemaV2 {
	schema := IdentityIndexerEventSchemaV2{
		SchemaVersion:		IdentityIndexerEventSchemaVersionV2,
		EventType:		eventType,
		RequiredKeys:		sortedBreakdownStringsV2(required),
		MaintainsIndexes:	sortedBreakdownStringsV2(indexes),
		RequiresProofReference:	true,
		ProofPassthroughFormat:	IdentityNodeAPIProofPassthroughVersionV2,
	}
	schema.SchemaHash = ComputeIdentityIndexerEventSchemaHashV2(schema)
	return schema
}

func findIdentityIndexerSchemaV2(eventType IdentityIndexerEventTypeV2, schemas []IdentityIndexerEventSchemaV2) (IdentityIndexerEventSchemaV2, bool) {
	for _, schema := range schemas {
		if schema.EventType == eventType {
			return schema, true
		}
	}
	return IdentityIndexerEventSchemaV2{}, false
}

func requiredIdentityNodeAPIEndpointsV2() []IdentityNodeAPIEndpointV2 {
	return []IdentityNodeAPIEndpointV2{
		IdentityNodeAPIQueryDelegationAuth,
		IdentityNodeAPIQueryDomainLifecycle,
		IdentityNodeAPIQueryRegistrationPrice,
		IdentityNodeAPIQueryRenewalPrice,
		IdentityNodeAPIResolveContractTarget,
		IdentityNodeAPIResolveInterface,
		IdentityNodeAPIResolvePrimary,
		IdentityNodeAPIResolveReverse,
		IdentityNodeAPIResolveRoutingMetadata,
		IdentityNodeAPIResolveServiceEndpoint,
	}
}

func requiredIdentityWalletSDKHelpersV2() []IdentityWalletSDKHelperNameV2 {
	return []IdentityWalletSDKHelperNameV2{
		IdentityWalletSDKBuildInvokeByNameTx,
		IdentityWalletSDKBuildSendByNameTx,
		IdentityWalletSDKCheckReverseResolution,
		IdentityWalletSDKNormalizeName,
		IdentityWalletSDKRenderVerifiedInterface,
		IdentityWalletSDKResolveContractTargetVerified,
		IdentityWalletSDKResolveInterfaceVerified,
		IdentityWalletSDKResolvePrimaryVerified,
		IdentityWalletSDKResolveServiceVerified,
		IdentityWalletSDKValidateName,
		IdentityWalletSDKVerifyResolutionProof,
	}
}

func validateNodeAPIEndpointSetV2(endpoints []IdentityNodeAPIEndpointV2) error {
	return validateRequiredTypedSetV2("identity node api endpoint", endpoints, requiredIdentityNodeAPIEndpointsV2(), IsIdentityNodeAPIEndpointV2)
}

func validateWalletSDKHelperSetV2(helpers []IdentityWalletSDKHelperNameV2) error {
	return validateRequiredTypedSetV2("identity wallet sdk helper", helpers, requiredIdentityWalletSDKHelpersV2(), IsIdentityWalletSDKHelperNameV2)
}

func validateRequiredTypedSetV2[T ~string](label string, values []T, required []T, valid func(T) bool) error {
	if len(values) != len(required) {
		return fmt.Errorf("%s entries mismatch", label)
	}
	seen := map[T]struct{}{}
	for _, value := range values {
		if !valid(value) {
			return fmt.Errorf("unsupported %s %q", label, value)
		}
		if _, found := seen[value]; found {
			return fmt.Errorf("duplicate %s %q", label, value)
		}
		seen[value] = struct{}{}
	}
	for _, value := range required {
		if _, found := seen[value]; !found {
			return fmt.Errorf("missing %s %q", label, value)
		}
	}
	return nil
}

func IsIdentityNodeAPIEndpointV2(endpoint IdentityNodeAPIEndpointV2) bool {
	switch endpoint {
	case IdentityNodeAPIResolvePrimary, IdentityNodeAPIResolveContractTarget, IdentityNodeAPIResolveServiceEndpoint,
		IdentityNodeAPIResolveInterface, IdentityNodeAPIResolveRoutingMetadata, IdentityNodeAPIResolveReverse,
		IdentityNodeAPIQueryDomainLifecycle, IdentityNodeAPIQueryRegistrationPrice, IdentityNodeAPIQueryRenewalPrice,
		IdentityNodeAPIQueryDelegationAuth:
		return true
	default:
		return false
	}
}

func IsIdentityWalletSDKHelperNameV2(helper IdentityWalletSDKHelperNameV2) bool {
	switch helper {
	case IdentityWalletSDKNormalizeName, IdentityWalletSDKValidateName, IdentityWalletSDKResolvePrimaryVerified,
		IdentityWalletSDKResolveContractTargetVerified, IdentityWalletSDKResolveServiceVerified, IdentityWalletSDKResolveInterfaceVerified,
		IdentityWalletSDKVerifyResolutionProof, IdentityWalletSDKBuildSendByNameTx, IdentityWalletSDKBuildInvokeByNameTx,
		IdentityWalletSDKRenderVerifiedInterface, IdentityWalletSDKCheckReverseResolution:
		return true
	default:
		return false
	}
}

func IsIdentityIndexerEventTypeV2(eventType IdentityIndexerEventTypeV2) bool {
	switch eventType {
	case IdentityIndexerEventDomainV2, IdentityIndexerEventOwnerV2, IdentityIndexerEventResolverV2,
		IdentityIndexerEventExpiryV2, IdentityIndexerEventDelegationV2, IdentityIndexerEventReverseV2:
		return true
	default:
		return false
	}
}

func sortedNodeAPIEndpointsV2(values []IdentityNodeAPIEndpointV2) []IdentityNodeAPIEndpointV2 {
	out := append([]IdentityNodeAPIEndpointV2(nil), values...)
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func sortedWalletSDKHelpersV2(values []IdentityWalletSDKHelperNameV2) []IdentityWalletSDKHelperNameV2 {
	out := append([]IdentityWalletSDKHelperNameV2(nil), values...)
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func sortedIdentityIndexerEventSchemasV2(values []IdentityIndexerEventSchemaV2) []IdentityIndexerEventSchemaV2 {
	out := append([]IdentityIndexerEventSchemaV2(nil), values...)
	sort.Slice(out, func(i, j int) bool { return out[i].EventType < out[j].EventType })
	return out
}

func sortedStringMapKeysV2(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func sortedStringMapKeysFromSliceMapV2(values map[string][]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func cloneStringMapV2(values map[string]string) map[string]string {
	out := make(map[string]string, len(values))
	for key, value := range values {
		out[key] = value
	}
	return out
}

func appendUniqueIndexValueV2(index map[string][]string, key string, value string) {
	if key == "" || value == "" {
		return
	}
	for _, existing := range index[key] {
		if existing == value {
			return
		}
	}
	index[key] = append(index[key], value)
	sort.Strings(index[key])
}

func sortIndexerReplayIndexesV2(result *IdentityIndexerReplayResultV2) {
	for _, index := range []map[string][]string{result.OwnerIndex, result.ResolverIndex, result.ExpiryIndex, result.DelegationIndex} {
		for key := range index {
			sort.Strings(index[key])
		}
	}
}
