package types

import (
	"errors"
	"fmt"

	postypes "github.com/sovereign-l1/l1/x/pos/types"
)

type HookEvent string

const (
	HookEventEpochBegin		HookEvent	= "epoch_begin"
	HookEventPhaseTransition	HookEvent	= "epoch_phase_transition"
	HookEventEpochEnd		HookEvent	= "epoch_end"
)

type HookRecord struct {
	Event			HookEvent
	EpochID			uint64
	Height			uint64
	UnixSeconds		uint64
	FromPhase		postypes.EpochPhase
	ToPhase			postypes.EpochPhase
	ValidatorSetHash	string
}

type EpochState struct {
	Params			postypes.Params
	Current			postypes.EpochRecord
	CurrentStartUnixSeconds	uint64
	EpochHeightSpan		uint64
	History			[]postypes.EpochRecord
	HookLog			[]HookRecord
}

func (s EpochState) Validate() error {
	if err := s.Params.Validate(); err != nil {
		return err
	}
	if s.EpochHeightSpan == 0 {
		return errors.New("epoch height span must be positive")
	}
	if s.Current.Seed != "" {
		if err := s.Current.Validate(); err != nil {
			return err
		}
		if s.CurrentStartUnixSeconds == 0 {
			return errors.New("current epoch start time must be positive")
		}
	}
	seen := make(map[uint64]struct{}, len(s.History))
	var previousID uint64
	var previousEnd uint64
	for i, record := range s.History {
		if err := record.Validate(); err != nil {
			return err
		}
		if record.Phase != postypes.EpochPhaseClosed {
			return errors.New("historical epochs must be closed")
		}
		if _, found := seen[record.EpochID]; found {
			return fmt.Errorf("duplicate historical epoch %d", record.EpochID)
		}
		seen[record.EpochID] = struct{}{}
		if i > 0 {
			if record.EpochID <= previousID {
				return errors.New("historical epochs must be sorted by epoch id")
			}
			if record.StartHeight <= previousEnd {
				return errors.New("historical epochs must be sorted by height")
			}
		}
		previousID = record.EpochID
		previousEnd = record.EndHeight
	}
	for _, hook := range s.HookLog {
		if err := hook.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (h HookRecord) Validate() error {
	switch h.Event {
	case HookEventEpochBegin, HookEventPhaseTransition, HookEventEpochEnd:
	default:
		return fmt.Errorf("unsupported epoch hook event %q", h.Event)
	}
	if h.Height == 0 {
		return errors.New("epoch hook height must be positive")
	}
	if h.UnixSeconds == 0 {
		return errors.New("epoch hook time must be positive")
	}
	if err := validateHookPhase(h.FromPhase); err != nil {
		return err
	}
	if err := validateHookPhase(h.ToPhase); err != nil {
		return err
	}
	switch h.Event {
	case HookEventEpochBegin:
		if h.ToPhase != postypes.EpochPhaseDelegation {
			return errors.New("epoch begin hook must target delegation phase")
		}
	case HookEventPhaseTransition:
		if err := postypes.ValidateEpochPhaseTransition(h.FromPhase, h.ToPhase); err != nil {
			return err
		}
	case HookEventEpochEnd:
		if h.ToPhase != postypes.EpochPhaseClosed {
			return errors.New("epoch end hook must target closed phase")
		}
	}
	return nil
}

func CloneState(state EpochState) EpochState {
	out := state
	out.History = make([]postypes.EpochRecord, len(state.History))
	copy(out.History, state.History)
	out.HookLog = make([]HookRecord, len(state.HookLog))
	copy(out.HookLog, state.HookLog)
	return out
}

func validateHookPhase(phase postypes.EpochPhase) error {
	if phase == "" {
		return nil
	}
	for _, allowed := range postypes.EpochPhaseValues() {
		if phase == allowed {
			return nil
		}
	}
	return fmt.Errorf("unsupported epoch hook phase %q", phase)
}
