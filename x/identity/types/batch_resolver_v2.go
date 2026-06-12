package types

import (
	"errors"
	"fmt"
)

const (
	MinIdentityBatchResolverUpdateGasV2	= uint64(1_000)
	MaxIdentityBatchResolverUpdateGasV2	= uint64(1_000_000)

	IdentityBatchFailureAtomicV2	IdentityBatchFailureModeV2	= "atomic"
	IdentityBatchFailurePartialV2	IdentityBatchFailureModeV2	= "partial"

	IdentityBatchUpdateSuccessV2		IdentityBatchUpdateStatusV2	= "success"
	IdentityBatchUpdateUnauthorizedV2	IdentityBatchUpdateStatusV2	= "unauthorized"
	IdentityBatchUpdateVersionErrorV2	IdentityBatchUpdateStatusV2	= "version_error"
	IdentityBatchUpdateGasErrorV2		IdentityBatchUpdateStatusV2	= "gas_error"
	IdentityBatchUpdateInvalidV2		IdentityBatchUpdateStatusV2	= "invalid"
)

type IdentityBatchFailureModeV2 string

type IdentityBatchUpdateStatusV2 string

type IdentityBatchResolverUpdateOptionsV2 struct {
	Mode		IdentityBatchFailureModeV2
	Height		uint64
	GasPerUpdate	uint64
	GasLimit	uint64
}

type IdentityBatchResolverUpdateResultV2 struct {
	Index		uint32
	Name		string
	NameHash	string
	Status		IdentityBatchUpdateStatusV2
	Error		string
	GasWanted	uint64
	GasUsed		uint64
	PreviousVersion	uint64
	NewVersion	uint64
}

type IdentityBatchResolverUpdateResponseV2 struct {
	Mode		IdentityBatchFailureModeV2
	Atomic		bool
	Results		[]IdentityBatchResolverUpdateResultV2
	Successes	uint32
	Failures	uint32
	GasWanted	uint64
	GasUsed		uint64
	ResultOrder	string
}

func ExecuteBatchResolverUpdatesV2(state IdentityState, msg MsgBatchUpdateResolversV2, options IdentityBatchResolverUpdateOptionsV2) (IdentityState, IdentityBatchResolverUpdateResponseV2, error) {
	if err := msg.ValidateBasic(); err != nil {
		return IdentityState{}, IdentityBatchResolverUpdateResponseV2{}, err
	}
	if err := validateBatchResolverOptionsV2(options, len(msg.Updates)); err != nil {
		return IdentityState{}, IdentityBatchResolverUpdateResponseV2{}, err
	}
	mode := options.Mode
	if mode == "" {
		mode = IdentityBatchFailureAtomicV2
	}
	next := state.Export()
	response := IdentityBatchResolverUpdateResponseV2{
		Mode:		mode,
		Atomic:		mode == IdentityBatchFailureAtomicV2,
		Results:	make([]IdentityBatchResolverUpdateResultV2, 0, len(msg.Updates)),
		GasWanted:	options.GasLimit,
		ResultOrder:	"input_index",
	}
	for i, update := range msg.Updates {
		result := IdentityBatchResolverUpdateResultV2{
			Index:		uint32(i),
			Name:		update.Name,
			NameHash:	update.NameHash,
			GasWanted:	options.GasPerUpdate,
		}
		if response.GasUsed+options.GasPerUpdate > options.GasLimit {
			result.Status = IdentityBatchUpdateGasErrorV2
			result.Error = "identity v2 resolver batch gas limit exceeded"
			response.Results = append(response.Results, result)
			response.Failures++
			if mode == IdentityBatchFailureAtomicV2 {
				return state.Export(), response, errors.New(result.Error)
			}
			continue
		}
		response.GasUsed += options.GasPerUpdate
		result.GasUsed = options.GasPerUpdate
		current, found := findResolver(next, update.Name)
		if found {
			result.PreviousVersion = ResolverRecordVersionV2(current)
		} else {
			result.PreviousVersion = update.ExpectedRecordVersion
		}
		if err := ValidateResolverRecordVersionForUpdateV2(result.PreviousVersion, update.ExpectedRecordVersion); err != nil {
			result.Status = IdentityBatchUpdateVersionErrorV2
			result.Error = err.Error()
			response.Results = append(response.Results, result)
			response.Failures++
			if mode == IdentityBatchFailureAtomicV2 {
				return state.Export(), response, err
			}
			continue
		}
		updated, record, err := PatchIdentityResolver(next, update.Name, msg.Auth.Signer, update.Patch, options.Height)
		if err != nil {
			result.Status = batchResolverErrorStatusV2(err)
			result.Error = err.Error()
			response.Results = append(response.Results, result)
			response.Failures++
			if mode == IdentityBatchFailureAtomicV2 {
				return state.Export(), response, err
			}
			continue
		}
		next = updated
		result.Status = IdentityBatchUpdateSuccessV2
		result.NewVersion = ResolverRecordVersionV2(record)
		response.Results = append(response.Results, result)
		response.Successes++
	}
	return next.Export(), response, nil
}

func ValidateBatchResolverUpdateResponseV2(response IdentityBatchResolverUpdateResponseV2) error {
	if err := validateBatchFailureModeV2(response.Mode); err != nil {
		return err
	}
	if response.ResultOrder != "input_index" {
		return errors.New("identity v2 resolver batch result order must be input_index")
	}
	if response.GasWanted == 0 {
		return errors.New("identity v2 resolver batch gas_wanted is required")
	}
	var successes uint32
	var failures uint32
	for i, result := range response.Results {
		if int(result.Index) != i {
			return errors.New("identity v2 resolver batch result index must match deterministic order")
		}
		if result.GasWanted == 0 && result.Status != IdentityBatchUpdateGasErrorV2 {
			return errors.New("identity v2 resolver batch result gas_wanted is required")
		}
		switch result.Status {
		case IdentityBatchUpdateSuccessV2:
			if result.Error != "" {
				return errors.New("identity v2 resolver batch success must not include error")
			}
			if result.NewVersion == 0 {
				return errors.New("identity v2 resolver batch success new_version is required")
			}
			successes++
		case IdentityBatchUpdateUnauthorizedV2, IdentityBatchUpdateVersionErrorV2, IdentityBatchUpdateGasErrorV2, IdentityBatchUpdateInvalidV2:
			if result.Error == "" {
				return errors.New("identity v2 resolver batch failure error is required")
			}
			failures++
		default:
			return fmt.Errorf("unsupported identity v2 resolver batch result status %q", result.Status)
		}
	}
	if successes != response.Successes || failures != response.Failures {
		return errors.New("identity v2 resolver batch success/failure counters mismatch")
	}
	return nil
}

func validateBatchResolverOptionsV2(options IdentityBatchResolverUpdateOptionsV2, updateCount int) error {
	if err := validateBatchFailureModeV2(options.Mode); err != nil {
		return err
	}
	if options.Height == 0 {
		return errors.New("identity v2 resolver batch height is required")
	}
	if options.GasPerUpdate < MinIdentityBatchResolverUpdateGasV2 || options.GasPerUpdate > MaxIdentityBatchResolverUpdateGasV2 {
		return fmt.Errorf("identity v2 resolver batch gas_per_update must be between %d and %d", MinIdentityBatchResolverUpdateGasV2, MaxIdentityBatchResolverUpdateGasV2)
	}
	if options.GasLimit < options.GasPerUpdate*uint64(updateCount) && options.Mode == IdentityBatchFailureAtomicV2 {
		return errors.New("identity v2 resolver batch atomic gas limit is insufficient")
	}
	if options.GasLimit == 0 {
		return errors.New("identity v2 resolver batch gas_limit is required")
	}
	return nil
}

func validateBatchFailureModeV2(mode IdentityBatchFailureModeV2) error {
	switch mode {
	case "", IdentityBatchFailureAtomicV2, IdentityBatchFailurePartialV2:
		return nil
	default:
		return fmt.Errorf("unsupported identity v2 resolver batch failure mode %q", mode)
	}
}

func batchResolverErrorStatusV2(err error) IdentityBatchUpdateStatusV2 {
	if err == nil {
		return IdentityBatchUpdateSuccessV2
	}
	message := err.Error()
	switch {
	case containsSubstringV2(message, "requires owner"), containsSubstringV2(message, "unauthorized"), containsSubstringV2(message, "delegate"):
		return IdentityBatchUpdateUnauthorizedV2
	case containsSubstringV2(message, "version"):
		return IdentityBatchUpdateVersionErrorV2
	default:
		return IdentityBatchUpdateInvalidV2
	}
}

func containsSubstringV2(value string, needle string) bool {
	if len(needle) == 0 {
		return true
	}
	if len(value) < len(needle) {
		return false
	}
	for i := 0; i <= len(value)-len(needle); i++ {
		if value[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
