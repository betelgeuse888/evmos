package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/ethereum/go-ethereum/common"
	ethermint "github.com/tharsis/ethermint/types"
	"github.com/tharsis/evmos/x/incentives/types"
)

var _ types.QueryServer = Keeper{}

// Incentives return registered incentives
func (k Keeper) Incentives(
	c context.Context,
	req *types.QueryIncentivesRequest,
) (*types.QueryIncentivesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	var incentives []types.Incentive
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixIncentive)

	pageRes, err := query.Paginate(
		store,
		req.Pagination,
		func(_, value []byte) error {
			var incentive types.Incentive
			if err := k.cdc.Unmarshal(value, &incentive); err != nil {
				return err
			}
			incentives = append(incentives, incentive)
			return nil
		},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryIncentivesResponse{
		Incentives: incentives,
		Pagination: pageRes,
	}, nil
}

// Incentive returns a given registered incentive
func (k Keeper) Incentive(
	c context.Context,
	req *types.QueryIncentiveRequest,
) (*types.QueryIncentiveResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	// check if the contract is a hex address
	if err := ethermint.ValidateAddress(req.Contract); err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			"invalid format for contract %s, should be hex ('0x...')", req.Contract,
		)
	}

	incentive, found := k.GetIncentive(ctx, common.HexToAddress(req.Contract))
	if !found {
		return nil, status.Errorf(
			codes.NotFound,
			"incentive with contract'%s'",
			req.Contract,
		)
	}

	return &types.QueryIncentiveResponse{Incentive: incentive}, nil
}

// GasMeters return active gas meters
func (k Keeper) GasMeters(
	c context.Context,
	req *types.QueryGasMetersRequest,
) (*types.QueryGasMetersResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	// check if the contract is a hex address
	if err := ethermint.ValidateAddress(req.Contract); err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid contract address %s", req.Contract).Error(),
		)
	}

	ctx := sdk.UnwrapSDKContext(c)
	contract := common.HexToAddress(req.Contract)
	store := prefix.NewStore(ctx.KVStore(k.storeKey), append(types.KeyPrefixGasMeter, contract.Bytes()...))

	gms := []types.GasMeter{}

	pageRes, err := query.Paginate(
		store,
		req.Pagination,
		func(key, value []byte) error {
			participant := common.BytesToAddress(key)
			cumulativeGas := sdk.BigEndianToUint64(value)

			gm := types.GasMeter{
				Contract:      req.Contract,
				Participant:   participant.Hex(),
				CumulativeGas: cumulativeGas,
			}

			gms = append(gms, gm)
			return nil
		},
	)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryGasMetersResponse{
		GasMeters:  gms,
		Pagination: pageRes,
	}, nil
}

// GasMeter returns a given registered gas meter
func (k Keeper) GasMeter(
	c context.Context,
	req *types.QueryGasMeterRequest,
) (*types.QueryGasMeterResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	if len(req.Contract) == 0 {
		return nil, status.Error(
			codes.InvalidArgument,
			"contract address is empty",
		)
	}

	// check if the contract is a hex address
	if err := ethermint.ValidateAddress(req.Contract); err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid contract address %s", req.Contract).Error(),
		)
	}

	if len(req.Participant) == 0 {
		return nil, status.Error(
			codes.InvalidArgument,
			"participant address is empty",
		)
	}

	// check if the participant is a hex address
	if err := ethermint.ValidateAddress(req.Participant); err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid participant address %s", req.Participant).Error(),
		)
	}

	gm, found := k.GetIncentiveGasMeter(
		ctx,
		common.HexToAddress(req.Contract),
		common.HexToAddress(req.Participant),
	)

	if !found {
		return nil, status.Errorf(
			codes.NotFound,
			"gas meter with contract '%s' and user '%s'",
			req.Contract,
			req.Participant,
		)
	}

	return &types.QueryGasMeterResponse{GasMeter: gm}, nil
}

// Params return hub contract param
func (k Keeper) Params(
	c context.Context,
	_ *types.QueryParamsRequest,
) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	params := k.GetParams(ctx)
	return &types.QueryParamsResponse{Params: params}, nil
}
