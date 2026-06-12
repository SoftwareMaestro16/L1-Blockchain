package types

import "fmt"

const (
	ModuleName	= "aetracore"
	StoreKey	= ModuleName

	AetherKernelParamsKey	= "aek/params"
	CoreParamsKey		= "core/params"
)

func CoreZoneKey(zoneID ZoneID) (string, error) {
	if err := ValidateZoneID(zoneID); err != nil {
		return "", err
	}
	return fmt.Sprintf("core/zones/%s", zoneID), nil
}

func CoreZoneRootKey(height uint64, zoneID ZoneID) (string, error) {
	if height == 0 {
		return "", fmt.Errorf("aetracore zone root key height must be positive")
	}
	if err := ValidateZoneID(zoneID); err != nil {
		return "", err
	}
	return fmt.Sprintf("core/zone_roots/%020d/%s", height, zoneID), nil
}

func CoreMessageRootKey(height uint64) (string, error) {
	return coreHeightKey("core/message_roots", height)
}

func CoreShardLayoutKey(zoneID ZoneID, layoutEpoch uint64) (string, error) {
	if err := ValidateZoneID(zoneID); err != nil {
		return "", err
	}
	if layoutEpoch == 0 {
		return "", fmt.Errorf("aetracore shard layout key epoch must be positive")
	}
	return fmt.Sprintf("core/shard_layouts/%s/%020d", zoneID, layoutEpoch), nil
}

func CoreRoutingTableKey(routingEpoch uint64) (string, error) {
	if routingEpoch == 0 {
		return "", fmt.Errorf("aetracore routing table key epoch must be positive")
	}
	return fmt.Sprintf("core/routing_table/%020d", routingEpoch), nil
}

func CoreProofRootKey(height uint64, rootType RootType) (string, error) {
	if height == 0 {
		return "", fmt.Errorf("aetracore proof root key height must be positive")
	}
	if err := validateToken("aetracore proof root key type", string(rootType), MaxScopeLength); err != nil {
		return "", err
	}
	return fmt.Sprintf("core/proof_roots/%020d/%s", height, rootType), nil
}

func CoreFinalityKey(height uint64) (string, error) {
	return coreHeightKey("core/finality", height)
}

func ShardMetaKey(zoneID ZoneID, shardID ShardID) (string, error) {
	if err := ValidateZoneID(zoneID); err != nil {
		return "", err
	}
	if err := ValidateShardID(shardID); err != nil {
		return "", err
	}
	return fmt.Sprintf("zones/%s/shards/%s/meta", zoneID, shardID), nil
}

func ShardMetricsKey(zoneID ZoneID, shardID ShardID, height uint64) (string, error) {
	if height == 0 {
		return "", fmt.Errorf("aetracore shard metrics key height must be positive")
	}
	return shardHeightKey("metrics", zoneID, shardID, height)
}

func ShardInboxKey(zoneID ZoneID, shardID ShardID, msgID string) (string, error) {
	return shardObjectKey("inbox", "message id", zoneID, shardID, msgID)
}

func ShardOutboxKey(zoneID ZoneID, shardID ShardID, msgID string) (string, error) {
	return shardObjectKey("outbox", "message id", zoneID, shardID, msgID)
}

func ShardReceiptKey(zoneID ZoneID, shardID ShardID, msgID string) (string, error) {
	return shardObjectKey("receipts", "message id", zoneID, shardID, msgID)
}

func ShardLockKey(zoneID ZoneID, shardID ShardID, objectID string) (string, error) {
	return shardObjectKey("locks", "object id", zoneID, shardID, objectID)
}

func ShardRootKey(zoneID ZoneID, shardID ShardID, height uint64) (string, error) {
	if height == 0 {
		return "", fmt.Errorf("aetracore shard root key height must be positive")
	}
	return shardHeightKey("root", zoneID, shardID, height)
}

func AetherKernelZoneKey(zoneID ZoneID) (string, error) {
	if err := ValidateZoneID(zoneID); err != nil {
		return "", err
	}
	return fmt.Sprintf("aek/zones/%s", zoneID), nil
}

func AetherKernelZoneCommitmentKey(height uint64, zoneID ZoneID) (string, error) {
	if height == 0 {
		return "", fmt.Errorf("aetracore zone commitment key height must be positive")
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
		return "", fmt.Errorf("aetracore routing table key epoch must be positive")
	}
	return fmt.Sprintf("aek/routing/table/%020d", epoch), nil
}

func AetherKernelExportKey(height uint64) (string, error) {
	return aetherKernelHeightKey("aek/export", height)
}

func aetherKernelHeightKey(prefix string, height uint64) (string, error) {
	if height == 0 {
		return "", fmt.Errorf("aetracore %s key height must be positive", prefix)
	}
	return fmt.Sprintf("%s/%020d", prefix, height), nil
}

func coreHeightKey(prefix string, height uint64) (string, error) {
	if height == 0 {
		return "", fmt.Errorf("aetracore %s key height must be positive", prefix)
	}
	return fmt.Sprintf("%s/%020d", prefix, height), nil
}

func shardHeightKey(scope string, zoneID ZoneID, shardID ShardID, height uint64) (string, error) {
	if err := ValidateZoneID(zoneID); err != nil {
		return "", err
	}
	if err := ValidateShardID(shardID); err != nil {
		return "", err
	}
	if err := validateToken("aetracore shard key scope", scope, MaxScopeLength); err != nil {
		return "", err
	}
	return fmt.Sprintf("zones/%s/shards/%s/%s/%020d", zoneID, shardID, scope, height), nil
}

func shardObjectKey(scope string, objectField string, zoneID ZoneID, shardID ShardID, objectID string) (string, error) {
	if err := ValidateZoneID(zoneID); err != nil {
		return "", err
	}
	if err := ValidateShardID(shardID); err != nil {
		return "", err
	}
	if err := validateToken("aetracore shard key scope", scope, MaxScopeLength); err != nil {
		return "", err
	}
	if err := validateToken("aetracore shard "+objectField, objectID, MaxScopeLength); err != nil {
		return "", err
	}
	return fmt.Sprintf("zones/%s/shards/%s/%s/%s", zoneID, shardID, scope, objectID), nil
}
