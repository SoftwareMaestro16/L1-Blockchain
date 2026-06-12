package addressing

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"strings"
)

type RawAddressPolicyVersion uint32

const (
	RawAddressPolicyVersionLegacyPadded	RawAddressPolicyVersion	= 1
	RawAddressPolicyVersionV2		RawAddressPolicyVersion	= 2
)

type RawAddressClass string

const (
	RawAddressClassUnknown		RawAddressClass	= "unknown"
	RawAddressClassSystemFixed	RawAddressClass	= "system_fixed"
	RawAddressClassLegacyPadded	RawAddressClass	= "legacy_padded"
	RawAddressClassV2		RawAddressClass	= "v2_256_bit"
)

const (
	rawAddressV2DomainSeparator	= "aetra-raw-address-v2"
	rawAddressV2MaxAttempts		= 256
)

func ClassifyRawAddressText(text string) (RawAddressClass, error) {
	raw, err := Parse(text)
	if err != nil {
		return RawAddressClassUnknown, err
	}
	return ClassifyRawAddressBytes(raw), nil
}

func ClassifyRawAddressBytes(raw []byte) RawAddressClass {
	if len(raw) == 0 {
		return RawAddressClassUnknown
	}
	if _, found := SystemAddressByBytes(raw); found {
		return RawAddressClassSystemFixed
	}
	if len(raw) != rawPayloadLength {
		return RawAddressClassUnknown
	}
	if hasLegacyRawAddressPrefix(raw) {
		return RawAddressClassLegacyPadded
	}
	return RawAddressClassV2
}

func IsLegacyPaddedRawAddress(raw []byte) bool {
	return ClassifyRawAddressBytes(raw) == RawAddressClassLegacyPadded
}

func IsV2RawAddress(raw []byte) bool {
	return ClassifyRawAddressBytes(raw) == RawAddressClassV2
}

func ValidateRawAddressPolicy(raw []byte, policyVersion RawAddressPolicyVersion) error {
	class := ClassifyRawAddressBytes(raw)
	if class == RawAddressClassSystemFixed {
		return nil
	}
	switch policyVersion {
	case RawAddressPolicyVersionLegacyPadded:
		if class != RawAddressClassLegacyPadded {
			return fmt.Errorf("raw address must use legacy padded policy version %d or a reserved system address", policyVersion)
		}
	case RawAddressPolicyVersionV2:
		if class != RawAddressClassV2 {
			return fmt.Errorf("raw address must use raw address policy version %d and avoid legacy padding", policyVersion)
		}
	default:
		return fmt.Errorf("unsupported raw address policy version %d", policyVersion)
	}
	return nil
}

func NormalizeV2RawAddress(domain string, raw []byte) ([]byte, error) {
	domain = strings.TrimSpace(domain)
	if domain == "" {
		return nil, fmt.Errorf("raw address domain is required")
	}
	candidate, err := ToRawPayload(raw)
	if err != nil {
		return nil, err
	}
	class := ClassifyRawAddressBytes(candidate)
	switch class {
	case RawAddressClassSystemFixed, RawAddressClassV2:
		return append([]byte(nil), candidate...), nil
	case RawAddressClassLegacyPadded:
		return deriveV2RawAddress(domain, candidate)
	default:
		return nil, fmt.Errorf("unsupported raw address candidate length %d", len(raw))
	}
}

func deriveV2RawAddress(domain string, seed []byte) ([]byte, error) {
	for attempt := uint32(0); attempt < rawAddressV2MaxAttempts; attempt++ {
		h := sha256.New()
		writeRawAddressPart(h.Write, []byte(rawAddressV2DomainSeparator))
		writeRawAddressPart(h.Write, []byte(domain))
		writeRawAddressPart(h.Write, seed)
		var attemptBytes [4]byte
		binary.BigEndian.PutUint32(attemptBytes[:], attempt)
		_, _ = h.Write(attemptBytes[:])
		sum := h.Sum(nil)
		if ClassifyRawAddressBytes(sum) == RawAddressClassV2 {
			return sum, nil
		}
	}
	return nil, fmt.Errorf("failed to derive v2 raw address for %q without legacy padding", domain)
}

func hasLegacyRawAddressPrefix(raw []byte) bool {
	if len(raw) != rawPayloadLength {
		return false
	}
	for i := 0; i < longAddressPadLength; i++ {
		if raw[i] != 0 {
			return false
		}
	}
	return true
}

func writeRawAddressPart(write func([]byte) (int, error), bz []byte) {
	var length [4]byte
	binary.BigEndian.PutUint32(length[:], uint32(len(bz)))
	_, _ = write(length[:])
	_, _ = write(bz)
}
