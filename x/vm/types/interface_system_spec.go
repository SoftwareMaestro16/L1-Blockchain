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
	AVMInterfaceExecutionSync	AVMInterfaceExecutionMode	= "sync"
	AVMInterfaceExecutionAsync	AVMInterfaceExecutionMode	= "async"
	AVMInterfaceExecutionScheduled	AVMInterfaceExecutionMode	= "scheduled"
	AVMInterfaceExecutionGet	AVMInterfaceExecutionMode	= "get"

	AVMInterfaceTargetContract	AVMInterfaceTargetType	= "contract"
	AVMInterfaceTargetService	AVMInterfaceTargetType	= "service"
	AVMInterfaceTargetNativeModule	AVMInterfaceTargetType	= "native_module"
	AVMInterfaceTargetWASM		AVMInterfaceTargetType	= "wasm_contract"
	AVMInterfaceTargetActor		AVMInterfaceTargetType	= "actor_contract"

	AVMInterfaceSchemaJSONSchema	AVMInterfaceSchemaEncoding	= "json_schema"
	AVMInterfaceSchemaProtobuf	AVMInterfaceSchemaEncoding	= "protobuf"
	AVMInterfaceSchemaTLB		AVMInterfaceSchemaEncoding	= "tlb"
	AVMInterfaceSchemaBinary	AVMInterfaceSchemaEncoding	= "binary"

	AVMInterfaceUseUIGeneration		AVMInterfaceUseCase	= "ui_generation"
	AVMInterfaceUseWalletForms		AVMInterfaceUseCase	= "wallet_forms"
	AVMInterfaceUseCLIAutoBinding		AVMInterfaceUseCase	= "cli_auto_binding"
	AVMInterfaceUseRPCIntrospection		AVMInterfaceUseCase	= "rpc_introspection"
	AVMInterfaceUseSDKCallBuilders		AVMInterfaceUseCase	= "sdk_call_builders"
	AVMInterfaceUseCapabilityDiscovery	AVMInterfaceUseCase	= "capability_discovery"

	AVMInterfaceSDKGo		AVMInterfaceSDKCodegenFormat	= "go"
	AVMInterfaceSDKTypeScript	AVMInterfaceSDKCodegenFormat	= "typescript"
	AVMInterfaceSDKJSON		AVMInterfaceSDKCodegenFormat	= "json"

	MaxAVMInterfaceTokenLength	= 128
	MaxAVMInterfaceVersionLength	= 32
	MaxAVMInterfaceDescriptors	= 512
)

type AVMInterfaceExecutionMode string
type AVMInterfaceTargetType string
type AVMInterfaceSchemaEncoding string
type AVMInterfaceUseCase string
type AVMInterfaceSDKCodegenFormat string

type AVMMethodDescriptor struct {
	MethodID			string
	Name				string
	InputSchemaHash			string
	OutputSchemaHash		string
	ExecutionMode			AVMInterfaceExecutionMode
	GasHint				uint64
	PaymentRequirementOptional	string
	ProofRequirementOptional	string
}

type AVMEventDescriptor struct {
	EventID		string
	Name		string
	SchemaHash	string
}

type AVMAsyncHandlerDescriptor struct {
	HandlerID		string
	Name			string
	InputSchemaHash		string
	OutputSchemaHash	string
	GasHint			uint64
	RetryPolicyOptional	string
	CallbackBehavior	string
	TimeoutHeight		uint64
}

type AVMGetMethodDescriptor struct {
	MethodID		string
	Name			string
	InputSchemaHash		string
	OutputSchemaHash	string
	GasHint			uint64
	ReadOnly		bool
}

type AVMInterfaceDescriptor struct {
	InterfaceHash		string
	InterfaceVersion	string
	Owner			string
	TargetType		AVMInterfaceTargetType
	MethodDescriptors	[]AVMMethodDescriptor
	EventDescriptors	[]AVMEventDescriptor
	AsyncHandlerDescriptors	[]AVMAsyncHandlerDescriptor
	GetMethodDescriptors	[]AVMGetMethodDescriptor
	SchemaEncoding		AVMInterfaceSchemaEncoding
	MetadataHashOptional	string
	MetadataGrantsAuth	bool
}

type AVMInterfaceSchema struct {
	InterfaceHash	string
	SchemaEncoding	AVMInterfaceSchemaEncoding
	DescriptorRoot	string
	UseCases	[]AVMInterfaceUseCase
	SchemaHash	string
}

type AVMInterfaceBinding struct {
	TargetID	string
	TargetType	AVMInterfaceTargetType
	InterfaceHash	string
	BindingHash	string
}

type AVMSDKCodeGenerationFormat struct {
	InterfaceHash		string
	Format			AVMInterfaceSDKCodegenFormat
	PackageName		string
	MethodBindings		[]string
	GetMethodBindings	[]string
	GenerationHash		string
}

type AVMInterfaceRegistry struct {
	Interfaces	[]AVMInterfaceDescriptor
	Schemas		[]AVMInterfaceSchema
	Bindings	[]AVMInterfaceBinding
	Root		string
}

func NewAVMInterfaceDescriptor(descriptor AVMInterfaceDescriptor) (AVMInterfaceDescriptor, error) {
	descriptor = canonicalAVMInterfaceDescriptor(descriptor)
	descriptor.InterfaceHash = ComputeAVMInterfaceHash(descriptor)
	return descriptor, descriptor.Validate()
}

func NewAVMInterfaceRegistry(registry AVMInterfaceRegistry) (AVMInterfaceRegistry, error) {
	registry = canonicalAVMInterfaceRegistry(registry)
	registry.Root = ComputeAVMInterfaceRegistryRoot(registry)
	return registry, registry.Validate()
}

func NewAVMInterfaceSchema(schema AVMInterfaceSchema) (AVMInterfaceSchema, error) {
	schema = canonicalAVMInterfaceSchema(schema)
	schema.SchemaHash = ComputeAVMInterfaceSchemaHash(schema)
	return schema, schema.Validate()
}

func NewAVMInterfaceBinding(binding AVMInterfaceBinding) (AVMInterfaceBinding, error) {
	binding = canonicalAVMInterfaceBinding(binding)
	binding.BindingHash = ComputeAVMInterfaceBindingHash(binding)
	return binding, binding.Validate()
}

func NewAVMSDKCodeGenerationFormat(format AVMSDKCodeGenerationFormat) (AVMSDKCodeGenerationFormat, error) {
	format = canonicalAVMSDKCodeGenerationFormat(format)
	format.GenerationHash = ComputeAVMSDKCodeGenerationHash(format)
	return format, format.Validate()
}

func (d AVMMethodDescriptor) Validate() error {
	d = canonicalAVMMethodDescriptor(d)
	if err := validateInterfaceToken("AVM method id", d.MethodID); err != nil {
		return err
	}
	if err := validateInterfaceToken("AVM method name", d.Name); err != nil {
		return err
	}
	if err := zonestypes.ValidateHash("AVM method input schema hash", d.InputSchemaHash); err != nil {
		return err
	}
	if err := zonestypes.ValidateHash("AVM method output schema hash", d.OutputSchemaHash); err != nil {
		return err
	}
	if !IsAVMInterfaceExecutionMode(d.ExecutionMode) {
		return fmt.Errorf("invalid AVM method execution mode %q", d.ExecutionMode)
	}
	if d.GasHint == 0 {
		return errors.New("AVM method gas hint must be positive")
	}
	if err := validateRouterOptionalToken("AVM method payment requirement", d.PaymentRequirementOptional, MaxAVMInterfaceTokenLength); err != nil {
		return err
	}
	return validateRouterOptionalToken("AVM method proof requirement", d.ProofRequirementOptional, MaxAVMInterfaceTokenLength)
}

func (d AVMEventDescriptor) Validate() error {
	d = canonicalAVMEventDescriptor(d)
	if err := validateInterfaceToken("AVM event id", d.EventID); err != nil {
		return err
	}
	if err := validateInterfaceToken("AVM event name", d.Name); err != nil {
		return err
	}
	return zonestypes.ValidateHash("AVM event schema hash", d.SchemaHash)
}

func (d AVMAsyncHandlerDescriptor) Validate() error {
	d = canonicalAVMAsyncHandlerDescriptor(d)
	if err := validateInterfaceToken("AVM async handler id", d.HandlerID); err != nil {
		return err
	}
	if err := validateInterfaceToken("AVM async handler name", d.Name); err != nil {
		return err
	}
	if err := zonestypes.ValidateHash("AVM async handler input schema hash", d.InputSchemaHash); err != nil {
		return err
	}
	if err := zonestypes.ValidateHash("AVM async handler output schema hash", d.OutputSchemaHash); err != nil {
		return err
	}
	if d.GasHint == 0 {
		return errors.New("AVM async handler gas hint must be positive")
	}
	if err := validateRouterOptionalToken("AVM async handler retry policy", d.RetryPolicyOptional, MaxAVMInterfaceTokenLength); err != nil {
		return err
	}
	if err := validateInterfaceToken("AVM async handler callback behavior", d.CallbackBehavior); err != nil {
		return err
	}
	if d.TimeoutHeight == 0 {
		return errors.New("AVM async handler timeout height must be positive")
	}
	return nil
}

func (d AVMGetMethodDescriptor) Validate() error {
	d = canonicalAVMGetMethodDescriptor(d)
	if err := validateInterfaceToken("AVM get method id", d.MethodID); err != nil {
		return err
	}
	if err := validateInterfaceToken("AVM get method name", d.Name); err != nil {
		return err
	}
	if err := zonestypes.ValidateHash("AVM get method input schema hash", d.InputSchemaHash); err != nil {
		return err
	}
	if err := zonestypes.ValidateHash("AVM get method output schema hash", d.OutputSchemaHash); err != nil {
		return err
	}
	if d.GasHint == 0 {
		return errors.New("AVM get method gas hint must be positive")
	}
	if !d.ReadOnly {
		return errors.New("AVM get methods must be read-only")
	}
	return nil
}

func (d AVMInterfaceDescriptor) Validate() error {
	d = canonicalAVMInterfaceDescriptor(d)
	if err := zonestypes.ValidateHash("AVM interface hash", d.InterfaceHash); err != nil {
		return err
	}
	if err := validateRouterOptionalToken("AVM interface version", d.InterfaceVersion, MaxAVMInterfaceVersionLength); err != nil {
		return err
	}
	if d.InterfaceVersion == "" {
		return errors.New("AVM interface version is required")
	}
	if err := validateInterfaceToken("AVM interface owner", d.Owner); err != nil {
		return err
	}
	if !IsAVMInterfaceTargetType(d.TargetType) {
		return fmt.Errorf("invalid AVM interface target type %q", d.TargetType)
	}
	if !IsAVMInterfaceSchemaEncoding(d.SchemaEncoding) {
		return fmt.Errorf("invalid AVM interface schema encoding %q", d.SchemaEncoding)
	}
	if d.MetadataHashOptional != "" {
		if err := zonestypes.ValidateHash("AVM interface metadata hash", d.MetadataHashOptional); err != nil {
			return err
		}
	}
	if d.MetadataGrantsAuth {
		return errors.New("AVM interface metadata cannot grant authorization")
	}
	total := len(d.MethodDescriptors) + len(d.EventDescriptors) + len(d.AsyncHandlerDescriptors) + len(d.GetMethodDescriptors)
	if total == 0 {
		return errors.New("AVM interface descriptor must expose at least one descriptor")
	}
	if total > MaxAVMInterfaceDescriptors {
		return fmt.Errorf("AVM interface descriptor entries must be <= %d", MaxAVMInterfaceDescriptors)
	}
	if err := validateAVMMethods(d.MethodDescriptors); err != nil {
		return err
	}
	if err := validateAVMEvents(d.EventDescriptors); err != nil {
		return err
	}
	if err := validateAVMAsyncHandlers(d.AsyncHandlerDescriptors); err != nil {
		return err
	}
	if err := validateAVMGetMethods(d.GetMethodDescriptors); err != nil {
		return err
	}
	if d.InterfaceHash != ComputeAVMInterfaceHash(d) {
		return errors.New("AVM interface hash mismatch")
	}
	return nil
}

func (r AVMInterfaceRegistry) Validate() error {
	r = canonicalAVMInterfaceRegistry(r)
	if len(r.Interfaces) == 0 {
		return errors.New("AVM interface registry must contain interfaces")
	}
	seen := make(map[string]struct{}, len(r.Interfaces))
	for i, descriptor := range r.Interfaces {
		if err := descriptor.Validate(); err != nil {
			return err
		}
		if _, found := seen[descriptor.InterfaceHash]; found {
			return fmt.Errorf("duplicate AVM interface hash %q", descriptor.InterfaceHash)
		}
		seen[descriptor.InterfaceHash] = struct{}{}
		if i > 0 && r.Interfaces[i-1].InterfaceHash >= descriptor.InterfaceHash {
			return errors.New("AVM interface registry entries must be sorted canonically")
		}
	}
	if err := validateAVMInterfaceSchemas(r.Schemas, seen); err != nil {
		return err
	}
	if err := validateAVMInterfaceBindings(r.Bindings, seen); err != nil {
		return err
	}
	if err := zonestypes.ValidateHash("AVM interface registry root", r.Root); err != nil {
		return err
	}
	if r.Root != ComputeAVMInterfaceRegistryRoot(r) {
		return errors.New("AVM interface registry root mismatch")
	}
	return nil
}

func (s AVMInterfaceSchema) Validate() error {
	s = canonicalAVMInterfaceSchema(s)
	if err := zonestypes.ValidateHash("AVM interface schema interface hash", s.InterfaceHash); err != nil {
		return err
	}
	if !IsAVMInterfaceSchemaEncoding(s.SchemaEncoding) {
		return fmt.Errorf("invalid AVM interface schema encoding %q", s.SchemaEncoding)
	}
	if err := zonestypes.ValidateHash("AVM interface descriptor root", s.DescriptorRoot); err != nil {
		return err
	}
	if len(s.UseCases) == 0 {
		return errors.New("AVM interface schema must declare use cases")
	}
	seen := make(map[AVMInterfaceUseCase]struct{}, len(s.UseCases))
	for i, useCase := range s.UseCases {
		if !IsAVMInterfaceUseCase(useCase) {
			return fmt.Errorf("invalid AVM interface use case %q", useCase)
		}
		if _, found := seen[useCase]; found {
			return fmt.Errorf("duplicate AVM interface use case %q", useCase)
		}
		seen[useCase] = struct{}{}
		if i > 0 && s.UseCases[i-1] >= useCase {
			return errors.New("AVM interface use cases must be sorted canonically")
		}
	}
	return validateAdapterHash("AVM interface schema", s.SchemaHash, ComputeAVMInterfaceSchemaHash(s))
}

func (b AVMInterfaceBinding) Validate() error {
	b = canonicalAVMInterfaceBinding(b)
	if err := validateInterfaceToken("AVM interface binding target id", b.TargetID); err != nil {
		return err
	}
	if !IsAVMInterfaceTargetType(b.TargetType) {
		return fmt.Errorf("invalid AVM interface binding target type %q", b.TargetType)
	}
	if err := zonestypes.ValidateHash("AVM interface binding interface hash", b.InterfaceHash); err != nil {
		return err
	}
	return validateAdapterHash("AVM interface binding", b.BindingHash, ComputeAVMInterfaceBindingHash(b))
}

func (f AVMSDKCodeGenerationFormat) Validate() error {
	f = canonicalAVMSDKCodeGenerationFormat(f)
	if err := zonestypes.ValidateHash("AVM SDK codegen interface hash", f.InterfaceHash); err != nil {
		return err
	}
	if !IsAVMInterfaceSDKCodegenFormat(f.Format) {
		return fmt.Errorf("invalid AVM SDK code generation format %q", f.Format)
	}
	if err := validateInterfaceToken("AVM SDK codegen package name", f.PackageName); err != nil {
		return err
	}
	if len(f.MethodBindings)+len(f.GetMethodBindings) == 0 {
		return errors.New("AVM SDK code generation must expose bindings")
	}
	if err := validateInterfaceBindingNames("AVM SDK method binding", f.MethodBindings); err != nil {
		return err
	}
	if err := validateInterfaceBindingNames("AVM SDK get method binding", f.GetMethodBindings); err != nil {
		return err
	}
	return validateAdapterHash("AVM SDK code generation", f.GenerationHash, ComputeAVMSDKCodeGenerationHash(f))
}

func IsAVMInterfaceExecutionMode(mode AVMInterfaceExecutionMode) bool {
	switch mode {
	case AVMInterfaceExecutionSync, AVMInterfaceExecutionAsync, AVMInterfaceExecutionScheduled, AVMInterfaceExecutionGet:
		return true
	default:
		return false
	}
}

func IsAVMInterfaceTargetType(target AVMInterfaceTargetType) bool {
	switch target {
	case AVMInterfaceTargetContract, AVMInterfaceTargetService, AVMInterfaceTargetNativeModule, AVMInterfaceTargetWASM, AVMInterfaceTargetActor:
		return true
	default:
		return false
	}
}

func IsAVMInterfaceSchemaEncoding(encoding AVMInterfaceSchemaEncoding) bool {
	switch encoding {
	case AVMInterfaceSchemaJSONSchema, AVMInterfaceSchemaProtobuf, AVMInterfaceSchemaTLB, AVMInterfaceSchemaBinary:
		return true
	default:
		return false
	}
}

func IsAVMInterfaceUseCase(useCase AVMInterfaceUseCase) bool {
	switch useCase {
	case AVMInterfaceUseUIGeneration,
		AVMInterfaceUseWalletForms,
		AVMInterfaceUseCLIAutoBinding,
		AVMInterfaceUseRPCIntrospection,
		AVMInterfaceUseSDKCallBuilders,
		AVMInterfaceUseCapabilityDiscovery:
		return true
	default:
		return false
	}
}

func IsAVMInterfaceSDKCodegenFormat(format AVMInterfaceSDKCodegenFormat) bool {
	switch format {
	case AVMInterfaceSDKGo, AVMInterfaceSDKTypeScript, AVMInterfaceSDKJSON:
		return true
	default:
		return false
	}
}

func VerifyAVMInterfaceHash(descriptor AVMInterfaceDescriptor, expectedHash string) error {
	descriptor = canonicalAVMInterfaceDescriptor(descriptor)
	if err := zonestypes.ValidateHash("expected AVM interface hash", expectedHash); err != nil {
		return err
	}
	if err := descriptor.Validate(); err != nil {
		return err
	}
	if descriptor.InterfaceHash != expectedHash || ComputeAVMInterfaceHash(descriptor) != expectedHash {
		return errors.New("AVM interface hash verification failed")
	}
	return nil
}

func VerifyAVMInterfaceSchema(descriptor AVMInterfaceDescriptor, schema AVMInterfaceSchema) error {
	descriptor = canonicalAVMInterfaceDescriptor(descriptor)
	schema = canonicalAVMInterfaceSchema(schema)
	if err := descriptor.Validate(); err != nil {
		return err
	}
	if err := schema.Validate(); err != nil {
		return err
	}
	if schema.InterfaceHash != descriptor.InterfaceHash {
		return errors.New("AVM interface schema interface hash mismatch")
	}
	if schema.SchemaEncoding != descriptor.SchemaEncoding {
		return errors.New("AVM interface schema encoding mismatch")
	}
	if schema.DescriptorRoot != ComputeAVMInterfaceDescriptorRoot(descriptor) {
		return errors.New("AVM interface schema descriptor root mismatch")
	}
	return nil
}

func QueryAVMInterfaceByTarget(registry AVMInterfaceRegistry, targetType AVMInterfaceTargetType, targetID string) (AVMInterfaceDescriptor, AVMInterfaceBinding, error) {
	registry = canonicalAVMInterfaceRegistry(registry)
	if err := registry.Validate(); err != nil {
		return AVMInterfaceDescriptor{}, AVMInterfaceBinding{}, err
	}
	targetID = strings.TrimSpace(targetID)
	for _, binding := range registry.Bindings {
		if binding.TargetType == targetType && binding.TargetID == targetID {
			for _, descriptor := range registry.Interfaces {
				if descriptor.InterfaceHash == binding.InterfaceHash {
					return descriptor, binding, nil
				}
			}
			return AVMInterfaceDescriptor{}, AVMInterfaceBinding{}, errors.New("AVM interface binding points to missing descriptor")
		}
	}
	return AVMInterfaceDescriptor{}, AVMInterfaceBinding{}, errors.New("AVM interface binding not found")
}

func QueryAVMInterfaceByContract(registry AVMInterfaceRegistry, contractID string) (AVMInterfaceDescriptor, AVMInterfaceBinding, error) {
	return QueryAVMInterfaceByTarget(registry, AVMInterfaceTargetContract, contractID)
}

func QueryAVMInterfaceByService(registry AVMInterfaceRegistry, serviceID string) (AVMInterfaceDescriptor, AVMInterfaceBinding, error) {
	return QueryAVMInterfaceByTarget(registry, AVMInterfaceTargetService, serviceID)
}

func ComputeAVMInterfaceHash(descriptor AVMInterfaceDescriptor) string {
	descriptor = canonicalAVMInterfaceDescriptor(descriptor)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-interface-descriptor-v1")
	writeEnginePart(h, descriptor.InterfaceVersion)
	writeEnginePart(h, descriptor.Owner)
	writeEnginePart(h, string(descriptor.TargetType))
	writeEngineUint64(h, uint64(len(descriptor.MethodDescriptors)))
	for _, method := range descriptor.MethodDescriptors {
		writeAVMMethodDescriptor(h, method)
	}
	writeEngineUint64(h, uint64(len(descriptor.EventDescriptors)))
	for _, event := range descriptor.EventDescriptors {
		writeEnginePart(h, event.EventID)
		writeEnginePart(h, event.Name)
		writeEnginePart(h, event.SchemaHash)
	}
	writeEngineUint64(h, uint64(len(descriptor.AsyncHandlerDescriptors)))
	for _, handler := range descriptor.AsyncHandlerDescriptors {
		writeEnginePart(h, handler.HandlerID)
		writeEnginePart(h, handler.Name)
		writeEnginePart(h, handler.InputSchemaHash)
		writeEnginePart(h, handler.OutputSchemaHash)
		writeEngineUint64(h, handler.GasHint)
		writeEnginePart(h, handler.RetryPolicyOptional)
		writeEnginePart(h, handler.CallbackBehavior)
		writeEngineUint64(h, handler.TimeoutHeight)
	}
	writeEngineUint64(h, uint64(len(descriptor.GetMethodDescriptors)))
	for _, method := range descriptor.GetMethodDescriptors {
		writeEnginePart(h, method.MethodID)
		writeEnginePart(h, method.Name)
		writeEnginePart(h, method.InputSchemaHash)
		writeEnginePart(h, method.OutputSchemaHash)
		writeEngineUint64(h, method.GasHint)
		writeEngineBool(h, method.ReadOnly)
	}
	writeEnginePart(h, string(descriptor.SchemaEncoding))
	writeEnginePart(h, descriptor.MetadataHashOptional)
	writeEngineBool(h, descriptor.MetadataGrantsAuth)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMInterfaceRegistryRoot(registry AVMInterfaceRegistry) string {
	registry = canonicalAVMInterfaceRegistry(registry)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-interface-registry-v1")
	writeEngineUint64(h, uint64(len(registry.Interfaces)))
	for _, descriptor := range registry.Interfaces {
		writeEnginePart(h, descriptor.InterfaceHash)
	}
	writeEngineUint64(h, uint64(len(registry.Schemas)))
	for _, schema := range registry.Schemas {
		writeEnginePart(h, schema.SchemaHash)
	}
	writeEngineUint64(h, uint64(len(registry.Bindings)))
	for _, binding := range registry.Bindings {
		writeEnginePart(h, binding.BindingHash)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMInterfaceDescriptorRoot(descriptor AVMInterfaceDescriptor) string {
	descriptor = canonicalAVMInterfaceDescriptor(descriptor)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-interface-descriptor-root-v1")
	writeEngineUint64(h, uint64(len(descriptor.MethodDescriptors)))
	for _, method := range descriptor.MethodDescriptors {
		writeAVMMethodDescriptor(h, method)
	}
	writeEngineUint64(h, uint64(len(descriptor.EventDescriptors)))
	for _, event := range descriptor.EventDescriptors {
		writeEnginePart(h, event.EventID)
		writeEnginePart(h, event.SchemaHash)
	}
	writeEngineUint64(h, uint64(len(descriptor.AsyncHandlerDescriptors)))
	for _, handler := range descriptor.AsyncHandlerDescriptors {
		writeEnginePart(h, handler.HandlerID)
		writeEnginePart(h, handler.InputSchemaHash)
		writeEnginePart(h, handler.OutputSchemaHash)
	}
	writeEngineUint64(h, uint64(len(descriptor.GetMethodDescriptors)))
	for _, method := range descriptor.GetMethodDescriptors {
		writeEnginePart(h, method.MethodID)
		writeEnginePart(h, method.InputSchemaHash)
		writeEnginePart(h, method.OutputSchemaHash)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMInterfaceSchemaHash(schema AVMInterfaceSchema) string {
	schema = canonicalAVMInterfaceSchema(schema)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-interface-schema-v1")
	writeEnginePart(h, schema.InterfaceHash)
	writeEnginePart(h, string(schema.SchemaEncoding))
	writeEnginePart(h, schema.DescriptorRoot)
	writeEngineUint64(h, uint64(len(schema.UseCases)))
	for _, useCase := range schema.UseCases {
		writeEnginePart(h, string(useCase))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMInterfaceBindingHash(binding AVMInterfaceBinding) string {
	binding = canonicalAVMInterfaceBinding(binding)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-interface-binding-v1")
	writeEnginePart(h, binding.TargetID)
	writeEnginePart(h, string(binding.TargetType))
	writeEnginePart(h, binding.InterfaceHash)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMSDKCodeGenerationHash(format AVMSDKCodeGenerationFormat) string {
	format = canonicalAVMSDKCodeGenerationFormat(format)
	h := sha256.New()
	writeEnginePart(h, "aetra-avm-interface-sdk-codegen-v1")
	writeEnginePart(h, format.InterfaceHash)
	writeEnginePart(h, string(format.Format))
	writeEnginePart(h, format.PackageName)
	writeEngineUint64(h, uint64(len(format.MethodBindings)))
	for _, binding := range format.MethodBindings {
		writeEnginePart(h, binding)
	}
	writeEngineUint64(h, uint64(len(format.GetMethodBindings)))
	for _, binding := range format.GetMethodBindings {
		writeEnginePart(h, binding)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func canonicalAVMInterfaceDescriptor(descriptor AVMInterfaceDescriptor) AVMInterfaceDescriptor {
	descriptor.InterfaceHash = strings.TrimSpace(descriptor.InterfaceHash)
	descriptor.InterfaceVersion = strings.TrimSpace(descriptor.InterfaceVersion)
	descriptor.Owner = strings.TrimSpace(descriptor.Owner)
	descriptor.MetadataHashOptional = strings.TrimSpace(descriptor.MetadataHashOptional)
	descriptor.MethodDescriptors = append([]AVMMethodDescriptor(nil), descriptor.MethodDescriptors...)
	for i := range descriptor.MethodDescriptors {
		descriptor.MethodDescriptors[i] = canonicalAVMMethodDescriptor(descriptor.MethodDescriptors[i])
	}
	sort.SliceStable(descriptor.MethodDescriptors, func(i, j int) bool {
		return descriptor.MethodDescriptors[i].MethodID < descriptor.MethodDescriptors[j].MethodID
	})
	descriptor.EventDescriptors = append([]AVMEventDescriptor(nil), descriptor.EventDescriptors...)
	for i := range descriptor.EventDescriptors {
		descriptor.EventDescriptors[i] = canonicalAVMEventDescriptor(descriptor.EventDescriptors[i])
	}
	sort.SliceStable(descriptor.EventDescriptors, func(i, j int) bool {
		return descriptor.EventDescriptors[i].EventID < descriptor.EventDescriptors[j].EventID
	})
	descriptor.AsyncHandlerDescriptors = append([]AVMAsyncHandlerDescriptor(nil), descriptor.AsyncHandlerDescriptors...)
	for i := range descriptor.AsyncHandlerDescriptors {
		descriptor.AsyncHandlerDescriptors[i] = canonicalAVMAsyncHandlerDescriptor(descriptor.AsyncHandlerDescriptors[i])
	}
	sort.SliceStable(descriptor.AsyncHandlerDescriptors, func(i, j int) bool {
		return descriptor.AsyncHandlerDescriptors[i].HandlerID < descriptor.AsyncHandlerDescriptors[j].HandlerID
	})
	descriptor.GetMethodDescriptors = append([]AVMGetMethodDescriptor(nil), descriptor.GetMethodDescriptors...)
	for i := range descriptor.GetMethodDescriptors {
		descriptor.GetMethodDescriptors[i] = canonicalAVMGetMethodDescriptor(descriptor.GetMethodDescriptors[i])
	}
	sort.SliceStable(descriptor.GetMethodDescriptors, func(i, j int) bool {
		return descriptor.GetMethodDescriptors[i].MethodID < descriptor.GetMethodDescriptors[j].MethodID
	})
	return descriptor
}

func canonicalAVMMethodDescriptor(descriptor AVMMethodDescriptor) AVMMethodDescriptor {
	descriptor.MethodID = strings.TrimSpace(descriptor.MethodID)
	descriptor.Name = strings.TrimSpace(descriptor.Name)
	descriptor.InputSchemaHash = strings.TrimSpace(descriptor.InputSchemaHash)
	descriptor.OutputSchemaHash = strings.TrimSpace(descriptor.OutputSchemaHash)
	descriptor.PaymentRequirementOptional = strings.TrimSpace(descriptor.PaymentRequirementOptional)
	descriptor.ProofRequirementOptional = strings.TrimSpace(descriptor.ProofRequirementOptional)
	return descriptor
}

func canonicalAVMEventDescriptor(descriptor AVMEventDescriptor) AVMEventDescriptor {
	descriptor.EventID = strings.TrimSpace(descriptor.EventID)
	descriptor.Name = strings.TrimSpace(descriptor.Name)
	descriptor.SchemaHash = strings.TrimSpace(descriptor.SchemaHash)
	return descriptor
}

func canonicalAVMAsyncHandlerDescriptor(descriptor AVMAsyncHandlerDescriptor) AVMAsyncHandlerDescriptor {
	descriptor.HandlerID = strings.TrimSpace(descriptor.HandlerID)
	descriptor.Name = strings.TrimSpace(descriptor.Name)
	descriptor.InputSchemaHash = strings.TrimSpace(descriptor.InputSchemaHash)
	descriptor.OutputSchemaHash = strings.TrimSpace(descriptor.OutputSchemaHash)
	descriptor.RetryPolicyOptional = strings.TrimSpace(descriptor.RetryPolicyOptional)
	descriptor.CallbackBehavior = strings.TrimSpace(descriptor.CallbackBehavior)
	return descriptor
}

func canonicalAVMGetMethodDescriptor(descriptor AVMGetMethodDescriptor) AVMGetMethodDescriptor {
	descriptor.MethodID = strings.TrimSpace(descriptor.MethodID)
	descriptor.Name = strings.TrimSpace(descriptor.Name)
	descriptor.InputSchemaHash = strings.TrimSpace(descriptor.InputSchemaHash)
	descriptor.OutputSchemaHash = strings.TrimSpace(descriptor.OutputSchemaHash)
	return descriptor
}

func canonicalAVMInterfaceRegistry(registry AVMInterfaceRegistry) AVMInterfaceRegistry {
	registry.Root = strings.TrimSpace(registry.Root)
	registry.Interfaces = append([]AVMInterfaceDescriptor(nil), registry.Interfaces...)
	for i := range registry.Interfaces {
		registry.Interfaces[i] = canonicalAVMInterfaceDescriptor(registry.Interfaces[i])
	}
	sort.SliceStable(registry.Interfaces, func(i, j int) bool {
		return registry.Interfaces[i].InterfaceHash < registry.Interfaces[j].InterfaceHash
	})
	registry.Schemas = append([]AVMInterfaceSchema(nil), registry.Schemas...)
	for i := range registry.Schemas {
		registry.Schemas[i] = canonicalAVMInterfaceSchema(registry.Schemas[i])
	}
	sort.SliceStable(registry.Schemas, func(i, j int) bool {
		return registry.Schemas[i].InterfaceHash < registry.Schemas[j].InterfaceHash
	})
	registry.Bindings = append([]AVMInterfaceBinding(nil), registry.Bindings...)
	for i := range registry.Bindings {
		registry.Bindings[i] = canonicalAVMInterfaceBinding(registry.Bindings[i])
	}
	sort.SliceStable(registry.Bindings, func(i, j int) bool {
		if registry.Bindings[i].TargetType != registry.Bindings[j].TargetType {
			return registry.Bindings[i].TargetType < registry.Bindings[j].TargetType
		}
		return registry.Bindings[i].TargetID < registry.Bindings[j].TargetID
	})
	return registry
}

func canonicalAVMInterfaceSchema(schema AVMInterfaceSchema) AVMInterfaceSchema {
	schema.InterfaceHash = strings.TrimSpace(schema.InterfaceHash)
	schema.DescriptorRoot = strings.TrimSpace(schema.DescriptorRoot)
	schema.SchemaHash = strings.TrimSpace(schema.SchemaHash)
	schema.UseCases = append([]AVMInterfaceUseCase(nil), schema.UseCases...)
	sort.SliceStable(schema.UseCases, func(i, j int) bool { return schema.UseCases[i] < schema.UseCases[j] })
	return schema
}

func canonicalAVMInterfaceBinding(binding AVMInterfaceBinding) AVMInterfaceBinding {
	binding.TargetID = strings.TrimSpace(binding.TargetID)
	binding.InterfaceHash = strings.TrimSpace(binding.InterfaceHash)
	binding.BindingHash = strings.TrimSpace(binding.BindingHash)
	return binding
}

func canonicalAVMSDKCodeGenerationFormat(format AVMSDKCodeGenerationFormat) AVMSDKCodeGenerationFormat {
	format.InterfaceHash = strings.TrimSpace(format.InterfaceHash)
	format.PackageName = strings.TrimSpace(format.PackageName)
	format.GenerationHash = strings.TrimSpace(format.GenerationHash)
	format.MethodBindings = cloneSortedNativeModuleStrings(format.MethodBindings)
	format.GetMethodBindings = cloneSortedNativeModuleStrings(format.GetMethodBindings)
	return format
}

func validateAVMMethods(methods []AVMMethodDescriptor) error {
	seen := make(map[string]struct{}, len(methods))
	for i, method := range methods {
		if err := method.Validate(); err != nil {
			return err
		}
		if _, found := seen[method.MethodID]; found {
			return fmt.Errorf("duplicate AVM method id %q", method.MethodID)
		}
		seen[method.MethodID] = struct{}{}
		if i > 0 && methods[i-1].MethodID >= method.MethodID {
			return errors.New("AVM method descriptors must be sorted canonically")
		}
	}
	return nil
}

func validateAVMEvents(events []AVMEventDescriptor) error {
	seen := make(map[string]struct{}, len(events))
	for i, event := range events {
		if err := event.Validate(); err != nil {
			return err
		}
		if _, found := seen[event.EventID]; found {
			return fmt.Errorf("duplicate AVM event id %q", event.EventID)
		}
		seen[event.EventID] = struct{}{}
		if i > 0 && events[i-1].EventID >= event.EventID {
			return errors.New("AVM event descriptors must be sorted canonically")
		}
	}
	return nil
}

func validateAVMAsyncHandlers(handlers []AVMAsyncHandlerDescriptor) error {
	seen := make(map[string]struct{}, len(handlers))
	for i, handler := range handlers {
		if err := handler.Validate(); err != nil {
			return err
		}
		if _, found := seen[handler.HandlerID]; found {
			return fmt.Errorf("duplicate AVM async handler id %q", handler.HandlerID)
		}
		seen[handler.HandlerID] = struct{}{}
		if i > 0 && handlers[i-1].HandlerID >= handler.HandlerID {
			return errors.New("AVM async handler descriptors must be sorted canonically")
		}
	}
	return nil
}

func validateAVMGetMethods(methods []AVMGetMethodDescriptor) error {
	seen := make(map[string]struct{}, len(methods))
	for i, method := range methods {
		if err := method.Validate(); err != nil {
			return err
		}
		if _, found := seen[method.MethodID]; found {
			return fmt.Errorf("duplicate AVM get method id %q", method.MethodID)
		}
		seen[method.MethodID] = struct{}{}
		if i > 0 && methods[i-1].MethodID >= method.MethodID {
			return errors.New("AVM get method descriptors must be sorted canonically")
		}
	}
	return nil
}

func validateAVMInterfaceSchemas(schemas []AVMInterfaceSchema, interfaceHashes map[string]struct{}) error {
	seen := make(map[string]struct{}, len(schemas))
	for i, schema := range schemas {
		if err := schema.Validate(); err != nil {
			return err
		}
		if _, found := interfaceHashes[schema.InterfaceHash]; !found {
			return fmt.Errorf("AVM interface schema references missing interface %q", schema.InterfaceHash)
		}
		if _, found := seen[schema.InterfaceHash]; found {
			return fmt.Errorf("duplicate AVM interface schema %q", schema.InterfaceHash)
		}
		seen[schema.InterfaceHash] = struct{}{}
		if i > 0 && schemas[i-1].InterfaceHash >= schema.InterfaceHash {
			return errors.New("AVM interface schemas must be sorted canonically")
		}
	}
	return nil
}

func validateAVMInterfaceBindings(bindings []AVMInterfaceBinding, interfaceHashes map[string]struct{}) error {
	seen := make(map[string]struct{}, len(bindings))
	for i, binding := range bindings {
		if err := binding.Validate(); err != nil {
			return err
		}
		if _, found := interfaceHashes[binding.InterfaceHash]; !found {
			return fmt.Errorf("AVM interface binding references missing interface %q", binding.InterfaceHash)
		}
		key := string(binding.TargetType) + "/" + binding.TargetID
		if _, found := seen[key]; found {
			return fmt.Errorf("duplicate AVM interface binding %q", key)
		}
		seen[key] = struct{}{}
		if i > 0 {
			prev := string(bindings[i-1].TargetType) + "/" + bindings[i-1].TargetID
			if prev >= key {
				return errors.New("AVM interface bindings must be sorted canonically")
			}
		}
	}
	return nil
}

func validateInterfaceBindingNames(fieldName string, values []string) error {
	seen := make(map[string]struct{}, len(values))
	for i, value := range values {
		if err := validateInterfaceToken(fieldName, value); err != nil {
			return err
		}
		if _, found := seen[value]; found {
			return fmt.Errorf("duplicate %s %q", fieldName, value)
		}
		seen[value] = struct{}{}
		if i > 0 && values[i-1] >= value {
			return fmt.Errorf("%s values must be sorted canonically", fieldName)
		}
	}
	return nil
}

func validateInterfaceToken(fieldName, value string) error {
	return validateEngineToken(fieldName, value, MaxAVMInterfaceTokenLength)
}

func writeAVMMethodDescriptor(w engineByteWriter, method AVMMethodDescriptor) {
	writeEnginePart(w, method.MethodID)
	writeEnginePart(w, method.Name)
	writeEnginePart(w, method.InputSchemaHash)
	writeEnginePart(w, method.OutputSchemaHash)
	writeEnginePart(w, string(method.ExecutionMode))
	writeEngineUint64(w, method.GasHint)
	writeEnginePart(w, method.PaymentRequirementOptional)
	writeEnginePart(w, method.ProofRequirementOptional)
}
