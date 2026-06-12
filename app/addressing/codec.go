package addressing

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strings"

	coreaddress "cosmossdk.io/core/address"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	RawPrefix		= "4:"
	SystemRawPrefix		= "-7:"
	RawAddressLength	= 66
	SystemRawAddressLength	= 67
	UserFriendlyLength	= 48
	UserFriendlyPrefix	= "AE"
	ZeroRawAddress		= "4:0000000000000000000000000000000000000000000000000000000000000000"
	ZeroUserFriendly	= "AEAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	rawPayloadLength	= 32
	shortAddressLength	= 20
	longAddressPadLength	= rawPayloadLength - shortAddressLength
	userFriendlyVersion	= byte(1)
)

var (
	userFriendlyMagic	= [3]byte{0x00, 0x40, 0x00}
	rawAddressRe		= regexp.MustCompile(`^4:[0-9a-f]{64}$`)
	systemRawAddressRe	= regexp.MustCompile(`^-7:[0-9a-f]{64}$`)
)

type Codec struct{}

var _ coreaddress.Codec = Codec{}

func (Codec) BytesToString(bz []byte) (string, error) {
	if len(bz) == 0 {
		return "", nil
	}
	return FormatUserFriendly(bz)
}

func (Codec) StringToBytes(text string) ([]byte, error) {
	return Parse(text)
}

func Format(bz []byte) string {
	raw, err := ToRawPayload(bz)
	if err != nil {
		panic(err)
	}
	return RawPrefix + hex.EncodeToString(raw)
}

func IsSystemRawAddress(text string) bool {
	return systemRawAddressRe.MatchString(strings.TrimSpace(text))
}

func ParseSystemRawAddress(text string) ([]byte, error) {
	text = strings.TrimSpace(text)
	if !systemRawAddressRe.MatchString(text) {
		return nil, fmt.Errorf("invalid system raw address format: expected -7:<64 lowercase hex>")
	}
	raw, err := hex.DecodeString(text[len(SystemRawPrefix):])
	if err != nil {
		return nil, err
	}
	return FromRawPayload(raw), nil
}

func FormatSystemRawAddress(raw []byte) string {
	payload, err := ToRawPayload(raw)
	if err != nil {
		panic(err)
	}
	return SystemRawPrefix + hex.EncodeToString(payload)
}

func FormatAccAddress(addr sdk.AccAddress) string {
	return mustFormatUserFriendly(addr.Bytes())
}

func FormatValAddress(addr sdk.ValAddress) string {
	return mustFormatUserFriendly(addr.Bytes())
}

func FormatConsAddress(addr sdk.ConsAddress) string {
	return mustFormatUserFriendly(addr.Bytes())
}

func IsZero(bz []byte) bool {
	raw, err := ToRawPayload(bz)
	if err != nil {
		return false
	}
	for _, b := range raw {
		if b != 0 {
			return false
		}
	}
	return true
}

func IsZeroAccAddress(addr sdk.AccAddress) bool {
	return IsZero(addr.Bytes())
}

func FormatUserFriendly(bz []byte) (string, error) {
	raw, err := ToRawPayload(bz)
	if err != nil {
		return "", err
	}
	payload := make([]byte, 0, 36)
	payload = append(payload, userFriendlyMagic[:]...)
	payload = append(payload, userFriendlyVersion)
	payload = append(payload, raw...)
	return base64.RawURLEncoding.EncodeToString(payload), nil
}

func mustFormatUserFriendly(bz []byte) string {
	text, err := FormatUserFriendly(bz)
	if err != nil {
		panic(err)
	}
	return text
}

func ParseAccAddress(text string) (sdk.AccAddress, error) {
	bz, err := Parse(text)
	if err != nil {
		return nil, err
	}
	return sdk.AccAddress(bz), nil
}

func Parse(text string) ([]byte, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, errors.New("empty address string is not allowed")
	}
	if rawAddressRe.MatchString(text) {
		raw, err := hex.DecodeString(text[len(RawPrefix):])
		if err != nil {
			return nil, err
		}
		return FromRawPayload(raw), nil
	}
	if systemRawAddressRe.MatchString(text) {
		return ParseSystemRawAddress(text)
	}
	if strings.HasPrefix(text, UserFriendlyPrefix) {
		for _, address := range reservedSystemAddresses {
			if address.UserFriendly == text {
				if systemRawAddressRe.MatchString(address.Raw) {
					return ParseSystemRawAddress(address.Raw)
				}
				raw, err := hex.DecodeString(address.Raw[len(RawPrefix):])
				if err != nil {
					return nil, err
				}
				return FromRawPayload(raw), nil
			}
		}
	}
	if len(text) == UserFriendlyLength && strings.HasPrefix(text, UserFriendlyPrefix) {
		payload, err := base64.RawURLEncoding.DecodeString(text)
		if err != nil {
			return nil, err
		}
		if len(payload) != 36 ||
			payload[0] != userFriendlyMagic[0] ||
			payload[1] != userFriendlyMagic[1] ||
			payload[2] != userFriendlyMagic[2] ||
			payload[3] != userFriendlyVersion {
			return nil, fmt.Errorf("invalid AE userfriendly address header")
		}
		return FromRawPayload(payload[4:]), nil
	}
	for _, prefix := range []string{
		sdk.GetConfig().GetBech32AccountAddrPrefix(),
		sdk.GetConfig().GetBech32ValidatorAddrPrefix(),
		sdk.GetConfig().GetBech32ConsensusAddrPrefix(),
	} {
		if prefix == "" {
			continue
		}
		bz, err := sdk.GetFromBech32(text, prefix)
		if err == nil {
			if verifyErr := sdk.VerifyAddressFormat(bz); verifyErr != nil {
				return nil, verifyErr
			}
			return bz, nil
		}
	}
	return nil, fmt.Errorf("invalid address format: expected 4:<64 lowercase hex>, -7:<64 lowercase hex>, or 48-char AE userfriendly address")
}

func ToRawPayload(bz []byte) ([]byte, error) {
	switch len(bz) {
	case shortAddressLength:
		raw := make([]byte, rawPayloadLength)
		copy(raw[longAddressPadLength:], bz)
		return raw, nil
	case rawPayloadLength:
		raw := make([]byte, rawPayloadLength)
		copy(raw, bz)
		return raw, nil
	default:
		return nil, fmt.Errorf("unsupported address byte length %d", len(bz))
	}
}

func FromRawPayload(raw []byte) []byte {
	if len(raw) != rawPayloadLength {
		return nil
	}
	for i := 0; i < longAddressPadLength; i++ {
		if raw[i] != 0 {
			out := make([]byte, rawPayloadLength)
			copy(out, raw)
			return out
		}
	}
	out := make([]byte, shortAddressLength)
	copy(out, raw[longAddressPadLength:])
	return out
}
