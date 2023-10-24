package types

import (
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
	mintAmount sdk.Dec,
	burnShares sdk.Dec,
	rewardsShares sdk.Dec,
) Params {
	return Params{
		Denom:         Denom,
		MintAmount:    mintAmount,
		BurnShares:    burnShares,
		RewardsShares: rewardsShares,
	}
}

func DefaultParams() Params {
	return Params{
		Denom:         DefaultDenom,
		MintAmount:    sdk.NewDec(int64(1_000_000_000_000_000_000)),
		BurnShares:    sdk.NewDecWithPrec(20, 2),
		RewardsShares: sdk.NewDecWithPrec(40, 2),
	}
}

func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(ParamStoreKeyDenom, &p.Denom, validateDenom),
		paramtypes.NewParamSetPair(ParamStoreKeyMintAmount, &p.MintAmount, validateMintAmount),
		paramtypes.NewParamSetPair(ParamStoreKeyBurnShares, &p.BurnShares, validateBurnShares),
		paramtypes.NewParamSetPair(ParamStoreKeyRewardsShares, &p.RewardsShares, validateRewardsShares),
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

func validateRewardsShares(i interface{}) error {
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
	if err := validateRewardsShares(p.RewardsShares); err != nil {
		return err
	}

	return nil
}
