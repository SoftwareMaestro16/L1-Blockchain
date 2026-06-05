package types

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sovereign-l1/l1/app/addressing"
)

func NormalizeAETDomain(name string) (string, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "", errors.New("identity domain is required")
	}
	lower := strings.ToLower(trimmed)
	if strings.Contains(lower, ".") && !strings.HasSuffix(lower, DomainTLD) {
		return "", fmt.Errorf("identity domain must end with %s", DomainTLD)
	}
	if !strings.HasSuffix(lower, DomainTLD) {
		lower += DomainTLD
	}
	labelsPart := strings.TrimSuffix(lower, DomainTLD)
	if labelsPart == "" {
		return "", errors.New("identity domain label is required")
	}
	if len(lower) > MaxDomainFullBytes {
		return "", fmt.Errorf("identity domain must be <= %d bytes", MaxDomainFullBytes)
	}
	labels := strings.Split(labelsPart, ".")
	if len(labels) > MaxDomainLabels {
		return "", fmt.Errorf("identity domain must not exceed %d labels", MaxDomainLabels)
	}
	for _, label := range labels {
		if label == "" {
			return "", errors.New("identity domain contains empty label")
		}
		if len(label) > MaxDomainLabelBytes {
			return "", fmt.Errorf("identity domain label must be <= %d bytes", MaxDomainLabelBytes)
		}
		if err := validateDomainLabel(label); err != nil {
			return "", err
		}
	}
	return lower, nil
}

func ComputeRegistrationCommitment(name string, owner sdk.AccAddress, salt string) (string, error) {
	normalized, err := NormalizeAETDomain(name)
	if err != nil {
		return "", err
	}
	if err := validateSpecAddress("identity commitment owner", owner); err != nil {
		return "", err
	}
	if strings.TrimSpace(salt) == "" {
		return "", errors.New("identity commitment salt is required")
	}
	return identityHash("registration-commitment", normalized, string(owner), salt), nil
}

func ComputeAuctionCommitment(name string, bidder sdk.AccAddress, bid uint64, salt string) (string, error) {
	normalized, err := NormalizeAETDomain(name)
	if err != nil {
		return "", err
	}
	if err := validateSpecAddress("identity auction bidder", bidder); err != nil {
		return "", err
	}
	if bid == 0 {
		return "", errors.New("identity auction bid must be positive")
	}
	if strings.TrimSpace(salt) == "" {
		return "", errors.New("identity auction salt is required")
	}
	return identityHash("auction-commitment", normalized, string(bidder), fmt.Sprintf("%020d", bid), salt), nil
}

func DomainNFTID(name string) (string, error) {
	normalized, err := NormalizeAETDomain(name)
	if err != nil {
		return "", err
	}
	return "anft66:domain:" + normalized, nil
}

func identityHash(parts ...string) string {
	h := sha256.New()
	for _, part := range parts {
		var length [8]byte
		binary.BigEndian.PutUint64(length[:], uint64(len(part)))
		_, _ = h.Write(length[:])
		_, _ = h.Write([]byte(part))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func validateSpecAddress(field string, addr sdk.AccAddress) error {
	if len(addr) == 0 {
		return fmt.Errorf("%s is required", field)
	}
	return addressing.RejectZeroAddress(field, addr)
}

func cloneSpecAddress(addr sdk.AccAddress) sdk.AccAddress {
	if len(addr) == 0 {
		return nil
	}
	return append(sdk.AccAddress(nil), addr...)
}

func compareAddress(left, right sdk.AccAddress) int {
	return strings.Compare(string(left), string(right))
}

func sortStringSet(values []string) []string {
	out := append([]string(nil), values...)
	sort.Strings(out)
	return out
}
