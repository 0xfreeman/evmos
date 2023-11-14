package lisbon

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v12/x/lisbon/keeper"
	"github.com/evmos/evmos/v12/x/lisbon/types"
)

// InitGenesis import module genesis
func InitGenesis(
	ctx sdk.Context,
	k keeper.Keeper,
	ak types.AccountKeeper,
	data types.GenesisState,
) {
	// Ensure inflation module account is set on genesis
	if acc := ak.GetModuleAccount(ctx, types.ModuleName); acc == nil {
		panic("the inflation module account has not been set")
	}
	params := data.Params
	k.SetParams(ctx, params)
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	return &types.GenesisState{
		Params: k.GetParams(ctx),
	}
}
