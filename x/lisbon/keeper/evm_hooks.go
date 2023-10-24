package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/core"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	"github.com/evmos/evmos/v10/x/lisbon/types"
)

var _ evmtypes.EvmHooks = Hooks{}

// Hooks wrapper struct for fees keeper
type Hooks struct {
	k Keeper
}

// Hooks return the wrapper hooks struct for the Keeper
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}
func (h Hooks) PostTxProcessing(ctx sdk.Context, msg core.Message, receipt *ethtypes.Receipt) error {
	return h.k.PostTxProcessing(ctx, msg, receipt)
}

// PostTxProcessing implements EvmHooks.PostTxProcessing. After each successful
// interaction with a registered contract, the contract deployer (or, if set,
// the withdraw address) receives a share from the transaction fees paid by the
// transaction sender.
func (k Keeper) PostTxProcessing(
	ctx sdk.Context,
	msg core.Message,
	receipt *ethtypes.Receipt,
) error {
	contract := msg.To()
	if contract == nil {
		return nil
	}
	txFee := sdk.NewIntFromUint64(receipt.GasUsed).Mul(sdk.NewIntFromBigInt(msg.GasPrice()))
	evmDenom := k.evmKeeper.GetEVMDenom(ctx)
	burnDecCoin := sdk.NewDecWithPrec(20, 2).MulInt(txFee).TruncateInt()
	burnCoins := sdk.NewCoins(sdk.NewCoin(evmDenom, burnDecCoin))
	err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, k.feeCollectorName, types.ModuleName, burnCoins)
	if err == nil {
		err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, burnCoins)
		if err != nil {
			return errorsmod.Wrapf(
				err,
				"failed to burn %s from fee collector account. contract %s",
				burnCoins.String(), contract,
			)
		} else {
			k.Logger(ctx).Info(
				"@@BurnCoins success",
				"height", ctx.BlockHeight(),
				"txFee", txFee,
				"burnCoins", burnCoins.String(),
				"txHash", receipt.TxHash.Hex(),
			)
		}
	}

	return nil
}
