# `go-substrate-gen` Architecture
## Goal
The overall goal of `go-substrate-gen` is to generate code that exposes methods to access all storage items and generate `go-substrate-rpc-client` calls for all extrinsics.
As all substrate storage items and extrinsics are per-pallet, we generate them per-pallet, and contain them in a directory structure that looks like the below, if `runtime` is the name of the folder the user put all the generated code within
```
runtime/
    types/
        types.go
    pallet1
        calls.go
        storage.go
    pallet2
        calls.go
        storage.go
    ...
```
The user can then call methods within their pallets by importing them from those go files
```golang
import (
    ...
    "github.com/path/to/package/runtime/pallet1"
    "github.com/path/to/package/runtime/types"
    ...
)

// Create calls to use with the rpc client
// Using example pallet with extrinsic 'TransferAll'
call := pallet1.MakeTransferAllCall(..).AsCall()

// Get storage items
// Using storage item named 'Balance'
datatype := pallet1.GetBalanceLatest(..).AsCall()
```
which requires
- a type for every storage item could be retrieved
- an appropriate type for each extrinsic call that could be submitted

`go-substrate-gen` uses [Jennifer](https://github.com/dave/jennifer) to perform the actual code generation. 

## Overview of Code Structure
The overall approach used is very simple:

1. Parse the metadata returned from the `state.getMetadata` RPC endpoint by any substrate chain.
2. Create a type generator which caches generated types
3. For each pallet in the parsed metadata:
    - For each extrinsic in the pallet:
        - Look at all scale types needed, and recursively generate go code to represent them
        - Generate a go struct which contains all information needed for the extrinsic
        - Generate a function to create a call for this extrinsic
    - Write all of the extrinsic functions to `pallet/calls.go`
    - For each storage item in the pallet:
        - Look at all scale types needed, and recursively generate go code to represent them
        - Generate a go struct that contains the storage information in that storage item
        - Generate a function to retrieve the storage information using rpc
    - Write all of the storage item functions to `pallet/storage.go`
4. Write all of the generated types to `types/types.go`

However, there is some complexity involved in the structure of the returned metadata and the translation of scale types to golang.

### Metadata Structure
The human-readable metadata can be viewed in `json-gen/meta.json` and can be made a bit nicer with the `jq` utility.
```shell
cat meta.json | jq > formatted-meta.json
```
#### Types in the Metadata
The metadata first consists of a large array of types.
Each type in the metadata is described according to the [scale codec](https://docs.substrate.io/reference/scale-codec/).

Each element in the array describing the internal structure of every included type, where types reference each other by index in the array (which is the same as their ID). We will go through a small example that should make this more clear.
```json
{
    "id": "0",
    "type": {
        "path": [
        "sp_core",
        "crypto",
        "AccountId32"
        ],
        "params": [],
        "def": {
            "Composite": {
                "fields": [
                {
                    "name": null,
                    "type": "1",
                    "typeName": "[u8; 32]",
                    "docs": []
                }
                ]
            }
        },
        "docs": []
    }
},
```
Above is the type with id 0. The metadata tells us that the type has a path `sp_core::crypto::AccountId32` in Rust. 
It also tells us that it has no generics ("params").
Then, it has the definition, which is a composite type (a struct that can have anonymous fields).
The composite type has one field, unnamed, with type 1, and the type is named `[u8; 32]`. 

It is basically a wrapper around whatever the type with ID 1 is.

In order to finish understanding this type, we need to see which type has ID 1.
```json
{
    "id": "1",
    "type": {
        "path": [],
        "params": [],
        "def": {
        "Array": {
            "len": "32",
            "type": "2"
        }
        },
        "docs": []
    }
},
```
Above is the type with id 1. We can see that this one has no path, which means it is not imported, but defined.
Instead of being a composite type, it is defined as an array of length 32, with the elements being of the type with id 2.

Finally, we can finish our type construction by figuring out what the type with id 2 is.

```json
{
    "id": "2",
    "type": {
        "path": [],
        "params": [],
        "def": {
        "Primitive": "U8"
        },
        "docs": []
    }
},
```
The type with id 2 is a simple primitive U8.
So in total, an `accountId32` contains a 32-byte array.

The entire metadata is structured similarly, with reference-based structures in the types.

#### Pallets in the Metadata
In the metadata, pallets are also listed in an array, with each having some fields. These list the storage and calls of that pallet, and each storage or call will also list the type of storage. For the sake of brevity, those are omitted from this documentation, but they reference the same type IDs from the types array.
```json
"pallets": [
  {
    "name": "pallet_name",
    "storage": {...},
    "calls": {...},
    "events": {...},
    "constants": {...},
    "errors": {...},
    "index": "..."
  },
  ...
]
```

### Generating the Code 
Upon receiving the metadata, it is parsed via the `go-substrate-rpc-client` module, which gives us a structure that follows the exact same format as the JSON metadata described above.

After parsing the metadata, a `TypeGenerator` is instantiated, which will act as a memoized cache of previously generated types. A type is considered "generated" once the code for it has been constructed, and it has been given a unique name.

The pallets within the metadata are then iterated over.
For each one, we create a `PalletGenerator`, which will then call out to a `CallGenerator` and `StorageGenerator` for each respective generation task.
All generators will ask the `TypeGenerator` for a generated type when it runs into a type ID reference in the metadata.
This results in constructing a type being a memoized DFS traversal of the type dependency graph, where the search starts from:
- storage values
- extrinsic arguments
- runtime events
- Errors

In terms of generating the specific go representations corresponding to scale-types, it is well-commented within the code under the `typegen/` directory, where there is a go file corresponding to each scale type.