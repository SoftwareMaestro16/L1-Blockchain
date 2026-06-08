param(
  [string]$OutputDir = ".devnet-1",
  [string]$Binary = "",
  [string]$ChainId = "aetra-devnet-1",
  [int]$BaseP2PPort = 26656,
  [int]$BaseRPCPort = 26657,
  [int]$BaseRESTPort = 1317,
  [int]$BaseGRPCPort = 9090,
  [switch]$SkipBuild,
  [switch]$Wait
)

$ErrorActionPreference = "Stop"
$RepoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot "..\.."))

$args = @{
  OutputDir      = $OutputDir
  Binary         = $Binary
  ChainId        = $ChainId
  ValidatorCount = 1
  BaseP2PPort    = $BaseP2PPort
  BaseRPCPort    = $BaseRPCPort
  BaseRESTPort   = $BaseRESTPort
  BaseGRPCPort   = $BaseGRPCPort
  TimeoutCommit  = "1s"
  LogLevel       = "info"
  Profile        = "base"
}
if ($SkipBuild) { $args.SkipBuild = $true }

& (Join-Path $RepoRoot "scripts\localnet\init.ps1") @args

$startArgs = $args.Clone()
$startArgs.Remove("SkipBuild")
$startArgs.NoInit = $true
if ($Wait) { $startArgs.Wait = $true }
& (Join-Path $RepoRoot "scripts\localnet\start.ps1") @startArgs
