package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v12/x/lisbon/types"
	gogotypes "github.com/gogo/protobuf/types"
)

// GetPreviousProposerConsAddr returns the proposer consensus address for the
// current block.
func (k Keeper) GetCurrentProposerConsAddr(ctx sdk.Context) sdk.ConsAddress {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ProposerKey)
	if bz == nil {
		panic("previous proposer not set")
	}

	addrValue := gogotypes.BytesValue{}
	k.cdc.MustUnmarshal(bz, &addrValue)
	return addrValue.GetValue()
}

// set the proposer public key for this block
func (k Keeper) SetCurrentProposerConsAddr(ctx sdk.Context, consAddr sdk.ConsAddress) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&gogotypes.BytesValue{Value: consAddr})
	store.Set(types.ProposerKey, bz)
}
