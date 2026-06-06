param(
  [string]$Doc = "docs\security\account-transaction-safety.md",
  [string]$Addressing = "app\addressing\validation.go",
  [string]$TokenParams = "app\params\token.go",
  [string]$FeeParams = "app\params\fees.go",
  [string]$TxLifecycle = "tests\integration\tx_lifecycle_test.go",
  [string]$Adversarial = "tests\adversarial\custom_modules_test.go",
  [string]$SecurityWorkflow = ".github\workflows\security.yml",
  [string]$DeterminismGate = "scripts\security\determinism-gate.ps1"
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

function Assert-NotContains {
  param([string]$Text, [string]$Pattern, [string]$Message)
  if ($Text -match $Pattern) { throw $Message }
}

$docText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Doc)
$addressingText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Addressing)
$tokenParamsText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $TokenParams)
$feeParamsText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $FeeParams)
$txLifecycleText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $TxLifecycle)
$adversarialText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $Adversarial)
$securityWorkflowText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $SecurityWorkflow)
$determinismGateText = Get-Content -Raw -LiteralPath (Resolve-RepoPath $DeterminismGate)

foreach ($term in @(
    "Base Chain Safety Before Contracts",
    "AVM and CosmWasm must not be allowed to mutate production state",
    'address validation lives in `app/addressing`',
    'native token constants live in `app/params`',
    'BaseDenom = "naet"',
    "ValidateNativeFeeDenomsV1",
    "zero address is rejected everywhere by default",
    'old public `0:`, `orb1`, and `ORB` formats',
    "bad transactions fail before message state mutation",
    "invalid signers cannot mutate state",
    "malformed tx bytes must not panic consensus code",
    "signed transaction replay test using identical signed bytes",
    "wrong chain-id signing test",
    "malformed protobuf transaction test",
    "invalid signer tests for bank, staking, gov, fees, contract-assets, DEX, AVM",
    "consensus panic tests for every custom module message and genesis type",
    "deterministic event contract tests",
    "go test ./...",
    "go vet ./...",
    "buf lint",
    "govulncheck",
    "gosec",
    "CodeQL",
    "gitleaks",
    "determinism gate"
  )) {
  Assert-Contains -Text $docText -Pattern ([regex]::Escape($term)) -Message "base chain safety doc missing: $term"
}

foreach ($term in @(
    "ParseUserAddress",
    "ParseAuthorityAddress",
    "ParseOptionalAdminAddress",
    "ParseContractAddress",
    "RejectZeroAddress",
    "IsZeroAccAddress"
  )) {
  Assert-Contains -Text $addressingText -Pattern ([regex]::Escape($term)) -Message "addressing helper missing: $term"
}

foreach ($term in @(
    'BaseDenom            = "naet"',
    'DisplayDenom         = "AET"',
    "NativeTokenMetadata"
  )) {
  Assert-Contains -Text $tokenParamsText -Pattern ([regex]::Escape($term)) -Message "native token params missing: $term"
}

foreach ($term in @(
    "ValidateNativeFeeDenomsV1",
    "BaseDenom",
    "v1 only accepts fee denom"
  )) {
  Assert-Contains -Text $feeParamsText -Pattern ([regex]::Escape($term)) -Message "fee params helper missing: $term"
}

foreach ($term in @(
    "TestSignedBankTxReplayIsRejectedAfterSequenceIncrement",
    "EncodeSignedTxWithChainID",
    "TestWrongChainIDSignedTxFailsBeforeBalanceMutation",
    "TestMalformedProtobufTxBytesFailWithoutFeeAccounting",
    "TestInvalidSignerTxFailsBeforeBalanceMutation",
    "TestMissingAndInvalidFeeTxsFailBeforeBalanceMutation",
    "TestUserCreatedTokenCannotPayProtocolFeesEvenWhenOwned",
    "TestInsufficientFeeFundsFailBeforeStateTransition"
  )) {
  Assert-Contains -Text $txLifecycleText -Pattern ([regex]::Escape($term)) -Message "tx lifecycle safety test missing: $term"
}

foreach ($term in @(
    "TestMalformedTxBytesFailSafely",
    "TestZeroAddressProtocolSafetyRules",
    "TestFeeAndGovernanceAbuseRejected",
    "TestRepeatedInvalidFeeSpamDoesNotAdvanceProtocolAccounting",
    "ErrUnauthorized",
    "ErrInvalidAddress",
    "ErrInvalidParams"
  )) {
  Assert-Contains -Text $adversarialText -Pattern ([regex]::Escape($term)) -Message "adversarial safety test missing: $term"
}

foreach ($term in @(
    "name: govulncheck",
    "name: gosec high severity",
    "name: gitleaks secrets",
    "dependency-review-action",
    "name: CodeQL"
  )) {
  Assert-Contains -Text $securityWorkflowText -Pattern ([regex]::Escape($term)) -Message "security workflow missing: $term"
}

foreach ($term in @(
    "Deterministic execution gate",
    "High",
    "Critical",
    "blocking"
  )) {
  Assert-Contains -Text $determinismGateText -Pattern ([regex]::Escape($term)) -Message "determinism gate missing: $term"
}

Assert-NotContains -Text $docText -Pattern 'norb|orbitalisd|orbitalis-local-1|\.orbitalis' -Message "base chain safety doc contains old runtime terms"

Write-Host "base chain safety doc test passed"
