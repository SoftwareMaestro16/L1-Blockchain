package types

import (
	"bytes"
	"errors"
	"sort"
)

type ProposalItem struct {
	ZoneID		ZoneID
	ShardID		ShardID
	TxHash		string
	PriorityClass	uint32
	AdmissionHeight	uint64
	TxIndex		uint32
	MessageIndex	uint32
}

type ProposalGroup struct {
	ZoneID	ZoneID
	ShardID	ShardID
	Items	[]ProposalItem
}

type ProposalSchedule struct {
	Height	uint64
	Groups	[]ProposalGroup
}

func BuildProposalSchedule(height uint64, items []ProposalItem, params AetraCoreParams) (ProposalSchedule, error) {
	if height == 0 {
		return ProposalSchedule{}, errors.New("aetracore proposal schedule height must be positive")
	}
	if err := params.Validate(); err != nil {
		return ProposalSchedule{}, err
	}
	if uint32(len(items)) > params.MaxProposalItemsPerBlock {
		return ProposalSchedule{}, errors.New("aetracore proposal item count exceeds block limit")
	}
	ordered := make([]ProposalItem, len(items))
	for i, item := range items {
		if err := item.Validate(); err != nil {
			return ProposalSchedule{}, err
		}
		ordered[i] = item
	}
	sortProposalItems(ordered)

	groups := make([]ProposalGroup, 0)
	for _, item := range ordered {
		if len(groups) == 0 || groups[len(groups)-1].ZoneID != item.ZoneID || groups[len(groups)-1].ShardID != item.ShardID {
			groups = append(groups, ProposalGroup{ZoneID: item.ZoneID, ShardID: item.ShardID})
		}
		groups[len(groups)-1].Items = append(groups[len(groups)-1].Items, item)
	}
	return ProposalSchedule{Height: height, Groups: groups}, nil
}

func ValidateProposalScheduleForState(schedule ProposalSchedule, state CoreState) error {
	if err := state.Validate(); err != nil {
		return err
	}
	if err := schedule.Validate(); err != nil {
		return err
	}
	for _, group := range schedule.Groups {
		descriptor, found := state.ZoneDescriptorByID(group.ZoneID)
		if !found {
			return errors.New("aetracore proposal group zone is not registered")
		}
		if !descriptor.Enabled {
			return errors.New("aetracore proposal group zone is disabled")
		}
		if group.ZoneID == ZoneIDAetraCore {
			continue
		}
		layout, found := state.LatestShardLayout(group.ZoneID, schedule.Height)
		if !found {
			return errors.New("aetracore proposal group missing shard layout")
		}
		if !layout.HasActiveShard(group.ShardID) {
			return errors.New("aetracore proposal group shard is not active in committed layout")
		}
	}
	return nil
}

func (i ProposalItem) Validate() error {
	if err := ValidateZoneID(ZoneID(i.ZoneID)); err != nil {
		return err
	}
	if err := ValidateShardID(i.ShardID); err != nil {
		return err
	}
	return ValidateHash("aetracore proposal tx hash", i.TxHash)
}

func (g ProposalGroup) Validate() error {
	if err := ValidateZoneID(ZoneID(g.ZoneID)); err != nil {
		return err
	}
	if err := ValidateShardID(g.ShardID); err != nil {
		return err
	}
	for i, item := range g.Items {
		if err := item.Validate(); err != nil {
			return err
		}
		if item.ZoneID != g.ZoneID || item.ShardID != g.ShardID {
			return errors.New("aetracore proposal group item route mismatch")
		}
		if i > 0 && compareProposalItems(g.Items[i-1], item) >= 0 {
			return errors.New("aetracore proposal group items must be sorted canonically")
		}
	}
	return nil
}

func (s ProposalSchedule) Validate() error {
	if s.Height == 0 {
		return errors.New("aetracore proposal schedule height must be positive")
	}
	for i, group := range s.Groups {
		if err := group.Validate(); err != nil {
			return err
		}
		if i > 0 && compareProposalGroupKey(s.Groups[i-1], group) >= 0 {
			return errors.New("aetracore proposal groups must be sorted canonically")
		}
	}
	return nil
}

func (l ShardLayout) HasActiveShard(shardID ShardID) bool {
	for _, shard := range l.ActiveShards {
		if shard.ShardID == shardID && shard.Available {
			return true
		}
	}
	return false
}

func sortProposalItems(items []ProposalItem) {
	sort.SliceStable(items, func(i, j int) bool {
		return compareProposalItems(items[i], items[j]) < 0
	})
}

func compareProposalGroupKey(left, right ProposalGroup) int {
	if left.ZoneID < right.ZoneID {
		return -1
	}
	if left.ZoneID > right.ZoneID {
		return 1
	}
	if left.ShardID < right.ShardID {
		return -1
	}
	if left.ShardID > right.ShardID {
		return 1
	}
	return 0
}

func compareProposalItems(left, right ProposalItem) int {
	if left.ZoneID < right.ZoneID {
		return -1
	}
	if left.ZoneID > right.ZoneID {
		return 1
	}
	if left.ShardID < right.ShardID {
		return -1
	}
	if left.ShardID > right.ShardID {
		return 1
	}
	if left.PriorityClass < right.PriorityClass {
		return -1
	}
	if left.PriorityClass > right.PriorityClass {
		return 1
	}
	if left.AdmissionHeight < right.AdmissionHeight {
		return -1
	}
	if left.AdmissionHeight > right.AdmissionHeight {
		return 1
	}
	if cmp := bytes.Compare([]byte(left.TxHash), []byte(right.TxHash)); cmp != 0 {
		return cmp
	}
	if left.TxIndex < right.TxIndex {
		return -1
	}
	if left.TxIndex > right.TxIndex {
		return 1
	}
	if left.MessageIndex < right.MessageIndex {
		return -1
	}
	if left.MessageIndex > right.MessageIndex {
		return 1
	}
	return 0
}
