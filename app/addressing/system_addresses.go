package addressing

import (
	"encoding/hex"
	"fmt"
	"strings"
)

const (
	SystemAddressStatusActive	= "active"

	ReservedUserWorkchain	= 4
	ReservedSystemWorkchain	= -7

	SystemAddressAETElectorName		= "AETElector"
	SystemAddressAETConfigName		= "AETConfig"
	SystemAddressAETMintName		= "AETMint"
	SystemAddressAETBurnName		= "AETBurn"
	SystemAddressAETElectorUserFriendly	= "AEAAAQEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEELECTOR"
	SystemAddressAETConfigUserFriendly	= "AEAAAQCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCONFIG"
	SystemAddressAETMintUserFriendly	= "AEAAAQMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMMINT"
	SystemAddressAETBurnUserFriendly	= "AEAAAQBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBURN"
	SystemAddressAETElectorRaw		= "-7:01041041041041041041041041041041041041041041041041041042c4093391"
	SystemAddressAETConfigRaw		= "-7:008208208208208208208208208208208208208208208208208208208e345206"
	SystemAddressAETMintRaw			= "4:030c30c30c30c30c30c30c30c30c30c30c30c30c30c30c30c30c30c30c308353"
	SystemAddressAETBurnRaw			= "4:004104104104104104104104104104104104104104104104104104104105444d"
)

type SystemAddress struct {
	Name			string
	ModuleName		string
	Raw			string
	UserFriendly		string
	Core			bool
	CanHoldFunds		bool
	CanReceiveUserFunds	bool
	CanSendFunds		bool
	Status			string
}

var reservedSystemAddresses = []SystemAddress{
	systemAddress(SystemAddressAETElectorName, "validator-election", SystemAddressAETElectorRaw, SystemAddressAETElectorUserFriendly, true, false, false, false),
	systemAddress(SystemAddressAETConfigName, "config", SystemAddressAETConfigRaw, SystemAddressAETConfigUserFriendly, true, false, false, false),
	systemAddress("AETConstitution", "constitution", "-7:034d34d34d34d34d34d34d34d34d34d34d34d34d34d34d34d34d34d34d34d34d", "AEAAAQNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNNN", true, false, false, false),
	systemAddress("AETSystemRegistry", "system-registry", "-7:0451451451451451451451451451451451451451451451451451451451451451", "AEAAAQRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRRR", true, false, false, false),
	systemAddress("AETValidatorRegistry", "validator-registry", "-7:0555555555555555555555555555555555555555555555555555555555555555", "AEAAAQVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVVV", true, false, false, false),
	systemAddress("AETConfigVoting", "config-voting", "-7:0186186186186186186186186186186186186186186186186186186186186186", "AEAAAQGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGGG", true, false, false, false),

	systemAddress(SystemAddressAETMintName, "mint-authority", SystemAddressAETMintRaw, SystemAddressAETMintUserFriendly, false, false, false, false),
	systemAddress(SystemAddressAETBurnName, "burn", SystemAddressAETBurnRaw, SystemAddressAETBurnUserFriendly, false, false, true, false),
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
		Name:			name,
		ModuleName:		moduleName,
		Raw:			raw,
		UserFriendly:		userFriendly,
		Core:			core,
		CanHoldFunds:		canHoldFunds,
		CanReceiveUserFunds:	canReceiveUserFunds,
		CanSendFunds:		canSendFunds,
		Status:			SystemAddressStatusActive,
	}
}

func AllSystemAddresses() []SystemAddress {
	out := make([]SystemAddress, len(reservedSystemAddresses))
	copy(out, reservedSystemAddresses)
	return out
}

func ValidateReservedSystemAddressCatalog() error {
	return ValidateSystemAddressCatalog(reservedSystemAddresses)
}

func ValidateSystemAddressCatalog(addresses []SystemAddress) error {
	seenNames := map[string]struct{}{}
	seenModules := map[string]struct{}{}
	seenBytes := map[string]string{}
	for _, address := range addresses {
		if strings.TrimSpace(address.Name) == "" {
			return fmt.Errorf("reserved system address name is required")
		}
		if strings.TrimSpace(address.ModuleName) == "" {
			return fmt.Errorf("reserved system address module is required for %s", address.Name)
		}
		if _, found := seenNames[address.Name]; found {
			return fmt.Errorf("duplicate reserved system address name %s", address.Name)
		}
		seenNames[address.Name] = struct{}{}
		if _, found := seenModules[address.ModuleName]; found {
			return fmt.Errorf("duplicate reserved system address module %s", address.ModuleName)
		}
		seenModules[address.ModuleName] = struct{}{}
		rawBytes, err := Parse(address.Raw)
		if err != nil {
			return fmt.Errorf("reserved system address %s raw address invalid: %w", address.Name, err)
		}
		if IsZero(rawBytes) {
			return fmt.Errorf("reserved system address %s must not use zero address", address.Name)
		}
		userBytes, err := Parse(address.UserFriendly)
		if err != nil {
			return fmt.Errorf("reserved system address %s user-friendly address invalid: %w", address.Name, err)
		}
		if IsZero(userBytes) {
			return fmt.Errorf("reserved system address %s user-friendly address must not be zero address", address.Name)
		}
		rawKey, err := addressTextKey(address.Raw)
		if err != nil {
			return err
		}
		userKey, err := addressTextKey(address.UserFriendly)
		if err != nil {
			return err
		}
		if rawKey != userKey {
			return fmt.Errorf("reserved system address %s raw and AE addresses mismatch", address.Name)
		}
		if other, found := seenBytes[rawKey]; found {
			return fmt.Errorf("duplicate reserved system address bytes used by %s and %s", other, address.Name)
		}
		seenBytes[rawKey] = address.Name
		if address.Status != SystemAddressStatusActive {
			return fmt.Errorf("reserved system address %s has invalid status %q", address.Name, address.Status)
		}
	}
	return nil
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

func SystemAddressByBytes(bz []byte) (SystemAddress, bool) {
	key, err := addressBytesKey(bz)
	if err != nil {
		return SystemAddress{}, false
	}
	for _, address := range reservedSystemAddresses {
		addressKey, err := addressTextKey(address.Raw)
		if err != nil {
			return SystemAddress{}, false
		}
		if addressKey == key {
			return address, true
		}
	}
	return SystemAddress{}, false
}

func SystemAddressByText(text string) (SystemAddress, bool) {
	bz, err := Parse(text)
	if err != nil {
		return SystemAddress{}, false
	}
	return SystemAddressByBytes(bz)
}

func IsReservedSystemAddressBytes(bz []byte) bool {
	_, found := SystemAddressByBytes(bz)
	return found
}

func IsReservedSystemAddressText(text string) bool {
	_, found := SystemAddressByText(text)
	return found
}

func ValidateNoUserControlledSystemAddresses(userAccounts []string) error {
	for _, account := range userAccounts {
		if err := ValidateUserSignerAddress(account); err != nil {
			return err
		}
	}
	return nil
}

func ValidateUserSignerAddress(account string) error {
	text := strings.TrimSpace(account)
	if text == "" {
		return nil
	}
	bz, err := Parse(text)
	if err != nil {
		return fmt.Errorf("invalid user-controlled account address %q: %w", text, err)
	}
	if IsZero(bz) {
		return fmt.Errorf("user-controlled account %q must not be zero address", text)
	}
	if IsReservedSystemAddressBytes(bz) {
		return fmt.Errorf("user-controlled account %q uses reserved system address", text)
	}
	return nil
}

func ValidateUserRecipientAddress(account string) error {
	text := strings.TrimSpace(account)
	if text == "" {
		return nil
	}
	bz, err := Parse(text)
	if err != nil {
		return fmt.Errorf("invalid user recipient address %q: %w", text, err)
	}
	if IsZero(bz) {
		return fmt.Errorf("user recipient %q must not be zero address", text)
	}
	address, found := SystemAddressByBytes(bz)
	if found && !address.CanReceiveUserFunds {
		return fmt.Errorf("user recipient %q is reserved system address and cannot receive user funds", text)
	}
	return nil
}

func ValidateNewUserAccountAddress(field, text string) error {
	if err := ValidateUserAddress(field, text); err != nil {
		return err
	}
	if IsReservedSystemAddressText(text) {
		return fmt.Errorf("%s must not use reserved system address", field)
	}
	return nil
}

func ValidateUserAdminAddress(field, text string) error {
	if strings.TrimSpace(text) == "" {
		return nil
	}
	if err := ValidateUserAddress(field, text); err != nil {
		return err
	}
	if IsReservedSystemAddressText(text) {
		return fmt.Errorf("%s must not use reserved system address", field)
	}
	return nil
}

func ValidateTxAuthorityAddress(field, text string) error {
	if err := ValidateAuthorityAddress(field, text); err != nil {
		return err
	}
	if IsReservedSystemAddressText(text) {
		return fmt.Errorf("%s must not use reserved system address", field)
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
