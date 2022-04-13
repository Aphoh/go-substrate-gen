package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/aphoh/go-substrate-gen/metadata"
	"github.com/aphoh/go-substrate-gen/palletgen"
	"github.com/aphoh/go-substrate-gen/typegen"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("%v\n", err.Error())
		os.Exit(1)
	}
}

func run() error {
	args := os.Args[1:]
	if len(args) < 2 {
		return fmt.Errorf("Expected two arguments (json path, package name)")
	}
	jsonPath := args[0]
	extPkgPath := args[1]

	// Parse metadata
	raw, err := ioutil.ReadFile(jsonPath)
	if err != nil {
		return fmt.Errorf("Error reading json: %v", err.Error())
	}
	meta, encResp, err := metadata.ParseMetadata(raw)
	if err != nil {
		return fmt.Errorf("Error parsing metadata: %v", err.Error())
	}
	// structure:
	// ./types/types.go
	// ./pallets/$PALLET/storage.go
	// ./pallets/$PALLET/calls.go

	typesPath := path.Join(extPkgPath, "/types")
	tg := typegen.NewTypeGenerator(meta, encResp, typesPath)

	for _, pallet := range meta.Pallets {
		lowerName := strings.ToLower(string(pallet.Name))
		palletPath := path.Join(extPkgPath, "/"+lowerName)
		pg := palletgen.NewPalletGenerator(&pallet, &tg)

		fp := filepath.Join(".", lowerName)
		err = os.MkdirAll(fp, os.ModePerm)
		if err != nil {
			return fmt.Errorf("Error creating pallet %v path: %v", pallet.Name, err)
		}

		storage, isSome, err := pg.GenerateStorage(palletPath)
		if err != nil {
			return fmt.Errorf("Error generating storage for pallet %v: %v", pallet.Name, err)
		}
		if isSome {
			err = ioutil.WriteFile(filepath.Join(fp, "storage.go"), []byte(storage), 0644)
			if err != nil {
				return fmt.Errorf("Error writing storage.go for pallet %v: %v", pallet.Name, err)
			}
		}

		calls, isSome, err := pg.GenerateCalls(palletPath)
		if err != nil {
			return fmt.Errorf("Error generating calls for pallet %v: %v", pallet.Name, err)
		}
		if isSome {
			err = ioutil.WriteFile(filepath.Join(fp, "calls.go"), []byte(calls), 0644)
			if err != nil {
				return fmt.Errorf("Error writing calls.go for pallet %v: %v", pallet.Name, err)
			}
		}
	}

	typesDir := filepath.Join(".", "types")
	os.MkdirAll(typesDir, os.ModePerm)
	typesFp := filepath.Join(typesDir, "types.go")
	typesGenerated := tg.GetGenerated()
	ioutil.WriteFile(typesFp, []byte(typesGenerated), 0644)

	return nil
}
