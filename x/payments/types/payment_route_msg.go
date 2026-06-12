package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	sdkmath "cosmossdk.io/math"

	"github.com/sovereign-l1/l1/app/addressing"
)

type PaymentRouteHop struct {
	ChannelID	string
	From		string
	To		string
	FeeAmount	string
	TimeoutHeight	uint64
}

type PaymentRouteCommitment struct {
	RouteID		string
	Committer	string
	CommitmentHash	string
	Signed		bool
	Reserved	bool
	ExpiresHeight	uint64
}

type PaymentRouteBalance struct {
	Participant	string
	Available	string
}

type PaymentRouteAdmission struct {
	CurrentHeight			uint64
	Commitments			[]PaymentRouteCommitment
	Balances			[]PaymentRouteBalance
	SupportedSettlementModes	[]ConditionSettlementMode
}

type MsgPaymentRoute struct {
	RouteID		string
	Payer		string
	Payee		string
	Amount		string
	MaxFee		string
	Hops		[]PaymentRouteHop
	ConditionRoot	string
	ExpiryHeight	uint64
	SettlementMode	ConditionSettlementMode
}

type PaymentRouteCongestionSnapshot struct {
	RouteID			string
	ChannelID		string
	HopIndex		uint32
	CongestionBps		uint32
	PendingMessageCount	uint32
	RetryCount		uint32
	ObservedHeight		uint64
}

type PaymentRouteReceiptStatus string

const (
	PaymentRouteReceiptDelivered	PaymentRouteReceiptStatus	= "DELIVERED"
	PaymentRouteReceiptRetry	PaymentRouteReceiptStatus	= "RETRY"
	PaymentRouteReceiptExpired	PaymentRouteReceiptStatus	= "EXPIRED"
	PaymentRouteReceiptBounced	PaymentRouteReceiptStatus	= "BOUNCED"
)

type PaymentRouteReceipt struct {
	RouteID		string
	Status		PaymentRouteReceiptStatus
	Amount		string
	FeeAmount	string
	ValueReturned	string
	Attempt		uint32
	RecordedHeight	uint64
	ExpiryHeight	uint64
	ReceiptHash	string
}

type PaymentRouteDeliveryTask struct {
	TaskID			string
	RouteID			string
	HopIndex		uint32
	ChannelID		string
	From			string
	To			string
	Amount			string
	FeeAmount		string
	Attempt			uint32
	DeliverAfterHeight	uint64
	ExpiryHeight		uint64
	TaskHash		string
}

type PaymentRouteDeliveryPlan struct {
	Tasks	[]PaymentRouteDeliveryTask
	Receipt	PaymentRouteReceipt
}

type PaymentRouteTableState struct {
	Epoch		uint64
	Routes		[]MsgPaymentRoute
	Tasks		[]PaymentRouteDeliveryTask
	Receipts	[]PaymentRouteReceipt
	RootHash	string
}

type PaymentRoutingEpochUpdate struct {
	Epoch		uint64
	CurrentHeight	uint64
	Routes		[]MsgPaymentRoute
	Tasks		[]PaymentRouteDeliveryTask
	Receipts	[]PaymentRouteReceipt
}

func (h PaymentRouteHop) Normalize() PaymentRouteHop {
	h.ChannelID = normalizeHash(h.ChannelID)
	h.From = strings.TrimSpace(h.From)
	h.To = strings.TrimSpace(h.To)
	h.FeeAmount = strings.TrimSpace(h.FeeAmount)
	return h
}

func (h PaymentRouteHop) Validate() error {
	hop := h.Normalize()
	if err := ValidateHash("payments route hop channel id", hop.ChannelID); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments route hop from", hop.From); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments route hop to", hop.To); err != nil {
		return err
	}
	if hop.From == hop.To {
		return errors.New("payments route hop endpoints must differ")
	}
	if err := validateNonNegativeInt("payments route hop fee", hop.FeeAmount); err != nil {
		return err
	}
	if hop.TimeoutHeight == 0 {
		return errors.New("payments route hop timeout must be positive")
	}
	return nil
}

func (m MsgPaymentRoute) Normalize() MsgPaymentRoute {
	m.RouteID = normalizeHash(m.RouteID)
	m.Payer = strings.TrimSpace(m.Payer)
	m.Payee = strings.TrimSpace(m.Payee)
	m.Amount = strings.TrimSpace(m.Amount)
	m.MaxFee = strings.TrimSpace(m.MaxFee)
	m.ConditionRoot = normalizeHash(m.ConditionRoot)
	for i := range m.Hops {
		m.Hops[i] = m.Hops[i].Normalize()
	}
	m.SettlementMode = ConditionSettlementMode(strings.TrimSpace(string(m.SettlementMode)))
	return m
}

func (m MsgPaymentRoute) ValidateBasic() error {
	route := m.Normalize()
	if err := ValidateHash("payments route id", route.RouteID); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments route payer", route.Payer); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments route payee", route.Payee); err != nil {
		return err
	}
	if route.Payer == route.Payee {
		return errors.New("payments route endpoints must differ")
	}
	if _, err := parsePositiveInt("payments route amount", route.Amount); err != nil {
		return err
	}
	if err := validateNonNegativeInt("payments route max fee", route.MaxFee); err != nil {
		return err
	}
	if len(route.Hops) == 0 {
		return errors.New("payments route requires hops")
	}
	if len(route.Hops) > MaxRoutingHops {
		return errors.New("payments route exceeds max hops")
	}
	if err := ValidateHash("payments route condition root", route.ConditionRoot); err != nil {
		return err
	}
	if route.ExpiryHeight == 0 {
		return errors.New("payments route expiry height must be positive")
	}
	if !IsConditionSettlementMode(route.SettlementMode) {
		return errors.New("payments route settlement mode is unsupported")
	}
	totalFee := sdkmath.ZeroInt()
	var previousTimeout uint64
	for i, hop := range route.Hops {
		if err := hop.Validate(); err != nil {
			return err
		}
		if i == 0 && hop.From != route.Payer {
			return errors.New("payments route first hop must start at payer")
		}
		if i == len(route.Hops)-1 && hop.To != route.Payee {
			return errors.New("payments route final hop must end at payee")
		}
		if i > 0 {
			if route.Hops[i-1].To != hop.From {
				return errors.New("payments route hops must be connected")
			}
			if hop.TimeoutHeight <= previousTimeout {
				return errors.New("payments route hop timeouts must be strictly increasing")
			}
		}
		if hop.TimeoutHeight > route.ExpiryHeight {
			return errors.New("payments route hop timeout exceeds route expiry")
		}
		fee, err := parseNonNegativeInt("payments route hop fee", hop.FeeAmount)
		if err != nil {
			return err
		}
		totalFee = totalFee.Add(fee)
		previousTimeout = hop.TimeoutHeight
	}
	maxFee, err := parseNonNegativeInt("payments route max fee", route.MaxFee)
	if err != nil {
		return err
	}
	if totalFee.GT(maxFee) {
		return errors.New("payments route hop fees exceed max fee")
	}
	return nil
}

func (m MsgPaymentRoute) Validate(admission PaymentRouteAdmission) error {
	route := m.Normalize()
	if err := route.ValidateBasic(); err != nil {
		return err
	}
	if admission.CurrentHeight == 0 {
		return errors.New("payments route admission height must be positive")
	}
	if admission.CurrentHeight > route.ExpiryHeight {
		return errors.New("payments route is expired")
	}
	if !admission.SupportsSettlementMode(route.SettlementMode) {
		return errors.New("payments route settlement mode is unsupported")
	}
	commitment, found := admission.CommitmentForRoute(route.RouteID)
	if !found {
		return errors.New("payments route commitment is required")
	}
	if err := commitment.ValidateForRoute(route, admission.CurrentHeight); err != nil {
		return err
	}
	available, found, err := admission.AvailableFor(route.Payer)
	if err != nil {
		return err
	}
	if !found {
		return errors.New("payments route payer balance is required")
	}
	amount, err := parsePositiveInt("payments route amount", route.Amount)
	if err != nil {
		return err
	}
	maxFee, err := parseNonNegativeInt("payments route max fee", route.MaxFee)
	if err != nil {
		return err
	}
	if available.LT(amount.Add(maxFee)) {
		return errors.New("payments route amount plus max fee is unavailable")
	}
	return nil
}

func BuildMsgPaymentRoute(route MsgPaymentRoute, admission PaymentRouteAdmission) (MsgPaymentRoute, error) {
	route = route.Normalize()
	if err := route.Validate(admission); err != nil {
		return MsgPaymentRoute{}, err
	}
	return route, nil
}

func (c PaymentRouteCommitment) Normalize() PaymentRouteCommitment {
	c.RouteID = normalizeHash(c.RouteID)
	c.Committer = strings.TrimSpace(c.Committer)
	c.CommitmentHash = normalizeHash(c.CommitmentHash)
	return c
}

func (c PaymentRouteCommitment) ValidateForRoute(route MsgPaymentRoute, currentHeight uint64) error {
	c = c.Normalize()
	route = route.Normalize()
	if c.RouteID != route.RouteID {
		return errors.New("payments route commitment route mismatch")
	}
	if err := addressing.ValidateUserAddress("payments route commitment committer", c.Committer); err != nil {
		return err
	}
	if !c.Signed && !c.Reserved {
		return errors.New("payments route commitment must be signed or reserved")
	}
	if c.ExpiresHeight == 0 || c.ExpiresHeight < currentHeight {
		return errors.New("payments route commitment is expired")
	}
	expected := ComputePaymentRouteCommitmentHash(route)
	if c.CommitmentHash != expected {
		return errors.New("payments route commitment hash mismatch")
	}
	return nil
}

func (a PaymentRouteAdmission) CommitmentForRoute(routeID string) (PaymentRouteCommitment, bool) {
	routeID = normalizeHash(routeID)
	for _, commitment := range a.Commitments {
		commitment = commitment.Normalize()
		if commitment.RouteID == routeID {
			return commitment, true
		}
	}
	return PaymentRouteCommitment{}, false
}

func (a PaymentRouteAdmission) AvailableFor(participant string) (sdkmath.Int, bool, error) {
	participant = strings.TrimSpace(participant)
	for _, balance := range a.Balances {
		normalized := balance.Normalize()
		if normalized.Participant != participant {
			continue
		}
		available, err := parseNonNegativeInt("payments route available balance", normalized.Available)
		return available, true, err
	}
	return sdkmath.Int{}, false, nil
}

func (a PaymentRouteAdmission) SupportsSettlementMode(mode ConditionSettlementMode) bool {
	supported := a.SupportedSettlementModes
	if len(supported) == 0 {
		supported = []ConditionSettlementMode{ConditionSettlementModePreimage, ConditionSettlementModeExpiry}
	}
	for _, candidate := range supported {
		if candidate == mode {
			return true
		}
	}
	return false
}

func (b PaymentRouteBalance) Normalize() PaymentRouteBalance {
	b.Participant = strings.TrimSpace(b.Participant)
	b.Available = strings.TrimSpace(b.Available)
	return b
}

func IsConditionSettlementMode(mode ConditionSettlementMode) bool {
	switch mode {
	case ConditionSettlementModePreimage, ConditionSettlementModeExpiry:
		return true
	default:
		return false
	}
}

func ComputePaymentRouteCommitmentHash(route MsgPaymentRoute) string {
	route = route.Normalize()
	parts := []string{
		"payment-route-commitment",
		route.RouteID,
		route.Payer,
		route.Payee,
		route.Amount,
		route.MaxFee,
		route.ConditionRoot,
		fmt.Sprintf("%020d", route.ExpiryHeight),
		string(route.SettlementMode),
	}
	for _, hop := range route.Hops {
		parts = append(parts,
			hop.ChannelID,
			hop.From,
			hop.To,
			hop.FeeAmount,
			fmt.Sprintf("%020d", hop.TimeoutHeight),
		)
	}
	return HashParts(parts...)
}

func DeterministicPaymentRouteScore(route MsgPaymentRoute, congestion []PaymentRouteCongestionSnapshot) (int64, error) {
	route = route.Normalize()
	if err := route.ValidateBasic(); err != nil {
		return 0, err
	}
	var score int64
	for i, hop := range route.Hops {
		fee, err := parseNonNegativeInt("payments route score hop fee", hop.FeeAmount)
		if err != nil {
			return 0, err
		}
		if !fee.IsInt64() {
			return 0, errors.New("payments route score fee is too large")
		}
		score += fee.Int64()
		score += int64(i+1) * 100
	}
	for _, snapshot := range normalizeRouteCongestionSnapshots(congestion) {
		if snapshot.RouteID != "" && snapshot.RouteID != route.RouteID {
			continue
		}
		if snapshot.HopIndex >= uint32(len(route.Hops)) {
			return 0, errors.New("payments route congestion hop index out of range")
		}
		hop := route.Hops[snapshot.HopIndex]
		if snapshot.ChannelID != "" && snapshot.ChannelID != hop.ChannelID {
			continue
		}
		score += int64(snapshot.CongestionBps) * 10
		score += int64(snapshot.PendingMessageCount) * 5
		score += int64(snapshot.RetryCount) * 25
	}
	return score, nil
}

func SchedulePaymentRouteDelivery(route MsgPaymentRoute, currentHeight uint64, previousTasks []PaymentRouteDeliveryTask, policy RouteRetryPolicy) (PaymentRouteDeliveryPlan, error) {
	route = route.Normalize()
	if err := route.ValidateBasic(); err != nil {
		return PaymentRouteDeliveryPlan{}, err
	}
	if currentHeight == 0 {
		return PaymentRouteDeliveryPlan{}, errors.New("payments route scheduler height must be positive")
	}
	if currentHeight > route.ExpiryHeight {
		receipt, err := BuildPaymentRouteBounceReceipt(route, currentHeight, PaymentRouteReceiptExpired)
		if err != nil {
			return PaymentRouteDeliveryPlan{}, err
		}
		return PaymentRouteDeliveryPlan{Receipt: receipt}, nil
	}
	policy = policy.Normalize()
	if err := policy.Validate(); err != nil {
		return PaymentRouteDeliveryPlan{}, err
	}
	attempt := nextRouteAttempt(previousTasks)
	if attempt > policy.MaxAttempts {
		receipt, err := BuildPaymentRouteBounceReceipt(route, currentHeight, PaymentRouteReceiptBounced)
		if err != nil {
			return PaymentRouteDeliveryPlan{}, err
		}
		receipt.Attempt = attempt - 1
		receipt.ReceiptHash = ComputePaymentRouteReceiptHash(receipt)
		return PaymentRouteDeliveryPlan{Receipt: receipt}, nil
	}
	deliverAfter := currentHeight
	if attempt > 1 {
		deliverAfter += policy.CongestionRetryDelay
	}
	tasks := make([]PaymentRouteDeliveryTask, 0, len(route.Hops))
	for i, hop := range route.Hops {
		task := PaymentRouteDeliveryTask{
			RouteID:		route.RouteID,
			HopIndex:		uint32(i),
			ChannelID:		hop.ChannelID,
			From:			hop.From,
			To:			hop.To,
			Amount:			route.Amount,
			FeeAmount:		hop.FeeAmount,
			Attempt:		attempt,
			DeliverAfterHeight:	deliverAfter,
			ExpiryHeight:		route.ExpiryHeight,
		}
		task.TaskHash = ComputePaymentRouteDeliveryTaskHash(task)
		task.TaskID = HashParts("payment-route-task-id", task.TaskHash)
		tasks = append(tasks, task.Normalize())
	}
	receipt := PaymentRouteReceipt{
		RouteID:	route.RouteID,
		Status:		PaymentRouteReceiptRetry,
		Amount:		route.Amount,
		FeeAmount:	totalRouteFee(route).String(),
		ValueReturned:	"0",
		Attempt:	attempt,
		RecordedHeight:	currentHeight,
		ExpiryHeight:	route.ExpiryHeight,
	}
	receipt.ReceiptHash = ComputePaymentRouteReceiptHash(receipt)
	return PaymentRouteDeliveryPlan{Tasks: tasks, Receipt: receipt.Normalize()}, nil
}

func BuildPaymentRouteBounceReceipt(route MsgPaymentRoute, currentHeight uint64, status PaymentRouteReceiptStatus) (PaymentRouteReceipt, error) {
	route = route.Normalize()
	if status != PaymentRouteReceiptExpired && status != PaymentRouteReceiptBounced {
		return PaymentRouteReceipt{}, errors.New("payments route bounce receipt requires expired or bounced status")
	}
	if err := route.ValidateBasic(); err != nil {
		return PaymentRouteReceipt{}, err
	}
	if currentHeight == 0 {
		return PaymentRouteReceipt{}, errors.New("payments route bounce height must be positive")
	}
	amount, err := parsePositiveInt("payments route bounce amount", route.Amount)
	if err != nil {
		return PaymentRouteReceipt{}, err
	}
	maxFee, err := parseNonNegativeInt("payments route bounce max fee", route.MaxFee)
	if err != nil {
		return PaymentRouteReceipt{}, err
	}
	receipt := PaymentRouteReceipt{
		RouteID:	route.RouteID,
		Status:		status,
		Amount:		route.Amount,
		FeeAmount:	route.MaxFee,
		ValueReturned:	amount.Add(maxFee).String(),
		RecordedHeight:	currentHeight,
		ExpiryHeight:	route.ExpiryHeight,
	}
	receipt.ReceiptHash = ComputePaymentRouteReceiptHash(receipt)
	return receipt.Normalize(), nil
}

func ValidatePaymentRouteBounceConservation(route MsgPaymentRoute, receipt PaymentRouteReceipt) error {
	route = route.Normalize()
	receipt = receipt.Normalize()
	if err := route.ValidateBasic(); err != nil {
		return err
	}
	if err := receipt.Validate(); err != nil {
		return err
	}
	if receipt.RouteID != route.RouteID {
		return errors.New("payments route bounce receipt route mismatch")
	}
	if receipt.Status != PaymentRouteReceiptExpired && receipt.Status != PaymentRouteReceiptBounced {
		return errors.New("payments route bounce receipt status mismatch")
	}
	amount, err := parsePositiveInt("payments route amount", route.Amount)
	if err != nil {
		return err
	}
	maxFee, err := parseNonNegativeInt("payments route max fee", route.MaxFee)
	if err != nil {
		return err
	}
	returned, err := parseNonNegativeInt("payments route bounce returned value", receipt.ValueReturned)
	if err != nil {
		return err
	}
	if !returned.Equal(amount.Add(maxFee)) {
		return errors.New("payments route bounce value conservation failed")
	}
	return nil
}

func ApplyPaymentRoutingEpochUpdate(previous PaymentRouteTableState, update PaymentRoutingEpochUpdate) (PaymentRouteTableState, error) {
	previous = previous.Normalize()
	if previous.Epoch > 0 && previous.RootHash != ComputePaymentRouteTableRoot(previous) {
		return PaymentRouteTableState{}, errors.New("payments route table previous root mismatch")
	}
	if update.Epoch <= previous.Epoch {
		return PaymentRouteTableState{}, errors.New("payments route epoch must increase")
	}
	if update.CurrentHeight == 0 {
		return PaymentRouteTableState{}, errors.New("payments route epoch height must be positive")
	}
	next := PaymentRouteTableState{
		Epoch:		update.Epoch,
		Routes:		append([]MsgPaymentRoute{}, update.Routes...),
		Tasks:		append([]PaymentRouteDeliveryTask{}, update.Tasks...),
		Receipts:	append([]PaymentRouteReceipt{}, update.Receipts...),
	}.Normalize()
	for _, route := range next.Routes {
		if err := route.ValidateBasic(); err != nil {
			return PaymentRouteTableState{}, err
		}
		if update.CurrentHeight > route.ExpiryHeight {
			return PaymentRouteTableState{}, errors.New("payments route epoch contains expired route")
		}
	}
	for _, task := range next.Tasks {
		if err := task.Validate(); err != nil {
			return PaymentRouteTableState{}, err
		}
	}
	for _, receipt := range next.Receipts {
		if err := receipt.Validate(); err != nil {
			return PaymentRouteTableState{}, err
		}
	}
	next.RootHash = ComputePaymentRouteTableRoot(next)
	return next, nil
}

func RecordPaymentRouteReceipt(state PaymentRouteTableState, receipt PaymentRouteReceipt) (PaymentRouteTableState, error) {
	state = state.Normalize()
	receipt = receipt.Normalize()
	if err := receipt.Validate(); err != nil {
		return PaymentRouteTableState{}, err
	}
	next := state
	replaced := false
	for i, existing := range next.Receipts {
		if existing.RouteID == receipt.RouteID && existing.Attempt == receipt.Attempt {
			next.Receipts[i] = receipt
			replaced = true
			break
		}
	}
	if !replaced {
		next.Receipts = append(next.Receipts, receipt)
	}
	next = next.Normalize()
	next.RootHash = ComputePaymentRouteTableRoot(next)
	return next, nil
}

func QueryPaymentRouteReceipt(state PaymentRouteTableState, routeID string, attempt uint32) (PaymentRouteReceipt, bool) {
	routeID = normalizeHash(routeID)
	state = state.Normalize()
	for _, receipt := range state.Receipts {
		if receipt.RouteID == routeID && (attempt == 0 || receipt.Attempt == attempt) {
			return receipt, true
		}
	}
	return PaymentRouteReceipt{}, false
}

func (t PaymentRouteDeliveryTask) Normalize() PaymentRouteDeliveryTask {
	t.TaskID = normalizeOptionalHash(t.TaskID)
	t.RouteID = normalizeHash(t.RouteID)
	t.ChannelID = normalizeHash(t.ChannelID)
	t.From = strings.TrimSpace(t.From)
	t.To = strings.TrimSpace(t.To)
	t.Amount = strings.TrimSpace(t.Amount)
	t.FeeAmount = strings.TrimSpace(t.FeeAmount)
	t.TaskHash = normalizeOptionalHash(t.TaskHash)
	return t
}

func (t PaymentRouteDeliveryTask) Validate() error {
	task := t.Normalize()
	if err := ValidateHash("payments route delivery task id", task.TaskID); err != nil {
		return err
	}
	if err := ValidateHash("payments route delivery route id", task.RouteID); err != nil {
		return err
	}
	if err := ValidateHash("payments route delivery channel id", task.ChannelID); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments route delivery from", task.From); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("payments route delivery to", task.To); err != nil {
		return err
	}
	if _, err := parsePositiveInt("payments route delivery amount", task.Amount); err != nil {
		return err
	}
	if err := validateNonNegativeInt("payments route delivery fee", task.FeeAmount); err != nil {
		return err
	}
	if task.Attempt == 0 {
		return errors.New("payments route delivery attempt must be positive")
	}
	if task.DeliverAfterHeight == 0 {
		return errors.New("payments route delivery height must be positive")
	}
	if task.ExpiryHeight < task.DeliverAfterHeight {
		return errors.New("payments route delivery expiry precedes delivery")
	}
	if task.TaskHash != ComputePaymentRouteDeliveryTaskHash(task) {
		return errors.New("payments route delivery task hash mismatch")
	}
	return nil
}

func ComputePaymentRouteDeliveryTaskHash(task PaymentRouteDeliveryTask) string {
	task = task.Normalize()
	return HashParts(
		"payment-route-delivery-task",
		task.RouteID,
		fmt.Sprintf("%020d", task.HopIndex),
		task.ChannelID,
		task.From,
		task.To,
		task.Amount,
		task.FeeAmount,
		fmt.Sprintf("%020d", uint64(task.Attempt)),
		fmt.Sprintf("%020d", task.DeliverAfterHeight),
		fmt.Sprintf("%020d", task.ExpiryHeight),
	)
}

func (r PaymentRouteReceipt) Normalize() PaymentRouteReceipt {
	r.RouteID = normalizeHash(r.RouteID)
	r.Amount = strings.TrimSpace(r.Amount)
	r.FeeAmount = strings.TrimSpace(r.FeeAmount)
	r.ValueReturned = strings.TrimSpace(r.ValueReturned)
	r.ReceiptHash = normalizeOptionalHash(r.ReceiptHash)
	return r
}

func (r PaymentRouteReceipt) Validate() error {
	receipt := r.Normalize()
	if err := ValidateHash("payments route receipt route id", receipt.RouteID); err != nil {
		return err
	}
	if !IsPaymentRouteReceiptStatus(receipt.Status) {
		return errors.New("payments route receipt status is unsupported")
	}
	if _, err := parsePositiveInt("payments route receipt amount", receipt.Amount); err != nil {
		return err
	}
	if err := validateNonNegativeInt("payments route receipt fee", receipt.FeeAmount); err != nil {
		return err
	}
	if err := validateNonNegativeInt("payments route receipt returned value", receipt.ValueReturned); err != nil {
		return err
	}
	if receipt.RecordedHeight == 0 {
		return errors.New("payments route receipt height must be positive")
	}
	if receipt.ExpiryHeight == 0 {
		return errors.New("payments route receipt expiry must be positive")
	}
	if receipt.ReceiptHash != ComputePaymentRouteReceiptHash(receipt) {
		return errors.New("payments route receipt hash mismatch")
	}
	return nil
}

func IsPaymentRouteReceiptStatus(status PaymentRouteReceiptStatus) bool {
	switch status {
	case PaymentRouteReceiptDelivered, PaymentRouteReceiptRetry, PaymentRouteReceiptExpired, PaymentRouteReceiptBounced:
		return true
	default:
		return false
	}
}

func ComputePaymentRouteReceiptHash(receipt PaymentRouteReceipt) string {
	receipt = receipt.Normalize()
	return HashParts(
		"payment-route-receipt",
		receipt.RouteID,
		string(receipt.Status),
		receipt.Amount,
		receipt.FeeAmount,
		receipt.ValueReturned,
		fmt.Sprintf("%020d", uint64(receipt.Attempt)),
		fmt.Sprintf("%020d", receipt.RecordedHeight),
		fmt.Sprintf("%020d", receipt.ExpiryHeight),
	)
}

func (s PaymentRouteTableState) Normalize() PaymentRouteTableState {
	for i := range s.Routes {
		s.Routes[i] = s.Routes[i].Normalize()
	}
	for i := range s.Tasks {
		s.Tasks[i] = s.Tasks[i].Normalize()
	}
	for i := range s.Receipts {
		s.Receipts[i] = s.Receipts[i].Normalize()
	}
	sort.SliceStable(s.Routes, func(i, j int) bool {
		return s.Routes[i].RouteID < s.Routes[j].RouteID
	})
	sort.SliceStable(s.Tasks, func(i, j int) bool {
		return s.Tasks[i].TaskID < s.Tasks[j].TaskID
	})
	sort.SliceStable(s.Receipts, func(i, j int) bool {
		if s.Receipts[i].RouteID == s.Receipts[j].RouteID {
			return s.Receipts[i].Attempt < s.Receipts[j].Attempt
		}
		return s.Receipts[i].RouteID < s.Receipts[j].RouteID
	})
	s.RootHash = normalizeOptionalHash(s.RootHash)
	return s
}

func ComputePaymentRouteTableRoot(state PaymentRouteTableState) string {
	state = state.Normalize()
	parts := []string{"payment-route-table", fmt.Sprintf("%020d", state.Epoch)}
	for _, route := range state.Routes {
		parts = append(parts, ComputePaymentRouteCommitmentHash(route))
	}
	for _, task := range state.Tasks {
		parts = append(parts, task.TaskHash)
	}
	for _, receipt := range state.Receipts {
		parts = append(parts, receipt.ReceiptHash)
	}
	return HashParts(parts...)
}

func normalizeRouteCongestionSnapshots(snapshots []PaymentRouteCongestionSnapshot) []PaymentRouteCongestionSnapshot {
	out := append([]PaymentRouteCongestionSnapshot{}, snapshots...)
	for i := range out {
		out[i].RouteID = normalizeOptionalHash(out[i].RouteID)
		out[i].ChannelID = normalizeOptionalHash(out[i].ChannelID)
	}
	sort.SliceStable(out, func(i, j int) bool {
		left := fmt.Sprintf("%s/%020d/%s/%020d", out[i].RouteID, out[i].HopIndex, out[i].ChannelID, out[i].ObservedHeight)
		right := fmt.Sprintf("%s/%020d/%s/%020d", out[j].RouteID, out[j].HopIndex, out[j].ChannelID, out[j].ObservedHeight)
		return left < right
	})
	return out
}

func nextRouteAttempt(tasks []PaymentRouteDeliveryTask) uint32 {
	var maxAttempt uint32
	for _, task := range tasks {
		if task.Attempt > maxAttempt {
			maxAttempt = task.Attempt
		}
	}
	return maxAttempt + 1
}

func totalRouteFee(route MsgPaymentRoute) sdkmath.Int {
	total := sdkmath.ZeroInt()
	for _, hop := range route.Normalize().Hops {
		fee, err := parseNonNegativeInt("payments route hop fee", hop.FeeAmount)
		if err == nil {
			total = total.Add(fee)
		}
	}
	return total
}
