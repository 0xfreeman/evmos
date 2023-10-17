package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/evmos/evmos/v10/x/happyoct/types"
	"github.com/tendermint/tendermint/libs/log"
)

type Keeper struct {
	storeKey         storetypes.StoreKey
	cdc              codec.BinaryCodec
	paramstore       paramtypes.Subspace
	bankKeeper       types.BankKeeper
	stakingKeeper    types.StakingKeeper
	accountKeeper    types.AccountKeeper
	feeCollectorName string
}

// NewKeeper creates a new mint Keeper instance
func NewKeeper(
	storeKey storetypes.StoreKey,
	cdc codec.BinaryCodec,
	ps paramtypes.Subspace,
	bk types.BankKeeper,
	sk types.StakingKeeper,
	ak types.AccountKeeper,
	feeCollectorName string,
) Keeper {
	// ensure mint module account is set
	if addr := ak.GetModuleAddress(types.ModuleName); addr == nil {
		panic("the mint module account has not been set")
	}

	// set KeyTable if it has not already been set
	//if !ps.HasKeyTable() {
	//	ps = ps.WithKeyTable(types.ParamKeyTable())
	//}

	return Keeper{
		storeKey:         storeKey,
		cdc:              cdc,
		paramstore:       ps,
		bankKeeper:       bk,
		stakingKeeper:    sk,
		accountKeeper:    ak,
		feeCollectorName: feeCollectorName,
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}
