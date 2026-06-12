package types

import (
	"errors"
	"fmt"
	"sort"

	"github.com/sovereign-l1/l1/x/internal/prototype"
)

type GenesisState struct {
	Version		uint64
	Accounts	[]Account
}

func DefaultGenesis() GenesisState {
	return GenesisState{Version: prototype.CurrentGenesisVersion}
}

func (gs GenesisState) Validate() error {
	if gs.Version != prototype.CurrentGenesisVersion {
		return fmt.Errorf("native account unsupported genesis version %d", gs.Version)
	}
	seenUser := make(map[string]struct{}, len(gs.Accounts))
	seenRaw := make(map[string]struct{}, len(gs.Accounts))
	seenNumber := make(map[uint64]struct{}, len(gs.Accounts))
	for _, account := range gs.Accounts {
		if err := ValidateAccountInvariant(account); err != nil {
			return err
		}
		if _, found := seenUser[account.AddressUser]; found {
			return fmt.Errorf("duplicate native account user address %s", account.AddressUser)
		}
		seenUser[account.AddressUser] = struct{}{}
		if _, found := seenRaw[account.AddressRaw]; found {
			return fmt.Errorf("duplicate native account raw address %s", account.AddressRaw)
		}
		seenRaw[account.AddressRaw] = struct{}{}
		if _, found := seenNumber[account.AccountNumber]; found {
			return fmt.Errorf("duplicate native account number %d", account.AccountNumber)
		}
		seenNumber[account.AccountNumber] = struct{}{}
	}
	return nil
}

type AccountReader interface {
	AccountsAfter(cursor string, limit uint64) ([]Account, bool, error)
}

type AccountWriter interface {
	SetAccount(account Account) error
}

type AccountStore interface {
	AccountReader
	AccountWriter
	AccountByUser(userAddress string) (Account, bool, error)
}

func ExportGenesis(store AccountReader) (GenesisState, error) {
	if store == nil {
		return GenesisState{}, errors.New("native account export store is required")
	}
	accounts, hasMore, err := store.AccountsAfter("", 0)
	if err != nil {
		return GenesisState{}, err
	}
	if hasMore {
		return GenesisState{}, errors.New("native account export requires an unbounded explicit export path, not paginated normal block execution")
	}
	gs := GenesisState{Version: prototype.CurrentGenesisVersion, Accounts: SortAccounts(accounts)}
	if err := gs.Validate(); err != nil {
		return GenesisState{}, err
	}
	return cloneGenesis(gs), nil
}

func ImportGenesis(store AccountWriter, gs GenesisState) error {
	if store == nil {
		return errors.New("native account import store is required")
	}
	gs = cloneGenesis(gs)
	if err := gs.Validate(); err != nil {
		return err
	}
	for _, account := range gs.Accounts {
		if err := store.SetAccount(account); err != nil {
			return err
		}
	}
	return nil
}

func SortAccounts(accounts []Account) []Account {
	out := make([]Account, 0, len(accounts))
	for _, account := range accounts {
		out = append(out, normalizeAccount(account))
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].AddressUser < out[j].AddressUser
	})
	return out
}

func cloneGenesis(gs GenesisState) GenesisState {
	gs.Accounts = SortAccounts(gs.Accounts)
	return gs
}
