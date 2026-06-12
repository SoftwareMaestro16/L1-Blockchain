package avm

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"sort"
	"strings"
)

const (
	MaxInterfaceNameLength		= 64
	MaxInterfaceDescriptionLength	= 256
	MaxInterfaceBindingTextLength	= 256
	MaxInterfacePayloadBytes	= 1 << 20
	MaxInterfaceMethods		= 128
	MaxInterfaceEvents		= 128
	MaxInterfaceAsyncHandlers	= 128
	MaxInterfaceGetMethods		= 128
	MaxInterfaceBindings		= 128
	MaxInterfaceFields		= 64
	MaxInterfaceAliases		= 16
	MaxInterfaceExamples		= 16
)

type InterfaceValueKind string

const (
	InterfaceValueBool	InterfaceValueKind	= "bool"
	InterfaceValueU64	InterfaceValueKind	= "u64"
	InterfaceValueBytes	InterfaceValueKind	= "bytes"
	InterfaceValueString	InterfaceValueKind	= "string"
	InterfaceValueAddress	InterfaceValueKind	= "address"
	InterfaceValueCoin	InterfaceValueKind	= "coin"
)

type InterfaceWalletRisk string

const (
	InterfaceWalletRiskLow		InterfaceWalletRisk	= "low"
	InterfaceWalletRiskMedium	InterfaceWalletRisk	= "medium"
	InterfaceWalletRiskHigh		InterfaceWalletRisk	= "high"
	InterfaceWalletRiskCritical	InterfaceWalletRisk	= "critical"
)

type InterfaceManifest struct {
	Name		string
	Version		uint16
	Methods		[]InterfaceMethod
	Events		[]InterfaceEvent
	AsyncHandlers	[]InterfaceAsyncHandler
	GetMethods	[]InterfaceGetMethod
	CLIBindings	[]InterfaceCLIBinding
	SDKBindings	[]InterfaceSDKBinding
	WalletActions	[]InterfaceWalletAction
}

type InterfaceMethod struct {
	Name		string
	Entrypoint	Entrypoint
	Opcode		uint32
	Async		bool
	Params		[]InterfaceParamDescriptor
	Results		[]InterfaceResultDescriptor
	Description	string
}

type InterfaceEvent struct {
	Name	string
	Opcode	uint32
	Fields	[]InterfaceParamDescriptor
}

type InterfaceAsyncHandler struct {
	Name		string
	Entrypoint	Entrypoint
	Opcode		uint32
	MessageType	string
	Bounced		bool
	Idempotent	bool
	Params		[]InterfaceParamDescriptor
	Results		[]InterfaceResultDescriptor
	Description	string
}

type InterfaceGetMethod struct {
	Name			string
	Entrypoint		Entrypoint
	Selector		uint32
	Params			[]InterfaceParamDescriptor
	Results			[]InterfaceResultDescriptor
	Cacheable		bool
	MaxResponseBytes	uint32
	Description		string
}

type InterfaceParamDescriptor struct {
	Name		string
	Kind		InterfaceValueKind
	Required	bool
	MaxBytes	uint32
	Description	string
}

type InterfaceResultDescriptor = InterfaceParamDescriptor

type InterfaceCLIBinding struct {
	Method		string
	Command		string
	Use		string
	Aliases		[]string
	Examples	[]string
	InputFormat	string
	OutputFormat	string
}

type InterfaceSDKBinding struct {
	Method		string
	Package		string
	Service		string
	MethodName	string
	RequestType	string
	ResponseType	string
	Async		bool
}

type InterfaceWalletAction struct {
	Method		string
	Title		string
	Description	string
	Category	string
	Icon		string
	Risk		InterfaceWalletRisk
	ConfirmLabel	string
	Inputs		[]InterfaceParamDescriptor
	Outputs		[]InterfaceResultDescriptor
}

type InterfaceDeveloperMetadata struct {
	Name		string
	Version		uint16
	MetadataHash	[MetadataHashLength]byte
	Methods		[]InterfaceMethod
	Events		[]InterfaceEvent
	AsyncHandlers	[]InterfaceAsyncHandler
	GetMethods	[]InterfaceGetMethod
	CLIBindings	[]InterfaceCLIBinding
	SDKBindings	[]InterfaceSDKBinding
	WalletActions	[]InterfaceWalletAction
}

func (m InterfaceManifest) Validate() error {
	if err := validateInterfaceName("interface", m.Name); err != nil {
		return err
	}
	if m.Version == 0 {
		return errors.New("AVM interface version must be positive")
	}
	if len(m.Methods)+len(m.AsyncHandlers)+len(m.GetMethods) == 0 {
		return errors.New("AVM interface must declare at least one callable descriptor")
	}
	if len(m.Methods) > MaxInterfaceMethods {
		return fmt.Errorf("AVM interface methods must be <= %d", MaxInterfaceMethods)
	}
	if len(m.Events) > MaxInterfaceEvents {
		return fmt.Errorf("AVM interface events must be <= %d", MaxInterfaceEvents)
	}
	if len(m.AsyncHandlers) > MaxInterfaceAsyncHandlers {
		return fmt.Errorf("AVM interface async handlers must be <= %d", MaxInterfaceAsyncHandlers)
	}
	if len(m.GetMethods) > MaxInterfaceGetMethods {
		return fmt.Errorf("AVM interface get methods must be <= %d", MaxInterfaceGetMethods)
	}
	if len(m.CLIBindings) > MaxInterfaceBindings {
		return fmt.Errorf("AVM interface CLI bindings must be <= %d", MaxInterfaceBindings)
	}
	if len(m.SDKBindings) > MaxInterfaceBindings {
		return fmt.Errorf("AVM interface SDK bindings must be <= %d", MaxInterfaceBindings)
	}
	if len(m.WalletActions) > MaxInterfaceBindings {
		return fmt.Errorf("AVM interface wallet actions must be <= %d", MaxInterfaceBindings)
	}

	callableNames := make(map[string]struct{}, len(m.Methods)+len(m.AsyncHandlers)+len(m.GetMethods))
	callableParams := make(map[string][]InterfaceParamDescriptor, len(m.Methods)+len(m.AsyncHandlers)+len(m.GetMethods))
	callableResults := make(map[string][]InterfaceResultDescriptor, len(m.Methods)+len(m.AsyncHandlers)+len(m.GetMethods))
	seenOpcodes := make(map[uint32]string, len(m.Methods)+len(m.AsyncHandlers))
	for _, method := range m.Methods {
		name := strings.TrimSpace(method.Name)
		if err := validateInterfaceName("interface method", name); err != nil {
			return err
		}
		if !IsValidEntrypoint(method.Entrypoint) {
			return fmt.Errorf("AVM interface method %q has invalid entrypoint", name)
		}
		if err := validateInterfaceText("interface method description", method.Description, MaxInterfaceDescriptionLength, false); err != nil {
			return err
		}
		if err := validateParamDescriptors("interface method parameter", method.Params); err != nil {
			return err
		}
		if err := validateParamDescriptors("interface method result", method.Results); err != nil {
			return err
		}
		if err := registerInterfaceCallable(callableNames, callableParams, callableResults, name, method.Params, method.Results); err != nil {
			return err
		}
		if err := registerInterfaceOpcode(seenOpcodes, "opcode", method.Opcode, name); err != nil {
			return err
		}
	}

	seenEvents := make(map[string]struct{}, len(m.Events))
	seenEventOpcodes := make(map[uint32]string, len(m.Events))
	for _, event := range m.Events {
		name := strings.TrimSpace(event.Name)
		if err := validateInterfaceName("interface event", name); err != nil {
			return err
		}
		if _, exists := seenEvents[name]; exists {
			return fmt.Errorf("duplicate AVM interface event %q", name)
		}
		seenEvents[name] = struct{}{}
		if err := registerInterfaceOpcode(seenEventOpcodes, "event opcode", event.Opcode, name); err != nil {
			return err
		}
		if err := validateParamDescriptors("interface event field", event.Fields); err != nil {
			return err
		}
	}

	for _, handler := range m.AsyncHandlers {
		name := strings.TrimSpace(handler.Name)
		if err := validateInterfaceName("interface async handler", name); err != nil {
			return err
		}
		if !IsValidEntrypoint(handler.Entrypoint) || !isAsyncHandlerEntrypoint(handler.Entrypoint) {
			return fmt.Errorf("AVM interface async handler %q has invalid entrypoint", name)
		}
		if handler.Bounced != (handler.Entrypoint == EntryReceiveBounced) {
			return fmt.Errorf("AVM interface async handler %q bounced flag must match receive-bounced entrypoint", name)
		}
		if err := validateInterfaceText("interface async handler message type", handler.MessageType, MaxInterfaceNameLength, false); err != nil {
			return err
		}
		if err := validateInterfaceText("interface async handler description", handler.Description, MaxInterfaceDescriptionLength, false); err != nil {
			return err
		}
		if err := validateParamDescriptors("interface async handler parameter", handler.Params); err != nil {
			return err
		}
		if err := validateParamDescriptors("interface async handler result", handler.Results); err != nil {
			return err
		}
		if err := registerInterfaceCallable(callableNames, callableParams, callableResults, name, handler.Params, handler.Results); err != nil {
			return err
		}
		if err := registerInterfaceOpcode(seenOpcodes, "opcode", handler.Opcode, name); err != nil {
			return err
		}
	}

	seenSelectors := make(map[uint32]string, len(m.GetMethods))
	for _, getter := range m.GetMethods {
		name := strings.TrimSpace(getter.Name)
		if err := validateInterfaceName("interface get method", name); err != nil {
			return err
		}
		if getter.Entrypoint != EntryQuery {
			return fmt.Errorf("AVM interface get method %q must use query entrypoint", name)
		}
		if getter.MaxResponseBytes > MaxInterfacePayloadBytes {
			return fmt.Errorf("AVM interface get method %q max response bytes must be <= %d", name, MaxInterfacePayloadBytes)
		}
		if err := validateInterfaceText("interface get method description", getter.Description, MaxInterfaceDescriptionLength, false); err != nil {
			return err
		}
		if err := validateParamDescriptors("interface get method parameter", getter.Params); err != nil {
			return err
		}
		if err := validateParamDescriptors("interface get method result", getter.Results); err != nil {
			return err
		}
		if err := registerInterfaceCallable(callableNames, callableParams, callableResults, name, getter.Params, getter.Results); err != nil {
			return err
		}
		if err := registerInterfaceOpcode(seenSelectors, "get selector", getter.Selector, name); err != nil {
			return err
		}
	}

	for _, binding := range m.CLIBindings {
		method := strings.TrimSpace(binding.Method)
		if err := validateBindingMethod("CLI binding", method, callableNames); err != nil {
			return err
		}
		if err := validateInterfaceName("interface CLI command", binding.Command); err != nil {
			return err
		}
		if err := validateInterfaceText("interface CLI use", binding.Use, MaxInterfaceBindingTextLength, false); err != nil {
			return err
		}
		if err := validateInterfaceText("interface CLI input format", binding.InputFormat, MaxInterfaceNameLength, false); err != nil {
			return err
		}
		if err := validateInterfaceText("interface CLI output format", binding.OutputFormat, MaxInterfaceNameLength, false); err != nil {
			return err
		}
		if err := validateStringList("interface CLI alias", binding.Aliases, MaxInterfaceAliases, MaxInterfaceNameLength); err != nil {
			return err
		}
		if err := validateStringList("interface CLI example", binding.Examples, MaxInterfaceExamples, MaxInterfaceBindingTextLength); err != nil {
			return err
		}
	}

	for _, binding := range m.SDKBindings {
		method := strings.TrimSpace(binding.Method)
		if err := validateBindingMethod("SDK binding", method, callableNames); err != nil {
			return err
		}
		for _, descriptor := range []struct {
			kind	string
			value	string
		}{
			{kind: "interface SDK package", value: binding.Package},
			{kind: "interface SDK service", value: binding.Service},
			{kind: "interface SDK method", value: binding.MethodName},
			{kind: "interface SDK request type", value: binding.RequestType},
			{kind: "interface SDK response type", value: binding.ResponseType},
		} {
			if err := validateInterfaceName(descriptor.kind, descriptor.value); err != nil {
				return err
			}
		}
	}

	for _, action := range m.WalletActions {
		method := strings.TrimSpace(action.Method)
		if err := validateBindingMethod("wallet action", method, callableNames); err != nil {
			return err
		}
		if err := validateInterfaceName("interface wallet title", action.Title); err != nil {
			return err
		}
		if err := validateInterfaceName("interface wallet confirm label", action.ConfirmLabel); err != nil {
			return err
		}
		if err := validateInterfaceText("interface wallet description", action.Description, MaxInterfaceDescriptionLength, false); err != nil {
			return err
		}
		if err := validateInterfaceText("interface wallet category", action.Category, MaxInterfaceNameLength, false); err != nil {
			return err
		}
		if err := validateInterfaceText("interface wallet icon", action.Icon, MaxInterfaceNameLength, false); err != nil {
			return err
		}
		if !IsValidInterfaceWalletRisk(action.Risk) {
			return fmt.Errorf("AVM interface wallet action %q has invalid risk %q", method, action.Risk)
		}
		if err := validateParamDescriptors("interface wallet input", action.Inputs); err != nil {
			return err
		}
		if err := validateParamDescriptors("interface wallet output", action.Outputs); err != nil {
			return err
		}
		if err := validateWalletDescriptors("input", method, action.Inputs, callableParams[method]); err != nil {
			return err
		}
		if err := validateWalletDescriptors("output", method, action.Outputs, callableResults[method]); err != nil {
			return err
		}
	}

	return nil
}

func InterfaceHash(manifest InterfaceManifest) ([MetadataHashLength]byte, error) {
	if err := manifest.Validate(); err != nil {
		return [MetadataHashLength]byte{}, err
	}
	manifest = canonicalInterfaceManifest(manifest)
	buf := bytes.NewBuffer(nil)
	writeString(buf, "AVM_INTERFACE_V2")
	writeString(buf, manifest.Name)
	writeU16(buf, manifest.Version)
	writeU16(buf, uint16(len(manifest.Methods)))
	for _, method := range manifest.Methods {
		writeMethodDescriptor(buf, method)
	}
	writeU16(buf, uint16(len(manifest.Events)))
	for _, event := range manifest.Events {
		writeString(buf, event.Name)
		writeU32(buf, event.Opcode)
		writeParamDescriptors(buf, event.Fields)
	}
	writeU16(buf, uint16(len(manifest.AsyncHandlers)))
	for _, handler := range manifest.AsyncHandlers {
		writeAsyncHandlerDescriptor(buf, handler)
	}
	writeU16(buf, uint16(len(manifest.GetMethods)))
	for _, getter := range manifest.GetMethods {
		writeGetMethodDescriptor(buf, getter)
	}
	writeU16(buf, uint16(len(manifest.CLIBindings)))
	for _, binding := range manifest.CLIBindings {
		writeCLIBinding(buf, binding)
	}
	writeU16(buf, uint16(len(manifest.SDKBindings)))
	for _, binding := range manifest.SDKBindings {
		writeSDKBinding(buf, binding)
	}
	writeU16(buf, uint16(len(manifest.WalletActions)))
	for _, action := range manifest.WalletActions {
		writeWalletAction(buf, action)
	}
	return sha256.Sum256(buf.Bytes()), nil
}

func VerifyInterface(module Module, manifest InterfaceManifest) error {
	hash, err := InterfaceHash(manifest)
	if err != nil {
		return err
	}
	if !bytes.Equal(module.MetadataHash[:], hash[:]) {
		return errors.New("AVM interface metadata hash mismatch")
	}
	for _, method := range manifest.Methods {
		if _, ok := module.Exports[method.Entrypoint]; !ok {
			return fmt.Errorf("AVM interface method %q entrypoint is not exported", method.Name)
		}
	}
	for _, handler := range manifest.AsyncHandlers {
		if _, ok := module.Exports[handler.Entrypoint]; !ok {
			return fmt.Errorf("AVM interface async handler %q entrypoint is not exported", handler.Name)
		}
	}
	for _, getter := range manifest.GetMethods {
		if _, ok := module.Exports[getter.Entrypoint]; !ok {
			return fmt.Errorf("AVM interface get method %q entrypoint is not exported", getter.Name)
		}
	}
	return nil
}

func BuildInterfaceDeveloperMetadata(manifest InterfaceManifest) (InterfaceDeveloperMetadata, error) {
	hash, err := InterfaceHash(manifest)
	if err != nil {
		return InterfaceDeveloperMetadata{}, err
	}
	manifest = canonicalInterfaceManifest(manifest)
	return InterfaceDeveloperMetadata{
		Name:		manifest.Name,
		Version:	manifest.Version,
		MetadataHash:	hash,
		Methods:	manifest.Methods,
		Events:		manifest.Events,
		AsyncHandlers:	manifest.AsyncHandlers,
		GetMethods:	manifest.GetMethods,
		CLIBindings:	manifest.CLIBindings,
		SDKBindings:	manifest.SDKBindings,
		WalletActions:	manifest.WalletActions,
	}, nil
}

func IsValidInterfaceValueKind(kind InterfaceValueKind) bool {
	switch kind {
	case InterfaceValueBool,
		InterfaceValueU64,
		InterfaceValueBytes,
		InterfaceValueString,
		InterfaceValueAddress,
		InterfaceValueCoin:
		return true
	default:
		return false
	}
}

func IsValidInterfaceWalletRisk(risk InterfaceWalletRisk) bool {
	switch risk {
	case InterfaceWalletRiskLow,
		InterfaceWalletRiskMedium,
		InterfaceWalletRiskHigh,
		InterfaceWalletRiskCritical:
		return true
	default:
		return false
	}
}

func canonicalInterfaceManifest(manifest InterfaceManifest) InterfaceManifest {
	manifest.Name = strings.TrimSpace(manifest.Name)
	manifest.Methods = cloneInterfaceMethods(manifest.Methods)
	manifest.Events = cloneInterfaceEvents(manifest.Events)
	manifest.AsyncHandlers = cloneInterfaceAsyncHandlers(manifest.AsyncHandlers)
	manifest.GetMethods = cloneInterfaceGetMethods(manifest.GetMethods)
	manifest.CLIBindings = cloneInterfaceCLIBindings(manifest.CLIBindings)
	manifest.SDKBindings = cloneInterfaceSDKBindings(manifest.SDKBindings)
	manifest.WalletActions = cloneInterfaceWalletActions(manifest.WalletActions)
	sort.SliceStable(manifest.Methods, func(i, j int) bool {
		if manifest.Methods[i].Name != manifest.Methods[j].Name {
			return manifest.Methods[i].Name < manifest.Methods[j].Name
		}
		if manifest.Methods[i].Entrypoint != manifest.Methods[j].Entrypoint {
			return manifest.Methods[i].Entrypoint < manifest.Methods[j].Entrypoint
		}
		return manifest.Methods[i].Opcode < manifest.Methods[j].Opcode
	})
	sort.SliceStable(manifest.Events, func(i, j int) bool {
		if manifest.Events[i].Name != manifest.Events[j].Name {
			return manifest.Events[i].Name < manifest.Events[j].Name
		}
		return manifest.Events[i].Opcode < manifest.Events[j].Opcode
	})
	sort.SliceStable(manifest.AsyncHandlers, func(i, j int) bool {
		if manifest.AsyncHandlers[i].Name != manifest.AsyncHandlers[j].Name {
			return manifest.AsyncHandlers[i].Name < manifest.AsyncHandlers[j].Name
		}
		if manifest.AsyncHandlers[i].Entrypoint != manifest.AsyncHandlers[j].Entrypoint {
			return manifest.AsyncHandlers[i].Entrypoint < manifest.AsyncHandlers[j].Entrypoint
		}
		return manifest.AsyncHandlers[i].Opcode < manifest.AsyncHandlers[j].Opcode
	})
	sort.SliceStable(manifest.GetMethods, func(i, j int) bool {
		if manifest.GetMethods[i].Name != manifest.GetMethods[j].Name {
			return manifest.GetMethods[i].Name < manifest.GetMethods[j].Name
		}
		if manifest.GetMethods[i].Entrypoint != manifest.GetMethods[j].Entrypoint {
			return manifest.GetMethods[i].Entrypoint < manifest.GetMethods[j].Entrypoint
		}
		return manifest.GetMethods[i].Selector < manifest.GetMethods[j].Selector
	})
	sort.SliceStable(manifest.CLIBindings, func(i, j int) bool {
		if manifest.CLIBindings[i].Method != manifest.CLIBindings[j].Method {
			return manifest.CLIBindings[i].Method < manifest.CLIBindings[j].Method
		}
		return manifest.CLIBindings[i].Command < manifest.CLIBindings[j].Command
	})
	sort.SliceStable(manifest.SDKBindings, func(i, j int) bool {
		if manifest.SDKBindings[i].Method != manifest.SDKBindings[j].Method {
			return manifest.SDKBindings[i].Method < manifest.SDKBindings[j].Method
		}
		if manifest.SDKBindings[i].Package != manifest.SDKBindings[j].Package {
			return manifest.SDKBindings[i].Package < manifest.SDKBindings[j].Package
		}
		if manifest.SDKBindings[i].Service != manifest.SDKBindings[j].Service {
			return manifest.SDKBindings[i].Service < manifest.SDKBindings[j].Service
		}
		return manifest.SDKBindings[i].MethodName < manifest.SDKBindings[j].MethodName
	})
	sort.SliceStable(manifest.WalletActions, func(i, j int) bool {
		if manifest.WalletActions[i].Method != manifest.WalletActions[j].Method {
			return manifest.WalletActions[i].Method < manifest.WalletActions[j].Method
		}
		return manifest.WalletActions[i].Title < manifest.WalletActions[j].Title
	})
	return manifest
}

func validateInterfaceName(kind, name string) error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return fmt.Errorf("AVM %s name is required", kind)
	}
	if len(trimmed) > MaxInterfaceNameLength {
		return fmt.Errorf("AVM %s name must be <= %d bytes", kind, MaxInterfaceNameLength)
	}
	return nil
}

func validateInterfaceText(kind, value string, max int, required bool) error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		if required {
			return fmt.Errorf("AVM %s is required", kind)
		}
		return nil
	}
	if len(trimmed) > max {
		return fmt.Errorf("AVM %s must be <= %d bytes", kind, max)
	}
	return nil
}

func validateStringList(kind string, values []string, maxItems, maxBytes int) error {
	if len(values) > maxItems {
		return fmt.Errorf("AVM %s count must be <= %d", kind, maxItems)
	}
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if err := validateInterfaceText(kind, trimmed, maxBytes, true); err != nil {
			return err
		}
		if _, exists := seen[trimmed]; exists {
			return fmt.Errorf("duplicate AVM %s %q", kind, trimmed)
		}
		seen[trimmed] = struct{}{}
	}
	return nil
}

func validateParamDescriptors(kind string, params []InterfaceParamDescriptor) error {
	if len(params) > MaxInterfaceFields {
		return fmt.Errorf("AVM %s count must be <= %d", kind, MaxInterfaceFields)
	}
	seen := make(map[string]struct{}, len(params))
	for _, param := range params {
		name := strings.TrimSpace(param.Name)
		if err := validateInterfaceName(kind, name); err != nil {
			return err
		}
		if _, exists := seen[name]; exists {
			return fmt.Errorf("duplicate AVM %s %q", kind, name)
		}
		seen[name] = struct{}{}
		if !IsValidInterfaceValueKind(param.Kind) {
			return fmt.Errorf("AVM %s %q has invalid kind %q", kind, name, param.Kind)
		}
		if param.MaxBytes > MaxInterfacePayloadBytes {
			return fmt.Errorf("AVM %s %q max bytes must be <= %d", kind, name, MaxInterfacePayloadBytes)
		}
		if err := validateInterfaceText(kind+" description", param.Description, MaxInterfaceDescriptionLength, false); err != nil {
			return err
		}
	}
	return nil
}

func validateBindingMethod(kind, method string, callableNames map[string]struct{}) error {
	if err := validateInterfaceName("interface "+kind+" method", method); err != nil {
		return err
	}
	if _, ok := callableNames[method]; !ok {
		return fmt.Errorf("AVM interface %s method %q is not declared", kind, method)
	}
	return nil
}

func validateWalletDescriptors(kind, method string, wallet []InterfaceParamDescriptor, declared []InterfaceParamDescriptor) error {
	if len(wallet) == 0 {
		return nil
	}
	declaredByName := make(map[string]InterfaceValueKind, len(declared))
	for _, descriptor := range declared {
		declaredByName[strings.TrimSpace(descriptor.Name)] = descriptor.Kind
	}
	for _, descriptor := range wallet {
		name := strings.TrimSpace(descriptor.Name)
		declaredKind, ok := declaredByName[name]
		if !ok {
			return fmt.Errorf("AVM interface wallet %s %q is not declared by method %q", kind, name, method)
		}
		if declaredKind != descriptor.Kind {
			return fmt.Errorf("AVM interface wallet %s %q kind mismatch for method %q", kind, name, method)
		}
	}
	return nil
}

func registerInterfaceCallable(
	callableNames map[string]struct{},
	callableParams map[string][]InterfaceParamDescriptor,
	callableResults map[string][]InterfaceResultDescriptor,
	name string,
	params []InterfaceParamDescriptor,
	results []InterfaceResultDescriptor,
) error {
	if _, exists := callableNames[name]; exists {
		return fmt.Errorf("duplicate AVM interface callable %q", name)
	}
	callableNames[name] = struct{}{}
	callableParams[name] = cloneParamDescriptors(params)
	callableResults[name] = cloneResultDescriptors(results)
	return nil
}

func registerInterfaceOpcode(seen map[uint32]string, kind string, opcode uint32, name string) error {
	if opcode == 0 {
		return nil
	}
	if previous, exists := seen[opcode]; exists {
		return fmt.Errorf("duplicate AVM interface %s %d for %q and %q", kind, opcode, previous, name)
	}
	seen[opcode] = name
	return nil
}

func isAsyncHandlerEntrypoint(entry Entrypoint) bool {
	switch entry {
	case EntryReceiveExternal, EntryReceiveInternal, EntryReceiveBounced:
		return true
	default:
		return false
	}
}

func writeMethodDescriptor(buf *bytes.Buffer, method InterfaceMethod) {
	writeString(buf, method.Name)
	buf.WriteByte(byte(method.Entrypoint))
	writeU32(buf, method.Opcode)
	writeBool(buf, method.Async)
	writeParamDescriptors(buf, method.Params)
	writeParamDescriptors(buf, method.Results)
	writeString(buf, method.Description)
}

func writeAsyncHandlerDescriptor(buf *bytes.Buffer, handler InterfaceAsyncHandler) {
	writeString(buf, handler.Name)
	buf.WriteByte(byte(handler.Entrypoint))
	writeU32(buf, handler.Opcode)
	writeString(buf, handler.MessageType)
	writeBool(buf, handler.Bounced)
	writeBool(buf, handler.Idempotent)
	writeParamDescriptors(buf, handler.Params)
	writeParamDescriptors(buf, handler.Results)
	writeString(buf, handler.Description)
}

func writeGetMethodDescriptor(buf *bytes.Buffer, getter InterfaceGetMethod) {
	writeString(buf, getter.Name)
	buf.WriteByte(byte(getter.Entrypoint))
	writeU32(buf, getter.Selector)
	writeParamDescriptors(buf, getter.Params)
	writeParamDescriptors(buf, getter.Results)
	writeBool(buf, getter.Cacheable)
	writeU32(buf, getter.MaxResponseBytes)
	writeString(buf, getter.Description)
}

func writeParamDescriptors(buf *bytes.Buffer, params []InterfaceParamDescriptor) {
	writeU16(buf, uint16(len(params)))
	for _, param := range params {
		writeString(buf, param.Name)
		writeString(buf, string(param.Kind))
		writeBool(buf, param.Required)
		writeU32(buf, param.MaxBytes)
		writeString(buf, param.Description)
	}
}

func writeCLIBinding(buf *bytes.Buffer, binding InterfaceCLIBinding) {
	writeString(buf, binding.Method)
	writeString(buf, binding.Command)
	writeString(buf, binding.Use)
	writeStringList(buf, binding.Aliases)
	writeStringList(buf, binding.Examples)
	writeString(buf, binding.InputFormat)
	writeString(buf, binding.OutputFormat)
}

func writeSDKBinding(buf *bytes.Buffer, binding InterfaceSDKBinding) {
	writeString(buf, binding.Method)
	writeString(buf, binding.Package)
	writeString(buf, binding.Service)
	writeString(buf, binding.MethodName)
	writeString(buf, binding.RequestType)
	writeString(buf, binding.ResponseType)
	writeBool(buf, binding.Async)
}

func writeWalletAction(buf *bytes.Buffer, action InterfaceWalletAction) {
	writeString(buf, action.Method)
	writeString(buf, action.Title)
	writeString(buf, action.Description)
	writeString(buf, action.Category)
	writeString(buf, action.Icon)
	writeString(buf, string(action.Risk))
	writeString(buf, action.ConfirmLabel)
	writeParamDescriptors(buf, action.Inputs)
	writeParamDescriptors(buf, action.Outputs)
}

func writeStringList(buf *bytes.Buffer, values []string) {
	writeU16(buf, uint16(len(values)))
	for _, value := range values {
		writeString(buf, value)
	}
}

func writeBool(buf *bytes.Buffer, value bool) {
	if value {
		buf.WriteByte(1)
	} else {
		buf.WriteByte(0)
	}
}

func cloneInterfaceMethods(methods []InterfaceMethod) []InterfaceMethod {
	out := make([]InterfaceMethod, len(methods))
	for i, method := range methods {
		out[i] = method
		out[i].Name = strings.TrimSpace(method.Name)
		out[i].Description = strings.TrimSpace(method.Description)
		out[i].Params = cloneParamDescriptors(method.Params)
		out[i].Results = cloneResultDescriptors(method.Results)
	}
	return out
}

func cloneInterfaceEvents(events []InterfaceEvent) []InterfaceEvent {
	out := make([]InterfaceEvent, len(events))
	for i, event := range events {
		out[i] = event
		out[i].Name = strings.TrimSpace(event.Name)
		out[i].Fields = cloneParamDescriptors(event.Fields)
	}
	return out
}

func cloneInterfaceAsyncHandlers(handlers []InterfaceAsyncHandler) []InterfaceAsyncHandler {
	out := make([]InterfaceAsyncHandler, len(handlers))
	for i, handler := range handlers {
		out[i] = handler
		out[i].Name = strings.TrimSpace(handler.Name)
		out[i].MessageType = strings.TrimSpace(handler.MessageType)
		out[i].Description = strings.TrimSpace(handler.Description)
		out[i].Params = cloneParamDescriptors(handler.Params)
		out[i].Results = cloneResultDescriptors(handler.Results)
	}
	return out
}

func cloneInterfaceGetMethods(getters []InterfaceGetMethod) []InterfaceGetMethod {
	out := make([]InterfaceGetMethod, len(getters))
	for i, getter := range getters {
		out[i] = getter
		out[i].Name = strings.TrimSpace(getter.Name)
		out[i].Description = strings.TrimSpace(getter.Description)
		out[i].Params = cloneParamDescriptors(getter.Params)
		out[i].Results = cloneResultDescriptors(getter.Results)
	}
	return out
}

func cloneInterfaceCLIBindings(bindings []InterfaceCLIBinding) []InterfaceCLIBinding {
	out := make([]InterfaceCLIBinding, len(bindings))
	for i, binding := range bindings {
		out[i] = binding
		out[i].Method = strings.TrimSpace(binding.Method)
		out[i].Command = strings.TrimSpace(binding.Command)
		out[i].Use = strings.TrimSpace(binding.Use)
		out[i].Aliases = cloneSortedStringList(binding.Aliases)
		out[i].Examples = cloneStringList(binding.Examples)
		out[i].InputFormat = strings.TrimSpace(binding.InputFormat)
		out[i].OutputFormat = strings.TrimSpace(binding.OutputFormat)
	}
	return out
}

func cloneInterfaceSDKBindings(bindings []InterfaceSDKBinding) []InterfaceSDKBinding {
	out := make([]InterfaceSDKBinding, len(bindings))
	for i, binding := range bindings {
		out[i] = binding
		out[i].Method = strings.TrimSpace(binding.Method)
		out[i].Package = strings.TrimSpace(binding.Package)
		out[i].Service = strings.TrimSpace(binding.Service)
		out[i].MethodName = strings.TrimSpace(binding.MethodName)
		out[i].RequestType = strings.TrimSpace(binding.RequestType)
		out[i].ResponseType = strings.TrimSpace(binding.ResponseType)
	}
	return out
}

func cloneInterfaceWalletActions(actions []InterfaceWalletAction) []InterfaceWalletAction {
	out := make([]InterfaceWalletAction, len(actions))
	for i, action := range actions {
		out[i] = action
		out[i].Method = strings.TrimSpace(action.Method)
		out[i].Title = strings.TrimSpace(action.Title)
		out[i].Description = strings.TrimSpace(action.Description)
		out[i].Category = strings.TrimSpace(action.Category)
		out[i].Icon = strings.TrimSpace(action.Icon)
		out[i].ConfirmLabel = strings.TrimSpace(action.ConfirmLabel)
		out[i].Inputs = cloneParamDescriptors(action.Inputs)
		out[i].Outputs = cloneResultDescriptors(action.Outputs)
	}
	return out
}

func cloneParamDescriptors(params []InterfaceParamDescriptor) []InterfaceParamDescriptor {
	out := make([]InterfaceParamDescriptor, len(params))
	for i, param := range params {
		out[i] = param
		out[i].Name = strings.TrimSpace(param.Name)
		out[i].Description = strings.TrimSpace(param.Description)
	}
	return out
}

func cloneResultDescriptors(results []InterfaceResultDescriptor) []InterfaceResultDescriptor {
	return cloneParamDescriptors(results)
}

func cloneStringList(values []string) []string {
	out := make([]string, len(values))
	for i, value := range values {
		out[i] = strings.TrimSpace(value)
	}
	return out
}

func cloneSortedStringList(values []string) []string {
	out := cloneStringList(values)
	sort.Strings(out)
	return out
}
