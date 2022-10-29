# Go Substrate Code Generator

A tool that generates boilerplate code for substrate-based chains (calls, storage access, events, etc) using [`go-substrate-rpc-client`](github.com/centrifuge/go-substrate-rpc-client).

This uses the metadata Json RPC provided by substrate-based chains, which you can get by doing:
```
curl -L -X POST -H "Content-Type: application/json" -d '{"id":1, "jsonrpc":"2.0", "method": "state_getMetadata"}' https://rpc.polkadot.io > polkadot-meta.json
```
Just replace `rpc.polkadot.io` with your own server.
### Installation
Clone the repo, run `go install ./...` and make sure it's on your path.


### Using `go generate`

Make a submodule in your go project. Place a `metadata.json` and the following `mymodule.go` file in it
```
//go:generate go-substrate-gen metadata.json "github.com/my/package/mymodule"

package mymodule
```

Then run `go generate ./...` and it'll generate code for each pallet in the `mymodule` directory.

### Calling manually
```
go-substrate-gen meta.json "github.com/my/package/submodule/for/code" 
```

### Getting Metadata
There is code included under `json-gen` to fetch a human-readable version of the json from a locally running substrate node in dev mode.
View [the readme](json-gen/README.md) for instructions.

### Architecture and Design
Please read the [documentation](doc/architecture.md), which describes the received metadata's structure, and also gives an overview of the code's structure.

### Calling code
Example for `pallet_balances`

```golang
// Transfer some liquid free balance to another account.
//
// `transfer` will set the `FreeBalance` of the sender and receiver.
// If the sender's account is below the existential deposit as a result
// of the transfer, the account will be reaped.
//
// The dispatch origin for this call must be `Signed` by the transactor.
//
// # <weight>
// - Dependent on arguments but not critical, given proper implementations for input config
//   types. See related functions below.
// - It contains a limited number of reads and writes internally and no complex
//   computation.
//
// Related functions:
//
//   - `ensure_can_withdraw` is always called internally but has a bounded complexity.
//   - Transferring balances to accounts that did not exist before will cause
//     `T::OnNewAccount::on_new_account` to be called.
//   - Removing enough funds from an account will trigger `T::DustRemoval::on_unbalanced`.
//   - `transfer_keep_alive` works the same way as `transfer`, but has an additional check
//     that the transfer will not kill the origin account.
// ---------------------------------
// - Origin account is already in memory, so no DB operations for them.
// # </weight>
func MakeTransferCall(dest0 *MultiAddress, value1 *types.UCompact) (types.Call, error) {...}
...
```

### Storage code

```golang
// Make a storage key for Account
//  The Balances pallet example of storing the balance of an account.
//
//  # Example
//
//  
//   impl pallet_balances::Config for Runtime {
//     type AccountStore = StorageMapShim<Self::Account<Runtime>, frame_system::Provider<Runtime>, AccountId, Self::AccountData<Balance>>
//   }
//  
//
//  You can also store the balance of an account in the `System` pallet.
//
//  # Example
//
//  ```nocompile
//   impl pallet_balances::Config for Runtime {
//    type AccountStore = System
//   }
//  ```
//
//  But this comes with tradeoffs, storing account balances in the system pallet stores
//  `frame_system` data alongside the account data contrary to storing account balances in the
//  `Balances` pallet, which uses a `StorageMap` to store balances data only.
//  NOTE: This is only used in the case that this pallet is used to store balances.
func MakeAccountStorageKey(byteArray0 [32]byte) (types.StorageKey, error) {...}

func GetAccount(state *state.State, bhash types.Hash, byteArray0 [32]byte) (ret AccountData, err error) {...}
...
```

### Types

```golang
// Generated pallet_balances_AccountData
type AccountData struct {
	// Field 0 with TypeId=6
	Free types.U128
	// Field 1 with TypeId=6
	Reserved types.U128
	// Field 2 with TypeId=6
	MiscFrozen types.U128
	// Field 3 with TypeId=6
	FeeFrozen types.U128
}

// Generated SpRuntimeMultiaddressMultiAddress with id=188
type MultiAddress struct {
	IsId              bool
	AsIdField0        [32]byte
	IsIndex           bool
	AsIndexField0     struct{}
	IsRaw             bool
	AsRawField0       []byte
	IsAddress32       bool
	AsAddress32Field0 [32]byte
	IsAddress20       bool
	AsAddress20Field0 [20]byte
}

func (ty MultiAddress) Encode(encoder scale.Encoder) (err error) {...}

func (ty *MultiAddress) Decode(decoder scale.Decoder) (err error) {...}
```

