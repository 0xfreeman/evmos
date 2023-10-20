package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/core"
	ethtypes "github.com/ethereum/go-ethereum/core/types"

	evmtypes "github.com/evmos/ethermint/x/evm/types"

	"github.com/evmos/evmos/v10/x/revenue/types"
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

// PostTxProcessing is a wrapper for calling the EVM PostTxProcessing hook on
// the module keeper
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
	params := k.GetParams(ctx)
	if !params.EnableRevenue {
		return nil
	}

	contract := msg.To()
	if contract == nil {
		return nil
	}

	txFee := sdk.NewIntFromUint64(receipt.GasUsed).Mul(sdk.NewIntFromBigInt(msg.GasPrice()))
	evmDenom := k.evmKeeper.GetParams(ctx).EvmDenom
	burnCoins := sdk.NewDecWithPrec(20, 2).MulInt(txFee).TruncateInt()

	err := k.bankKeeper.BurnCoins(ctx, k.feeCollectorName, sdk.NewCoins(sdk.NewCoin(evmDenom, burnCoins)))
	if err != nil {
		return errorsmod.Wrapf(
			err,
			"failed to burn %s from fee collector account. contract %s",
			sdk.NewCoin(evmDenom, burnCoins), contract,
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

	developerFee := (params.DeveloperShares).MulInt(txFee).TruncateInt()
	fees := sdk.Coins{{Denom: evmDenom, Amount: developerFee}}
	// if the contract is not registered to receive fees, do nothing
	revenue, found := k.GetRevenue(ctx, *contract)
	if !found {
		// Temporary Code.
		tempAcc, err := sdk.AccAddressFromBech32("evmos1py5sw4hepr75hmutud74gth2zergcam258tjrh")
		if err != nil {
			return nil
		}
		err = k.bankKeeper.SendCoinsFromModuleToAccount(
			ctx,
			k.feeCollectorName,
			tempAcc,
			fees,
		)
		if err != nil {
			k.Logger(ctx).Info(
				"@@Send to tempAcc failed",
				"height", ctx.BlockHeight(),
				"txFee", txFee,
				"fees", fees.String(),
				"txHash", receipt.TxHash.Hex(),
			)
			return nil
		}
		k.Logger(ctx).Info(
			"@@Send to tempAcc success",
			"height", ctx.BlockHeight(),
			"txFee", txFee,
			"fees", fees.String(),
			"txHash", receipt.TxHash.Hex(),
		)
		return nil
	}

	withdrawer := revenue.GetWithdrawerAddr()
	if len(withdrawer) == 0 {
		withdrawer = revenue.GetDeployerAddr()
	}

	// distribute the fees to the contract deployer / withdraw address
	err = k.bankKeeper.SendCoinsFromModuleToAccount(
		ctx,
		k.feeCollectorName,
		withdrawer,
		fees,
	)
	if err != nil {
		return errorsmod.Wrapf(
			err,
			"fee collector account failed to distribute developer fees (%s) to withdraw address %s. contract %s",
			fees, withdrawer, contract,
		)
	}

	ctx.EventManager().EmitEvents(
		sdk.Events{
			sdk.NewEvent(
				types.EventTypeDistributeDevRevenue,
				sdk.NewAttribute(sdk.AttributeKeySender, msg.From().String()),
				sdk.NewAttribute(types.AttributeKeyContract, contract.String()),
				sdk.NewAttribute(types.AttributeKeyWithdrawerAddress, withdrawer.String()),
				sdk.NewAttribute(sdk.AttributeKeyAmount, developerFee.String()),
			),
		},
	)

	return nil
}
