package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v10/x/lisbon/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

func (k Keeper) MintCoins(ctx sdk.Context, coin sdk.Coin) error {
	coins := sdk.NewCoins(coin)

	// skip as no coins need to be minted
	if coins.Empty() {
		return nil
	}

	return k.bankKeeper.MintCoins(ctx, types.ModuleName, coins)
}

func (k Keeper) BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) {
	consAddr := sdk.ConsAddress(req.Header.ProposerAddress)
	k.SetCurrentProposerConsAddr(ctx, consAddr)
	k.Logger(ctx).Info(
		"@@Set Current Proposer Cons Addr.",
		"height", ctx.BlockHeight(),
		"consAddr", consAddr.String(),
	)

}

func (k Keeper) EndBlocker(ctx sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	proposerConsAddr := k.GetCurrentProposerConsAddr(ctx)
	k.Logger(ctx).Info(
		"@@Get Current Proposer Cons Addr.",
		"height", ctx.BlockHeight(),
		"proposerConsAddr", proposerConsAddr.String(),
	)
	k.MintAndAllocateInflation(ctx, proposerConsAddr)
	k.AllocateTokens(ctx, proposerConsAddr)
	return []abci.ValidatorUpdate{}
}

// MintAndAllocateInflation performs inflation minting and allocation
func (k Keeper) MintAndAllocateInflation(ctx sdk.Context, proposer sdk.ConsAddress) (err error) {
	// Mint coins for distribution
	currentValidator := k.stakingKeeper.ValidatorByConsAddr(ctx, proposer)
	params := k.GetParams(ctx)
	//evmDenom := k.evmKeeper.GetEVMDenom(ctx)
	coin := sdk.NewCoin(params.Denom, params.MintAmount)
	if err := k.MintCoins(ctx, coin); err != nil {
		return err
	}

	k.Logger(ctx).Info(
		"@@Mint And Allocate Inflation",
		"height", ctx.BlockHeight(),
		"consAddr", proposer.String(),
		"validatorAddr", currentValidator.GetOperator().String(),
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
	oldBalance := k.bankKeeper.GetBalance(ctx, sdk.AccAddress(validatorAddr), mintedCoin.Denom)
	k.Logger(ctx).Info(
		"@@Validator Balance",
		"height", ctx.BlockHeight(),
		"old balance", oldBalance.String(),
	)
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
	newBalance := k.bankKeeper.GetBalance(ctx, sdk.AccAddress(validatorAddr), mintedCoin.Denom)
	k.Logger(ctx).Info(
		"@@Allocate Inflation",
		"height", ctx.BlockHeight(),
		"validatorAddr", validatorAddr.String(),
		"mintedRewards", mintedRewards.String(),
		"new balance", newBalance.String(),
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
		"@@Allocate Fee To Validator.",
		"height", ctx.BlockHeight(),
		"currentValidator", currentValidator.GetOperator().String(),
		"feesCollectedInt", feesCollectedInt.String(),
	)
}
