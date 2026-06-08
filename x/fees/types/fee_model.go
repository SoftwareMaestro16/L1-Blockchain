package types

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/sovereign-l1/l1/app/params"
)

const (
	BasisPoints                   = uint64(10_000)
	DefaultBaseFeeAmount          = MinDefaultFeeAmount
	DefaultMaxFeeAmount           = "1000"
	DefaultTargetUtilizationBps   = uint32(5_000)
	DefaultCongestionBps          = uint32(8_000)
	DefaultMaxTxGas               = uint64(1_000_000)
	DefaultMaxBlockGas            = uint64(20_000_000)
	DefaultMaxBlockTxs            = uint64(5_000)
	DefaultMaxTxBytes             = uint64(256 * 1024)
	DefaultMaxMemoBytes           = uint64(1024)
	DefaultMaxMessagesPerTx       = uint64(16)
	DefaultMinGasPriceUnitGas     = uint64(100_000)
	DefaultSenderTxsPerBlock      = uint64(25)
	DefaultStakeAllowanceStep     = "1000000000"
	DefaultStakeSenderTxsPerBlock = uint64(250)
	DefaultFeePriorityWeightBps   = uint32(1_000)
	DefaultStakePriorityWeightBps = uint32(9_000)
)

type FeeQuote struct {
	RequiredFee       sdk.Coin
	BaseFee           sdk.Coin
	MaxFee            sdk.Coin
	UtilizationBps    uint32
	Congested         bool
	AtHardCap         bool
	AcceptedFeeAmount sdkmath.Int
	EconomicControl   appparams.BalanceControllerOutput
}

type AdmissionInput struct {
	Fee              sdk.Coins
	GasLimit         uint64
	BlockGasConsumed uint64
	BlockTxCount     uint64
	SenderTxCount    uint64
	SenderStake      sdkmath.Int
}

type TxEnvelopeLimits struct {
	MaxTxBytes       uint64
	MaxMemoBytes     uint64
	MaxMessagesPerTx uint64
}

type TxEnvelopeInput struct {
	TxBytes  uint64
	Memo     string
	MsgCount uint64
}

func DefaultTxEnvelopeLimits() TxEnvelopeLimits {
	return TxEnvelopeLimits{
		MaxTxBytes:       DefaultMaxTxBytes,
		MaxMemoBytes:     DefaultMaxMemoBytes,
		MaxMessagesPerTx: DefaultMaxMessagesPerTx,
	}
}

func ValidateTxEnvelope(limits TxEnvelopeLimits, in TxEnvelopeInput) error {
	if limits.MaxTxBytes == 0 {
		limits.MaxTxBytes = DefaultMaxTxBytes
	}
	if limits.MaxMemoBytes == 0 {
		limits.MaxMemoBytes = DefaultMaxMemoBytes
	}
	if limits.MaxMessagesPerTx == 0 {
		limits.MaxMessagesPerTx = DefaultMaxMessagesPerTx
	}
	if in.TxBytes > limits.MaxTxBytes {
		return ErrInvalidFee.Wrapf("tx size %d exceeds max_tx_bytes %d", in.TxBytes, limits.MaxTxBytes)
	}
	if uint64(len([]byte(in.Memo))) > limits.MaxMemoBytes {
		return ErrInvalidFee.Wrapf("memo size %d exceeds max_memo_bytes %d", len([]byte(in.Memo)), limits.MaxMemoBytes)
	}
	if in.MsgCount > limits.MaxMessagesPerTx {
		return ErrInvalidFee.Wrapf("message count %d exceeds max_messages_per_tx %d", in.MsgCount, limits.MaxMessagesPerTx)
	}
	return nil
}

func QuoteFee(params Params, gasLimit, blockGasConsumed uint64) (FeeQuote, error) {
	params = NormalizeParams(params)
	if err := params.Validate(); err != nil {
		return FeeQuote{}, err
	}
	baseFee, err := params.BaseFeeInt()
	if err != nil {
		return FeeQuote{}, err
	}
	maxFee, err := params.MaxFeeInt()
	if err != nil {
		return FeeQuote{}, err
	}
	utilization := BlockUtilizationBps(blockGasConsumed, gasLimit, params.MaxBlockGas)
	required := DynamicFeeAmount(baseFee, maxFee, params.TargetBlockUtilizationBps, utilization)
	minGasFee, err := MinimumGasPriceFee(params, gasLimit)
	if err != nil {
		return FeeQuote{}, err
	}
	if required.LT(minGasFee) {
		required = minGasFee
	}
	accepted := required
	if accepted.GT(maxFee) {
		accepted = maxFee
	}
	economicControl, err := appparams.BalanceController(appparams.BalanceControllerInput{
		CurrentInflationBps: appparams.DefaultTargetInflationBps,
		StakeRatioBps:       appparams.DefaultTargetStakeBps,
		BlockLoadBps:        int64(utilization),
	})
	if err != nil {
		return FeeQuote{}, err
	}
	return FeeQuote{
		RequiredFee:       sdk.NewCoin(BondDenom, accepted),
		BaseFee:           sdk.NewCoin(BondDenom, baseFee),
		MaxFee:            sdk.NewCoin(BondDenom, maxFee),
		UtilizationBps:    utilization,
		Congested:         utilization >= params.CongestionThresholdBps,
		AtHardCap:         accepted.Equal(maxFee),
		AcceptedFeeAmount: accepted,
		EconomicControl:   economicControl,
	}, nil
}

func MinimumGasPriceFee(params Params, gasLimit uint64) (sdkmath.Int, error) {
	params = NormalizeParams(params)
	minFee, err := params.MinFeeInt()
	if err != nil {
		return sdkmath.Int{}, err
	}
	if gasLimit == 0 {
		return minFee, nil
	}
	units := (gasLimit + DefaultMinGasPriceUnitGas - 1) / DefaultMinGasPriceUnitGas
	return minFee.MulRaw(int64(units)), nil // #nosec G115 -- gas units are bounded by MaxTxGas validation before admission.
}

func ValidateAdmission(params Params, in AdmissionInput) (FeeQuote, error) {
	params = NormalizeParams(params)
	if err := params.Validate(); err != nil {
		return FeeQuote{}, err
	}
	if in.GasLimit == 0 {
		return FeeQuote{}, ErrInvalidFee.Wrap("gas limit must be positive")
	}
	if in.GasLimit > params.MaxTxGas {
		return FeeQuote{}, ErrInvalidFee.Wrapf("gas limit %d exceeds max_tx_gas %d", in.GasLimit, params.MaxTxGas)
	}
	if in.BlockTxCount >= params.MaxBlockTxs {
		return FeeQuote{}, ErrInvalidFee.Wrapf("block transaction limit %d reached", params.MaxBlockTxs)
	}
	if in.BlockGasConsumed+in.GasLimit < in.BlockGasConsumed || in.BlockGasConsumed+in.GasLimit > params.MaxBlockGas {
		return FeeQuote{}, ErrInvalidFee.Wrapf("block gas limit %d reached", params.MaxBlockGas)
	}
	senderLimit, err := SenderTxLimit(params, in.SenderStake)
	if err != nil {
		return FeeQuote{}, err
	}
	if in.SenderTxCount >= senderLimit {
		return FeeQuote{}, ErrInvalidFee.Wrapf("sender rate limit %d transactions per block reached", senderLimit)
	}
	quote, err := QuoteFee(params, in.GasLimit, in.BlockGasConsumed)
	if err != nil {
		return FeeQuote{}, err
	}
	if err := ValidateFeeCoins(params, in.Fee, true); err != nil {
		return FeeQuote{}, err
	}
	amount := in.Fee.AmountOf(BondDenom)
	if amount.GT(quote.MaxFee.Amount) {
		return FeeQuote{}, ErrInvalidFee.Wrapf("fee must not exceed hard cap %s%s", quote.MaxFee.Amount.String(), BondDenom)
	}
	if amount.LT(quote.RequiredFee.Amount) {
		return FeeQuote{}, ErrInvalidFee.Wrapf("fee must be at least dynamic requirement %s", quote.RequiredFee.String())
	}
	return quote, nil
}

func DynamicFeeAmount(baseFee, maxFee sdkmath.Int, targetBps, utilizationBps uint32) sdkmath.Int {
	if utilizationBps <= targetBps || baseFee.GTE(maxFee) {
		return baseFee
	}
	remainingBps := uint64(BasisPoints) - uint64(targetBps)
	if remainingBps == 0 {
		return maxFee
	}
	overBps := uint64(utilizationBps - targetBps)
	rangeAmount := maxFee.Sub(baseFee)
	numerator := rangeAmount.MulRaw(int64(overBps)).MulRaw(int64(overBps))
	denominator := sdkmath.NewIntFromUint64(remainingBps * remainingBps)
	increment := ceilQuo(numerator, denominator)
	required := baseFee.Add(increment)
	if required.GT(maxFee) {
		return maxFee
	}
	return required
}

func BlockUtilizationBps(blockGasConsumed, txGasLimit, maxBlockGas uint64) uint32 {
	if maxBlockGas == 0 {
		return 0
	}
	used := blockGasConsumed + txGasLimit
	if used < blockGasConsumed || used >= maxBlockGas {
		return uint32(BasisPoints)
	}
	return uint32((used * BasisPoints) / maxBlockGas)
}

func SenderTxLimit(params Params, stake sdkmath.Int) (uint64, error) {
	params = NormalizeParams(params)
	step, err := params.StakeTxAllowanceStepInt()
	if err != nil {
		return 0, err
	}
	limit := params.MaxSenderTxsPerBlock
	if stake.IsPositive() && step.IsPositive() {
		bonus := stake.Quo(step).Uint64()
		limit += bonus
	}
	if limit > params.MaxSenderTxsPerBlockWithStake {
		return params.MaxSenderTxsPerBlockWithStake, nil
	}
	return limit, nil
}

func PriorityScore(params Params, paidFee sdk.Coin, requiredFee sdk.Coin, stake sdkmath.Int) (int64, error) {
	params = NormalizeParams(params)
	if err := params.Validate(); err != nil {
		return 0, err
	}
	if paidFee.Denom != BondDenom || requiredFee.Denom != BondDenom {
		return 0, fmt.Errorf("priority fee denom must be %s", BondDenom)
	}
	if !paidFee.IsPositive() || !requiredFee.IsPositive() {
		return 0, fmt.Errorf("priority fees must be positive")
	}
	feeCredit := paidFee.Amount
	if feeCredit.GT(requiredFee.Amount) {
		feeCredit = requiredFee.Amount
	}
	feeScore := feeCredit.MulRaw(int64(params.FeePriorityWeightBps)).Quo(requiredFee.Amount).Int64()
	stakeStep, err := params.StakeTxAllowanceStepInt()
	if err != nil {
		return 0, err
	}
	stakeUnits := sdkmath.ZeroInt()
	if stake.IsPositive() && stakeStep.IsPositive() {
		stakeUnits = stake.Quo(stakeStep)
	}
	stakeScore := stakeUnits.MulRaw(int64(params.StakePriorityWeightBps)).Int64()
	return feeScore + stakeScore, nil
}

func ceilQuo(numerator, denominator sdkmath.Int) sdkmath.Int {
	if numerator.IsZero() {
		return sdkmath.ZeroInt()
	}
	quotient := numerator.Quo(denominator)
	if numerator.Mod(denominator).IsPositive() {
		return quotient.AddRaw(1)
	}
	return quotient
}
