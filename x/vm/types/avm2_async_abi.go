package types

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"

	sdkmath "cosmossdk.io/math"

	zonestypes "github.com/sovereign-l1/l1/x/zones/types"
)

const (
	AVM2PayloadTypePromiseResolution = "promise.resolution"
	AVM2PayloadTypePromiseTimeout    = "promise.timeout"
	MaxAVM2ABIEncodingLength         = 64
)

type AVM2AsyncCallPlan struct {
	Height             uint64
	CallerContract     string
	Message            AVMAsyncMessage
	Promise            AVM2PromiseState
	AwaitNonBlocking   bool
	PersistedPromiseID string
	OutboxRoot         string
	PromiseRoot        string
	PlanHash           string
}

type AVM2PromiseResolution struct {
	PromiseID         string
	OriginalMessageID string
	ResolutionMessage AVMAsyncMessage
	Status            AVM2PromiseStatus
	ReceiptHash       string
	ReturnHash        string
	DeliveryHeight    uint64
	ResolutionHash    string
}

type AVM2PromiseTimeoutTask struct {
	PromiseID          string
	Contract           string
	DueHeight          uint64
	HandlerPayloadHash string
	TimeoutMessage     AVMAsyncMessage
	TaskHash           string
}

type AVM2MethodDescriptor struct {
	Selector           string
	Name               string
	ArgumentSchemaHash string
	ReturnSchemaHash   string
	ArgumentEncoding   string
	GasHint            uint64
	MethodHash         string
}

type AVM2EventDescriptor struct {
	Name       string
	SchemaHash string
	EventHash  string
}

type AVM2ErrorDescriptor struct {
	Code       string
	SchemaHash string
	ErrorHash  string
}

type AVM2FundRequirement struct {
	Selector        string
	Denom           string
	Minimum         sdkmath.Int
	RequirementHash string
}

type AVM2GasHint struct {
	Selector string
	Estimate uint64
	HintHash string
}

type AVM2ABIIntrospectionDescriptor struct {
	ABIVersion           uint64
	CodeID               uint64
	CodeHash             string
	Methods              []AVM2MethodDescriptor
	Events               []AVM2EventDescriptor
	Errors               []AVM2ErrorDescriptor
	RequiredFunds        []AVM2FundRequirement
	GasHints             []AVM2GasHint
	IdentityNameOptional string
	InterfaceHash        string
}

type AVM2CallFund struct {
	Denom  string
	Amount sdkmath.Int
}

type AVM2ABIMethodCall struct {
	InterfaceHash    string
	MethodSelector   string
	ArgumentEncoding string
	ArgumentHash     string
	Funds            []AVM2CallFund
	CallHash         string
}

type AVM2ABIIdentityBinding struct {
	Name               string
	InterfaceHash      string
	ResolverRecordHash string
	BindingHash        string
}

func NewAVM2AsyncCallPlan(plan AVM2AsyncCallPlan) (AVM2AsyncCallPlan, error) {
	plan = canonicalAVM2AsyncCallPlan(plan)
	plan.OutboxRoot = ComputeAVM2MessageRoot([]AVMAsyncMessage{plan.Message})
	plan.PromiseRoot = ComputeAVM2PromiseRoot([]AVM2PromiseState{plan.Promise})
	plan.PlanHash = ComputeAVM2AsyncCallPlanHash(plan)
	return plan, plan.Validate()
}

func NewAVM2PromiseResolution(resolution AVM2PromiseResolution) (AVM2PromiseResolution, error) {
	resolution = canonicalAVM2PromiseResolution(resolution)
	resolution.ResolutionHash = ComputeAVM2PromiseResolutionHash(resolution)
	return resolution, resolution.Validate()
}

func NewAVM2PromiseTimeoutTask(task AVM2PromiseTimeoutTask) (AVM2PromiseTimeoutTask, error) {
	task = canonicalAVM2PromiseTimeoutTask(task)
	task.TaskHash = ComputeAVM2PromiseTimeoutTaskHash(task)
	return task, task.Validate()
}

func NewAVM2ABIIntrospectionDescriptor(descriptor AVM2ABIIntrospectionDescriptor) (AVM2ABIIntrospectionDescriptor, error) {
	descriptor = canonicalAVM2ABIIntrospectionDescriptor(descriptor)
	for i := range descriptor.Methods {
		descriptor.Methods[i].MethodHash = ComputeAVM2MethodDescriptorHash(descriptor.Methods[i])
	}
	for i := range descriptor.Events {
		descriptor.Events[i].EventHash = ComputeAVM2EventDescriptorHash(descriptor.Events[i])
	}
	for i := range descriptor.Errors {
		descriptor.Errors[i].ErrorHash = ComputeAVM2ErrorDescriptorHash(descriptor.Errors[i])
	}
	for i := range descriptor.RequiredFunds {
		descriptor.RequiredFunds[i].RequirementHash = ComputeAVM2FundRequirementHash(descriptor.RequiredFunds[i])
	}
	for i := range descriptor.GasHints {
		descriptor.GasHints[i].HintHash = ComputeAVM2GasHintHash(descriptor.GasHints[i])
	}
	descriptor = canonicalAVM2ABIIntrospectionDescriptor(descriptor)
	descriptor.InterfaceHash = ComputeAVM2ABIIntrospectionHash(descriptor)
	return descriptor, descriptor.Validate()
}

func NewAVM2ABIMethodCall(call AVM2ABIMethodCall) (AVM2ABIMethodCall, error) {
	call = canonicalAVM2ABIMethodCall(call)
	call.CallHash = ComputeAVM2ABIMethodCallHash(call)
	return call, call.Validate()
}

func NewAVM2ABIIdentityBinding(binding AVM2ABIIdentityBinding) (AVM2ABIIdentityBinding, error) {
	binding = canonicalAVM2ABIIdentityBinding(binding)
	binding.BindingHash = ComputeAVM2ABIIdentityBindingHash(binding)
	return binding, binding.Validate()
}

func ApplyAVM2PromiseResolution(promise AVM2PromiseState, resolution AVM2PromiseResolution) (AVM2PromiseState, error) {
	promise = canonicalAVM2PromiseState(promise)
	resolution = canonicalAVM2PromiseResolution(resolution)
	if err := promise.Validate(); err != nil {
		return AVM2PromiseState{}, err
	}
	if err := resolution.Validate(); err != nil {
		return AVM2PromiseState{}, err
	}
	if promise.PromiseID != resolution.PromiseID {
		return AVM2PromiseState{}, errors.New("AVM 2.0 promise resolution id mismatch")
	}
	if promise.MessageID != resolution.OriginalMessageID {
		return AVM2PromiseState{}, errors.New("AVM 2.0 promise resolution message mismatch")
	}
	if resolution.DeliveryHeight <= promise.CreatedHeight {
		return AVM2PromiseState{}, errors.New("AVM 2.0 promise resolution must be delivered in future height")
	}
	updated := promise
	updated.Status = resolution.Status
	updated.ReceiptHash = resolution.ReceiptHash
	updated.ReturnHash = resolution.ReturnHash
	updated.PromiseHash = ComputeAVM2PromiseHash(updated)
	return updated, updated.Validate()
}

func ScheduleAVM2PromiseTimeout(promise AVM2PromiseState, timeoutMessage AVMAsyncMessage) (AVM2PromiseTimeoutTask, error) {
	promise = canonicalAVM2PromiseState(promise)
	timeoutMessage = canonicalAVMAsyncMessage(timeoutMessage)
	if err := promise.Validate(); err != nil {
		return AVM2PromiseTimeoutTask{}, err
	}
	if promise.Status != AVM2PromisePending {
		return AVM2PromiseTimeoutTask{}, errors.New("AVM 2.0 only pending promises can schedule timeout")
	}
	task := AVM2PromiseTimeoutTask{
		PromiseID:          promise.PromiseID,
		Contract:           promise.Contract,
		DueHeight:          promise.ExpiryHeight,
		HandlerPayloadHash: ComputeAVM2BytesHash([]byte(promise.PromiseID + ":timeout")),
		TimeoutMessage:     timeoutMessage,
	}
	return NewAVM2PromiseTimeoutTask(task)
}

func BindAVM2ABIToCode(code AVM2CodeRecord, descriptor AVM2ABIIntrospectionDescriptor) error {
	code = canonicalAVM2CodeRecord(code)
	descriptor = canonicalAVM2ABIIntrospectionDescriptor(descriptor)
	if err := code.Validate(); err != nil {
		return err
	}
	if err := descriptor.Validate(); err != nil {
		return err
	}
	if code.CodeID != descriptor.CodeID {
		return errors.New("AVM 2.0 ABI code id mismatch")
	}
	if code.CodeHash != descriptor.CodeHash {
		return errors.New("AVM 2.0 ABI code hash mismatch")
	}
	if code.ABIHash != descriptor.InterfaceHash {
		return errors.New("AVM 2.0 code ABI hash must commit introspection descriptor")
	}
	return nil
}

func ValidateAVM2ABIMethodCall(descriptor AVM2ABIIntrospectionDescriptor, call AVM2ABIMethodCall) error {
	descriptor = canonicalAVM2ABIIntrospectionDescriptor(descriptor)
	call = canonicalAVM2ABIMethodCall(call)
	if err := descriptor.Validate(); err != nil {
		return err
	}
	if err := call.Validate(); err != nil {
		return err
	}
	if call.InterfaceHash != descriptor.InterfaceHash {
		return errors.New("AVM 2.0 method call interface hash mismatch")
	}
	method, found := descriptor.FindMethod(call.MethodSelector)
	if !found {
		return errors.New("AVM 2.0 method selector not found")
	}
	if method.ArgumentEncoding != call.ArgumentEncoding {
		return errors.New("AVM 2.0 method argument encoding mismatch")
	}
	return validateAVM2CallFunds(descriptor.RequiredFundsForSelector(call.MethodSelector), call.Funds)
}

func (p AVM2AsyncCallPlan) Validate() error {
	p = canonicalAVM2AsyncCallPlan(p)
	if p.Height == 0 {
		return errors.New("AVM 2.0 async call height must be positive")
	}
	if err := validateEngineToken("AVM 2.0 async caller contract", p.CallerContract, MaxAVMStateKeySegmentLength); err != nil {
		return err
	}
	if err := p.Message.Validate(); err != nil {
		return err
	}
	if err := p.Promise.Validate(); err != nil {
		return err
	}
	if !p.AwaitNonBlocking {
		return errors.New("AVM 2.0 async await must be non-blocking")
	}
	if p.Promise.Status != AVM2PromisePending {
		return errors.New("AVM 2.0 async call must persist pending promise")
	}
	if p.Promise.Contract != p.CallerContract {
		return errors.New("AVM 2.0 async promise contract mismatch")
	}
	if p.Promise.MessageID != p.Message.ID {
		return errors.New("AVM 2.0 async promise message mismatch")
	}
	if p.PersistedPromiseID != p.Promise.PromiseID {
		return errors.New("AVM 2.0 persisted promise id mismatch")
	}
	if p.Message.SourceZone == p.Message.DestinationZone {
		return errors.New("AVM 2.0 cross-zone contract call must use cross-zone message")
	}
	if p.Message.CreatedHeight != p.Height {
		return errors.New("AVM 2.0 async message created height must match plan height")
	}
	if p.Promise.CreatedHeight != p.Height {
		return errors.New("AVM 2.0 async promise created height must match plan height")
	}
	if p.Promise.ExpiryHeight != p.Message.ExpiryHeight {
		return errors.New("AVM 2.0 async promise expiry must match message expiry")
	}
	if p.OutboxRoot != ComputeAVM2MessageRoot([]AVMAsyncMessage{p.Message}) {
		return errors.New("AVM 2.0 async outbox root mismatch")
	}
	if p.PromiseRoot != ComputeAVM2PromiseRoot([]AVM2PromiseState{p.Promise}) {
		return errors.New("AVM 2.0 async promise root mismatch")
	}
	if err := zonestypes.ValidateHash("AVM 2.0 async call plan hash", p.PlanHash); err != nil {
		return err
	}
	if p.PlanHash != ComputeAVM2AsyncCallPlanHash(p) {
		return errors.New("AVM 2.0 async call plan hash mismatch")
	}
	return nil
}

func (r AVM2PromiseResolution) Validate() error {
	r = canonicalAVM2PromiseResolution(r)
	if err := zonestypes.ValidateHash("AVM 2.0 promise resolution promise id", r.PromiseID); err != nil {
		return err
	}
	if err := zonestypes.ValidateHash("AVM 2.0 promise resolution original message id", r.OriginalMessageID); err != nil {
		return err
	}
	if err := r.ResolutionMessage.Validate(); err != nil {
		return err
	}
	if !IsAVM2PromiseStatus(r.Status) || r.Status == AVM2PromisePending {
		return errors.New("AVM 2.0 promise resolution status must be terminal")
	}
	if err := zonestypes.ValidateHash("AVM 2.0 promise resolution receipt hash", r.ReceiptHash); err != nil {
		return err
	}
	if r.ReturnHash != "" {
		if err := zonestypes.ValidateHash("AVM 2.0 promise resolution return hash", r.ReturnHash); err != nil {
			return err
		}
	}
	if r.DeliveryHeight == 0 {
		return errors.New("AVM 2.0 promise resolution delivery height must be positive")
	}
	if r.ResolutionMessage.PayloadType != AVM2PayloadTypePromiseResolution {
		return errors.New("AVM 2.0 promise resolution must use resolution payload type")
	}
	if err := zonestypes.ValidateHash("AVM 2.0 promise resolution hash", r.ResolutionHash); err != nil {
		return err
	}
	if r.ResolutionHash != ComputeAVM2PromiseResolutionHash(r) {
		return errors.New("AVM 2.0 promise resolution hash mismatch")
	}
	return nil
}

func (t AVM2PromiseTimeoutTask) Validate() error {
	t = canonicalAVM2PromiseTimeoutTask(t)
	if err := zonestypes.ValidateHash("AVM 2.0 timeout promise id", t.PromiseID); err != nil {
		return err
	}
	if err := validateEngineToken("AVM 2.0 timeout contract", t.Contract, MaxAVMStateKeySegmentLength); err != nil {
		return err
	}
	if t.DueHeight == 0 {
		return errors.New("AVM 2.0 timeout due height must be positive")
	}
	if err := zonestypes.ValidateHash("AVM 2.0 timeout handler payload hash", t.HandlerPayloadHash); err != nil {
		return err
	}
	if err := t.TimeoutMessage.Validate(); err != nil {
		return err
	}
	if t.TimeoutMessage.PayloadType != AVM2PayloadTypePromiseTimeout {
		return errors.New("AVM 2.0 timeout message must use timeout payload type")
	}
	if t.TimeoutMessage.CreatedHeight < t.DueHeight {
		return errors.New("AVM 2.0 timeout message cannot be created before due height")
	}
	if err := zonestypes.ValidateHash("AVM 2.0 timeout task hash", t.TaskHash); err != nil {
		return err
	}
	if t.TaskHash != ComputeAVM2PromiseTimeoutTaskHash(t) {
		return errors.New("AVM 2.0 timeout task hash mismatch")
	}
	return nil
}

func (d AVM2MethodDescriptor) Validate() error {
	d = canonicalAVM2MethodDescriptor(d)
	if err := validateEngineToken("AVM 2.0 method selector", d.Selector, MaxAVM2TokenLength); err != nil {
		return err
	}
	if err := validateEngineToken("AVM 2.0 method name", d.Name, MaxAVM2TokenLength); err != nil {
		return err
	}
	if err := zonestypes.ValidateHash("AVM 2.0 method argument schema hash", d.ArgumentSchemaHash); err != nil {
		return err
	}
	if err := zonestypes.ValidateHash("AVM 2.0 method return schema hash", d.ReturnSchemaHash); err != nil {
		return err
	}
	if err := validateEngineToken("AVM 2.0 method argument encoding", d.ArgumentEncoding, MaxAVM2ABIEncodingLength); err != nil {
		return err
	}
	if d.GasHint == 0 {
		return errors.New("AVM 2.0 method gas hint must be positive")
	}
	if err := zonestypes.ValidateHash("AVM 2.0 method hash", d.MethodHash); err != nil {
		return err
	}
	if d.MethodHash != ComputeAVM2MethodDescriptorHash(d) {
		return errors.New("AVM 2.0 method hash mismatch")
	}
	return nil
}

func (d AVM2EventDescriptor) Validate() error {
	d = canonicalAVM2EventDescriptor(d)
	if err := validateEngineToken("AVM 2.0 ABI event name", d.Name, MaxAVM2TokenLength); err != nil {
		return err
	}
	if err := zonestypes.ValidateHash("AVM 2.0 ABI event schema hash", d.SchemaHash); err != nil {
		return err
	}
	if err := zonestypes.ValidateHash("AVM 2.0 ABI event hash", d.EventHash); err != nil {
		return err
	}
	if d.EventHash != ComputeAVM2EventDescriptorHash(d) {
		return errors.New("AVM 2.0 ABI event hash mismatch")
	}
	return nil
}

func (d AVM2ErrorDescriptor) Validate() error {
	d = canonicalAVM2ErrorDescriptor(d)
	if err := validateEngineToken("AVM 2.0 ABI error code", d.Code, MaxAVM2TokenLength); err != nil {
		return err
	}
	if err := zonestypes.ValidateHash("AVM 2.0 ABI error schema hash", d.SchemaHash); err != nil {
		return err
	}
	if err := zonestypes.ValidateHash("AVM 2.0 ABI error hash", d.ErrorHash); err != nil {
		return err
	}
	if d.ErrorHash != ComputeAVM2ErrorDescriptorHash(d) {
		return errors.New("AVM 2.0 ABI error hash mismatch")
	}
	return nil
}

func (r AVM2FundRequirement) Validate() error {
	r = canonicalAVM2FundRequirement(r)
	if err := validateEngineToken("AVM 2.0 fund selector", r.Selector, MaxAVM2TokenLength); err != nil {
		return err
	}
	if err := validateEngineToken("AVM 2.0 fund denom", r.Denom, MaxAVM2TokenLength); err != nil {
		return err
	}
	if r.Minimum.IsNil() || r.Minimum.IsNegative() {
		return errors.New("AVM 2.0 fund minimum must be non-negative")
	}
	if err := zonestypes.ValidateHash("AVM 2.0 fund requirement hash", r.RequirementHash); err != nil {
		return err
	}
	if r.RequirementHash != ComputeAVM2FundRequirementHash(r) {
		return errors.New("AVM 2.0 fund requirement hash mismatch")
	}
	return nil
}

func (h AVM2GasHint) Validate() error {
	h = canonicalAVM2GasHint(h)
	if err := validateEngineToken("AVM 2.0 gas hint selector", h.Selector, MaxAVM2TokenLength); err != nil {
		return err
	}
	if h.Estimate == 0 {
		return errors.New("AVM 2.0 gas hint estimate must be positive")
	}
	if err := zonestypes.ValidateHash("AVM 2.0 gas hint hash", h.HintHash); err != nil {
		return err
	}
	if h.HintHash != ComputeAVM2GasHintHash(h) {
		return errors.New("AVM 2.0 gas hint hash mismatch")
	}
	return nil
}

func (d AVM2ABIIntrospectionDescriptor) Validate() error {
	d = canonicalAVM2ABIIntrospectionDescriptor(d)
	if d.ABIVersion == 0 || d.CodeID == 0 {
		return errors.New("AVM 2.0 ABI introspection version and code id must be positive")
	}
	if err := zonestypes.ValidateHash("AVM 2.0 ABI code hash", d.CodeHash); err != nil {
		return err
	}
	if len(d.Methods) == 0 {
		return errors.New("AVM 2.0 ABI must declare methods")
	}
	if err := validateAVM2Methods(d.Methods); err != nil {
		return err
	}
	if err := validateAVM2Events(d.Events); err != nil {
		return err
	}
	if err := validateAVM2Errors(d.Errors); err != nil {
		return err
	}
	if err := validateAVM2FundRequirements(d.RequiredFunds, d.Methods); err != nil {
		return err
	}
	if err := validateAVM2GasHints(d.GasHints, d.Methods); err != nil {
		return err
	}
	if d.IdentityNameOptional != "" {
		if err := validateEngineToken("AVM 2.0 ABI identity name", d.IdentityNameOptional, MaxAVM2TokenLength); err != nil {
			return err
		}
	}
	if err := zonestypes.ValidateHash("AVM 2.0 ABI introspection interface hash", d.InterfaceHash); err != nil {
		return err
	}
	if d.InterfaceHash != ComputeAVM2ABIIntrospectionHash(d) {
		return errors.New("AVM 2.0 ABI introspection hash mismatch")
	}
	return nil
}

func (c AVM2ABIMethodCall) Validate() error {
	c = canonicalAVM2ABIMethodCall(c)
	if err := zonestypes.ValidateHash("AVM 2.0 call interface hash", c.InterfaceHash); err != nil {
		return err
	}
	if err := validateEngineToken("AVM 2.0 call method selector", c.MethodSelector, MaxAVM2TokenLength); err != nil {
		return err
	}
	if err := validateEngineToken("AVM 2.0 call argument encoding", c.ArgumentEncoding, MaxAVM2ABIEncodingLength); err != nil {
		return err
	}
	if err := zonestypes.ValidateHash("AVM 2.0 call argument hash", c.ArgumentHash); err != nil {
		return err
	}
	if err := validateAVM2CallFundRecords(c.Funds); err != nil {
		return err
	}
	if err := zonestypes.ValidateHash("AVM 2.0 method call hash", c.CallHash); err != nil {
		return err
	}
	if c.CallHash != ComputeAVM2ABIMethodCallHash(c) {
		return errors.New("AVM 2.0 method call hash mismatch")
	}
	return nil
}

func (b AVM2ABIIdentityBinding) Validate() error {
	b = canonicalAVM2ABIIdentityBinding(b)
	if err := validateEngineToken("AVM 2.0 ABI identity name", b.Name, MaxAVM2TokenLength); err != nil {
		return err
	}
	if !strings.HasSuffix(b.Name, ".aet") {
		return errors.New("AVM 2.0 ABI identity binding must use .aet name")
	}
	if err := zonestypes.ValidateHash("AVM 2.0 ABI identity interface hash", b.InterfaceHash); err != nil {
		return err
	}
	if err := zonestypes.ValidateHash("AVM 2.0 ABI identity resolver record hash", b.ResolverRecordHash); err != nil {
		return err
	}
	if err := zonestypes.ValidateHash("AVM 2.0 ABI identity binding hash", b.BindingHash); err != nil {
		return err
	}
	if b.BindingHash != ComputeAVM2ABIIdentityBindingHash(b) {
		return errors.New("AVM 2.0 ABI identity binding hash mismatch")
	}
	return nil
}

func (d AVM2ABIIntrospectionDescriptor) FindMethod(selector string) (AVM2MethodDescriptor, bool) {
	d = canonicalAVM2ABIIntrospectionDescriptor(d)
	selector = strings.TrimSpace(selector)
	for _, method := range d.Methods {
		if method.Selector == selector {
			return method, true
		}
	}
	return AVM2MethodDescriptor{}, false
}

func (d AVM2ABIIntrospectionDescriptor) RequiredFundsForSelector(selector string) []AVM2FundRequirement {
	d = canonicalAVM2ABIIntrospectionDescriptor(d)
	selector = strings.TrimSpace(selector)
	var out []AVM2FundRequirement
	for _, requirement := range d.RequiredFunds {
		if requirement.Selector == selector {
			out = append(out, requirement)
		}
	}
	return out
}

func ComputeAVM2AsyncCallPlanHash(plan AVM2AsyncCallPlan) string {
	plan = canonicalAVM2AsyncCallPlan(plan)
	plan.PlanHash = ""
	h := sha256.New()
	writeEnginePart(h, "aetra-avm2-async-call-plan-v1")
	writeEngineUint64(h, plan.Height)
	writeEnginePart(h, plan.CallerContract)
	writeEnginePart(h, plan.Message.ID)
	writeEnginePart(h, plan.Promise.PromiseID)
	writeEngineBool(h, plan.AwaitNonBlocking)
	writeEnginePart(h, plan.PersistedPromiseID)
	writeEnginePart(h, plan.OutboxRoot)
	writeEnginePart(h, plan.PromiseRoot)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVM2PromiseResolutionHash(resolution AVM2PromiseResolution) string {
	resolution = canonicalAVM2PromiseResolution(resolution)
	resolution.ResolutionHash = ""
	h := sha256.New()
	writeEnginePart(h, "aetra-avm2-promise-resolution-v1")
	writeEnginePart(h, resolution.PromiseID)
	writeEnginePart(h, resolution.OriginalMessageID)
	writeEnginePart(h, resolution.ResolutionMessage.ID)
	writeEnginePart(h, string(resolution.Status))
	writeEnginePart(h, resolution.ReceiptHash)
	writeEnginePart(h, resolution.ReturnHash)
	writeEngineUint64(h, resolution.DeliveryHeight)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVM2PromiseTimeoutTaskHash(task AVM2PromiseTimeoutTask) string {
	task = canonicalAVM2PromiseTimeoutTask(task)
	task.TaskHash = ""
	h := sha256.New()
	writeEnginePart(h, "aetra-avm2-promise-timeout-task-v1")
	writeEnginePart(h, task.PromiseID)
	writeEnginePart(h, task.Contract)
	writeEngineUint64(h, task.DueHeight)
	writeEnginePart(h, task.HandlerPayloadHash)
	writeEnginePart(h, task.TimeoutMessage.ID)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVM2MethodDescriptorHash(method AVM2MethodDescriptor) string {
	method = canonicalAVM2MethodDescriptor(method)
	method.MethodHash = ""
	h := sha256.New()
	writeEnginePart(h, "aetra-avm2-method-descriptor-v1")
	writeEnginePart(h, method.Selector)
	writeEnginePart(h, method.Name)
	writeEnginePart(h, method.ArgumentSchemaHash)
	writeEnginePart(h, method.ReturnSchemaHash)
	writeEnginePart(h, method.ArgumentEncoding)
	writeEngineUint64(h, method.GasHint)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVM2EventDescriptorHash(event AVM2EventDescriptor) string {
	event = canonicalAVM2EventDescriptor(event)
	event.EventHash = ""
	h := sha256.New()
	writeEnginePart(h, "aetra-avm2-abi-event-descriptor-v1")
	writeEnginePart(h, event.Name)
	writeEnginePart(h, event.SchemaHash)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVM2ErrorDescriptorHash(errDesc AVM2ErrorDescriptor) string {
	errDesc = canonicalAVM2ErrorDescriptor(errDesc)
	errDesc.ErrorHash = ""
	h := sha256.New()
	writeEnginePart(h, "aetra-avm2-error-descriptor-v1")
	writeEnginePart(h, errDesc.Code)
	writeEnginePart(h, errDesc.SchemaHash)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVM2FundRequirementHash(requirement AVM2FundRequirement) string {
	requirement = canonicalAVM2FundRequirement(requirement)
	requirement.RequirementHash = ""
	h := sha256.New()
	writeEnginePart(h, "aetra-avm2-fund-requirement-v1")
	writeEnginePart(h, requirement.Selector)
	writeEnginePart(h, requirement.Denom)
	writeEnginePart(h, requirement.Minimum.String())
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVM2GasHintHash(hint AVM2GasHint) string {
	hint = canonicalAVM2GasHint(hint)
	hint.HintHash = ""
	h := sha256.New()
	writeEnginePart(h, "aetra-avm2-gas-hint-v1")
	writeEnginePart(h, hint.Selector)
	writeEngineUint64(h, hint.Estimate)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVM2ABIIntrospectionHash(descriptor AVM2ABIIntrospectionDescriptor) string {
	descriptor = canonicalAVM2ABIIntrospectionDescriptor(descriptor)
	descriptor.InterfaceHash = ""
	h := sha256.New()
	writeEnginePart(h, "aetra-avm2-abi-introspection-v1")
	writeEngineUint64(h, descriptor.ABIVersion)
	writeEngineUint64(h, descriptor.CodeID)
	writeEnginePart(h, descriptor.CodeHash)
	writeEngineUint64(h, uint64(len(descriptor.Methods)))
	for _, method := range descriptor.Methods {
		writeEnginePart(h, method.MethodHash)
	}
	writeEngineUint64(h, uint64(len(descriptor.Events)))
	for _, event := range descriptor.Events {
		writeEnginePart(h, event.EventHash)
	}
	writeEngineUint64(h, uint64(len(descriptor.Errors)))
	for _, errDesc := range descriptor.Errors {
		writeEnginePart(h, errDesc.ErrorHash)
	}
	writeEngineUint64(h, uint64(len(descriptor.RequiredFunds)))
	for _, requirement := range descriptor.RequiredFunds {
		writeEnginePart(h, requirement.RequirementHash)
	}
	writeEngineUint64(h, uint64(len(descriptor.GasHints)))
	for _, hint := range descriptor.GasHints {
		writeEnginePart(h, hint.HintHash)
	}
	writeEnginePart(h, descriptor.IdentityNameOptional)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVM2ABIMethodCallHash(call AVM2ABIMethodCall) string {
	call = canonicalAVM2ABIMethodCall(call)
	call.CallHash = ""
	h := sha256.New()
	writeEnginePart(h, "aetra-avm2-method-call-v1")
	writeEnginePart(h, call.InterfaceHash)
	writeEnginePart(h, call.MethodSelector)
	writeEnginePart(h, call.ArgumentEncoding)
	writeEnginePart(h, call.ArgumentHash)
	writeEngineUint64(h, uint64(len(call.Funds)))
	for _, fund := range call.Funds {
		writeEnginePart(h, fund.Denom)
		writeEnginePart(h, fund.Amount.String())
	}
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVM2ABIIdentityBindingHash(binding AVM2ABIIdentityBinding) string {
	binding = canonicalAVM2ABIIdentityBinding(binding)
	binding.BindingHash = ""
	h := sha256.New()
	writeEnginePart(h, "aetra-avm2-abi-identity-binding-v1")
	writeEnginePart(h, binding.Name)
	writeEnginePart(h, binding.InterfaceHash)
	writeEnginePart(h, binding.ResolverRecordHash)
	return hex.EncodeToString(h.Sum(nil))
}

func canonicalAVM2AsyncCallPlan(plan AVM2AsyncCallPlan) AVM2AsyncCallPlan {
	plan.CallerContract = strings.TrimSpace(plan.CallerContract)
	plan.Message = canonicalAVMAsyncMessage(plan.Message)
	plan.Promise = canonicalAVM2PromiseState(plan.Promise)
	plan.PersistedPromiseID = strings.TrimSpace(plan.PersistedPromiseID)
	plan.OutboxRoot = strings.TrimSpace(plan.OutboxRoot)
	plan.PromiseRoot = strings.TrimSpace(plan.PromiseRoot)
	plan.PlanHash = strings.TrimSpace(plan.PlanHash)
	return plan
}

func canonicalAVM2PromiseResolution(resolution AVM2PromiseResolution) AVM2PromiseResolution {
	resolution.PromiseID = strings.TrimSpace(resolution.PromiseID)
	resolution.OriginalMessageID = strings.TrimSpace(resolution.OriginalMessageID)
	resolution.ResolutionMessage = canonicalAVMAsyncMessage(resolution.ResolutionMessage)
	resolution.ReceiptHash = strings.TrimSpace(resolution.ReceiptHash)
	resolution.ReturnHash = strings.TrimSpace(resolution.ReturnHash)
	resolution.ResolutionHash = strings.TrimSpace(resolution.ResolutionHash)
	return resolution
}

func canonicalAVM2PromiseTimeoutTask(task AVM2PromiseTimeoutTask) AVM2PromiseTimeoutTask {
	task.PromiseID = strings.TrimSpace(task.PromiseID)
	task.Contract = strings.TrimSpace(task.Contract)
	task.HandlerPayloadHash = strings.TrimSpace(task.HandlerPayloadHash)
	task.TimeoutMessage = canonicalAVMAsyncMessage(task.TimeoutMessage)
	task.TaskHash = strings.TrimSpace(task.TaskHash)
	return task
}

func canonicalAVM2MethodDescriptor(method AVM2MethodDescriptor) AVM2MethodDescriptor {
	method.Selector = strings.TrimSpace(method.Selector)
	method.Name = strings.TrimSpace(method.Name)
	method.ArgumentSchemaHash = strings.TrimSpace(method.ArgumentSchemaHash)
	method.ReturnSchemaHash = strings.TrimSpace(method.ReturnSchemaHash)
	method.ArgumentEncoding = strings.TrimSpace(method.ArgumentEncoding)
	method.MethodHash = strings.TrimSpace(method.MethodHash)
	return method
}

func canonicalAVM2EventDescriptor(event AVM2EventDescriptor) AVM2EventDescriptor {
	event.Name = strings.TrimSpace(event.Name)
	event.SchemaHash = strings.TrimSpace(event.SchemaHash)
	event.EventHash = strings.TrimSpace(event.EventHash)
	return event
}

func canonicalAVM2ErrorDescriptor(errDesc AVM2ErrorDescriptor) AVM2ErrorDescriptor {
	errDesc.Code = strings.TrimSpace(errDesc.Code)
	errDesc.SchemaHash = strings.TrimSpace(errDesc.SchemaHash)
	errDesc.ErrorHash = strings.TrimSpace(errDesc.ErrorHash)
	return errDesc
}

func canonicalAVM2FundRequirement(requirement AVM2FundRequirement) AVM2FundRequirement {
	requirement.Selector = strings.TrimSpace(requirement.Selector)
	requirement.Denom = strings.TrimSpace(requirement.Denom)
	requirement.RequirementHash = strings.TrimSpace(requirement.RequirementHash)
	return requirement
}

func canonicalAVM2GasHint(hint AVM2GasHint) AVM2GasHint {
	hint.Selector = strings.TrimSpace(hint.Selector)
	hint.HintHash = strings.TrimSpace(hint.HintHash)
	return hint
}

func canonicalAVM2ABIIntrospectionDescriptor(descriptor AVM2ABIIntrospectionDescriptor) AVM2ABIIntrospectionDescriptor {
	descriptor.CodeHash = strings.TrimSpace(descriptor.CodeHash)
	descriptor.IdentityNameOptional = strings.TrimSpace(descriptor.IdentityNameOptional)
	descriptor.InterfaceHash = strings.TrimSpace(descriptor.InterfaceHash)
	descriptor.Methods = append([]AVM2MethodDescriptor(nil), descriptor.Methods...)
	for i := range descriptor.Methods {
		descriptor.Methods[i] = canonicalAVM2MethodDescriptor(descriptor.Methods[i])
	}
	sort.SliceStable(descriptor.Methods, func(i, j int) bool { return descriptor.Methods[i].Selector < descriptor.Methods[j].Selector })
	descriptor.Events = append([]AVM2EventDescriptor(nil), descriptor.Events...)
	for i := range descriptor.Events {
		descriptor.Events[i] = canonicalAVM2EventDescriptor(descriptor.Events[i])
	}
	sort.SliceStable(descriptor.Events, func(i, j int) bool { return descriptor.Events[i].Name < descriptor.Events[j].Name })
	descriptor.Errors = append([]AVM2ErrorDescriptor(nil), descriptor.Errors...)
	for i := range descriptor.Errors {
		descriptor.Errors[i] = canonicalAVM2ErrorDescriptor(descriptor.Errors[i])
	}
	sort.SliceStable(descriptor.Errors, func(i, j int) bool { return descriptor.Errors[i].Code < descriptor.Errors[j].Code })
	descriptor.RequiredFunds = append([]AVM2FundRequirement(nil), descriptor.RequiredFunds...)
	for i := range descriptor.RequiredFunds {
		descriptor.RequiredFunds[i] = canonicalAVM2FundRequirement(descriptor.RequiredFunds[i])
	}
	sort.SliceStable(descriptor.RequiredFunds, func(i, j int) bool {
		if descriptor.RequiredFunds[i].Selector == descriptor.RequiredFunds[j].Selector {
			return descriptor.RequiredFunds[i].Denom < descriptor.RequiredFunds[j].Denom
		}
		return descriptor.RequiredFunds[i].Selector < descriptor.RequiredFunds[j].Selector
	})
	descriptor.GasHints = append([]AVM2GasHint(nil), descriptor.GasHints...)
	for i := range descriptor.GasHints {
		descriptor.GasHints[i] = canonicalAVM2GasHint(descriptor.GasHints[i])
	}
	sort.SliceStable(descriptor.GasHints, func(i, j int) bool { return descriptor.GasHints[i].Selector < descriptor.GasHints[j].Selector })
	return descriptor
}

func canonicalAVM2CallFund(fund AVM2CallFund) AVM2CallFund {
	fund.Denom = strings.TrimSpace(fund.Denom)
	return fund
}

func canonicalAVM2ABIMethodCall(call AVM2ABIMethodCall) AVM2ABIMethodCall {
	call.InterfaceHash = strings.TrimSpace(call.InterfaceHash)
	call.MethodSelector = strings.TrimSpace(call.MethodSelector)
	call.ArgumentEncoding = strings.TrimSpace(call.ArgumentEncoding)
	call.ArgumentHash = strings.TrimSpace(call.ArgumentHash)
	call.CallHash = strings.TrimSpace(call.CallHash)
	call.Funds = append([]AVM2CallFund(nil), call.Funds...)
	for i := range call.Funds {
		call.Funds[i] = canonicalAVM2CallFund(call.Funds[i])
	}
	sort.SliceStable(call.Funds, func(i, j int) bool { return call.Funds[i].Denom < call.Funds[j].Denom })
	return call
}

func canonicalAVM2ABIIdentityBinding(binding AVM2ABIIdentityBinding) AVM2ABIIdentityBinding {
	binding.Name = strings.TrimSpace(binding.Name)
	binding.InterfaceHash = strings.TrimSpace(binding.InterfaceHash)
	binding.ResolverRecordHash = strings.TrimSpace(binding.ResolverRecordHash)
	binding.BindingHash = strings.TrimSpace(binding.BindingHash)
	return binding
}

func validateAVM2Methods(methods []AVM2MethodDescriptor) error {
	seen := make(map[string]struct{}, len(methods))
	for i, method := range methods {
		if err := method.Validate(); err != nil {
			return err
		}
		if _, found := seen[method.Selector]; found {
			return errors.New("duplicate AVM 2.0 method selector")
		}
		seen[method.Selector] = struct{}{}
		if i > 0 && methods[i-1].Selector >= method.Selector {
			return errors.New("AVM 2.0 methods must be sorted canonically")
		}
	}
	return nil
}

func validateAVM2Events(events []AVM2EventDescriptor) error {
	seen := make(map[string]struct{}, len(events))
	for i, event := range events {
		if err := event.Validate(); err != nil {
			return err
		}
		if _, found := seen[event.Name]; found {
			return errors.New("duplicate AVM 2.0 ABI event")
		}
		seen[event.Name] = struct{}{}
		if i > 0 && events[i-1].Name >= event.Name {
			return errors.New("AVM 2.0 ABI events must be sorted canonically")
		}
	}
	return nil
}

func validateAVM2Errors(errorsList []AVM2ErrorDescriptor) error {
	seen := make(map[string]struct{}, len(errorsList))
	for i, errDesc := range errorsList {
		if err := errDesc.Validate(); err != nil {
			return err
		}
		if _, found := seen[errDesc.Code]; found {
			return errors.New("duplicate AVM 2.0 ABI error")
		}
		seen[errDesc.Code] = struct{}{}
		if i > 0 && errorsList[i-1].Code >= errDesc.Code {
			return errors.New("AVM 2.0 ABI errors must be sorted canonically")
		}
	}
	return nil
}

func validateAVM2FundRequirements(requirements []AVM2FundRequirement, methods []AVM2MethodDescriptor) error {
	selectors := make(map[string]struct{}, len(methods))
	for _, method := range methods {
		selectors[method.Selector] = struct{}{}
	}
	seen := make(map[string]struct{}, len(requirements))
	for _, requirement := range requirements {
		if err := requirement.Validate(); err != nil {
			return err
		}
		if _, found := selectors[requirement.Selector]; !found {
			return errors.New("AVM 2.0 fund requirement references missing method")
		}
		key := requirement.Selector + "/" + requirement.Denom
		if _, found := seen[key]; found {
			return errors.New("duplicate AVM 2.0 fund requirement")
		}
		seen[key] = struct{}{}
	}
	return nil
}

func validateAVM2GasHints(hints []AVM2GasHint, methods []AVM2MethodDescriptor) error {
	selectors := make(map[string]struct{}, len(methods))
	for _, method := range methods {
		selectors[method.Selector] = struct{}{}
	}
	seen := make(map[string]struct{}, len(hints))
	for _, hint := range hints {
		if err := hint.Validate(); err != nil {
			return err
		}
		if _, found := selectors[hint.Selector]; !found {
			return errors.New("AVM 2.0 gas hint references missing method")
		}
		if _, found := seen[hint.Selector]; found {
			return errors.New("duplicate AVM 2.0 gas hint")
		}
		seen[hint.Selector] = struct{}{}
	}
	return nil
}

func validateAVM2CallFundRecords(funds []AVM2CallFund) error {
	seen := make(map[string]struct{}, len(funds))
	for i, fund := range funds {
		fund = canonicalAVM2CallFund(fund)
		if err := validateEngineToken("AVM 2.0 call fund denom", fund.Denom, MaxAVM2TokenLength); err != nil {
			return err
		}
		if fund.Amount.IsNil() || fund.Amount.IsNegative() {
			return errors.New("AVM 2.0 call fund amount must be non-negative")
		}
		if _, found := seen[fund.Denom]; found {
			return errors.New("duplicate AVM 2.0 call fund")
		}
		seen[fund.Denom] = struct{}{}
		if i > 0 && funds[i-1].Denom >= fund.Denom {
			return errors.New("AVM 2.0 call funds must be sorted canonically")
		}
	}
	return nil
}

func validateAVM2CallFunds(requirements []AVM2FundRequirement, funds []AVM2CallFund) error {
	available := make(map[string]sdkmath.Int, len(funds))
	for _, fund := range funds {
		fund = canonicalAVM2CallFund(fund)
		available[fund.Denom] = fund.Amount
	}
	for _, requirement := range requirements {
		have, found := available[requirement.Denom]
		if !found {
			return fmt.Errorf("AVM 2.0 method call missing required fund %q", requirement.Denom)
		}
		if have.LT(requirement.Minimum) {
			return fmt.Errorf("AVM 2.0 method call fund %q below requirement", requirement.Denom)
		}
	}
	return nil
}
