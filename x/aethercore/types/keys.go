package types

import "fmt"

const (
	ModuleName = "aethercore"
	StoreKey   = ModuleName

	AetherKernelParamsKey = "aek/params"
)

func AetherKernelZoneKey(zoneID ZoneID) (string, error) {
	if err := ValidateZoneID(zoneID); err != nil {
		return "", err
	}
	return fmt.Sprintf("aek/zones/%s", zoneID), nil
}

func AetherKernelZoneCommitmentKey(height uint64, zoneID ZoneID) (string, error) {
	if height == 0 {
		return "", fmt.Errorf("aethercore zone commitment key height must be positive")
	}
	if err := ValidateZoneID(zoneID); err != nil {
		return "", err
	}
	return fmt.Sprintf("aek/zone_commitments/%020d/%s", height, zoneID), nil
}

func AetherKernelMessageRootKey(height uint64) (string, error) {
	return aetherKernelHeightKey("aek/messages/root", height)
}

func AetherKernelReceiptsRootKey(height uint64) (string, error) {
	return aetherKernelHeightKey("aek/receipts/root", height)
}

func AetherKernelServicesRootKey(height uint64) (string, error) {
	return aetherKernelHeightKey("aek/services/root", height)
}

func AetherKernelIdentityRootKey(height uint64) (string, error) {
	return aetherKernelHeightKey("aek/identity/root", height)
}

func AetherKernelStorageRootKey(height uint64) (string, error) {
	return aetherKernelHeightKey("aek/storage/root", height)
}

func AetherKernelRoutingTableKey(epoch uint64) (string, error) {
	if epoch == 0 {
		return "", fmt.Errorf("aethercore routing table key epoch must be positive")
	}
	return fmt.Sprintf("aek/routing/table/%020d", epoch), nil
}

func AetherKernelExportKey(height uint64) (string, error) {
	return aetherKernelHeightKey("aek/export", height)
}

func aetherKernelHeightKey(prefix string, height uint64) (string, error) {
	if height == 0 {
		return "", fmt.Errorf("aethercore %s key height must be positive", prefix)
	}
	return fmt.Sprintf("%s/%020d", prefix, height), nil
}
