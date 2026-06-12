package types

import (
	"errors"
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sovereign-l1/l1/app/addressing"
)

const (
	ResourceCompute		= "compute"
	ResourceStorage		= "storage"
	ResourcePriority	= "priority"

	BasisPoints	= uint64(10_000)
)

type Params struct {
	MaxPremiumNaet		uint64
	FairnessPriorityCap	uint64
	MinNormalSlots		uint32
	MaxAccountShareBps	uint64
}

type Order struct {
	ID		string
	Account		sdk.AccAddress
	Resource	string
	Quantity	uint64
	PremiumNaet	uint64
	BaseFeePaid	bool
	NormalUser	bool
	Sequence	uint64
}

type Selection struct {
	Accepted	[]Order
	Rejected	[]RejectedOrder
}

type RejectedOrder struct {
	Order	Order
	Reason	string
}

func DefaultParams() Params {
	return Params{
		MaxPremiumNaet:		1_000,
		FairnessPriorityCap:	100,
		MinNormalSlots:		1,
		MaxAccountShareBps:	5_000,
	}
}

func (p Params) Validate() error {
	if p.FairnessPriorityCap == 0 {
		return errors.New("market fairness priority cap must be positive")
	}
	if p.MaxAccountShareBps == 0 || p.MaxAccountShareBps > BasisPoints {
		return errors.New("market max account share must be within 1..10000 bps")
	}
	return nil
}

func Select(params Params, orders []Order, capacity uint32) (Selection, error) {
	if err := params.Validate(); err != nil {
		return Selection{}, err
	}
	if capacity == 0 {
		return Selection{}, errors.New("market capacity must be positive")
	}
	valid := make([]Order, 0, len(orders))
	selection := Selection{}
	for _, order := range orders {
		if err := order.Validate(params); err != nil {
			selection.Rejected = append(selection.Rejected, RejectedOrder{Order: order.Clone(), Reason: err.Error()})
			continue
		}
		valid = append(valid, order.Clone())
	}
	sortOrders(valid, params)

	accountLimit := maxUint32(1, uint32((uint64(capacity)*params.MaxAccountShareBps+BasisPoints-1)/BasisPoints))
	acceptedByAccount := make(map[string]uint32)
	used := make(map[string]bool)

	normalReserve := minUint32(params.MinNormalSlots, capacity)
	if normalReserve > 0 {
		for i, order := range valid {
			if uint32(len(selection.Accepted)) >= normalReserve {
				break
			}
			if !order.NormalUser {
				continue
			}
			if acceptedByAccount[string(order.Account)] >= accountLimit {
				continue
			}
			selection.Accepted = append(selection.Accepted, order.Clone())
			acceptedByAccount[string(order.Account)]++
			used[order.ID] = true
			valid[i].ID = ""
		}
	}

	for _, order := range valid {
		if uint32(len(selection.Accepted)) >= capacity {
			break
		}
		if order.ID == "" || used[order.ID] {
			continue
		}
		if acceptedByAccount[string(order.Account)] >= accountLimit {
			selection.Rejected = append(selection.Rejected, RejectedOrder{Order: order.Clone(), Reason: "market account share cap"})
			continue
		}
		selection.Accepted = append(selection.Accepted, order.Clone())
		acceptedByAccount[string(order.Account)]++
	}
	sort.SliceStable(selection.Accepted, func(i, j int) bool {
		return selection.Accepted[i].Sequence < selection.Accepted[j].Sequence
	})
	return selection, nil
}

func (o Order) Validate(params Params) error {
	if o.ID == "" {
		return errors.New("market order id is required")
	}
	if len(o.Account) == 0 {
		return errors.New("market order account is required")
	}
	if err := addressing.RejectZeroAddress("market order account", o.Account); err != nil {
		return err
	}
	if !IsResource(o.Resource) {
		return fmt.Errorf("invalid market resource %q", o.Resource)
	}
	if o.Quantity == 0 {
		return errors.New("market order quantity must be positive")
	}
	if !o.BaseFeePaid {
		return errors.New("market order cannot replace base naet fee")
	}
	if o.PremiumNaet > params.MaxPremiumNaet {
		return fmt.Errorf("market premium exceeds cap %d", params.MaxPremiumNaet)
	}
	return nil
}

func (o Order) Clone() Order {
	out := o
	out.Account = append(sdk.AccAddress(nil), o.Account...)
	return out
}

func PriorityScore(params Params, order Order) uint64 {
	score := order.PremiumNaet
	if score > params.FairnessPriorityCap {
		return params.FairnessPriorityCap
	}
	return score
}

func CanReplaceBaseFee() bool {
	return false
}

func IsResource(resource string) bool {
	switch resource {
	case ResourceCompute, ResourceStorage, ResourcePriority:
		return true
	default:
		return false
	}
}

func sortOrders(orders []Order, params Params) {
	sort.SliceStable(orders, func(i, j int) bool {
		left := PriorityScore(params, orders[i])
		right := PriorityScore(params, orders[j])
		if left != right {
			return left > right
		}
		if orders[i].NormalUser != orders[j].NormalUser {
			return orders[i].NormalUser
		}
		if orders[i].Sequence != orders[j].Sequence {
			return orders[i].Sequence < orders[j].Sequence
		}
		return orders[i].ID < orders[j].ID
	})
}

func minUint32(a, b uint32) uint32 {
	if a < b {
		return a
	}
	return b
}

func maxUint32(a, b uint32) uint32 {
	if a > b {
		return a
	}
	return b
}
