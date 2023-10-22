package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v10/x/happyoct/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

//func (k Keeper) PrintLog(ctx sdk.Context, req abci.RequestBeginBlock) error {
//	if err := k.MintAndAllocateInflation(ctx, req); err != nil {
//		return err
//	}
//	//k.AllocateTokens(ctx, req)
//	return nil
//}

func (k Keeper) BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) {
	consAddr := sdk.ConsAddress(req.Header.ProposerAddress)
	k.SetCurrentProposerConsAddr(ctx, consAddr)
	k.Logger(ctx).Info(
		"Set Current Proposer Cons Addr.",
		"height", ctx.BlockHeight(),
		"consAddr", consAddr.String(),
	)

}

func (k Keeper) EndBlocker(ctx sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	proposerConsAddr := k.GetCurrentProposerConsAddr(ctx)
	k.Logger(ctx).Info(
		"Get Current Proposer Cons Addr.",
		"height", ctx.BlockHeight(),
		"proposerConsAddr", proposerConsAddr.String(),
	)
	k.MintAndAllocateInflation(ctx, proposerConsAddr)
	return []abci.ValidatorUpdate{}
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
func (k Keeper) MintAndAllocateInflation(ctx sdk.Context, proposer sdk.ConsAddress) (err error) {
	// Mint coins for distribution
	currentValidator := k.stakingKeeper.ValidatorByConsAddr(ctx, proposer)
	evmDenom := k.evmKeeper.GetEVMDenom(ctx)
	coin := sdk.NewCoin(evmDenom, sdk.NewInt(1000000000000000000))
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
	return k.AllocateInflation(ctx, coin, currentValidator.GetOperator())
}

// AllocateInflation allocates coins from the inflation to external
// modules according to allocation proportions:
func (k Keeper) AllocateInflation(
	ctx sdk.Context,
	mintedCoin sdk.Coin,
	validatorAddr sdk.ValAddress,
) (err error) {
	// Allocate staking rewards into fee collector account
	mintedRewards := sdk.NewCoins(mintedCoin)
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

func (k Keeper) AllocateTokens(ctx sdk.Context, proposer sdk.ConsAddress) {
	currentValidator := k.stakingKeeper.ValidatorByConsAddr(ctx, proposer)
	feeCollector := k.accountKeeper.GetModuleAccount(ctx, k.feeCollectorName)
	feesCollectedInt := k.bankKeeper.GetAllBalances(ctx, feeCollector.GetAddress())
	// transfer collected fees to the distribution module account
	err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, k.feeCollectorName, types.ModuleName, feesCollectedInt)
	if err != nil {
		panic(err)
	}
	//k.bankKeeper.BurnCoins(ctx, types.ModuleName, coins)
	err = k.bankKeeper.SendCoinsFromModuleToAccount(
		ctx,
		types.ModuleName,
		sdk.AccAddress(currentValidator.GetOperator()),
		feesCollectedInt,
	)
	k.Logger(ctx).Info(
		"Allocate Fee Tokens",
		"height", ctx.BlockHeight(),
		"currentValidator", currentValidator.GetOperator().String(),
		"feesCollectedInt", feesCollectedInt.String(),
	)
}
