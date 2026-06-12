package keeper

import (
	"errors"
	"fmt"

	postypes "github.com/sovereign-l1/l1/x/pos/types"
	taskgrouptypes "github.com/sovereign-l1/l1/x/taskgroups/types"
)

type Keeper struct {
	groups		map[string]postypes.TaskGroup
	receipts	map[string]taskgrouptypes.VerificationReceiptSet
}

func NewKeeper(groups []postypes.TaskGroup) (Keeper, error) {
	k := Keeper{
		groups:		make(map[string]postypes.TaskGroup, len(groups)),
		receipts:	make(map[string]taskgrouptypes.VerificationReceiptSet),
	}
	for _, group := range groups {
		if err := k.SetTaskGroup(group); err != nil {
			return Keeper{}, err
		}
	}
	return k, nil
}

func (k *Keeper) SetTaskGroup(group postypes.TaskGroup) error {
	if err := group.Validate(); err != nil {
		return err
	}
	if k.groups == nil {
		k.groups = make(map[string]postypes.TaskGroup)
	}
	if k.receipts == nil {
		k.receipts = make(map[string]taskgrouptypes.VerificationReceiptSet)
	}
	k.groups[group.TaskGroupID] = group
	return nil
}

func (k *Keeper) SubmitVerificationReceipt(receipt taskgrouptypes.VerificationReceipt) error {
	group, found := k.groups[receipt.TaskGroupID]
	if !found {
		return errors.New("task group not found for verification receipt")
	}
	if err := receipt.Validate(group); err != nil {
		return err
	}
	existing := k.receipts[receipt.TaskGroupID].Receipts
	next := make([]taskgrouptypes.VerificationReceipt, 0, len(existing)+1)
	next = append(next, existing...)
	next = append(next, receipt)
	set, err := taskgrouptypes.NewVerificationReceiptSet(group, next)
	if err != nil {
		return err
	}
	k.receipts[receipt.TaskGroupID] = set
	return nil
}

func (k Keeper) VerificationReceiptSet(taskGroupID string) (taskgrouptypes.VerificationReceiptSet, bool) {
	set, found := k.receipts[taskGroupID]
	return set, found
}

func (k Keeper) AggregateVerificationReceipts(taskGroupID string, verifiedObjectHash string, quorumBps uint32) (taskgrouptypes.VerificationAggregation, bool, error) {
	group, found := k.groups[taskGroupID]
	if !found {
		return taskgrouptypes.VerificationAggregation{}, false, nil
	}
	set, found := k.receipts[taskGroupID]
	if !found {
		return taskgrouptypes.VerificationAggregation{}, true, fmt.Errorf("verification receipts not found for task group %q", taskGroupID)
	}
	aggregation, err := taskgrouptypes.AggregateVerificationReceipts(group, set, verifiedObjectHash, quorumBps)
	if err != nil {
		return taskgrouptypes.VerificationAggregation{}, true, err
	}
	return aggregation, true, nil
}
