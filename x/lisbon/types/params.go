package types

import (
	sdk_math "cosmossdk.io/math"
	"errors"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	evm "github.com/evmos/ethermint/x/evm/types"
	"strings"
)

var DefaultDenom = evm.DefaultEVMDenom

var (
	ParamStoreKeyDenom         = []byte("ParamStoreKeyDenom")
	ParamStoreKeyMintAmount    = []byte("ParamStoreKeyMintAmount")
	ParamStoreKeyBurnShares    = []byte("ParamStoreKeyBurnShares")
	ParamStoreKeyRewardsShares = []byte("ParamStoreKeyRewardsShares")
)

func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(
	Denom string,
	mintAmount sdk_math.Int,
	burnShares sdk.Dec,
) Params {
	return Params{
		Denom:      Denom,
		MintAmount: mintAmount,
		BurnShares: burnShares,
	}
}

func DefaultParams() Params {
	return Params{
		Denom:      DefaultDenom,
		MintAmount: sdk_math.NewInt(1000000000000000000),
		BurnShares: sdk.NewDecWithPrec(20, 2),
	}
}

func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(ParamStoreKeyDenom, &p.Denom, validateDenom),
		paramtypes.NewParamSetPair(ParamStoreKeyMintAmount, &p.MintAmount, validateMintAmount),
		paramtypes.NewParamSetPair(ParamStoreKeyBurnShares, &p.BurnShares, validateBurnShares),
	}
}

func validateDenom(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if strings.TrimSpace(v) == "" {
		return errors.New("mint denom cannot be blank")
	}
	if err := sdk.ValidateDenom(v); err != nil {
		return err
	}

	return nil
}

func validateMintAmount(i interface{}) error {
	v, ok := i.(sdk_math.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	// validate initial value
	if v.IsNegative() {
		return fmt.Errorf("initial value cannot be negative")
	}
	return nil

}

func validateBurnShares(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	// validate initial value
	if v.IsNegative() {
		return fmt.Errorf("initial value cannot be negative")
	}
	if v.GT(sdk.NewDec(1)) {
		return fmt.Errorf("reduction factor cannot be greater than 1")
	}
	return nil
}

func (p Params) Validate() error {
	if err := validateDenom(p.Denom); err != nil {
		return err
	}
	if err := validateMintAmount(p.MintAmount); err != nil {
		return err
	}
	if err := validateBurnShares(p.BurnShares); err != nil {
		return err
	}

	return nil
}
