package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v10/x/happyoct/types"
)

func (k Keeper) PrintLog(ctx sdk.Context) error {
	if err := k.MintAndAllocateInflation(ctx); err != nil {
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
func (k Keeper) MintAndAllocateInflation(ctx sdk.Context) (err error) {
	// Mint coins for distribution
	coin := sdk.NewCoin("aevmos", sdk.NewInt(1000000000000000000))
	if err := k.MintCoins(ctx, coin); err != nil {
		return err
	}

	// Allocate minted coins according to allocation proportions (staking, usage
	// incentives, community pool)
	return k.AllocateExponentialInflation(ctx, coin)
}

// AllocateExponentialInflation allocates coins from the inflation to external
// modules according to allocation proportions:
func (k Keeper) AllocateExponentialInflation(
	ctx sdk.Context,
	mintedCoin sdk.Coin,
) (
	err error,
) {
	// Allocate staking rewards into fee collector account
	mintedRewards := sdk.NewCoins(mintedCoin)
	err = k.bankKeeper.SendCoinsFromModuleToModule(
		ctx,
		types.ModuleName,
		k.feeCollectorName,
		mintedRewards,
	)
	if err != nil {
		return err
	}
	k.Logger(ctx).Info(
		"AllocateExponentialInflation",
		"height", ctx.BlockHeight(),
		"mintedRewards", mintedRewards.String(),
	)
	return nil
}
