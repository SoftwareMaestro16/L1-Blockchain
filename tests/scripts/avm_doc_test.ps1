param(
  [string]$Doc = "docs\architecture\avm.md",
  [string]$Direction = "docs\architecture\vm-direction.md",
  [string]$Boundaries = "docs\module-boundaries.md",
  [string]$Code = "x\aetravm\avm\avm.go",
  [string]$Tests = "x\aetravm\avm\avm_test.go",
  [string]$FuzzTests = "x\aetravm\avm\fuzz_test.go"
)

$ErrorActionPreference = "Stop"

$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))

function Resolve-RepoPath {
  param([string]$Path)
  if ([System.IO.Path]::IsPathRooted($Path)) { return $Path }
  return Join-Path $RepoRoot $Path
}

function Assert-Contains {
  param([string]$Text, [string]$Pattern, [string]$Message)
  if ($Text -notmatch $Pattern) { throw $Message }
}

$docText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Doc)
$directionText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Direction)
$boundariesText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Boundaries)
$codeText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Code)
$testText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Tests)
$fuzzTestText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $FuzzTests)

foreach ($term in @(
    'Aetra Virtual Machine',
    'x/aetravm/avm',
    'binary serialization spec',
    'message ABI',
    'storage ABI',
    'deterministic execution proof',
    'gas schedule',
    'memory limits',
    'code size limits',
    'stack/register limits',
    'host function allowlist',
    'fuzz tests',
    'differential tests',
    'upgrade and migration policy',
    'adversarial audit',
    'magic',
    'version',
    'metadata_hash',
    'imports',
    'exports',
    'instruction_count',
    'sha256(encoded_module)',
    'deploy',
    'receive external',
    'receive internal',
    'receive bounced',
    'query/getter',
    'migrate',
    'value in `naet`',
    'read storage',
    'write storage',
    'emit internal message',
    'inspect message envelope',
    'get block context',
    'charge gas',
    'return result code',
    'wall-clock time',
    'random host entropy',
    'filesystem or network access',
    'floating point',
    'unbounded iteration',
    'nondeterministic map iteration',
    'bytecode verifier',
    'disassembler',
    'local runner',
    'gas profiler',
    'contract test harness',
    'state snapshot inspector',
    'store code',
    'instantiate contract',
    'route external message',
    'process internal queue',
    'execute getters',
    'export/import state',
    'go test ./x/aetravm/avm',
    'AVM can execute a minimal contract deterministically'
  )) {
  Assert-Contains -Text $docText -Pattern ([regex]::Escape($term)) -Message "AVM doc missing: $term"
}

foreach ($term in @(
    'Magic',
    'Version',
    'DecodeModule',
    'EncodeModule',
    'CodeHash',
    'EntryDeploy',
    'EntryReceiveExternal',
    'EntryReceiveInternal',
    'EntryReceiveBounced',
    'EntryQuery',
    'EntryMigrate',
    'HostReadStorage',
    'HostWriteStorage',
    'HostEmitInternal',
    'HostInspectMsg',
    'HostBlockContext',
    'HostChargeGas',
    'HostReturn',
    'OpWallClock',
    'OpRandom',
    'OpFileRead',
    'OpFloatAdd',
    'OpIterMap',
    'IsForbiddenOpcode',
    'AsyncHandler'
  )) {
  Assert-Contains -Text $codeText -Pattern ([regex]::Escape($term)) -Message "AVM code missing: $term"
}

foreach ($term in @(
    'TestDeployValidModuleAndRejectMalformedBytecode',
    'TestBytecodeEncodeDecodeDifferentialRoundTrip',
    'TestVerifierRejectsOversizedCodeAndNondeterministicOpcode',
    'TestRunSimpleCounterDeterministicallyAndBoundsGas',
    'TestAVMEmitsInternalMessageIntoAsyncQueue',
    'TestAVMAsyncFailedSendBouncesAndQueueSurvivesExportImport'
  )) {
  Assert-Contains -Text $testText -Pattern ([regex]::Escape($term)) -Message "AVM test missing: $term"
}

Assert-Contains -Text $fuzzTestText -Pattern ([regex]::Escape('FuzzDecodeModuleRejectsMalformedWithoutPanic')) -Message "AVM fuzz test missing decoder fuzz target"

foreach ($term in @(
    'x/aetravm/avm',
    'Aetra Virtual Machine',
    'go test ./x/aetravm/avm'
  )) {
  Assert-Contains -Text $directionText -Pattern ([regex]::Escape($term)) -Message "VM direction missing AVM term: $term"
}

foreach ($term in @(
    'x/aetravm/avm',
    'pure Go executable specification',
    'No SDK stores, keepers, module accounts, genesis, CLI, or ABCI hooks'
  )) {
  Assert-Contains -Text $boundariesText -Pattern ([regex]::Escape($term)) -Message "module boundaries missing AVM term: $term"
}

Write-Host "AVM doc test passed"
