package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/sovereign-l1/l1/app/addressing"
)

const (
	ScopeContractExtension	= "contract_extension"
	ScopeResolverDelegate	= "resolver_delegate"
	ScopeDomainManager	= "domain_manager"
	ScopeModuleACL		= "module_acl"
	ScopeGovernance		= "governance"
	ScopeEmergency		= "emergency"
)

type Permission struct {
	ID		string
	Owner		sdk.AccAddress
	Grantee		sdk.AccAddress
	Scope		string
	Resource	string
	GrantedAtHeight	uint64
	ExpiresAtHeight	uint64
	RevokedAtHeight	uint64
}

type Registry struct {
	permissions map[string]Permission
}

func NewRegistry() *Registry {
	return &Registry{permissions: make(map[string]Permission)}
}

func Grant(permission Permission) (Permission, error) {
	if err := permission.Validate(); err != nil {
		return Permission{}, err
	}
	return permission.Clone(), nil
}

func (r *Registry) Grant(permission Permission) error {
	permission, err := Grant(permission)
	if err != nil {
		return err
	}
	if _, ok := r.permissions[permission.ID]; ok {
		return fmt.Errorf("permission %q already exists", permission.ID)
	}
	r.permissions[permission.ID] = permission
	return nil
}

func (r *Registry) Revoke(id string, owner sdk.AccAddress, height uint64) error {
	permission, ok := r.permissions[id]
	if !ok {
		return fmt.Errorf("permission %q not found", id)
	}
	if string(permission.Owner) != string(owner) {
		return errors.New("permission revoke requires owner")
	}
	if height == 0 {
		return errors.New("permission revoke height must be positive")
	}
	permission.RevokedAtHeight = height
	r.permissions[id] = permission
	return nil
}

func (r *Registry) Check(grantee sdk.AccAddress, scope, resource string, height uint64) bool {
	for _, permission := range r.permissions {
		if permission.Matches(grantee, scope, resource, height) {
			return true
		}
	}
	return false
}

func (r *Registry) List() []Permission {
	out := make([]Permission, 0, len(r.permissions))
	for _, permission := range r.permissions {
		out = append(out, permission.Clone())
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].ID < out[j].ID
	})
	return out
}

func (p Permission) Validate() error {
	if strings.TrimSpace(p.ID) == "" {
		return errors.New("permission id is required")
	}
	if len(p.Owner) == 0 {
		return errors.New("permission owner is required")
	}
	if err := addressing.RejectZeroAddress("permission owner", p.Owner); err != nil {
		return err
	}
	if len(p.Grantee) == 0 {
		return errors.New("permission grantee is required")
	}
	if err := addressing.RejectZeroAddress("permission grantee", p.Grantee); err != nil {
		return err
	}
	if !IsScope(p.Scope) {
		return fmt.Errorf("invalid permission scope %q", p.Scope)
	}
	if strings.TrimSpace(p.Resource) == "" {
		return errors.New("permission resource is required")
	}
	if p.ExpiresAtHeight == 0 {
		return errors.New("permission expiry is required")
	}
	if p.GrantedAtHeight >= p.ExpiresAtHeight {
		return errors.New("permission expiry must be after grant height")
	}
	if p.RevokedAtHeight != 0 && p.RevokedAtHeight < p.GrantedAtHeight {
		return errors.New("permission revocation cannot precede grant height")
	}
	return nil
}

func (p Permission) IsActive(height uint64) bool {
	if height == 0 {
		return false
	}
	if height < p.GrantedAtHeight {
		return false
	}
	if height >= p.ExpiresAtHeight {
		return false
	}
	if p.RevokedAtHeight != 0 && height >= p.RevokedAtHeight {
		return false
	}
	return true
}

func (p Permission) Matches(grantee sdk.AccAddress, scope, resource string, height uint64) bool {
	if string(p.Grantee) != string(grantee) {
		return false
	}
	if p.Scope != scope {
		return false
	}
	if p.Resource != resource {
		return false
	}
	return p.IsActive(height)
}

func (p Permission) Clone() Permission {
	out := p
	out.Owner = append(sdk.AccAddress(nil), p.Owner...)
	out.Grantee = append(sdk.AccAddress(nil), p.Grantee...)
	return out
}

func IsScope(scope string) bool {
	switch scope {
	case ScopeContractExtension,
		ScopeResolverDelegate,
		ScopeDomainManager,
		ScopeModuleACL,
		ScopeGovernance,
		ScopeEmergency:
		return true
	default:
		return false
	}
}

func HasHiddenSuperuserBypass() bool {
	return false
}
