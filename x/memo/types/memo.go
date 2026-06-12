package types

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"unicode"
	"unicode/utf8"

	sdkmath "cosmossdk.io/math"

	appparams "github.com/sovereign-l1/l1/app/params"
)

const (
	DefaultMaxMemoChars	= uint32(200)
	HardMaxMemoChars	= uint32(500)
	DefaultMaxMemoBytes	= uint32(1024)
	HardMaxMemoBytes	= uint32(4096)
	MemoHashBytes		= 32

	DefaultMemoBaseFee	= int64(0)
	DefaultMemoByteFee	= int64(1)

	MemoReputationEliteBps		= uint32(7_500)
	MemoReputationNormalBps		= uint32(10_000)
	MemoReputationNewBps		= uint32(15_000)
	MemoReputationRestrictedBps	= uint32(30_000)
	DefaultCongestionBps		= uint32(10_000)
	BpsDenominator			= uint32(10_000)
)

type TxMetadata struct {
	Memo		string
	MemoHash	[]byte
	MemoVisible	bool
}

type MemoParams struct {
	MaxMemoChars	uint32
	MaxMemoBytes	uint32
	MemoBaseFee	sdkmath.Int
	MemoByteFee	sdkmath.Int
}

func DefaultMemoParams() MemoParams {
	return MemoParams{
		MaxMemoChars:	DefaultMaxMemoChars,
		MaxMemoBytes:	DefaultMaxMemoBytes,
		MemoBaseFee:	sdkmath.NewInt(DefaultMemoBaseFee),
		MemoByteFee:	sdkmath.NewInt(DefaultMemoByteFee),
	}
}

func ValidateMemoParams(params MemoParams) error {
	if params.MaxMemoChars > HardMaxMemoChars {
		return fmt.Errorf("max memo chars must not exceed hard bound %d", HardMaxMemoChars)
	}
	if params.MaxMemoBytes > HardMaxMemoBytes {
		return fmt.Errorf("max memo bytes must not exceed hard bound %d", HardMaxMemoBytes)
	}
	if params.MaxMemoBytes == 0 && params.MaxMemoChars > 0 {
		return errors.New("max memo bytes must be positive when memo chars are allowed")
	}
	if params.MemoBaseFee.IsNegative() {
		return errors.New("memo base fee must not be negative")
	}
	if params.MemoByteFee.IsNegative() {
		return errors.New("memo byte fee must not be negative")
	}
	return nil
}

func ValidateTxMetadata(metadata TxMetadata, params MemoParams) error {
	if err := ValidateMemoParams(params); err != nil {
		return err
	}
	if len(metadata.MemoHash) > 0 && len(metadata.MemoHash) != MemoHashBytes {
		return fmt.Errorf("memo hash must be %d bytes", MemoHashBytes)
	}
	if metadata.Memo == "" {
		return nil
	}
	if !utf8.ValidString(metadata.Memo) {
		return errors.New("memo must be valid UTF-8")
	}
	if utf8.RuneCountInString(metadata.Memo) > int(params.MaxMemoChars) {
		return fmt.Errorf("memo character count must not exceed %d", params.MaxMemoChars)
	}
	if len([]byte(metadata.Memo)) > int(params.MaxMemoBytes) {
		return fmt.Errorf("memo byte length must not exceed %d", params.MaxMemoBytes)
	}
	for _, r := range metadata.Memo {
		if unicode.IsControl(r) {
			return fmt.Errorf("memo contains prohibited control character U+%04X", r)
		}
	}
	if len(metadata.MemoHash) > 0 {
		expected := MemoHash(metadata.Memo)
		if string(metadata.MemoHash) != string(expected) {
			return errors.New("memo hash does not match memo")
		}
	}
	return nil
}

func MemoFee(metadata TxMetadata, params MemoParams, reputationScore uint8, congestionMultiplierBps uint32) (sdkmath.Int, string, error) {
	if err := ValidateTxMetadata(metadata, params); err != nil {
		return sdkmath.Int{}, "", err
	}
	if metadata.Memo == "" {
		return sdkmath.ZeroInt(), appparams.BaseDenom, nil
	}
	if congestionMultiplierBps == 0 {
		return sdkmath.Int{}, "", errors.New("congestion multiplier bps must be positive")
	}
	raw := params.MemoBaseFee.Add(params.MemoByteFee.MulRaw(int64(len([]byte(metadata.Memo)))))
	raw = applyBps(raw, MemoReputationMultiplierBps(reputationScore))
	raw = applyBps(raw, congestionMultiplierBps)
	return raw, appparams.BaseDenom, nil
}

func MemoReputationMultiplierBps(score uint8) uint32 {
	switch {
	case score >= 80:
		return MemoReputationEliteBps
	case score >= 50:
		return MemoReputationNormalBps
	case score >= 20:
		return MemoReputationNewBps
	default:
		return MemoReputationRestrictedBps
	}
}

func MemoHash(memo string) []byte {
	sum := sha256.Sum256([]byte(memo))
	return sum[:]
}

func MetadataAffectsExecution(TxMetadata) bool {
	return false
}

func applyBps(amount sdkmath.Int, bps uint32) sdkmath.Int {
	return amount.MulRaw(int64(bps)).AddRaw(int64(BpsDenominator - 1)).QuoRaw(int64(BpsDenominator))
}
