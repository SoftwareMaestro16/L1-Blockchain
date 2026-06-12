package async

import (
	sdkmath "cosmossdk.io/math"

	appparams "github.com/sovereign-l1/l1/app/params"
)

func (e *Executor) EconomicActivity() appparams.ProtocolEconomicActivity {
	storageFees := sdkmath.ZeroInt()
	forwardingFees := sdkmath.ZeroInt()
	for _, receipt := range e.receipts {
		storageFees = storageFees.Add(normalizeEconomicInt(receipt.StorageFeeNaet))
		forwardingFees = forwardingFees.Add(normalizeEconomicInt(receipt.ForwardFeeNaet))
	}
	return appparams.ProtocolEconomicActivity{
		AVMStorageFeeNaet:	storageFees,
		AVMForwardingFeeNaet:	forwardingFees,
		AVMDeploymentCostNaet:	sdkmath.NewIntFromUint64(e.metrics.DeploymentCostsNaet),
	}
}

func (e *Executor) EconomicFlow(control appparams.BalanceControllerOutput) (appparams.ProtocolEconomicFlowOutput, error) {
	return appparams.ComputeProtocolEconomicFlow(appparams.ProtocolEconomicFlowInput{
		Activity:		e.EconomicActivity(),
		BurnRatioBps:		control.BurnRatioBps,
		TreasuryRatioBps:	appparams.TreasuryFeeRatioBps,
	})
}

func addNaetMetric(current uint64, amount sdkmath.Int) uint64 {
	amount = normalizeEconomicInt(amount)
	if !amount.IsPositive() {
		return current
	}
	remaining := ^uint64(0) - current
	if amount.GT(sdkmath.NewIntFromUint64(remaining)) {
		return ^uint64(0)
	}
	return current + amount.Uint64()
}

func normalizeEconomicInt(value sdkmath.Int) sdkmath.Int {
	if value.IsNil() {
		return sdkmath.ZeroInt()
	}
	return value
}
