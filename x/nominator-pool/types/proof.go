package types

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/sovereign-l1/l1/app/addressing"
	reputationtypes "github.com/sovereign-l1/l1/x/reputation/types"
)

type StakingProofKind string

const (
	StakingProofDeposit	StakingProofKind	= "deposit"
	StakingProofShare	StakingProofKind	= "share"
	StakingProofAllocation	StakingProofKind	= "allocation"
	StakingProofReward	StakingProofKind	= "reward"
	StakingProofReputation	StakingProofKind	= "reputation"
)

type StakingProofRequest struct {
	Kind		StakingProofKind
	Height		uint64
	PoolID		string
	Account		string
	Epoch		uint64
	AppHash		string
	RootHash	string
}

type StakingProofMetadata struct {
	Kind		StakingProofKind
	Height		uint64
	StoreKey	string
	StateKey	string
	AppHash		string
	RootHash	string
	ProofPath	[]ProofPathMetadata
	MetadataHash	string
	BoundedLookup	bool
}

type ProofPathMetadata struct {
	Step		uint32
	FromRoot	string
	ToRoot		string
	StoreKey	string
	StateKey	string
	ProofKind	string
}

func BuildStakingProofMetadata(req StakingProofRequest) (StakingProofMetadata, error) {
	req.PoolID = strings.TrimSpace(req.PoolID)
	req.Account = strings.TrimSpace(req.Account)
	req.AppHash = strings.TrimSpace(req.AppHash)
	req.RootHash = strings.TrimSpace(req.RootHash)
	if req.Height == 0 {
		return StakingProofMetadata{}, errors.New("staking proof height must be positive")
	}
	if req.AppHash == "" || req.RootHash == "" {
		return StakingProofMetadata{}, errors.New("staking proof app hash and root hash metadata are required")
	}
	storeKey := StoreKey
	var stateKey string
	switch req.Kind {
	case StakingProofDeposit:
		account, err := canonicalUserAccount(req.Account)
		if err != nil {
			return StakingProofMetadata{}, err
		}
		if err := validateID("staking proof pool id", req.PoolID, MaxPoolIDBytesV1); err != nil {
			return StakingProofMetadata{}, err
		}
		stateKey = PoolDepositProofStateKey(req.PoolID, account)
	case StakingProofShare:
		account, err := canonicalUserAccount(req.Account)
		if err != nil {
			return StakingProofMetadata{}, err
		}
		if err := validateID("staking proof pool id", req.PoolID, MaxPoolIDBytesV1); err != nil {
			return StakingProofMetadata{}, err
		}
		stateKey = PoolShareProofStateKey(req.PoolID, account)
	case StakingProofAllocation:
		if err := validateID("staking proof pool id", req.PoolID, MaxPoolIDBytesV1); err != nil {
			return StakingProofMetadata{}, err
		}
		if req.Epoch == 0 {
			return StakingProofMetadata{}, errors.New("staking proof allocation epoch must be positive")
		}
		stateKey = PoolAllocationProofStateKey(req.PoolID, req.Epoch)
	case StakingProofReward:
		account, err := canonicalUserAccount(req.Account)
		if err != nil {
			return StakingProofMetadata{}, err
		}
		if err := validateID("staking proof pool id", req.PoolID, MaxPoolIDBytesV1); err != nil {
			return StakingProofMetadata{}, err
		}
		stateKey = PoolRewardProofStateKey(req.PoolID, account)
	case StakingProofReputation:
		account, err := canonicalUserAccount(req.Account)
		if err != nil {
			return StakingProofMetadata{}, err
		}
		storeKey = reputationtypes.StoreKey
		stateKey = StakeReputationProofStateKey(account)
	default:
		return StakingProofMetadata{}, fmt.Errorf("unsupported staking proof kind %q", req.Kind)
	}
	out := StakingProofMetadata{
		Kind:		req.Kind,
		Height:		req.Height,
		StoreKey:	storeKey,
		StateKey:	stateKey,
		AppHash:	req.AppHash,
		RootHash:	req.RootHash,
		BoundedLookup:	true,
		ProofPath: []ProofPathMetadata{
			{
				Step:		0,
				FromRoot:	req.AppHash,
				ToRoot:		req.RootHash,
				StoreKey:	storeKey,
				StateKey:	"",
				ProofKind:	"app-to-store-root",
			},
			{
				Step:		1,
				FromRoot:	req.RootHash,
				ToRoot:		req.RootHash,
				StoreKey:	storeKey,
				StateKey:	stateKey,
				ProofKind:	"store-key-membership",
			},
		},
	}
	out.MetadataHash = ComputeStakingProofMetadataHash(out)
	return out, out.Validate()
}

func (m StakingProofMetadata) Validate() error {
	if m.Kind == "" {
		return errors.New("staking proof kind is required")
	}
	if m.Height == 0 {
		return errors.New("staking proof height must be positive")
	}
	if strings.TrimSpace(m.StoreKey) == "" || strings.TrimSpace(m.StateKey) == "" {
		return errors.New("staking proof store key and state key are required")
	}
	if strings.TrimSpace(m.AppHash) == "" || strings.TrimSpace(m.RootHash) == "" {
		return errors.New("staking proof app hash and root hash metadata are required")
	}
	if len(m.ProofPath) == 0 {
		return errors.New("staking proof path metadata is required")
	}
	for idx, step := range m.ProofPath {
		if step.Step != uint32(idx) {
			return errors.New("staking proof path steps must be sequential")
		}
		if strings.TrimSpace(step.ProofKind) == "" {
			return errors.New("staking proof path kind is required")
		}
	}
	if !m.BoundedLookup {
		return errors.New("staking proof metadata must use bounded lookup")
	}
	if m.MetadataHash != ComputeStakingProofMetadataHash(m) {
		return errors.New("staking proof metadata hash mismatch")
	}
	return nil
}

func PoolDepositProofStateKey(poolID, account string) string {
	return "nominator-pool/pools/" + poolID + "/deposits/" + account
}

func PoolShareProofStateKey(poolID, account string) string {
	return "nominator-pool/pools/" + poolID + "/shares/" + account
}

func PoolAllocationProofStateKey(poolID string, epoch uint64) string {
	return fmt.Sprintf("nominator-pool/pools/%s/allocations/%020d", poolID, epoch)
}

func PoolRewardProofStateKey(poolID, account string) string {
	return "nominator-pool/pools/" + poolID + "/rewards/" + account
}

func StakeReputationProofStateKey(account string) string {
	return "reputation/stake/" + account
}

func ComputeStakingProofMetadataHash(m StakingProofMetadata) string {
	parts := []string{
		"staking-proof-metadata-v1",
		string(m.Kind),
		fmt.Sprint(m.Height),
		m.StoreKey,
		m.StateKey,
		m.AppHash,
		m.RootHash,
		fmt.Sprint(m.BoundedLookup),
	}
	for _, step := range m.ProofPath {
		parts = append(parts,
			fmt.Sprint(step.Step),
			step.FromRoot,
			step.ToRoot,
			step.StoreKey,
			step.StateKey,
			step.ProofKind,
		)
	}
	return hashProofMetadataParts(parts...)
}

func canonicalUserAccount(account string) (string, error) {
	addr, err := addressing.ParseUserAddress("staking proof account", account)
	if err != nil {
		return "", err
	}
	return addressing.FormatAccAddress(addr), nil
}

func hashProofMetadataParts(parts ...string) string {
	h := sha256.New()
	for _, part := range parts {
		_, _ = h.Write([]byte(part))
		_, _ = h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))
}
