package stride

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/evmos/evmos/v14/precompiles/ics20"
)

const (
	// LiquidStakeMethod is the name of the liquidStake method
	LiquidStakeMethod = "liquidStake"
	// RedeemMethod is the name of the redeem method
	RedeemMethod = "redeem"
	// LiquidStakeAction is the action name needed in the memo field
	LiquidStakeAction = "LiquidStake"
)

// LiquidStake is a transaction that liquid stakes tokens using
// a ICS20 transfer with a custom memo field that will trigger Stride's Autopilot middleware
func (p Precompile) LiquidStake(
	ctx sdk.Context,
	origin common.Address,
	stateDB vm.StateDB,
	contract *vm.Contract,
	method *abi.Method,
	args []interface{},
) ([]byte, error) {
	sender, token, amount, receiver, err := ParseLiquidStakeArgs(args)
	if err != nil {
		return nil, err
	}

	// The provided sender address should always be equal to the origin address.
	// In case the contract caller address is the same as the sender address provided,
	// update the sender address to be equal to the origin address.
	// Otherwise, if the provided sender address is different from the origin address,
	// return an error because is a forbidden operation
	sender, err = ics20.CheckOriginAndSender(contract, origin, sender)
	if err != nil {
		return nil, err
	}

	evmDenom := p.evmKeeper.GetParams(ctx).EvmDenom

	tokenPairID := p.erc20Keeper.GetDenomMap(ctx, evmDenom)
	tokenPair, found := p.erc20Keeper.GetTokenPair(ctx, tokenPairID)
	// NOTE this should always exist
	if !found {
		return nil, fmt.Errorf("token pair not found")
	}

	// NOTE: for v1 we only support the native EVM (and staking) denomination (WEVMOS/WTEVMOS).
	if token != tokenPair.GetERC20Contract() {
		return nil, fmt.Errorf("unsupported token %s. The only supported token contract for Stride Outpost v1 is %s", token, tokenPair.Erc20Address)
	}

	coin := sdk.Coin{Denom: tokenPair.Denom, Amount: sdk.NewIntFromBigInt(amount)}

	// Create the memo for the ICS20 transfer
	memo := p.createMemo(LiquidStakeAction, receiver)

	// Build the MsgTransfer with the memo and coin
	// TODO: move out this function
	msg, err := NewMsgTransfer(p.channelID, sdk.AccAddress(sender.Bytes()).String(), receiver, memo, coin)
	if err != nil {
		return nil, err
	}

	// no need to have authorization when the contract caller is the same as origin (owner of funds)
	// and the sender is the origin
	accept, expiration, err := ics20.CheckAndAcceptAuthorizationIfNeeded(ctx, contract, origin, p.AuthzKeeper, msg)
	if err != nil {
		return nil, err
	}

	// Execute the ICS20 Transfer
	res, err := p.transferKeeper.Transfer(sdk.WrapSDKContext(ctx), msg)
	if err != nil {
		return nil, err
	}

	// Update grant only if is needed
	if err := ics20.UpdateGrantIfNeeded(ctx, contract, p.AuthzKeeper, origin, expiration, accept); err != nil {
		return nil, err
	}

	// Emit the IBC transfer Event
	if err = ics20.EmitIBCTransferEvent(ctx, stateDB, p.ABI.Events, sender, p.Address(), msg); err != nil {
		return nil, err
	}

	// Emit the custom LiquidStake Event
	if err = p.EmitLiquidStakeEvent(ctx, stateDB, sender, token, amount); err != nil {
		return nil, err
	}

	return method.Outputs.Pack(res.Sequence, true)
}

// Redeem is a transaction that redeems the native tokens using the liquid stake
// tokens. It executes a ICS20 transfer with a custom memo field that will
// trigger Stride's Autopilot middleware
//func (p Precompile) Redeem(
//	ctx sdk.Context,
//	origin common.Address,
//	stateDB vm.StateDB,
//	contract *vm.Contract,
//	method *abi.Method,
//	args []interface{},
//) ([]byte, error) {
//	sender, token, amount, receiver, err := ParseLiquidStakeArgs(args)
//	if err != nil {
//		return nil, err
//	}
//
//	// The provided sender address should always be equal to the origin address.
//	// In case the contract caller address is the same as the sender address provided,
//	// update the sender address to be equal to the origin address.
//	// Otherwise, if the provided sender address is different from the origin address,
//	// return an error because is a forbidden operation
//	sender, err = ics20.CheckOriginAndSender(contract, origin, sender)
//	if err != nil {
//		return nil, err
//	}
//
//	evmDenom := p.evmKeeper.GetParams(ctx).EvmDenom
//	stToken := fmt.Sprintf("st%s", evmDenom)
//
//	// TODO: move this into a separate function
//	denomTrace := ibctransfertypes.DenomTrace{
//		Path:      fmt.Sprintf("%s/%s", p.portID, p.channelID),
//		BaseDenom: stToken,
//	}
//
//	ibcDenom := denomTrace.IBCDenom()
//
//	tokenPairID := p.erc20Keeper.GetDenomMap(ctx, ibcDenom)
//	tokenPair, found := p.erc20Keeper.GetTokenPair(ctx, tokenPairID)
//	if !found {
//		return nil, fmt.Errorf("token pair not found for %s", ibcDenom)
//	}
//
//	if token != tokenPair.GetERC20Contract() {
//		return nil, fmt.Errorf("unsupported token %s. The only supported token contract for Stride Outpost v1 is %s", token, tokenPair.Erc20Address)
//	}
//
//	coin := sdk.Coin{Denom: tokenPair.Denom, Amount: sdk.NewIntFromBigInt(amount)}
//
//	// Create the memo for the ICS20 transfer
//	memo := p.createRedeemMemo(receiver)
//
//	// Build the MsgTransfer with the memo and coin
//	// TODO: move out this function
//	msg, err := NewMsgTransfer(p.channelID, sdk.AccAddress(sender.Bytes()).String(), receiver, memo, coin)
//	if err != nil {
//		return nil, err
//	}
//
//	// no need to have authorization when the contract caller is the same as origin (owner of funds)
//	// and the sender is the origin
//	accept, expiration, err := ics20.CheckAndAcceptAuthorizationIfNeeded(ctx, contract, origin, p.AuthzKeeper, msg)
//	if err != nil {
//		return nil, err
//	}
//
//	// Execute the ICS20 Transfer
//	res, err := p.transferKeeper.Transfer(sdk.WrapSDKContext(ctx), msg)
//	if err != nil {
//		return nil, err
//	}
//
//	// Update grant only if is needed
//	if err := ics20.UpdateGrantIfNeeded(ctx, contract, p.AuthzKeeper, origin, expiration, accept); err != nil {
//		return nil, err
//	}
//
//	// Emit the IBC transfer Event
//	if err = ics20.EmitIBCTransferEvent(ctx, stateDB, p.ABI.Events, sender, p.Address(), msg); err != nil {
//		return nil, err
//	}
//
//	// Emit the custom Redeem Event
//	if err = p.EmitRedeemEvent(ctx, stateDB, sender, token, amount); err != nil {
//		return nil, err
//	}
//
//	return method.Outputs.Pack(res.Sequence, true)
//}