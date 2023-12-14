package keeper

import (
	context "context"

	"cosmossdk.io/x/circuit/types"
)

func (k *Keeper) ExportGenesis(ctx context.Context) (data *types.GenesisState) {
	var (
		permissions  []*types.GenesisAccountPermissions
		disabledMsgs map[string]*types.FilteredUrl = make(map[string]*types.FilteredUrl)
	)

	err := k.Permissions.Walk(ctx, nil, func(address []byte, perm types.Permissions) (stop bool, err error) {
		add, err := k.addressCodec.BytesToString(address)
		if err != nil {
			return true, err
		}
		// Convert the Permissions struct to a GenesisAccountPermissions struct
		// and add it to the permissions slice
		permissions = append(permissions, &types.GenesisAccountPermissions{
			Address:     add,
			Permissions: &perm,
		})
		return false, nil
	})
	if err != nil {
		panic(err)
	}

	err = k.DisableList.Walk(ctx, nil, func(key string, value types.FilteredUrl) (stop bool, err error) {
		disabledMsgs[key] = &value
		return false, nil
	})
	if err != nil {
		panic(err)
	}

	return &types.GenesisState{
		AccountPermissions: permissions,
		DisabledTypeUrls:   disabledMsgs,
	}
}

// InitGenesis initializes the circuit module's state from a given genesis state.
func (k *Keeper) InitGenesis(ctx context.Context, genState *types.GenesisState) {
	for _, accounts := range genState.AccountPermissions {
		add, err := k.addressCodec.StringToBytes(accounts.Address)
		if err != nil {
			panic(err)
		}

		// Set the permissions for the account
		if err := k.Permissions.Set(ctx, add, *accounts.Permissions); err != nil {
			panic(err)
		}
	}
	for url, value := range genState.DisabledTypeUrls {
		// Set the disabled type urls
		if err := k.DisableList.Set(ctx, url, *value); err != nil {
			panic(err)
		}
	}
}
