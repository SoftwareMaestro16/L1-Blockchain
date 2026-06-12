package types

import "strings"

const (
	ModuleName		= "services"
	StoreKey		= ModuleName
	ServiceStorePrefix	= "services/"
)

func IsServiceStoreKey(key string) bool {
	return strings.HasPrefix(key, ServiceStorePrefix) && !strings.Contains(key, "//")
}
