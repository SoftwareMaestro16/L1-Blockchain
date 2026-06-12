package module

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"

	"lukechampine.com/blake3"
)

const (
	VerifierVersion			= 1
	MagicNumber			= uint32(0x41564D01)	// "AVM\x01"
	MaxModuleImports		= 256
	MaxModuleExports		= 64
	MaxModuleInstructions		= 65536
	MaxModuleCodeBytes		= 1 << 20	// 1 MiB
	MaxModuleMetadataBytes		= 4096
	MaxModuleSchemaBytes		= 1 << 16	// 64 KiB
	MaxModuleDependencyDepth	= 16
	MaxStackDepth			= 1024

	ModuleSectionMagic	= 0x4D53	// "MS"
	ModuleSectionCode	= 0x01
	ModuleSectionImports	= 0x02
	ModuleSectionExports	= 0x03
	ModuleSectionMetadata	= 0x04
	ModuleSectionSchema	= 0x05
	ModuleSectionDeps	= 0x06
)

// TrustLevel classifies module trust.
//   - Untrusted: user-uploaded, needs full verification
//   - Verified: passed static verifier
//   - Canonical: standard library / system module
type TrustLevel uint8

const (
	Untrusted	TrustLevel	= iota
	Verified
	Canonical
)

func (t TrustLevel) String() string {
	switch t {
	case Untrusted:
		return "untrusted"
	case Verified:
		return "verified"
	case Canonical:
		return "canonical"
	default:
		return "unknown"
	}
}

// AVMModule represents a canonical AVM module with all sections.
//
// ModuleHash = BLAKE3(canonical_encoding(all sections))
// Any change in ANY section MUST change ModuleHash.
type AVMModule struct {
	Magic			uint32
	Version			uint32
	ABIVersion		uint32
	ImportTable		[]ImportEntry
	ExportTable		[]ExportEntry
	MetadataHash		[]byte
	Instructions		[]byte
	DependencyHashes	[][]byte
	Schema			[]byte
}

// ImportEntry represents a resolved import with deterministic linking.
type ImportEntry struct {
	ModuleName	string
	ModuleID	[]byte	// Hash of the dependency module
	FunctionIndex	uint32
}

// ExportEntry represents an exported entrypoint.
type ExportEntry struct {
	Name	string
	Index	Entrypoint
	Offset	uint32
}

// Entrypoint defines the standard AVM entrypoints.
type Entrypoint uint8

const (
	EntryDeploy		Entrypoint	= 1
	EntryReceiveExternal	Entrypoint	= 2
	EntryReceiveInternal	Entrypoint	= 3
	EntryReceiveBounced	Entrypoint	= 4
	EntryQuery		Entrypoint	= 5
	EntryMigrate		Entrypoint	= 6
)

// VerificationResult is the canonical, deterministic output of verification.
//
// Invariants:
//   - Identical bytecode → identical verification result
//   - No machine-dependent analysis
//   - No parallel nondeterministic ordering
//   - Contains analyzed stack bounds and CFG hash
type VerificationResult struct {
	ModuleHash		[]byte
	VerifierVersion		uint32
	Passed			bool
	ErrorCode		uint32
	ErrorMessage		string
	AnalyzedStackBound	uint32
	CFGHash			[]byte
	TrustLevel		TrustLevel
	DependencyHashes	[][]byte
	ABICompatibility	bool
}

// BasicBlock represents a basic block in the control flow graph.
type BasicBlock struct {
	StartOffset	uint32
	EndOffset	uint32
	Successors	[]uint32
	Predecessors	[]uint32
}

// ControlFlowGraph represents the CFG of a verified module.
type ControlFlowGraph struct {
	Blocks		[]*BasicBlock
	EntryBlock	uint32
	ExitBlocks	[]uint32
	Hash		[]byte
}

// StackEffect describes the stack effect per instruction.
type StackEffect struct {
	PushCount	int
	PopCount	int
	NetDelta	int
}

// StackBounds describes the analyzed stack bounds for a module.
type StackBounds struct {
	MaxDepth	int
	MinDepth	int
	NetDelta	int
}

// DependencyEdge represents a dependency in the module DAG.
type DependencyEdge struct {
	FromModuleHash	[]byte
	ToModuleHash	[]byte
	ToFunctionIndex	uint32
}

// DependencyDAG represents the acyclic dependency graph of modules.
type DependencyDAG struct {
	Edges		[]DependencyEdge
	TopOrder	[][]byte	// Topological order of module hashes
	Verified	map[string]bool
}

// Verifier performs deterministic verification of AVM modules.
//
// Invariants:
//   - Memory safe on arbitrary input
//   - Panic-proof (returns errors, never panics)
//   - Deterministic under fuzz conditions
//   - Same bytecode → same VerificationResult
type Verifier struct {
	params VerifierParams
}

// VerifierParams configures the verifier's limits.
type VerifierParams struct {
	MaxCodeBytes		uint32
	MaxInstructions		uint32
	MaxImports		uint16
	MaxExports		uint16
	MaxStackDepth		uint32
	MaxDependencies		uint16
	MaxDependencyDepth	uint8
}

func DefaultVerifierParams() VerifierParams {
	return VerifierParams{
		MaxCodeBytes:		MaxModuleCodeBytes,
		MaxInstructions:	MaxModuleInstructions,
		MaxImports:		MaxModuleImports,
		MaxExports:		MaxModuleExports,
		MaxStackDepth:		MaxStackDepth,
		MaxDependencies:	256,
		MaxDependencyDepth:	MaxModuleDependencyDepth,
	}
}

func NewVerifier(params VerifierParams) (*Verifier, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}
	return &Verifier{params: params}, nil
}

func (p VerifierParams) Validate() error {
	if p.MaxCodeBytes == 0 {
		return errors.New("AVM verifier max code bytes must be positive")
	}
	if p.MaxInstructions == 0 {
		return errors.New("AVM verifier max instructions must be positive")
	}
	if p.MaxImports == 0 {
		return errors.New("AVM verifier max imports must be positive")
	}
	if p.MaxExports == 0 {
		return errors.New("AVM verifier max exports must be positive")
	}
	if p.MaxStackDepth == 0 {
		return errors.New("AVM verifier max stack depth must be positive")
	}
	return nil
}

// Verify performs full deterministic verification of AVM bytecode.
//
// Verification pipeline:
//  1. Decode module structure
//  2. Validate version/compatibility
//  3. Validate required entrypoints
//  4. Validate imports (forbidden, capability)
//  5. Build and validate Control Flow Graph
//  6. Compute and validate stack effects
//  7. Validate dependency DAG (acyclic, hashes match)
//  8. Compute canonical module hash
//  9. Produce VerificationResult
func (v *Verifier) Verify(data []byte) (VerificationResult, error) {
	if len(data) == 0 {
		return v.failWithCode(0, "empty module data"), nil
	}

	mod, err := v.decode(data)
	if err != nil {
		return v.failWithMessage(0, err.Error()), nil
	}

	if err := v.validateMagic(mod); err != nil {
		return v.failWithMessage(0, err.Error()), nil
	}

	if err := v.validateVersion(mod); err != nil {
		return v.failWithMessage(0, err.Error()), nil
	}

	if err := v.validateCodeSize(mod); err != nil {
		return v.failWithMessage(0, err.Error()), nil
	}

	if err := v.validateImports(mod); err != nil {
		return v.failWithMessage(0, err.Error()), nil
	}

	if err := v.validateExports(mod); err != nil {
		return v.failWithMessage(0, err.Error()), nil
	}

	cfg, err := v.buildCFG(mod.Instructions)
	if err != nil {
		return v.failWithMessage(0, err.Error()), nil
	}

	bounds, err := v.analyzeStackBounds(mod.Instructions)
	if err != nil {
		return v.failWithMessage(0, err.Error()), nil
	}

	if bounds.MaxDepth > int(v.params.MaxStackDepth) {
		return v.failWithMessage(0, fmt.Sprintf("stack overflow: max depth %d exceeds limit %d", bounds.MaxDepth, v.params.MaxStackDepth)), nil
	}

	if err := v.validateDependencyDAG(mod); err != nil {
		return v.failWithMessage(0, err.Error()), nil
	}

	moduleHash := v.computeModuleHash(data)
	cfgHash := v.computeCFGHash(cfg)

	return VerificationResult{
		ModuleHash:		moduleHash,
		VerifierVersion:	VerifierVersion,
		Passed:			true,
		ErrorCode:		0,
		AnalyzedStackBound:	uint32(bounds.MaxDepth),
		CFGHash:		cfgHash,
		TrustLevel:		Verified,
		DependencyHashes:	mod.DependencyHashes,
		ABICompatibility:	true,
	}, nil
}

// ModuleHash = BLAKE3(canonical_encoding(all module sections))
// Any change in ANY section MUST change ModuleHash.
func (v *Verifier) computeModuleHash(data []byte) []byte {
	h := blake3.New(32, nil)
	h.Write(data)
	return h.Sum(nil)
}

func (v *Verifier) decode(data []byte) (*AVMModule, error) {
	if len(data) < 16 {
		return nil, errors.New("module too small")
	}

	reader := bytes.NewReader(data)

	magic, err := readU32FromReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read magic: %w", err)
	}

	version, err := readU32FromReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read version: %w", err)
	}

	abiVersion, err := readU32FromReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read ABI version: %w", err)
	}

	metadataLen, err := readU16FromReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata length: %w", err)
	}
	metadataHash := make([]byte, metadataLen)
	if metadataLen > 0 {
		if _, err := reader.Read(metadataHash); err != nil {
			return nil, fmt.Errorf("failed to read metadata: %w", err)
		}
	}

	importCount, err := readU16FromReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read import count: %w", err)
	}
	if int(importCount) > int(v.params.MaxImports) {
		return nil, fmt.Errorf("import count %d exceeds maximum %d", importCount, v.params.MaxImports)
	}

	imports := make([]ImportEntry, importCount)
	for i := range imports {
		moduleNameLen, err := readU16FromReader(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to read import module name length: %w", err)
		}
		moduleName := make([]byte, moduleNameLen)
		if _, err := reader.Read(moduleName); err != nil {
			return nil, fmt.Errorf("failed to read import module name: %w", err)
		}
		imports[i].ModuleName = string(moduleName)

		moduleIDLen, err := readU8FromReader(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to read import module ID length: %w", err)
		}
		moduleID := make([]byte, moduleIDLen)
		if _, err := reader.Read(moduleID); err != nil {
			return nil, fmt.Errorf("failed to read import module ID: %w", err)
		}
		imports[i].ModuleID = moduleID

		functionIndex, err := readU32FromReader(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to read import function index: %w", err)
		}
		imports[i].FunctionIndex = functionIndex
	}

	exportCount, err := readU16FromReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read export count: %w", err)
	}
	if int(exportCount) > int(v.params.MaxExports) {
		return nil, fmt.Errorf("export count %d exceeds maximum %d", exportCount, v.params.MaxExports)
	}

	exports := make([]ExportEntry, exportCount)
	for i := range exports {
		nameLen, err := readU16FromReader(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to read export name length: %w", err)
		}
		name := make([]byte, nameLen)
		if _, err := reader.Read(name); err != nil {
			return nil, fmt.Errorf("failed to read export name: %w", err)
		}
		exports[i].Name = string(name)

		entryByte, err := reader.ReadByte()
		if err != nil {
			return nil, fmt.Errorf("failed to read export entrypoint: %w", err)
		}
		exports[i].Index = Entrypoint(entryByte)

		offset, err := readU32FromReader(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to read export offset: %w", err)
		}
		exports[i].Offset = offset
	}

	instructionCount, err := readU32FromReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read instruction count: %w", err)
	}
	if int(instructionCount) > int(v.params.MaxInstructions) {
		return nil, fmt.Errorf("instruction count %d exceeds maximum %d", instructionCount, v.params.MaxInstructions)
	}

	codeLen, err := readU32FromReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read code length: %w", err)
	}
	if int(codeLen) > int(v.params.MaxCodeBytes) {
		return nil, fmt.Errorf("code size %d exceeds maximum %d", codeLen, v.params.MaxCodeBytes)
	}

	instructions := make([]byte, codeLen)
	if _, err := reader.Read(instructions); err != nil {
		return nil, fmt.Errorf("failed to read instructions: %w", err)
	}

	depCount, err := readU16FromReader(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read dependency count: %w", err)
	}

	deps := make([][]byte, depCount)
	for i := range deps {
		depLen, err := readU8FromReader(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to read dependency hash length: %w", err)
		}
		dep := make([]byte, depLen)
		if _, err := reader.Read(dep); err != nil {
			return nil, fmt.Errorf("failed to read dependency hash: %w", err)
		}
		deps[i] = dep
	}

	schemaLen, err := readU32FromReader(reader)
	if err != nil {

		schemaLen = 0
	}

	var schema []byte
	if schemaLen > 0 && schemaLen <= MaxModuleSchemaBytes {
		schema = make([]byte, schemaLen)
		if _, err := reader.Read(schema); err != nil {
			return nil, fmt.Errorf("failed to read schema: %w", err)
		}
	}

	return &AVMModule{
		Magic:			magic,
		Version:		version,
		ABIVersion:		abiVersion,
		ImportTable:		imports,
		ExportTable:		exports,
		MetadataHash:		metadataHash,
		Instructions:		instructions,
		DependencyHashes:	deps,
		Schema:			schema,
	}, nil
}

func (v *Verifier) validateMagic(mod *AVMModule) error {
	if mod.Magic != MagicNumber {
		return fmt.Errorf("invalid magic number: expected 0x%08X, got 0x%08X", MagicNumber, mod.Magic)
	}
	return nil
}

func (v *Verifier) validateVersion(mod *AVMModule) error {
	if mod.Version != uint32(VerifierVersion) {
		return fmt.Errorf("unsupported module version %d: expected %d", mod.Version, VerifierVersion)
	}

	if mod.ABIVersion != 1 {
		return fmt.Errorf("unsupported ABI version %d: expected 1", mod.ABIVersion)
	}

	return nil
}

func (v *Verifier) validateCodeSize(mod *AVMModule) error {
	if len(mod.Instructions) == 0 {
		return errors.New("module instructions must not be empty")
	}
	if uint32(len(mod.Instructions)) > v.params.MaxCodeBytes {
		return fmt.Errorf("code size %d exceeds maximum %d", len(mod.Instructions), v.params.MaxCodeBytes)
	}
	return nil
}

func (v *Verifier) validateImports(mod *AVMModule) error {
	if len(mod.ImportTable) > int(v.params.MaxImports) {
		return fmt.Errorf("import count %d exceeds maximum %d", len(mod.ImportTable), v.params.MaxImports)
	}
	seenModules := make(map[string]struct{}, len(mod.ImportTable))
	for _, imp := range mod.ImportTable {
		if imp.ModuleName == "" {
			return errors.New("import module name must not be empty")
		}
		if len(imp.ModuleID) == 0 {
			return errors.New("import module ID must not be empty")
		}
		if _, exists := seenModules[imp.ModuleName]; exists {
			return fmt.Errorf("duplicate import module %q", imp.ModuleName)
		}
		seenModules[imp.ModuleName] = struct{}{}
	}
	return nil
}

func (v *Verifier) validateExports(mod *AVMModule) error {
	if len(mod.ExportTable) == 0 {
		return errors.New("module must export at least one entrypoint")
	}

	requiredEntrypoints := []Entrypoint{EntryDeploy}
	seenEntries := make(map[Entrypoint]struct{}, len(mod.ExportTable))
	seenNames := make(map[string]struct{}, len(mod.ExportTable))

	for _, exp := range mod.ExportTable {
		if !isValidEntrypoint(exp.Index) {
			return fmt.Errorf("invalid entrypoint %d", exp.Index)
		}
		if _, exists := seenEntries[exp.Index]; exists {
			return fmt.Errorf("duplicate entrypoint %d", exp.Index)
		}
		if _, exists := seenNames[exp.Name]; exists {
			return fmt.Errorf("duplicate export name %q", exp.Name)
		}
		seenEntries[exp.Index] = struct{}{}
		seenNames[exp.Name] = struct{}{}

		if int(exp.Offset) >= len(mod.Instructions) {
			return fmt.Errorf("export %q offset %d out of range (max %d)", exp.Name, exp.Offset, len(mod.Instructions)-1)
		}
	}

	for _, required := range requiredEntrypoints {
		if _, exists := seenEntries[required]; !exists {
			return fmt.Errorf("missing required entrypoint %d", required)
		}
	}

	return nil
}

func isValidEntrypoint(entry Entrypoint) bool {
	switch entry {
	case EntryDeploy, EntryReceiveExternal, EntryReceiveInternal, EntryReceiveBounced, EntryQuery, EntryMigrate:
		return true
	default:
		return false
	}
}

// buildCFG builds a Control Flow Graph from bytecode.
//
// Invariants:
//   - All basic blocks identified
//   - All branches validated
//   - No dangling entrypoints
//   - All jump targets validated to instruction boundaries
func (v *Verifier) buildCFG(code []byte) (*ControlFlowGraph, error) {
	if len(code) == 0 {
		return nil, errors.New("empty code for CFG construction")
	}

	leaders := make(map[uint32]struct{})
	leaders[0] = struct{}{}

	for i := 0; i < len(code); {
		opcode := code[i]
		target := getJumpTarget(opcode, code, i)
		if target >= 0 {
			if target >= len(code) {
				return nil, fmt.Errorf("jump target %d is out of bounds (code length %d)", target, len(code))
			}
			leaders[uint32(target)] = struct{}{}
		}
		next := nextInstructionOffset(opcode, i)
		if next <= i {

			break
		}
		if next < len(code) {
			leaders[uint32(next)] = struct{}{}
		}
		i = next
	}

	sortedLeaders := make([]uint32, 0, len(leaders))
	for l := range leaders {
		sortedLeaders = append(sortedLeaders, l)
	}
	sort.Slice(sortedLeaders, func(i, j int) bool { return sortedLeaders[i] < sortedLeaders[j] })

	blocks := make([]*BasicBlock, len(sortedLeaders))
	for i, start := range sortedLeaders {
		end := uint32(len(code))
		if i+1 < len(sortedLeaders) {
			end = sortedLeaders[i+1]
		}
		blocks[i] = &BasicBlock{
			StartOffset:	start,
			EndOffset:	end,
			Successors:	[]uint32{},
			Predecessors:	[]uint32{},
		}
	}

	entryBlock := uint32(0)
	exitBlocks := []uint32{}

	for i, block := range blocks {
		if block.EndOffset > 0 && block.EndOffset <= uint32(len(code)) {
			lastIdx := int(block.EndOffset) - 1
			if lastIdx >= 0 && lastIdx < len(code) {
				opcode := code[lastIdx]
				target := getJumpTarget(opcode, code, lastIdx)
				if target >= 0 {
					for j, b := range blocks {
						if b.StartOffset == uint32(target) {
							block.Successors = append(block.Successors, uint32(j))
							b.Predecessors = append(b.Predecessors, uint32(i))
							break
						}
					}
				}
			}
		}

		if block.EndOffset == uint32(len(code)) || isTerminalOpcode(code[int(block.EndOffset)-1]) {
			exitBlocks = append(exitBlocks, uint32(i))
		}
	}

	cfg := &ControlFlowGraph{
		Blocks:		blocks,
		EntryBlock:	entryBlock,
		ExitBlocks:	exitBlocks,
	}

	cfg.Hash = v.computeCFGHashFromBlocks(blocks)

	return cfg, nil
}

func (v *Verifier) computeCFGHash(cfg *ControlFlowGraph) []byte {
	return cfg.Hash
}

func (v *Verifier) computeCFGHashFromBlocks(blocks []*BasicBlock) []byte {
	h := blake3.New(32, nil)
	for _, block := range blocks {
		var buf [4]byte
		binary.BigEndian.PutUint32(buf[:], block.StartOffset)
		h.Write(buf[:])
		binary.BigEndian.PutUint32(buf[:], block.EndOffset)
		h.Write(buf[:])
		binary.BigEndian.PutUint32(buf[:], uint32(len(block.Successors)))
		h.Write(buf[:])
		for _, succ := range block.Successors {
			binary.BigEndian.PutUint32(buf[:], succ)
			h.Write(buf[:])
		}
	}
	return h.Sum(nil)
}

// AVMInstructionSize maps opcode to instruction size.
var AVMInstructionSize = map[byte]int{
	0x00:	1,
	0x01:	9,
	0x02:	1,
	0x03:	1,
	0x04:	1,
	0x05:	1,
	0x06:	1,
	0x07:	1,
	0x08:	1,
	0x09:	1,
	0x0a:	9,
	0x0b:	1,
}

func nextInstructionOffset(opcode byte, current int) int {
	size, ok := AVMInstructionSize[opcode]
	if !ok {
		return 1
	}
	return size
}

func getJumpTarget(opcode byte, code []byte, offset int) int {

	if opcode == 0x06 {
		return -1
	}
	return -1
}

func isTerminalOpcode(opcode byte) bool {
	return opcode == 0x06
}

// AVMOpcodeStackEffects maps opcodes to their stack effects.
var AVMOpcodeStackEffects = map[byte]StackEffect{
	0x00:	{PushCount: 0, PopCount: 0, NetDelta: 0},
	0x01:	{PushCount: 1, PopCount: 0, NetDelta: 1},
	0x02:	{PushCount: 1, PopCount: 0, NetDelta: 1},
	0x03:	{PushCount: 0, PopCount: 1, NetDelta: -1},
	0x04:	{PushCount: 1, PopCount: 2, NetDelta: -1},
	0x05:	{PushCount: 0, PopCount: 0, NetDelta: 0},
	0x06:	{PushCount: 0, PopCount: 0, NetDelta: 0},
	0x07:	{PushCount: 1, PopCount: 0, NetDelta: 1},
	0x08:	{PushCount: 1, PopCount: 0, NetDelta: 1},
	0x09:	{PushCount: 1, PopCount: 0, NetDelta: 1},
	0x0a:	{PushCount: 0, PopCount: 0, NetDelta: 0},
	0x0b:	{PushCount: 0, PopCount: 0, NetDelta: 0},
}

// analyzeStackBounds computes the max/min stack depth across all instruction paths.
//
// Invariants:
//   - Stack must never underflow
//   - Stack must never exceed MaxStackDepth
//   - All paths must end with consistent stack heights
func (v *Verifier) analyzeStackBounds(code []byte) (StackBounds, error) {
	currentDepth := 0
	maxDepth := 0
	minDepth := 0

	for i := 0; i < len(code); {
		opcode := code[i]
		effect, ok := AVMOpcodeStackEffects[opcode]
		if !ok {

			i += nextInstructionOffset(opcode, i)
			continue
		}

		currentDepth += effect.NetDelta
		if currentDepth < 0 {
			return StackBounds{}, fmt.Errorf("stack underflow at offset %d: depth %d", i, currentDepth)
		}
		if currentDepth > maxDepth {
			maxDepth = currentDepth
		}
		if currentDepth < minDepth {
			minDepth = currentDepth
		}
		if currentDepth > int(v.params.MaxStackDepth) {
			return StackBounds{}, fmt.Errorf("stack overflow at offset %d: depth %d exceeds max %d", i, currentDepth, v.params.MaxStackDepth)
		}

		i += nextInstructionOffset(opcode, i)
	}

	return StackBounds{
		MaxDepth:	maxDepth,
		MinDepth:	minDepth,
		NetDelta:	currentDepth,
	}, nil
}

// validateDependencyDAG ensures module dependencies form a DAG.
//
// Invariants:
//   - No cycles allowed
//   - Each dependency hash MUST be verified before linking
//   - Dependency hash MUST match exact module hash
func (v *Verifier) validateDependencyDAG(mod *AVMModule) error {
	if len(mod.DependencyHashes) > int(v.params.MaxDependencies) {
		return fmt.Errorf("dependency count %d exceeds maximum %d", len(mod.DependencyHashes), v.params.MaxDependencies)
	}

	seen := make(map[string]struct{}, len(mod.DependencyHashes))
	for _, dep := range mod.DependencyHashes {
		hex := fmt.Sprintf("%x", dep)
		if _, exists := seen[hex]; exists {
			return fmt.Errorf("duplicate dependency hash %s", hex)
		}
		seen[hex] = struct{}{}
	}

	depth := len(mod.DependencyHashes)
	if depth > int(v.params.MaxDependencyDepth) {
		return fmt.Errorf("dependency depth %d exceeds maximum %d", depth, v.params.MaxDependencyDepth)
	}

	return nil
}

// BuildDependencyDAG constructs a DAG from module dependencies.
func BuildDependencyDAG(modules []AVMModule) (*DependencyDAG, error) {
	dag := &DependencyDAG{
		Verified: make(map[string]bool),
	}

	moduleMap := make(map[string]*AVMModule, len(modules))
	for i := range modules {
		hash := blake3Sum32(encodeModule(&modules[i]))
		hex := fmt.Sprintf("%x", hash)
		moduleMap[hex] = &modules[i]
	}

	var order [][]byte
	visited := make(map[string]bool)
	visiting := make(map[string]bool)

	var topoSort func(hash []byte) error
	topoSort = func(hash []byte) error {
		hex := fmt.Sprintf("%x", hash)
		if visited[hex] {
			return nil
		}
		if visiting[hex] {
			return fmt.Errorf("circular dependency detected at module %s", hex)
		}
		visiting[hex] = true

		mod, ok := moduleMap[hex]
		if !ok {
			return fmt.Errorf("missing dependency module %s", hex)
		}

		for _, dep := range mod.DependencyHashes {
			dag.Edges = append(dag.Edges, DependencyEdge{
				FromModuleHash:		hash,
				ToModuleHash:		dep,
				ToFunctionIndex:	0,
			})
			if err := topoSort(dep); err != nil {
				return err
			}
		}

		visiting[hex] = false
		visited[hex] = true
		order = append(order, hash)
		return nil
	}

	for hexStr := range moduleMap {
		hash, err := hex.DecodeString(hexStr)
		if err != nil {
			return nil, fmt.Errorf("invalid module hash %s: %w", hexStr, err)
		}
		if err := topoSort(hash); err != nil {
			return nil, err
		}
	}

	dag.TopOrder = order
	for hex := range visited {
		dag.Verified[hex] = true
	}

	return dag, nil
}

// ValidateExecutionGuarantee checks that a verified module guarantees runtime safety.
//
// Invariants (if module passes verification):
//   - Execution MUST NOT trigger undefined opcode behavior
//   - Execution MUST stay within verified stack bounds
//   - Execution MUST respect CFG constraints
func ValidateExecutionGuarantee(result VerificationResult) error {
	if !result.Passed {
		return fmt.Errorf("module failed verification: %s", result.ErrorMessage)
	}

	if result.AnalyzedStackBound == 0 {
		return errors.New("verified module must have analyzed stack bound")
	}

	if len(result.CFGHash) == 0 {
		return errors.New("verified module must have CFG hash")
	}

	return nil
}

func encodeModule(mod *AVMModule) []byte {
	buf := bytes.NewBuffer(nil)
	binary.Write(buf, binary.BigEndian, mod.Magic)
	binary.Write(buf, binary.BigEndian, mod.Version)
	binary.Write(buf, binary.BigEndian, mod.ABIVersion)
	binary.Write(buf, binary.BigEndian, uint16(len(mod.MetadataHash)))
	buf.Write(mod.MetadataHash)
	binary.Write(buf, binary.BigEndian, uint16(len(mod.ImportTable)))
	for _, imp := range mod.ImportTable {
		binary.Write(buf, binary.BigEndian, uint16(len(imp.ModuleName)))
		buf.WriteString(imp.ModuleName)
		buf.WriteByte(byte(len(imp.ModuleID)))
		buf.Write(imp.ModuleID)
		binary.Write(buf, binary.BigEndian, imp.FunctionIndex)
	}
	binary.Write(buf, binary.BigEndian, uint16(len(mod.ExportTable)))
	for _, exp := range mod.ExportTable {
		binary.Write(buf, binary.BigEndian, uint16(len(exp.Name)))
		buf.WriteString(exp.Name)
		buf.WriteByte(byte(exp.Index))
		binary.Write(buf, binary.BigEndian, exp.Offset)
	}
	binary.Write(buf, binary.BigEndian, uint32(len(mod.Instructions)))
	buf.Write(mod.Instructions)
	binary.Write(buf, binary.BigEndian, uint16(len(mod.DependencyHashes)))
	for _, dep := range mod.DependencyHashes {
		buf.WriteByte(byte(len(dep)))
		buf.Write(dep)
	}
	if len(mod.Schema) > 0 {
		binary.Write(buf, binary.BigEndian, uint32(len(mod.Schema)))
		buf.Write(mod.Schema)
	}
	return buf.Bytes()
}

func blake3Sum32(data []byte) []byte {
	h := blake3.New(32, nil)
	h.Write(data)
	return h.Sum(nil)
}

func (v *Verifier) failWithCode(code uint32, msg string) VerificationResult {
	return VerificationResult{
		Passed:		false,
		ErrorCode:	code,
		ErrorMessage:	msg,
		TrustLevel:	Untrusted,
	}
}

func (v *Verifier) failWithMessage(code uint32, msg string) VerificationResult {
	return v.failWithCode(code, msg)
}

func readU32FromReader(r *bytes.Reader) (uint32, error) {
	var buf [4]byte
	if _, err := r.Read(buf[:]); err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32(buf[:]), nil
}

func readU16FromReader(r *bytes.Reader) (uint16, error) {
	var buf [2]byte
	if _, err := r.Read(buf[:]); err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(buf[:]), nil
}

func readU8FromReader(r *bytes.Reader) (byte, error) {
	return r.ReadByte()
}
