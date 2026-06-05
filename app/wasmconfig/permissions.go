package wasmconfig

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/sovereign-l1/l1/app/addressing"
)

func CanUpload(actor string, p Policy) error {
	if err := requireEnabled(p); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("wasm upload actor", actor); err != nil {
		return err
	}
	if same, err := sameAddress(actor, p.GovernanceAuthority); err != nil {
		return err
	} else if same {
		return nil
	}
	if p.UploadPermission != UploadPermissionAllowlist {
		return errors.New("wasm upload requires governance authority")
	}
	for _, allowed := range p.UploadAllowlist {
		same, err := sameAddress(actor, allowed)
		if err != nil {
			return err
		}
		if same {
			return nil
		}
	}
	return errors.New("wasm upload actor is not allowlisted")
}

func CanInstantiate(actor, codeOwner string, p Policy) error {
	if err := requireEnabled(p); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("wasm instantiate actor", actor); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("wasm code owner", codeOwner); err != nil {
		return err
	}
	if p.InstantiatePermission == InstantiatePermissionEverybody {
		return nil
	}
	same, err := sameAddress(actor, codeOwner)
	if err != nil {
		return err
	}
	if !same {
		return errors.New("wasm instantiate requires code owner")
	}
	return nil
}

func ValidateInstantiateAddresses(admin, recipient string, p Policy) error {
	if err := requireEnabled(p); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("wasm contract admin", admin); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("wasm instantiate recipient", recipient); err != nil {
		return err
	}
	return nil
}

func CanExecute(actor, contract string, p Policy) error {
	if err := requireEnabled(p); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("wasm execute actor", actor); err != nil {
		return err
	}
	return addressing.ValidateContractAddress("wasm contract address", contract)
}

func CanMigrate(actor, admin string, p Policy) error {
	if err := requireEnabled(p); err != nil {
		return err
	}
	if !p.MigrationsEnabled {
		return errors.New("wasm migrations are disabled by governance")
	}
	if err := addressing.ValidateUserAddress("wasm migrate actor", actor); err != nil {
		return err
	}
	if err := addressing.ValidateUserAddress("wasm contract admin", admin); err != nil {
		return err
	}
	same, err := sameAddress(actor, admin)
	if err != nil {
		return err
	}
	if !same {
		return errors.New("wasm migrate requires contract admin")
	}
	return nil
}

func CanPinCode(actor string, pinnedCount uint32, p Policy) error {
	if err := requireEnabled(p); err != nil {
		return err
	}
	if p.PinnedCodePolicy == PinnedCodePolicyDisabled {
		return errors.New("wasm pinned code is disabled")
	}
	if same, err := sameAddress(actor, p.GovernanceAuthority); err != nil {
		return err
	} else if !same {
		return errors.New("wasm pinned code requires governance authority")
	}
	if pinnedCount >= p.MaxPinnedCodes {
		return fmt.Errorf("wasm pinned code count must be < %d", p.MaxPinnedCodes)
	}
	return nil
}

func sameAddress(left, right string) (bool, error) {
	leftAddr, err := addressing.ParseUserAddress("left address", left)
	if err != nil {
		return false, err
	}
	rightAddr, err := addressing.ParseUserAddress("right address", right)
	if err != nil {
		return false, err
	}
	return bytes.Equal(leftAddr.Bytes(), rightAddr.Bytes()), nil
}
