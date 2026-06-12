package types

import (
	"errors"
	"fmt"
	"sort"
)

type ZoneRegistryState struct {
	Zones		[]Zone
	ActiveZones	[]ZoneID
	Commitments	[]ZoneCommitment
}

func EmptyState() ZoneRegistryState {
	return ZoneRegistryState{
		Zones:		[]Zone{},
		ActiveZones:	[]ZoneID{},
		Commitments:	[]ZoneCommitment{},
	}
}

func RegisterZone(state ZoneRegistryState, zone Zone) (ZoneRegistryState, error) {
	if err := state.Validate(); err != nil {
		return ZoneRegistryState{}, err
	}
	if err := zone.Validate(); err != nil {
		return ZoneRegistryState{}, err
	}
	if _, found := state.ZoneByID(zone.ID); found {
		return ZoneRegistryState{}, fmt.Errorf("zone %s already registered", zone.ID)
	}
	next := state.Clone()
	next.Zones = append(next.Zones, zone)
	sortZones(next.Zones)
	return next, next.Validate()
}

func ActivateZone(state ZoneRegistryState, id ZoneID, currentHeight uint64) (ZoneRegistryState, error) {
	if err := state.Validate(); err != nil {
		return ZoneRegistryState{}, err
	}
	zone, found := state.ZoneByID(id)
	if !found {
		return ZoneRegistryState{}, fmt.Errorf("zone %s is not registered", id)
	}
	if currentHeight < zone.ActivationHeight {
		return ZoneRegistryState{}, fmt.Errorf("zone %s cannot activate before height %d", id, zone.ActivationHeight)
	}
	if state.IsActive(id) {
		return state.Clone(), nil
	}
	next := state.Clone()
	next.ActiveZones = append(next.ActiveZones, id)
	sortZoneIDs(next.ActiveZones)
	return next, next.Validate()
}

func AppendCommitment(state ZoneRegistryState, commitment ZoneCommitment) (ZoneRegistryState, error) {
	if err := state.Validate(); err != nil {
		return ZoneRegistryState{}, err
	}
	if err := commitment.ValidateHash(); err != nil {
		return ZoneRegistryState{}, err
	}
	if _, found := state.ZoneByID(commitment.ZoneID); !found {
		return ZoneRegistryState{}, fmt.Errorf("zone %s is not registered", commitment.ZoneID)
	}
	previous, hasPrevious := state.LastCommitment(commitment.ZoneID)
	if hasPrevious {
		if commitment.ZoneHeight <= previous.ZoneHeight {
			return ZoneRegistryState{}, errors.New("zone commitment height must increase")
		}
		if commitment.PreviousCommitment != previous.CommitmentHash {
			return ZoneRegistryState{}, errors.New("zone commitment missing previous commitment")
		}
	} else if commitment.PreviousCommitment != "" {
		return ZoneRegistryState{}, errors.New("zone commitment references missing previous commitment")
	}
	for _, existing := range state.Commitments {
		if existing.ZoneID == commitment.ZoneID && existing.ZoneHeight == commitment.ZoneHeight {
			return ZoneRegistryState{}, errors.New("duplicate zone commitment height")
		}
	}
	next := state.Clone()
	next.Commitments = append(next.Commitments, commitment)
	sortCommitments(next.Commitments)
	return next, next.Validate()
}

func ImportState(state ZoneRegistryState) (ZoneRegistryState, error) {
	if err := state.Validate(); err != nil {
		return ZoneRegistryState{}, err
	}
	return state.Clone(), nil
}

func (s ZoneRegistryState) Export() ZoneRegistryState {
	out := s.Clone()
	sortZones(out.Zones)
	sortZoneIDs(out.ActiveZones)
	sortCommitments(out.Commitments)
	return out
}

func (s ZoneRegistryState) Clone() ZoneRegistryState {
	out := ZoneRegistryState{
		Zones:		make([]Zone, len(s.Zones)),
		ActiveZones:	make([]ZoneID, len(s.ActiveZones)),
		Commitments:	make([]ZoneCommitment, len(s.Commitments)),
	}
	copy(out.Zones, s.Zones)
	copy(out.ActiveZones, s.ActiveZones)
	copy(out.Commitments, s.Commitments)
	return out
}

func (s ZoneRegistryState) Validate() error {
	if err := validateZones(s.Zones); err != nil {
		return err
	}
	registered := make(map[ZoneID]Zone, len(s.Zones))
	for _, zone := range s.Zones {
		registered[zone.ID] = zone
	}
	if err := validateActiveZones(s.ActiveZones, registered); err != nil {
		return err
	}
	return validateCommitments(s.Commitments, registered)
}

func (s ZoneRegistryState) ZoneByID(id ZoneID) (Zone, bool) {
	for _, zone := range s.Zones {
		if zone.ID == id {
			return zone, true
		}
	}
	return Zone{}, false
}

func (s ZoneRegistryState) IsActive(id ZoneID) bool {
	for _, active := range s.ActiveZones {
		if active == id {
			return true
		}
	}
	return false
}

func (s ZoneRegistryState) LastCommitment(id ZoneID) (ZoneCommitment, bool) {
	for i := len(s.Commitments) - 1; i >= 0; i-- {
		if s.Commitments[i].ZoneID == id {
			return s.Commitments[i], true
		}
	}
	return ZoneCommitment{}, false
}

func validateZones(zones []Zone) error {
	var previous ZoneID
	seen := make(map[ZoneID]struct{}, len(zones))
	for i, zone := range zones {
		if err := zone.Validate(); err != nil {
			return err
		}
		if _, found := seen[zone.ID]; found {
			return fmt.Errorf("duplicate zone id %s", zone.ID)
		}
		seen[zone.ID] = struct{}{}
		if i > 0 && string(previous) >= string(zone.ID) {
			return errors.New("zones must be sorted canonically by id")
		}
		previous = zone.ID
	}
	return nil
}

func validateActiveZones(active []ZoneID, registered map[ZoneID]Zone) error {
	var previous ZoneID
	seen := make(map[ZoneID]struct{}, len(active))
	for i, id := range active {
		if err := ValidateZoneID(id); err != nil {
			return err
		}
		if _, found := registered[id]; !found {
			return fmt.Errorf("active zone %s is not registered", id)
		}
		if _, found := seen[id]; found {
			return fmt.Errorf("duplicate active zone %s", id)
		}
		seen[id] = struct{}{}
		if i > 0 && string(previous) >= string(id) {
			return errors.New("active zones must be sorted canonically by id")
		}
		previous = id
	}
	return nil
}

func validateCommitments(commitments []ZoneCommitment, registered map[ZoneID]Zone) error {
	latestByZone := make(map[ZoneID]ZoneCommitment)
	seenHeights := make(map[string]struct{}, len(commitments))
	for i, commitment := range commitments {
		if err := commitment.ValidateHash(); err != nil {
			return err
		}
		if _, found := registered[commitment.ZoneID]; !found {
			return fmt.Errorf("zone %s is not registered", commitment.ZoneID)
		}
		if i > 0 && compareCommitments(commitments[i-1], commitment) >= 0 {
			return errors.New("zone commitments must be sorted canonically by zone id and height")
		}
		heightKey := fmt.Sprintf("%s/%020d", commitment.ZoneID, commitment.ZoneHeight)
		if _, found := seenHeights[heightKey]; found {
			return errors.New("duplicate zone commitment height")
		}
		seenHeights[heightKey] = struct{}{}
		previous, found := latestByZone[commitment.ZoneID]
		if found {
			if commitment.ZoneHeight <= previous.ZoneHeight {
				return errors.New("zone commitment height must increase")
			}
			if commitment.PreviousCommitment != previous.CommitmentHash {
				return errors.New("zone commitment missing previous commitment")
			}
		} else if commitment.PreviousCommitment != "" {
			return errors.New("zone commitment references missing previous commitment")
		}
		latestByZone[commitment.ZoneID] = commitment
	}
	return nil
}

func sortZones(zones []Zone) {
	sort.SliceStable(zones, func(i, j int) bool {
		return zones[i].ID < zones[j].ID
	})
}

func sortZoneIDs(ids []ZoneID) {
	sort.SliceStable(ids, func(i, j int) bool {
		return ids[i] < ids[j]
	})
}

func sortCommitments(commitments []ZoneCommitment) {
	sort.SliceStable(commitments, func(i, j int) bool {
		return compareCommitments(commitments[i], commitments[j]) < 0
	})
}

func compareCommitments(left, right ZoneCommitment) int {
	if left.ZoneID < right.ZoneID {
		return -1
	}
	if left.ZoneID > right.ZoneID {
		return 1
	}
	if left.ZoneHeight < right.ZoneHeight {
		return -1
	}
	if left.ZoneHeight > right.ZoneHeight {
		return 1
	}
	return 0
}
