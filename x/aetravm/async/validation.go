package async

import (
	"errors"
	"fmt"

	appparams "github.com/sovereign-l1/l1/app/params"

	aetraaddress "github.com/sovereign-l1/l1/app/addressing"
)

func (c ContractAccount) Validate(params Params) error {
	if err := aetraaddress.RejectZeroAddress("contract account", c.Address); err != nil {
		return err
	}
	if len(c.CodeHash) != CodeHashLength {
		return fmt.Errorf("contract code hash must be %d bytes", CodeHashLength)
	}
	if len(c.State) > int(params.MaxStateSize) {
		return fmt.Errorf("contract state size must be <= %d", params.MaxStateSize)
	}
	if c.BalanceNaet.IsNil() || c.BalanceNaet.IsNegative() {
		return errors.New("contract naet balance must be non-negative")
	}
	return nil
}

func (m MessageEnvelope) Validate(params Params) error {
	if err := aetraaddress.RejectZeroAddress("message source", m.Source); err != nil {
		return err
	}
	if err := aetraaddress.RejectZeroAddress("message destination", m.Destination); err != nil {
		return err
	}
	if m.Value.Denom != appparams.BaseDenom {
		return fmt.Errorf("message value denom must be %s", appparams.BaseDenom)
	}
	if !m.Value.IsValid() || m.Value.Amount.IsNegative() {
		return errors.New("message value must be valid and non-negative")
	}
	if len(m.Body) > int(params.MaxBodySize) {
		return fmt.Errorf("message body size must be <= %d", params.MaxBodySize)
	}
	if m.DeliverAtBlock != 0 && m.DeadlineBlock != 0 && m.DeliverAtBlock > m.DeadlineBlock {
		return errors.New("message deliver block must not exceed deadline block")
	}
	if m.RetryCount > m.MaxRetries {
		return errors.New("message retry count must not exceed max retries")
	}
	if m.MaxRetries > params.MaxRetriesPerMessage {
		return fmt.Errorf("message max retries must be <= %d", params.MaxRetriesPerMessage)
	}
	if m.RetryCount > 0 && m.MaxRetries == 0 {
		return errors.New("message retry count requires max retries")
	}
	if m.RetryDelayBlocks > params.MaxRetryDelayBlocks {
		return fmt.Errorf("message retry delay blocks must be <= %d", params.MaxRetryDelayBlocks)
	}
	if m.GasLimit == 0 {
		return errors.New("message gas limit must be positive")
	}
	if m.ForwardFee.Denom != appparams.BaseDenom {
		return fmt.Errorf("message forward fee denom must be %s", appparams.BaseDenom)
	}
	if !m.ForwardFee.IsValid() || m.ForwardFee.Amount.IsNegative() {
		return errors.New("message forward fee must be valid and non-negative")
	}
	if m.Depth > params.MaxRecursionDepth {
		return fmt.Errorf("message depth must be <= %d", params.MaxRecursionDepth)
	}
	return nil
}
