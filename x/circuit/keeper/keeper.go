package keeper

import (
	context "context"
	"time"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/store"
	"cosmossdk.io/x/circuit/types"

	"github.com/cosmos/cosmos-sdk/codec"
)

// Keeper defines the circuit module's keeper.
type Keeper struct {
	cdc          codec.BinaryCodec
	storeService store.KVStoreService

	authority []byte

	addressCodec address.Codec

	Schema collections.Schema
	// Permissions contains the permissions for each account
	Permissions collections.Map[[]byte, types.Permissions]
	// DisableList contains the message URLs that are disabled, and associated parameters
	DisableList collections.Map[string, types.FilteredUrl]
}

// NewKeeper constructs a new Circuit Keeper instance
func NewKeeper(cdc codec.BinaryCodec, storeService store.KVStoreService, authority string, addressCodec address.Codec) Keeper {
	auth, err := addressCodec.StringToBytes(authority)
	if err != nil {
		panic(err)
	}

	sb := collections.NewSchemaBuilder(storeService)

	k := Keeper{
		cdc:          cdc,
		storeService: storeService,
		authority:    auth,
		addressCodec: addressCodec,
		Permissions: collections.NewMap(
			sb,
			types.AccountPermissionPrefix,
			"permissions",
			collections.BytesKey,
			codec.CollValue[types.Permissions](cdc),
		),
		DisableList: collections.NewMap(
			sb,
			types.DisableListPrefix,
			"disable_list",
			collections.StringKey,
			codec.CollValue[types.FilteredUrl](cdc),
		),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema

	return k
}

func (k *Keeper) GetAuthority() []byte {
	return k.authority
}

// IsAllowed returns true when msg URL is not found in the DisableList for given context, else false.
func (k *Keeper) IsAllowed(ctx context.Context, blockTime time.Time, msgURL string) (bool, error) {
	filteredURL, err := k.DisableList.Get(ctx, msgURL)
	if err == collections.ErrNotFound {
		// key not found, so the url is implicitly allowed
		return true, nil
	}
	if filteredURL.ExpiresAt > 0 && blockTime.Unix() >= filteredURL.ExpiresAt {
		// tripped circuit has expired so remove
		return true, k.DisableList.Remove(ctx, msgURL)
	}
	// TODO: check BypassSet
	return false, nil
}
