package addressing

import (
	"encoding/hex"
	"fmt"
	"strings"
)

const SystemAddressStatusActive = "active"

type SystemAddress struct {
	Name                string
	ModuleName          string
	Raw                 string
	UserFriendly        string
	Core                bool
	CanHoldFunds        bool
	CanReceiveUserFunds bool
	CanSendFunds        bool
	Status              string
}

var reservedSystemAddresses = []SystemAddress{
	systemAddress("AETElector", "validator-election", "-7:01041041041041041041041041041041041041041041041041041042c4093391", "AEAAAQEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEELECTOR", true, false, false, false),
	systemAddress("AETConfig", "config", "-7:008208208208208208208208208208208208208208208208208208208e345206", "AEAAAQCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCONFIG", true, false, false, false),
	systemAddress("AETConstitution", "constitution", "-7:034d34d34d34d34d34d34d34d34d34d34d34d34d34d34d34d34d34d34d34d34d", "AEAAAQNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNN", true, false, false, false),
	systemAddress("AETSystemRegistry", "system-registry", "-7:0451451451451451451451451451451451451451451451451451451451451451", "AEAAAQRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRR", true, false, false, false),
	systemAddress("AETValidatorRegistry", "validator-registry", "-7:0555555555555555555555555555555555555555555555555555555555555555", "AEAAAQVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVV", true, false, false, false),
	systemAddress("AETConfigVoting", "config-voting", "-7:0186186186186186186186186186186186186186186186186186186186186186", "AEAAAQGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGG", true, false, false, false),

	systemAddress("AETMint", "mint-authority", "4:030c30c30c30c30c30c30c30c30c30c30c30c30c30c30c30c30c30c30c308353", "AEAAAQMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMINT", false, false, false, false),
	systemAddress("AETBurn", "burn", "4:004104104104104104104104104104104104104104104104104104104105444d", "AEAAAQBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBURN", false, false, true, false),
	systemAddress("AETEvidence", "evidence", "4:00c30c30c30c30c30c30c30c30c30c30c30c30c30c30c30c30c30c30c30c30c3", "AEAAAQDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDDD", false, false, false, false),
	systemAddress("AETReporterRewards", "reporter", "4:03cf3cf3cf3cf3cf3cf3cf3cf3cf3cf3cf3cf3cf3cf3cf3cf3cf3cf3cf3cf3cf", "AEAAAQPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPPP", false, true, false, false),
	systemAddress("AETNominatorPool", "nominator-pool", "4:038e38e38e38e38e38e38e38e38e38e38e38e38e38e38e38e38e38e38e38e38e", "AEAAAQOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOOO", false, false, false, false),
	systemAddress("AETSingleNominatorPool", "single-nominator-pool", "4:0492492492492492492492492492492492492492492492492492492492492492", "AEAAAQSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSS", false, false, false, false),
	systemAddress("AETValidatorInsurance", "validator-insurance", "4:0208208208208208208208208208208208208208208208208208208208208208", "AEAAAQIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIIII", false, true, false, false),
	systemAddress("AETDelegatorProtection", "delegator-protection", "4:02cb2cb2cb2cb2cb2cb2cb2cb2cb2cb2cb2cb2cb2cb2cb2cb2cb2cb2cb2cb2cb", "AEAAAQLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLLL", false, true, false, false),
	systemAddress("AETReputation", "reputation", "4:0514514514514514514514514514514514514514514514514514514514514514", "AEAAAQUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUUU", false, false, false, false),
	systemAddress("AETPerformanceOracle", "performance-oracle", "4:0145145145145145145145145145145145145145145145145145145145145145", "AEAAAQFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF", false, false, false, false),
	systemAddress("AETStakeConcentration", "stake-concentration", "4:028a28a28a28a28a28a28a28a28a28a28a28a28a28a28a28a28a28a28a28a28a", "AEAAAQKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKKK", false, false, false, false),
	systemAddress("AETDynamicCommission", "dynamic-commission", "4:0249249249249249249249249249249249249249249249249249249249249249", "AEAAAQJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJJ", false, false, false, false),
	systemAddress("AETEmissions", "emissions", "4:0618618618618618618618618618618618618618618618618618618618618618", "AEAAAQYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYY", false, false, false, false),
	systemAddress("AETFeeCollector", "fee-collector", "4:0410410410410410410410410410410410410410410410410410410410410410", "AEAAAQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQQ", false, true, false, false),
	systemAddress("AETTreasury", "treasury", "4:04d34d34d34d34d34d34d34d34d34d34d34d34d34d34d34d34d34d34d34d34d3", "AEAAAQTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTTT", false, true, false, false),
	systemAddress("AETScheduler", "scheduler", "4:01c71c71c71c71c71c71c71c71c71c71c71c71c71c71c71c71c71c71c71c71c7", "AEAAAQHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHH", false, false, false, false),
	systemAddress("AETAVMScheduler", "avm-scheduler", "4:0596596596596596596596596596596596596596596596596596596596596596", "AEAAAQWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWW", false, false, false, false),
	systemAddress("AETActorRegistry", "actor-registry", "4:05d75d75d75d75d75d75d75d75d75d75d75d75d75d75d75d75d75d75d75d75d7", "AEAAAQXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX", false, false, false, false),
	systemAddress("AETStorageRent", "storage-rent", "4:0659659659659659659659659659659659659659659659659659659659659659", "AEAAAQZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZZ", false, true, false, false),
	systemAddress("AETIdentityRoot", "identity-root", "4:0211211211211211211211211211211211211211211211211211211211211211", "AEAAAQIRIRIRIRIRIRIRIRIRIRIRIRIRIRIRIRIRIRIRIRIR", false, false, false, false),
	systemAddress("AETBridgeHub", "bridge-hub", "4:0047047047047047047047047047047047047047047047047047047047047047", "AEAAAQBHBHBHBHBHBHBHBHBHBHBHBHBHBHBHBHBHBHBHBHBH", false, false, false, false),
	systemAddress("AETCrossChainRegistry", "cross-chain-registry", "4:0082082082082082082082082082082082082082082082082082082082082082", "AEAAAQCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCC", false, false, false, false),
	systemAddress("AETShardingCoordinator", "sharding-coordinator", "4:0487487487487487487487487487487487487487487487487487487487487487", "AEAAAQSHSHSHSHSHSHSHSHSHSHSHSHSHSHSHSHSHSHSHSHSH", false, false, false, false),
}

func systemAddress(name, moduleName, raw, userFriendly string, core, canHoldFunds, canReceiveUserFunds, canSendFunds bool) SystemAddress {
	return SystemAddress{
		Name:                name,
		ModuleName:          moduleName,
		Raw:                 raw,
		UserFriendly:        userFriendly,
		Core:                core,
		CanHoldFunds:        canHoldFunds,
		CanReceiveUserFunds: canReceiveUserFunds,
		CanSendFunds:        canSendFunds,
		Status:              SystemAddressStatusActive,
	}
}

func AllSystemAddresses() []SystemAddress {
	out := make([]SystemAddress, len(reservedSystemAddresses))
	copy(out, reservedSystemAddresses)
	return out
}

func SystemAddressByName(name string) (SystemAddress, bool) {
	name = strings.TrimSpace(name)
	for _, address := range reservedSystemAddresses {
		if address.Name == name {
			return address, true
		}
	}
	return SystemAddress{}, false
}

func SystemAddressByRaw(raw string) (SystemAddress, bool) {
	raw = strings.TrimSpace(raw)
	for _, address := range reservedSystemAddresses {
		if address.Raw == raw {
			return address, true
		}
	}
	return SystemAddress{}, false
}

func SystemAddressByUserFriendly(uf string) (SystemAddress, bool) {
	uf = strings.TrimSpace(uf)
	for _, address := range reservedSystemAddresses {
		if address.UserFriendly == uf {
			return address, true
		}
	}
	return SystemAddress{}, false
}

func IsReservedSystemAddressBytes(bz []byte) bool {
	key, err := addressBytesKey(bz)
	if err != nil {
		return false
	}
	for _, address := range reservedSystemAddresses {
		addressKey, err := addressTextKey(address.Raw)
		if err != nil {
			return false
		}
		if addressKey == key {
			return true
		}
	}
	return false
}

func IsReservedSystemAddressText(text string) bool {
	bz, err := Parse(text)
	if err != nil {
		return false
	}
	return IsReservedSystemAddressBytes(bz)
}

func ValidateNoUserControlledSystemAddresses(userAccounts []string) error {
	for _, account := range userAccounts {
		text := strings.TrimSpace(account)
		if text == "" {
			continue
		}
		bz, err := Parse(text)
		if err != nil {
			return fmt.Errorf("invalid user-controlled account address %q: %w", text, err)
		}
		if IsReservedSystemAddressBytes(bz) {
			return fmt.Errorf("user-controlled account %q uses reserved system address", text)
		}
	}
	return nil
}

func SystemAddressBytesKey(address SystemAddress) (string, error) {
	return addressTextKey(address.Raw)
}

func AddressTextBytesKey(text string) (string, error) {
	return addressTextKey(text)
}

func addressTextKey(text string) (string, error) {
	bz, err := Parse(text)
	if err != nil {
		return "", err
	}
	return addressBytesKey(bz)
}

func addressBytesKey(bz []byte) (string, error) {
	raw, err := ToRawPayload(bz)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(raw), nil
}
