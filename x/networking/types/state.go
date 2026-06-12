package types

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type NetworkingState struct {
	LayerStack		[]LayerSpec
	Adapter			AetherNetworkingAdapter
	ChannelPolicies		[]ChannelPolicy
	NodeRecords		[]NodeRecord
	RoleCommitments		[]RoleCommitment
	Sessions		[]SessionChannel
	IdentityTransitions	[]IdentityTransitionRecord
	OverlayDescriptors	[]OverlayDescriptor
}

func EmptyState() NetworkingState {
	return NetworkingState{
		LayerStack:		DefaultLayerStack(),
		Adapter:		DefaultAetherNetworkingAdapter(),
		ChannelPolicies:	DefaultChannelPolicies(),
		NodeRecords:		[]NodeRecord{},
		RoleCommitments:	[]RoleCommitment{},
		Sessions:		[]SessionChannel{},
		IdentityTransitions:	[]IdentityTransitionRecord{},
		OverlayDescriptors:	DefaultOverlayDescriptors(),
	}
}

func RegisterNodeRecord(state NetworkingState, record NodeRecord, networkSalt []byte, currentHeight uint64) (NetworkingState, error) {
	state = state.Export()
	if err := state.Validate(); err != nil {
		return NetworkingState{}, err
	}
	record = NormalizeNodeRecord(record)
	if err := record.Validate(networkSalt, currentHeight); err != nil {
		return NetworkingState{}, err
	}
	next := state.Clone()
	replaced := false
	for i, existing := range next.NodeRecords {
		if existing.NodeID == record.NodeID {
			next.NodeRecords[i] = record
			replaced = true
			break
		}
	}
	if !replaced {
		next.NodeRecords = append(next.NodeRecords, record)
	}
	sortNodeRecords(next.NodeRecords)
	return next, next.Validate()
}

func ApplyIdentityTransition(state NetworkingState, transition IdentityTransitionRecord, newRecord NodeRecord, networkSalt []byte, currentHeight uint64) (NetworkingState, error) {
	state = state.Export()
	if err := state.Validate(); err != nil {
		return NetworkingState{}, err
	}
	if currentHeight == 0 {
		return NetworkingState{}, errors.New("networking current height must be positive")
	}
	transition = NormalizeIdentityTransition(transition)
	newRecord = NormalizeNodeRecord(newRecord)

	var oldRecord NodeRecord
	foundOld := false
	for _, record := range state.NodeRecords {
		record = NormalizeNodeRecord(record)
		if record.NodeID == transition.FromNodeID {
			oldRecord = record
			foundOld = true
			break
		}
	}
	if !foundOld {
		return NetworkingState{}, errors.New("networking identity transition old node is not registered")
	}
	if containsNode(state.NodeRecords, transition.ToNodeID) {
		return NetworkingState{}, errors.New("networking identity transition target node already exists")
	}
	if err := ValidateIdentityTransition(oldRecord, newRecord, transition, networkSalt, currentHeight); err != nil {
		return NetworkingState{}, err
	}

	next := state.Clone()
	next.NodeRecords = next.NodeRecords[:0]
	for _, record := range state.NodeRecords {
		record = NormalizeNodeRecord(record)
		if record.NodeID == transition.FromNodeID {
			next.NodeRecords = append(next.NodeRecords, newRecord)
			continue
		}
		next.NodeRecords = append(next.NodeRecords, record)
	}
	next.RoleCommitments = pruneRoleCommitmentsForNode(next.RoleCommitments, transition.FromNodeID)
	next.Sessions = pruneSessionsForNode(next.Sessions, transition.FromNodeID)
	next.IdentityTransitions = append(next.IdentityTransitions, transition)
	sortNodeRecords(next.NodeRecords)
	sortRoleCommitments(next.RoleCommitments)
	sortSessions(next.Sessions)
	sortIdentityTransitions(next.IdentityTransitions)
	return next, next.Validate()
}

func OpenSession(state NetworkingState, session SessionChannel, currentHeight uint64) (NetworkingState, error) {
	state = state.Export()
	if err := state.Validate(); err != nil {
		return NetworkingState{}, err
	}
	if currentHeight == 0 {
		return NetworkingState{}, errors.New("networking current height must be positive")
	}
	if err := session.Validate(); err != nil {
		return NetworkingState{}, err
	}
	if currentHeight > session.ExpiresHeight {
		return NetworkingState{}, errors.New("networking session is expired")
	}
	if !state.hasNode(session.LocalNodeID) || !state.hasNode(session.RemoteNodeID) {
		return NetworkingState{}, errors.New("networking session requires registered endpoints")
	}
	for _, existing := range state.Sessions {
		if existing.SessionID == session.SessionID {
			return NetworkingState{}, errors.New("networking session already exists")
		}
	}
	next := state.Clone()
	next.Sessions = append(next.Sessions, session)
	sortSessions(next.Sessions)
	return next, next.Validate()
}

func PruneExpired(state NetworkingState, currentHeight uint64) (NetworkingState, error) {
	state = state.Export()
	if err := state.Validate(); err != nil {
		return NetworkingState{}, err
	}
	next := NetworkingState{
		LayerStack:		cloneLayerStack(state.LayerStack),
		Adapter:		cloneAdapter(state.Adapter),
		ChannelPolicies:	cloneChannelPolicies(state.ChannelPolicies),
		IdentityTransitions:	[]IdentityTransitionRecord{},
		OverlayDescriptors:	[]OverlayDescriptor{},
	}
	for _, record := range state.NodeRecords {
		if currentHeight == 0 || currentHeight <= record.ExpiresHeight {
			next.NodeRecords = append(next.NodeRecords, record)
		}
	}
	for _, session := range state.Sessions {
		if currentHeight == 0 || currentHeight <= session.ExpiresHeight {
			if containsNode(next.NodeRecords, session.LocalNodeID) && containsNode(next.NodeRecords, session.RemoteNodeID) {
				next.Sessions = append(next.Sessions, session)
			}
		}
	}
	for _, commitment := range state.RoleCommitments {
		if currentHeight == 0 || currentHeight <= commitment.ExpiresHeight {
			if containsNode(next.NodeRecords, commitment.NodeID) {
				next.RoleCommitments = append(next.RoleCommitments, commitment)
			}
		}
	}
	for _, transition := range state.IdentityTransitions {
		if currentHeight == 0 || currentHeight <= transition.ExpiresHeight {
			next.IdentityTransitions = append(next.IdentityTransitions, transition)
		}
	}
	for _, desc := range state.OverlayDescriptors {
		if currentHeight == 0 || desc.ExpiresHeight == 0 || currentHeight <= desc.ExpiresHeight {
			next.OverlayDescriptors = append(next.OverlayDescriptors, desc)
		}
	}
	sortNodeRecords(next.NodeRecords)
	sortRoleCommitments(next.RoleCommitments)
	sortSessions(next.Sessions)
	sortIdentityTransitions(next.IdentityTransitions)
	sortOverlayDescriptors(next.OverlayDescriptors)
	return next, next.Validate()
}

func ImportState(state NetworkingState) (NetworkingState, error) {
	state = state.Export()
	if err := state.Validate(); err != nil {
		return NetworkingState{}, err
	}
	return state, nil
}

func (s NetworkingState) Export() NetworkingState {
	out := s.Clone()
	if len(out.LayerStack) == 0 {
		out.LayerStack = DefaultLayerStack()
	}
	if out.Adapter.BaselineTransport == "" {
		out.Adapter = DefaultAetherNetworkingAdapter()
	}
	if len(out.ChannelPolicies) == 0 {
		out.ChannelPolicies = DefaultChannelPolicies()
	}
	if len(out.OverlayDescriptors) == 0 {
		out.OverlayDescriptors = DefaultOverlayDescriptors()
	}
	sortChannelPolicies(out.ChannelPolicies)
	sortNodeRecords(out.NodeRecords)
	sortRoleCommitments(out.RoleCommitments)
	sortSessions(out.Sessions)
	sortIdentityTransitions(out.IdentityTransitions)
	sortOverlayDescriptors(out.OverlayDescriptors)
	return out
}

func (s NetworkingState) Clone() NetworkingState {
	out := NetworkingState{
		LayerStack:		cloneLayerStack(s.LayerStack),
		Adapter:		cloneAdapter(s.Adapter),
		ChannelPolicies:	cloneChannelPolicies(s.ChannelPolicies),
		NodeRecords:		make([]NodeRecord, len(s.NodeRecords)),
		RoleCommitments:	cloneRoleCommitments(s.RoleCommitments),
		Sessions:		make([]SessionChannel, len(s.Sessions)),
		IdentityTransitions:	cloneIdentityTransitions(s.IdentityTransitions),
		OverlayDescriptors:	cloneOverlayDescriptors(s.OverlayDescriptors),
	}
	for i, record := range s.NodeRecords {
		out.NodeRecords[i] = NormalizeNodeRecord(record)
	}
	for i, session := range s.Sessions {
		out.Sessions[i] = cloneSession(session)
	}
	return out
}

func (s NetworkingState) Validate() error {
	if err := ValidateLayerStack(s.LayerStack); err != nil {
		return err
	}
	if err := ValidateAetherNetworkingAdapter(s.Adapter); err != nil {
		return err
	}
	if err := ValidateChannelPolicies(s.ChannelPolicies); err != nil {
		return err
	}
	if err := validateNodeRecords(s.NodeRecords); err != nil {
		return err
	}
	if err := validateRoleCommitments(s.NodeRecords, s.RoleCommitments); err != nil {
		return err
	}
	if err := validateSessions(s.NodeRecords, s.Sessions); err != nil {
		return err
	}
	if err := validateIdentityTransitions(s.IdentityTransitions); err != nil {
		return err
	}
	return validateOverlayDescriptors(s.OverlayDescriptors, 0)
}

func (s NetworkingState) hasNode(nodeID string) bool {
	return containsNode(s.NodeRecords, nodeID)
}

func validateNodeRecords(records []NodeRecord) error {
	seen := make(map[string]struct{}, len(records))
	var previous string
	for i, record := range records {
		normalized := NormalizeNodeRecord(record)
		if err := normalized.ValidateBasic(); err != nil {
			return err
		}
		if _, found := seen[normalized.NodeID]; found {
			return errors.New("networking duplicate node record")
		}
		seen[normalized.NodeID] = struct{}{}
		if i > 0 && previous >= normalized.NodeID {
			return errors.New("networking node records must be sorted canonically")
		}
		previous = normalized.NodeID
	}
	return nil
}

func validateSessions(records []NodeRecord, sessions []SessionChannel) error {
	seen := make(map[string]struct{}, len(sessions))
	var previous string
	for i, session := range sessions {
		session.SessionID = normalizeHashText(session.SessionID)
		if err := session.Validate(); err != nil {
			return err
		}
		if !containsNode(records, session.LocalNodeID) || !containsNode(records, session.RemoteNodeID) {
			return errors.New("networking session references unknown node")
		}
		if _, found := seen[session.SessionID]; found {
			return errors.New("networking duplicate session")
		}
		seen[session.SessionID] = struct{}{}
		if i > 0 && previous >= session.SessionID {
			return errors.New("networking sessions must be sorted canonically")
		}
		previous = session.SessionID
	}
	return nil
}

func containsNode(records []NodeRecord, nodeID string) bool {
	needle := normalizeHashText(nodeID)
	for _, record := range records {
		if NormalizeNodeRecord(record).NodeID == needle {
			return true
		}
	}
	return false
}

func cloneChannelPolicies(policies []ChannelPolicy) []ChannelPolicy {
	out := make([]ChannelPolicy, len(policies))
	copy(out, policies)
	return out
}

func cloneSession(session SessionChannel) SessionChannel {
	session.LocalNodeID = normalizeHashText(session.LocalNodeID)
	session.RemoteNodeID = normalizeHashText(session.RemoteNodeID)
	session.SessionID = normalizeHashText(session.SessionID)
	session.SessionKeys = cloneSessionKeySet(session.SessionKeys)
	session.ProtocolVersions = append([]string(nil), session.ProtocolVersions...)
	sortStrings(session.ProtocolVersions)
	session.Streams = append([]StreamSpec(nil), session.Streams...)
	sort.SliceStable(session.Streams, func(i, j int) bool {
		if session.Streams[i].Priority != session.Streams[j].Priority {
			return session.Streams[i].Priority < session.Streams[j].Priority
		}
		leftRank := channelSortRank(session.Streams[i].Channel)
		rightRank := channelSortRank(session.Streams[j].Channel)
		if leftRank != rightRank {
			return leftRank < rightRank
		}
		return session.Streams[i].StreamID < session.Streams[j].StreamID
	})
	return session
}

func cloneSessionKeySet(keys SessionKeySet) SessionKeySet {
	keys.KeyID = normalizeHashText(keys.KeyID)
	keys.LocalEphemeralPubKey = cloneBytes(keys.LocalEphemeralPubKey)
	keys.RemoteEphemeralPubKey = cloneBytes(keys.RemoteEphemeralPubKey)
	keys.TranscriptHash = normalizeHashText(keys.TranscriptHash)
	keys.SecretCommitmentHash = normalizeHashText(keys.SecretCommitmentHash)
	return keys
}

func cloneIdentityTransitions(transitions []IdentityTransitionRecord) []IdentityTransitionRecord {
	out := make([]IdentityTransitionRecord, len(transitions))
	for i, transition := range transitions {
		out[i] = NormalizeIdentityTransition(transition)
	}
	return out
}

func sortChannelPolicies(policies []ChannelPolicy) {
	sort.SliceStable(policies, func(i, j int) bool {
		if policies[i].Priority != policies[j].Priority {
			return policies[i].Priority < policies[j].Priority
		}
		return channelSortRank(policies[i].Channel) < channelSortRank(policies[j].Channel)
	})
}

func sortNodeRecords(records []NodeRecord) {
	sort.SliceStable(records, func(i, j int) bool {
		return NormalizeNodeRecord(records[i]).NodeID < NormalizeNodeRecord(records[j]).NodeID
	})
}

func sortRoleCommitments(commitments []RoleCommitment) {
	sort.SliceStable(commitments, func(i, j int) bool {
		left := NormalizeRoleCommitment(commitments[i])
		right := NormalizeRoleCommitment(commitments[j])
		if left.NodeID != right.NodeID {
			return left.NodeID < right.NodeID
		}
		return left.Role < right.Role
	})
}

func sortSessions(sessions []SessionChannel) {
	sort.SliceStable(sessions, func(i, j int) bool {
		return sessions[i].SessionID < sessions[j].SessionID
	})
}

func sortIdentityTransitions(transitions []IdentityTransitionRecord) {
	sort.SliceStable(transitions, func(i, j int) bool {
		return NormalizeIdentityTransition(transitions[i]).TransitionID < NormalizeIdentityTransition(transitions[j]).TransitionID
	})
}

func validateIdentityTransitions(transitions []IdentityTransitionRecord) error {
	seen := make(map[string]struct{}, len(transitions))
	var previous string
	for i, transition := range transitions {
		normalized := NormalizeIdentityTransition(transition)
		if err := normalized.ValidateBasic(); err != nil {
			return err
		}
		if _, found := seen[normalized.TransitionID]; found {
			return errors.New("networking duplicate identity transition")
		}
		seen[normalized.TransitionID] = struct{}{}
		if i > 0 && previous >= normalized.TransitionID {
			return errors.New("networking identity transitions must be sorted canonically")
		}
		previous = normalized.TransitionID
	}
	return nil
}

func pruneRoleCommitmentsForNode(commitments []RoleCommitment, nodeID string) []RoleCommitment {
	needle := normalizeHashText(nodeID)
	out := make([]RoleCommitment, 0, len(commitments))
	for _, commitment := range commitments {
		if NormalizeRoleCommitment(commitment).NodeID != needle {
			out = append(out, commitment)
		}
	}
	return out
}

func pruneSessionsForNode(sessions []SessionChannel, nodeID string) []SessionChannel {
	needle := normalizeHashText(nodeID)
	out := make([]SessionChannel, 0, len(sessions))
	for _, session := range sessions {
		if normalizeHashText(session.LocalNodeID) != needle && normalizeHashText(session.RemoteNodeID) != needle {
			out = append(out, session)
		}
	}
	return out
}

func normalizeHashText(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func (s NetworkingState) DebugString() string {
	return fmt.Sprintf("networking nodes=%d sessions=%d channels=%d overlays=%d", len(s.NodeRecords), len(s.Sessions), len(s.ChannelPolicies), len(s.OverlayDescriptors))
}
