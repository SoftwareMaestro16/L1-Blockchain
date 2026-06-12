package types

import (
	"errors"
	"fmt"
	"strings"
)

type RoleCommitment struct {
	NodeID		string
	Role		NodeRole
	Bonded		bool
	CommitmentHash	string
	ExpiresHeight	uint64
}

type RoleScope struct {
	NodeID			string
	Role			NodeRole
	Advertised		bool
	Committed		bool
	ConsensusCritical	bool
}

func IsConsensusCriticalRole(role NodeRole, bondedCommitted bool) bool {
	switch role {
	case NodeRoleValidator:
		return true
	case NodeRoleService, NodeRoleRouting, NodeRoleStorageProvider:
		return bondedCommitted
	default:
		return false
	}
}

func RegisterRoleCommitment(state NetworkingState, commitment RoleCommitment, currentHeight uint64) (NetworkingState, error) {
	state = state.Export()
	if err := state.Validate(); err != nil {
		return NetworkingState{}, err
	}
	commitment = NormalizeRoleCommitment(commitment)
	if err := commitment.Validate(state.NodeRecords, currentHeight); err != nil {
		return NetworkingState{}, err
	}
	next := state.Clone()
	replaced := false
	for i, existing := range next.RoleCommitments {
		if existing.NodeID == commitment.NodeID && existing.Role == commitment.Role {
			next.RoleCommitments[i] = commitment
			replaced = true
			break
		}
	}
	if !replaced {
		next.RoleCommitments = append(next.RoleCommitments, commitment)
	}
	sortRoleCommitments(next.RoleCommitments)
	return next, next.Validate()
}

func RoleScopes(record NodeRecord, commitments []RoleCommitment, currentHeight uint64) ([]RoleScope, error) {
	record = NormalizeNodeRecord(record)
	if err := record.ValidateBasic(); err != nil {
		return nil, err
	}
	out := make([]RoleScope, 0, len(record.Roles))
	for _, role := range record.Roles {
		committed := hasActiveRoleCommitment(record.NodeID, role, commitments, currentHeight)
		out = append(out, RoleScope{
			NodeID:			record.NodeID,
			Role:			role,
			Advertised:		true,
			Committed:		committed,
			ConsensusCritical:	IsConsensusCriticalRole(role, committed),
		})
	}
	return out, nil
}

func NormalizeRoleCommitment(commitment RoleCommitment) RoleCommitment {
	commitment.NodeID = normalizeHashText(commitment.NodeID)
	commitment.Role = NodeRole(strings.ToUpper(strings.TrimSpace(string(commitment.Role))))
	commitment.CommitmentHash = normalizeHashText(commitment.CommitmentHash)
	return commitment
}

func (c RoleCommitment) Validate(records []NodeRecord, currentHeight uint64) error {
	commitment := NormalizeRoleCommitment(c)
	if err := ValidateHash("networking role commitment node id", commitment.NodeID); err != nil {
		return err
	}
	if !IsNodeRole(commitment.Role) {
		return fmt.Errorf("unknown networking committed role %q", commitment.Role)
	}
	if commitment.Role == NodeRoleValidator {
		return errors.New("networking validator role is consensus-critical through validator identity, not role commitment")
	}
	if !commitment.Bonded {
		return errors.New("networking consensus role commitment must be bonded")
	}
	if err := ValidateHash("networking role commitment hash", commitment.CommitmentHash); err != nil {
		return err
	}
	if commitment.ExpiresHeight == 0 {
		return errors.New("networking role commitment expires height must be positive")
	}
	if currentHeight > 0 && currentHeight > commitment.ExpiresHeight {
		return errors.New("networking role commitment is expired")
	}
	record, found := findNodeRecord(records, commitment.NodeID)
	if !found {
		return errors.New("networking role commitment references unknown node")
	}
	if !hasRole(record.Roles, commitment.Role) {
		return errors.New("networking role commitment must reference an advertised role")
	}
	if commitment.ExpiresHeight > record.ExpiresHeight {
		return errors.New("networking role commitment cannot outlive node record")
	}
	return nil
}

func validateRoleCommitments(records []NodeRecord, commitments []RoleCommitment) error {
	seen := make(map[string]struct{}, len(commitments))
	var previous string
	for i, commitment := range commitments {
		commitment = NormalizeRoleCommitment(commitment)
		if err := commitment.Validate(records, 0); err != nil {
			return err
		}
		key := commitment.NodeID + "/" + string(commitment.Role)
		if _, found := seen[key]; found {
			return errors.New("networking duplicate role commitment")
		}
		seen[key] = struct{}{}
		if i > 0 && previous >= key {
			return errors.New("networking role commitments must be sorted canonically")
		}
		previous = key
	}
	return nil
}

func hasActiveRoleCommitment(nodeID string, role NodeRole, commitments []RoleCommitment, currentHeight uint64) bool {
	nodeID = normalizeHashText(nodeID)
	role = NodeRole(strings.ToUpper(strings.TrimSpace(string(role))))
	for _, commitment := range commitments {
		commitment = NormalizeRoleCommitment(commitment)
		if commitment.NodeID != nodeID || commitment.Role != role || !commitment.Bonded {
			continue
		}
		if currentHeight > 0 && currentHeight > commitment.ExpiresHeight {
			continue
		}
		return true
	}
	return false
}

func findNodeRecord(records []NodeRecord, nodeID string) (NodeRecord, bool) {
	needle := normalizeHashText(nodeID)
	for _, record := range records {
		normalized := NormalizeNodeRecord(record)
		if normalized.NodeID == needle {
			return normalized, true
		}
	}
	return NodeRecord{}, false
}

func cloneRoleCommitments(commitments []RoleCommitment) []RoleCommitment {
	out := make([]RoleCommitment, len(commitments))
	for i, commitment := range commitments {
		out[i] = NormalizeRoleCommitment(commitment)
	}
	return out
}
