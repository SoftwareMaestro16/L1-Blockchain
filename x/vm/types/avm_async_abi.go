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
	AVMPayloadTypePromiseResolution	= "promise.resolution"
	AVMPayloadTypePromiseTimeout	= "promise.timeout"
	MaxAVMABIEncodingLength		= 64
)

type AVMAsyncCallPlan struct {
	Height			uint64
	CallerContract		string
	Message			AVMAsyncMessage
	Promise			AVMPromiseState
	AwaitNonBlocking	bool
	PersistedPromiseID	string
	OutboxRoot		string
	PromiseRoot		string
	PlanHash		string
}

type AVMPromiseResolution struct {
	PromiseID		string
	OriginalMessageID	string
	ResolutionMessage	AVMAsyncMessage
	Status			AVMPromiseStatus
	ReceiptHash		string
	ReturnHash		string
	DeliveryHeight		uint64
	ResolutionHash		string
}

type AVMPromiseTimeoutTask struct {
	PromiseID		string
	Contract		string
	DueHeight		uint64
	HandlerPayloadHash	string
	TimeoutMessage		AVMAsyncMessage
	TaskHash		string
}

type AVM2MethodDescriptor struct {
	Selector		string
	Name			string
	ArgumentSchemaHash	string
	ReturnSchemaHash	string
	ArgumentEncoding	string
	GasHint			uint64
	MethodHash		string
}

type AVM2EventDescriptor struct {
	Name		string
	SchemaHash	string
	EventHash	string
}

type AVMErrorDescriptor struct {
	Code		string
	SchemaHash	string
	ErrorHash	string
}

type AVMFundRequirement struct {
	Selector	string
	Denom		string
	Minimum		sdkmath.Int
	RequirementHash	string
}

type AVMGasHint struct {
	Selector	string
	Estimate	uint64
	HintHash	string
}

type AVMABIIntrospectionDescriptor struct {
	ABIVersion		uint64
	CodeID			uint64
	CodeHash		string
	Methods			[]AVM2MethodDescriptor
	Events			[]AVM2EventDescriptor
	Errors			[]AVMErrorDescriptor
	RequiredFunds		[]AVMFundRequirement
	GasHints		[]AVMGasHint
	IdentityNameOptional	string
	InterfaceHash		string
}

type AVMCallFund struct {
	Denom	string
	Amount	sdkmath.Int
}

type AVMABIMethodCall struct {
	InterfaceHash		string
	MethodSelector		string
	ArgumentEncoding	string
	ArgumentHash		string
	Funds			[]AVMCallFund
	CallHash		string
}

type AVMABIIdentityBinding struct {
	Name			string
	InterfaceHash		string
	ResolverRecordHash	string
	BindingHash		string
}

func NewAVMAsyncCallPlan(plan AVMAsyncCallPlan) (AVMAsyncCallPlan, error) {
	plan = canonicalAVMAsyncCallPlan(plan)
	plan.OutboxRoot = ComputeAVMMessageRoot([]AVMAsyncMessage{plan.Message})
	plan.PromiseRoot = ComputeAVMPromiseRoot([]AVMPromiseState{plan.Promise})
	plan.PlanHash = ComputeAVMAsyncCallPlanHash(plan)
	return plan, plan.Validate()
}

func NewAVMPromiseResolution(resolution AVMPromiseResolution) (AVMPromiseResolution, error) {
	resolution = canonicalAVMPromiseResolution(resolution)
	resolution.ResolutionHash = ComputeAVMPromiseResolutionHash(resolution)
	return resolution, resolution.Validate()
}

func NewAVMPromiseTimeoutTask(task AVMPromiseTimeoutTask) (AVMPromiseTimeoutTask, error) {
	task = canonicalAVMPromiseTimeoutTask(task)
	task.TaskHash = ComputeAVMPromiseTimeoutTaskHash(task)
	return task, task.Validate()
}

func NewAVMABIIntrospectionDescriptor(descriptor AVMABIIntrospectionDescriptor) (AVMABIIntrospectionDescriptor, error) {
	descriptor = canonicalAVMABIIntrospectionDescriptor(descriptor)
	for i := range descriptor.Methods {
		descriptor.Methods[i].MethodHash = ComputeAVM2MethodDescriptorHash(descriptor.Methods[i])
	}
	for i := range descriptor.Events {
		descriptor.Events[i].EventHash = ComputeAVM2EventDescriptorHash(descriptor.Events[i])
	}
	for i := range descriptor.Errors {
		descriptor.Errors[i].ErrorHash = ComputeAVMErrorDescriptorHash(descriptor.Errors[i])
	}
	for i := range descriptor.RequiredFunds {
		descriptor.RequiredFunds[i].RequirementHash = ComputeAVMFundRequirementHash(descriptor.RequiredFunds[i])
	}
	for i := range descriptor.GasHints {
		descriptor.GasHints[i].HintHash = ComputeAVMGasHintHash(descriptor.GasHints[i])
	}
	descriptor = canonicalAVMABIIntrospectionDescriptor(descriptor)
	descriptor.InterfaceHash = ComputeAVMABIIntrospectionHash(descriptor)
	return descriptor, descriptor.Validate()
}

func NewAVMABIMethodCall(call AVMABIMethodCall) (AVMABIMethodCall, error) {
	call = canonicalAVMABIMethodCall(call)
	call.CallHash = ComputeAVMABIMethodCallHash(call)
	return call, call.Validate()
}

func NewAVMABIIdentityBinding(binding AVMABIIdentityBinding) (AVMABIIdentityBinding, error) {
	binding = canonicalAVMABIIdentityBinding(binding)
	binding.BindingHash = ComputeAVMABIIdentityBindingHash(binding)
	return binding, binding.Validate()
}

func ApplyAVMPromiseResolution(promise AVMPromiseState, resolution AVMPromiseResolution) (AVMPromiseState, error) {
	promise = canonicalAVMPromiseState(promise)
	resolution = canonicalAVMPromiseResolution(resolution)
	if err := promise.Validate(); err != nil {
		return AVMPromiseState{}, err
	}
	if err := resolution.Validate(); err != nil {
		return AVMPromiseState{}, err
	}
	if promise.PromiseID != resolution.PromiseID {
		return AVMPromiseState{}, errors.New("AVM 2.0 promise resolution id mismatch")
	}
	if promise.MessageID != resolution.OriginalMessageID {
		return AVMPromiseState{}, errors.New("AVM 2.0 promise resolution message mismatch")
	}
	if resolution.DeliveryHeight <= promise.CreatedHeight {
		return AVMPromiseState{}, errors.New("AVM 2.0 promise resolution must be delivered in future height")
	}
	updated := promise
	updated.Status = resolution.Status
	updated.ReceiptHash = resolution.ReceiptHash
	updated.ReturnHash = resolution.ReturnHash
	updated.PromiseHash = ComputeAVMPromiseHash(updated)
	return updated, updated.Validate()
}

func ScheduleAVMPromiseTimeout(promise AVMPromiseState, timeoutMessage AVMAsyncMessage) (AVMPromiseTimeoutTask, error) {
	promise = canonicalAVMPromiseState(promise)
	timeoutMessage = canonicalAVMAsyncMessage(timeoutMessage)
	if err := promise.Validate(); err != nil {
		return AVMPromiseTimeoutTask{}, err
	}
	if promise.Status != AVMPromisePending {
		return AVMPromiseTimeoutTask{}, errors.New("AVM 2.0 only pending promises can schedule timeout")
	}
	task := AVMPromiseTimeoutTask{
		PromiseID:		promise.PromiseID,
		Contract:		promise.Contract,
		DueHeight:		promise.ExpiryHeight,
		HandlerPayloadHash:	ComputeAVMBytesHash([]byte(promise.PromiseID + ":timeout")),
		TimeoutMessage:		timeoutMessage,
	}
	return NewAVMPromiseTimeoutTask(task)
}

func BindAVMABIToCode(code AVMCodeRecord, descriptor AVMABIIntrospectionDescriptor) error {
	code = canonicalAVMCodeRecord(code)
	descriptor = canonicalAVMABIIntrospectionDescriptor(descriptor)
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

func ValidateAVMABIMethodCall(descriptor AVMABIIntrospectionDescriptor, call AVMABIMethodCall) error {
	descriptor = canonicalAVMABIIntrospectionDescriptor(descriptor)
	call = canonicalAVMABIMethodCall(call)
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
	return validateAVMCallFunds(descriptor.RequiredFundsForSelector(call.MethodSelector), call.Funds)
}

func (p AVMAsyncCallPlan) Validate() error {
	p = canonicalAVMAsyncCallPlan(p)
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
	if p.Promise.Status != AVMPromisePending {
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
	if p.OutboxRoot != ComputeAVMMessageRoot([]AVMAsyncMessage{p.Message}) {
		return errors.New("AVM 2.0 async outbox root mismatch")
	}
	if p.PromiseRoot != ComputeAVMPromiseRoot([]AVMPromiseState{p.Promise}) {
		return errors.New("AVM 2.0 async promise root mismatch")
	}
	if err := zonestypes.ValidateHash("AVM 2.0 async call plan hash", p.PlanHash); err != nil {
		return err
	}
	if p.PlanHash != ComputeAVMAsyncCallPlanHash(p) {
		return errors.New("AVM 2.0 async call plan hash mismatch")
	}
	return nil
}

func (r AVMPromiseResolution) Validate() error {
	r = canonicalAVMPromiseResolution(r)
	if err := zonestypes.ValidateHash("AVM 2.0 promise resolution promise id", r.PromiseID); err != nil {
		return err
	}
	if err := zonestypes.ValidateHash("AVM 2.0 promise resolution original message id", r.OriginalMessageID); err != nil {
		return err
	}
	if err := r.ResolutionMessage.Validate(); err != nil {
		return err
	}
	if !IsAVMPromiseStatus(r.Status) || r.Status == AVMPromisePending {
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
	if r.ResolutionMessage.PayloadType != AVMPayloadTypePromiseResolution {
		return errors.New("AVM 2.0 promise resolution must use resolution payload type")
	}
	if err := zonestypes.ValidateHash("AVM 2.0 promise resolution hash", r.ResolutionHash); err != nil {
		return err
	}
	if r.ResolutionHash != ComputeAVMPromiseResolutionHash(r) {
		return errors.New("AVM 2.0 promise resolution hash mismatch")
	}
	return nil
}

func (t AVMPromiseTimeoutTask) Validate() error {
	t = canonicalAVMPromiseTimeoutTask(t)
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
	if t.TimeoutMessage.PayloadType != AVMPayloadTypePromiseTimeout {
		return errors.New("AVM 2.0 timeout message must use timeout payload type")
	}
	if t.TimeoutMessage.CreatedHeight < t.DueHeight {
		return errors.New("AVM 2.0 timeout message cannot be created before due height")
	}
	if err := zonestypes.ValidateHash("AVM 2.0 timeout task hash", t.TaskHash); err != nil {
		return err
	}
	if t.TaskHash != ComputeAVMPromiseTimeoutTaskHash(t) {
		return errors.New("AVM 2.0 timeout task hash mismatch")
	}
	return nil
}

func (d AVM2MethodDescriptor) Validate() error {
	d = canonicalAVM2MethodDescriptor(d)
	if err := validateEngineToken("AVM 2.0 method selector", d.Selector, MaxAVMTokenLength); err != nil {
		return err
	}
	if err := validateEngineToken("AVM 2.0 method name", d.Name, MaxAVMTokenLength); err != nil {
		return err
	}
	if err := zonestypes.ValidateHash("AVM 2.0 method argument schema hash", d.ArgumentSchemaHash); err != nil {
		return err
	}
	if err := zonestypes.ValidateHash("AVM 2.0 method return schema hash", d.ReturnSchemaHash); err != nil {
		return err
	}
	if err := validateEngineToken("AVM 2.0 method argument encoding", d.ArgumentEncoding, MaxAVMABIEncodingLength); err != nil {
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
	if err := validateEngineToken("AVM 2.0 ABI event name", d.Name, MaxAVMTokenLength); err != nil {
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

func (d AVMErrorDescriptor) Validate() error {
	d = canonicalAVMErrorDescriptor(d)
	if err := validateEngineToken("AVM 2.0 ABI error code", d.Code, MaxAVMTokenLength); err != nil {
		return err
	}
	if err := zonestypes.ValidateHash("AVM 2.0 ABI error schema hash", d.SchemaHash); err != nil {
		return err
	}
	if err := zonestypes.ValidateHash("AVM 2.0 ABI error hash", d.ErrorHash); err != nil {
		return err
	}
	if d.ErrorHash != ComputeAVMErrorDescriptorHash(d) {
		return errors.New("AVM 2.0 ABI error hash mismatch")
	}
	return nil
}

func (r AVMFundRequirement) Validate() error {
	r = canonicalAVMFundRequirement(r)
	if err := validateEngineToken("AVM 2.0 fund selector", r.Selector, MaxAVMTokenLength); err != nil {
		return err
	}
	if err := validateEngineToken("AVM 2.0 fund denom", r.Denom, MaxAVMTokenLength); err != nil {
		return err
	}
	if r.Minimum.IsNil() || r.Minimum.IsNegative() {
		return errors.New("AVM 2.0 fund minimum must be non-negative")
	}
	if err := zonestypes.ValidateHash("AVM 2.0 fund requirement hash", r.RequirementHash); err != nil {
		return err
	}
	if r.RequirementHash != ComputeAVMFundRequirementHash(r) {
		return errors.New("AVM 2.0 fund requirement hash mismatch")
	}
	return nil
}

func (h AVMGasHint) Validate() error {
	h = canonicalAVMGasHint(h)
	if err := validateEngineToken("AVM 2.0 gas hint selector", h.Selector, MaxAVMTokenLength); err != nil {
		return err
	}
	if h.Estimate == 0 {
		return errors.New("AVM 2.0 gas hint estimate must be positive")
	}
	if err := zonestypes.ValidateHash("AVM 2.0 gas hint hash", h.HintHash); err != nil {
		return err
	}
	if h.HintHash != ComputeAVMGasHintHash(h) {
		return errors.New("AVM 2.0 gas hint hash mismatch")
	}
	return nil
}

func (d AVMABIIntrospectionDescriptor) Validate() error {
	d = canonicalAVMABIIntrospectionDescriptor(d)
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
	if err := validateAVMErrors(d.Errors); err != nil {
		return err
	}
	if err := validateAVMFundRequirements(d.RequiredFunds, d.Methods); err != nil {
		return err
	}
	if err := validateAVMGasHints(d.GasHints, d.Methods); err != nil {
		return err
	}
	if d.IdentityNameOptional != "" {
		if err := validateEngineToken("AVM 2.0 ABI identity name", d.IdentityNameOptional, MaxAVMTokenLength); err != nil {
			return err
		}
	}
	if err := zonestypes.ValidateHash("AVM 2.0 ABI introspection interface hash", d.InterfaceHash); err != nil {
		return err
	}
	if d.InterfaceHash != ComputeAVMABIIntrospectionHash(d) {
		return errors.New("AVM 2.0 ABI introspection hash mismatch")
	}
	return nil
}

func (c AVMABIMethodCall) Validate() error {
	c = canonicalAVMABIMethodCall(c)
	if err := zonestypes.ValidateHash("AVM 2.0 call interface hash", c.InterfaceHash); err != nil {
		return err
	}
	if err := validateEngineToken("AVM 2.0 call method selector", c.MethodSelector, MaxAVMTokenLength); err != nil {
		return err
	}
	if err := validateEngineToken("AVM 2.0 call argument encoding", c.ArgumentEncoding, MaxAVMABIEncodingLength); err != nil {
		return err
	}
	if err := zonestypes.ValidateHash("AVM 2.0 call argument hash", c.ArgumentHash); err != nil {
		return err
	}
	if err := validateAVMCallFundRecords(c.Funds); err != nil {
		return err
	}
	if err := zonestypes.ValidateHash("AVM 2.0 method call hash", c.CallHash); err != nil {
		return err
	}
	if c.CallHash != ComputeAVMABIMethodCallHash(c) {
		return errors.New("AVM 2.0 method call hash mismatch")
	}
	return nil
}

func (b AVMABIIdentityBinding) Validate() error {
	b = canonicalAVMABIIdentityBinding(b)
	if err := validateEngineToken("AVM 2.0 ABI identity name", b.Name, MaxAVMTokenLength); err != nil {
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
	if b.BindingHash != ComputeAVMABIIdentityBindingHash(b) {
		return errors.New("AVM 2.0 ABI identity binding hash mismatch")
	}
	return nil
}

func (d AVMABIIntrospectionDescriptor) FindMethod(selector string) (AVM2MethodDescriptor, bool) {
	d = canonicalAVMABIIntrospectionDescriptor(d)
	selector = strings.TrimSpace(selector)
	for _, method := range d.Methods {
		if method.Selector == selector {
			return method, true
		}
	}
	return AVM2MethodDescriptor{}, false
}

func (d AVMABIIntrospectionDescriptor) RequiredFundsForSelector(selector string) []AVMFundRequirement {
	d = canonicalAVMABIIntrospectionDescriptor(d)
	selector = strings.TrimSpace(selector)
	var out []AVMFundRequirement
	for _, requirement := range d.RequiredFunds {
		if requirement.Selector == selector {
			out = append(out, requirement)
		}
	}
	return out
}

func ComputeAVMAsyncCallPlanHash(plan AVMAsyncCallPlan) string {
	plan = canonicalAVMAsyncCallPlan(plan)
	h := sha256.New()
	writeEnginePart(h, "aetra-AVM-async-call-plan-v1")
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

func ComputeAVMPromiseResolutionHash(resolution AVMPromiseResolution) string {
	resolution = canonicalAVMPromiseResolution(resolution)
	h := sha256.New()
	writeEnginePart(h, "aetra-AVM-promise-resolution-v1")
	writeEnginePart(h, resolution.PromiseID)
	writeEnginePart(h, resolution.OriginalMessageID)
	writeEnginePart(h, resolution.ResolutionMessage.ID)
	writeEnginePart(h, string(resolution.Status))
	writeEnginePart(h, resolution.ReceiptHash)
	writeEnginePart(h, resolution.ReturnHash)
	writeEngineUint64(h, resolution.DeliveryHeight)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMPromiseTimeoutTaskHash(task AVMPromiseTimeoutTask) string {
	task = canonicalAVMPromiseTimeoutTask(task)
	h := sha256.New()
	writeEnginePart(h, "aetra-AVM-promise-timeout-task-v1")
	writeEnginePart(h, task.PromiseID)
	writeEnginePart(h, task.Contract)
	writeEngineUint64(h, task.DueHeight)
	writeEnginePart(h, task.HandlerPayloadHash)
	writeEnginePart(h, task.TimeoutMessage.ID)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVM2MethodDescriptorHash(method AVM2MethodDescriptor) string {
	method = canonicalAVM2MethodDescriptor(method)
	h := sha256.New()
	writeEnginePart(h, "aetra-AVM-method-descriptor-v1")
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
	h := sha256.New()
	writeEnginePart(h, "aetra-AVM-abi-event-descriptor-v1")
	writeEnginePart(h, event.Name)
	writeEnginePart(h, event.SchemaHash)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMErrorDescriptorHash(errDesc AVMErrorDescriptor) string {
	errDesc = canonicalAVMErrorDescriptor(errDesc)
	h := sha256.New()
	writeEnginePart(h, "aetra-AVM-error-descriptor-v1")
	writeEnginePart(h, errDesc.Code)
	writeEnginePart(h, errDesc.SchemaHash)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMFundRequirementHash(requirement AVMFundRequirement) string {
	requirement = canonicalAVMFundRequirement(requirement)
	h := sha256.New()
	writeEnginePart(h, "aetra-AVM-fund-requirement-v1")
	writeEnginePart(h, requirement.Selector)
	writeEnginePart(h, requirement.Denom)
	writeEnginePart(h, requirement.Minimum.String())
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMGasHintHash(hint AVMGasHint) string {
	hint = canonicalAVMGasHint(hint)
	h := sha256.New()
	writeEnginePart(h, "aetra-AVM-gas-hint-v1")
	writeEnginePart(h, hint.Selector)
	writeEngineUint64(h, hint.Estimate)
	return hex.EncodeToString(h.Sum(nil))
}

func ComputeAVMABIIntrospectionHash(descriptor AVMABIIntrospectionDescriptor) string {
	descriptor = canonicalAVMABIIntrospectionDescriptor(descriptor)
	h := sha256.New()
	writeEnginePart(h, "aetra-AVM-abi-introspection-v1")
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

func ComputeAVMABIMethodCallHash(call AVMABIMethodCall) string {
	call = canonicalAVMABIMethodCall(call)
	call.CallHash = ""
	h := sha256.New()
	writeEnginePart(h, "aetra-AVM-method-call-v1")
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

func ComputeAVMABIIdentityBindingHash(binding AVMABIIdentityBinding) string {
	binding = canonicalAVMABIIdentityBinding(binding)
	binding.BindingHash = ""
	h := sha256.New()
	writeEnginePart(h, "aetra-AVM-abi-identity-binding-v1")
	writeEnginePart(h, binding.Name)
	writeEnginePart(h, binding.InterfaceHash)
	writeEnginePart(h, binding.ResolverRecordHash)
	return hex.EncodeToString(h.Sum(nil))
}

func canonicalAVMAsyncCallPlan(plan AVMAsyncCallPlan) AVMAsyncCallPlan {
	plan.CallerContract = strings.TrimSpace(plan.CallerContract)
	plan.Message = canonicalAVMAsyncMessage(plan.Message)
	plan.Promise = canonicalAVMPromiseState(plan.Promise)
	plan.PersistedPromiseID = strings.TrimSpace(plan.PersistedPromiseID)
	plan.OutboxRoot = strings.TrimSpace(plan.OutboxRoot)
	plan.PromiseRoot = strings.TrimSpace(plan.PromiseRoot)
	plan.PlanHash = strings.TrimSpace(plan.PlanHash)
	return plan
}

func canonicalAVMPromiseResolution(resolution AVMPromiseResolution) AVMPromiseResolution {
	resolution.PromiseID = strings.TrimSpace(resolution.PromiseID)
	resolution.OriginalMessageID = strings.TrimSpace(resolution.OriginalMessageID)
	resolution.ResolutionMessage = canonicalAVMAsyncMessage(resolution.ResolutionMessage)
	resolution.ReceiptHash = strings.TrimSpace(resolution.ReceiptHash)
	resolution.ReturnHash = strings.TrimSpace(resolution.ReturnHash)
	resolution.ResolutionHash = strings.TrimSpace(resolution.ResolutionHash)
	return resolution
}

func canonicalAVMPromiseTimeoutTask(task AVMPromiseTimeoutTask) AVMPromiseTimeoutTask {
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

func canonicalAVMErrorDescriptor(errDesc AVMErrorDescriptor) AVMErrorDescriptor {
	errDesc.Code = strings.TrimSpace(errDesc.Code)
	errDesc.SchemaHash = strings.TrimSpace(errDesc.SchemaHash)
	errDesc.ErrorHash = strings.TrimSpace(errDesc.ErrorHash)
	return errDesc
}

func canonicalAVMFundRequirement(requirement AVMFundRequirement) AVMFundRequirement {
	requirement.Selector = strings.TrimSpace(requirement.Selector)
	requirement.Denom = strings.TrimSpace(requirement.Denom)
	requirement.RequirementHash = strings.TrimSpace(requirement.RequirementHash)
	return requirement
}

func canonicalAVMGasHint(hint AVMGasHint) AVMGasHint {
	hint.Selector = strings.TrimSpace(hint.Selector)
	hint.HintHash = strings.TrimSpace(hint.HintHash)
	return hint
}

func canonicalAVMABIIntrospectionDescriptor(descriptor AVMABIIntrospectionDescriptor) AVMABIIntrospectionDescriptor {
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
	descriptor.Errors = append([]AVMErrorDescriptor(nil), descriptor.Errors...)
	for i := range descriptor.Errors {
		descriptor.Errors[i] = canonicalAVMErrorDescriptor(descriptor.Errors[i])
	}
	sort.SliceStable(descriptor.Errors, func(i, j int) bool { return descriptor.Errors[i].Code < descriptor.Errors[j].Code })
	descriptor.RequiredFunds = append([]AVMFundRequirement(nil), descriptor.RequiredFunds...)
	for i := range descriptor.RequiredFunds {
		descriptor.RequiredFunds[i] = canonicalAVMFundRequirement(descriptor.RequiredFunds[i])
	}
	sort.SliceStable(descriptor.RequiredFunds, func(i, j int) bool {
		if descriptor.RequiredFunds[i].Selector == descriptor.RequiredFunds[j].Selector {
			return descriptor.RequiredFunds[i].Denom < descriptor.RequiredFunds[j].Denom
		}
		return descriptor.RequiredFunds[i].Selector < descriptor.RequiredFunds[j].Selector
	})
	descriptor.GasHints = append([]AVMGasHint(nil), descriptor.GasHints...)
	for i := range descriptor.GasHints {
		descriptor.GasHints[i] = canonicalAVMGasHint(descriptor.GasHints[i])
	}
	sort.SliceStable(descriptor.GasHints, func(i, j int) bool { return descriptor.GasHints[i].Selector < descriptor.GasHints[j].Selector })
	return descriptor
}

func canonicalAVMCallFund(fund AVMCallFund) AVMCallFund {
	fund.Denom = strings.TrimSpace(fund.Denom)
	return fund
}

func canonicalAVMABIMethodCall(call AVMABIMethodCall) AVMABIMethodCall {
	call.InterfaceHash = strings.TrimSpace(call.InterfaceHash)
	call.MethodSelector = strings.TrimSpace(call.MethodSelector)
	call.ArgumentEncoding = strings.TrimSpace(call.ArgumentEncoding)
	call.ArgumentHash = strings.TrimSpace(call.ArgumentHash)
	call.CallHash = strings.TrimSpace(call.CallHash)
	call.Funds = append([]AVMCallFund(nil), call.Funds...)
	for i := range call.Funds {
		call.Funds[i] = canonicalAVMCallFund(call.Funds[i])
	}
	sort.SliceStable(call.Funds, func(i, j int) bool { return call.Funds[i].Denom < call.Funds[j].Denom })
	return call
}

func canonicalAVMABIIdentityBinding(binding AVMABIIdentityBinding) AVMABIIdentityBinding {
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

func validateAVMErrors(errorsList []AVMErrorDescriptor) error {
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

func validateAVMFundRequirements(requirements []AVMFundRequirement, methods []AVM2MethodDescriptor) error {
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

func validateAVMGasHints(hints []AVMGasHint, methods []AVM2MethodDescriptor) error {
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

func validateAVMCallFundRecords(funds []AVMCallFund) error {
	seen := make(map[string]struct{}, len(funds))
	for i, fund := range funds {
		fund = canonicalAVMCallFund(fund)
		if err := validateEngineToken("AVM 2.0 call fund denom", fund.Denom, MaxAVMTokenLength); err != nil {
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

func validateAVMCallFunds(requirements []AVMFundRequirement, funds []AVMCallFund) error {
	available := make(map[string]sdkmath.Int, len(funds))
	for _, fund := range funds {
		fund = canonicalAVMCallFund(fund)
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
