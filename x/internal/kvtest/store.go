package kvtest

import (
	"context"
	"errors"

	corestore "cosmossdk.io/core/store"
	dbm "github.com/cosmos/cosmos-db"
)

type StoreService struct {
	store *Store
}

func NewStoreService() *StoreService {
	return &StoreService{store: &Store{values: map[string][]byte{}}}
}

func (s *StoreService) OpenKVStore(context.Context) corestore.KVStore {
	return s.store
}

func (s *StoreService) RawStore() *Store {
	return s.store
}

type Store struct {
	values map[string][]byte
}

func (s *Store) Get(key []byte) ([]byte, error) {
	if key == nil {
		return nil, errors.New("nil key")
	}
	value, found := s.values[string(key)]
	if !found {
		return nil, nil
	}
	return append([]byte(nil), value...), nil
}

func (s *Store) Has(key []byte) (bool, error) {
	if key == nil {
		return false, errors.New("nil key")
	}
	_, found := s.values[string(key)]
	return found, nil
}

func (s *Store) Set(key, value []byte) error {
	if key == nil || value == nil {
		return errors.New("nil key or value")
	}
	s.values[string(key)] = append([]byte(nil), value...)
	return nil
}

func (s *Store) Delete(key []byte) error {
	if key == nil {
		return errors.New("nil key")
	}
	delete(s.values, string(key))
	return nil
}

func (s *Store) Iterator(_, _ []byte) (dbm.Iterator, error) {
	return nil, errors.New("iterator not implemented")
}

func (s *Store) ReverseIterator(_, _ []byte) (dbm.Iterator, error) {
	return nil, errors.New("reverse iterator not implemented")
}
