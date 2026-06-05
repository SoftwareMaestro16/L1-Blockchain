param(
  [string]$OutputDir = "",
  [string]$Binary = "",
  [string]$ChainId = "aetheris-local-1",
  [int]$ValidatorCount = 3,
  [int]$MinHeight = 4,
  [int]$TimeoutSeconds = 120,
  [int]$BaseP2PPort = 26656,
  [int]$BaseRPCPort = 26657,
  [int]$BaseRESTPort = 1317,
  [int]$BaseGRPCPort = 9090,
  [int]$BasePprofPort = 6060,
  [int]$PortStride = 100,
  [string]$TimeoutCommit = "1s",
  [string]$LogLevel = "info",
  [ValidateSet("base", "execution-os-sim", "zones-prototype", "mesh-prototype", "identity-prototype")]
  [string]$ProfileName = "base",
  [bool]$EnableAPI = $true,
  [bool]$EnableGRPC = $true,
  [bool]$EnableRPC = $true,
  [string]$Node = "",
  [string]$Fees = "1000naet",
  [string]$WrongFees = "1000testtoken",
  [string]$FactorySubdenom = "smokegold",
  [string]$DelegationAmount = "5000000naet",
  [switch]$SkipBuild,
  [switch]$KeepLogsOnFailure
)

$ErrorActionPreference = "Stop"

$args = @{
  Profile          = "Smoke"
  OutputDir        = $OutputDir
  Binary           = $Binary
  ChainId          = $ChainId
  ValidatorCount   = $ValidatorCount
  MinHeight        = $MinHeight
  TimeoutSeconds   = $TimeoutSeconds
  BaseP2PPort      = $BaseP2PPort
  BaseRPCPort      = $BaseRPCPort
  BaseRESTPort     = $BaseRESTPort
  BaseGRPCPort     = $BaseGRPCPort
  BasePprofPort    = $BasePprofPort
  PortStride       = $PortStride
  TimeoutCommit    = $TimeoutCommit
  LogLevel         = $LogLevel
  ProfileName      = $ProfileName
  EnableAPI        = $EnableAPI
  EnableGRPC       = $EnableGRPC
  EnableRPC        = $EnableRPC
  Node             = $Node
  Fees             = $Fees
  WrongFees        = $WrongFees
  FactorySubdenom  = $FactorySubdenom
  DelegationAmount = $DelegationAmount
}

if ($SkipBuild) {
  $args.SkipBuild = $true
}
if ($KeepLogsOnFailure) {
  $args.KeepLogsOnFailure = $true
}

& (Join-Path $PSScriptRoot "prototype_acceptance.ps1") @args
