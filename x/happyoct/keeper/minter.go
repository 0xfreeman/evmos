package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v10/x/happyoct/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

func (k Keeper) PrintLog(ctx sdk.Context, req abci.RequestBeginBlock) error {
	if err := k.MintAndAllocateInflation(ctx, req); err != nil {
		return err
	}
	return nil
}

func (k Keeper) MintCoins(ctx sdk.Context, coin sdk.Coin) error {
	coins := sdk.NewCoins(coin)

	// skip as no coins need to be minted
	if coins.Empty() {
		return nil
	}

	return k.bankKeeper.MintCoins(ctx, types.ModuleName, coins)
}

// MintAndAllocateInflation performs inflation minting and allocation
func (k Keeper) MintAndAllocateInflation(ctx sdk.Context, req abci.RequestBeginBlock) (err error) {
	// Mint coins for distribution
	currentProposer := sdk.ConsAddress(req.Header.ProposerAddress)
	currentValidator := k.stakingKeeper.ValidatorByConsAddr(ctx, currentProposer)
	coin := sdk.NewCoin("aevmos", sdk.NewInt(1000000000000000000))
	if err := k.MintCoins(ctx, coin); err != nil {
		return err
	}
	k.Logger(ctx).Info(
		"MintAndAllocateInflation",
		"height", ctx.BlockHeight(),
		"consAddr", currentValidator.GetOperator().String(),
		"AccAddress", sdk.AccAddress(currentValidator.GetOperator()).String(),
	)

	// Allocate minted coins according to allocation proportions (staking, usage
	// incentives, community pool)
	return k.AllocateExponentialInflation(ctx, coin, currentValidator.GetOperator())
}

// AllocateExponentialInflation allocates coins from the inflation to external
// modules according to allocation proportions:
func (k Keeper) AllocateExponentialInflation(
	ctx sdk.Context,
	mintedCoin sdk.Coin,
	validatorAddr sdk.ValAddress,
) (
	err error,
) {
	// Allocate staking rewards into fee collector account
	mintedRewards := sdk.NewCoins(mintedCoin)
	//err = k.bankKeeper.SendCoinsFromModuleToModule(
	//	ctx,
	//	types.ModuleName,
	//	k.feeCollectorName,
	//	mintedRewards,
	//)
	err = k.bankKeeper.SendCoinsFromModuleToAccount(
		ctx,
		types.ModuleName,
		sdk.AccAddress(validatorAddr),
		mintedRewards,
	)
	if err != nil {
		return err
	}
	k.Logger(ctx).Info(
		"AllocateExponentialInflation",
		"height", ctx.BlockHeight(),
		"mintedRewards", mintedRewards.String(),
		"validatorAddr", validatorAddr.String(),
	)
	return nil
}
