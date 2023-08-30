package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/aphoh/go-substrate-gen/metadata"
	"github.com/aphoh/go-substrate-gen/palletgen"
	"github.com/aphoh/go-substrate-gen/typegen"
	"github.com/aphoh/go-substrate-gen/versiongen"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

const VERSION = "0.8.1"

func main() {
	if err := run(); err != nil {
		fmt.Printf("%v\n", err.Error())
		os.Exit(1)
	}
}

func run() error {
	var jsonPath string
	var extPkgPath string
	var retVersion bool
	var retHelp bool
	var chainVersionPath string

	flag.StringVar(&jsonPath, "json-path", "", "The path to the scale-encoded metadata.")
	flag.StringVar(&extPkgPath, "ext-pkg-path", "", "The fully-qualified external package path.")
	flag.BoolVar(&retVersion, "version", false, "Print the version of go-substrate-gen and return.")
	flag.BoolVar(&retHelp, "help", false, "Print the help for go-substrate-gen and return.")
	flag.StringVar(&chainVersionPath, "chain-version-path", "", "The path to the json blob containing the version of the substrate chain.")

	flag.Parse()

	if retVersion {
		fmt.Printf("go-substrate-gen version %s\n", VERSION)
		return nil
	}
	if retHelp {
		flag.Usage()
		return nil
	}

	argIdx := 0
	if jsonPath == "" {
		jsonPath = flag.Arg(argIdx)
		argIdx += 1
		if jsonPath == "" {
			return fmt.Errorf("json-path is mandatory as either a positional or named argument")
		}
	}

	if extPkgPath == "" {
		extPkgPath = flag.Arg(argIdx)
		argIdx += 1
		if extPkgPath == "" {
			return fmt.Errorf("ext-pkg-path is mandatory as either a positional or named argument")
		}
	}

	// Parse metadata
	raw, err := ioutil.ReadFile(jsonPath)
	if err != nil {
		return fmt.Errorf("error reading json: %v", err.Error())
	}
	// go-substrate-rpc-client parsed metadata
	meta, encResp, err := metadata.ParseMetadata(raw)
	if err != nil {
		return fmt.Errorf("error parsing metadata: %v", err.Error())
	}

	// Parse chain version
	var chainVersion *types.RuntimeVersion
	if chainVersionPath != "" {
		cvRaw, err := ioutil.ReadFile(chainVersionPath)
		if err != nil {
			return fmt.Errorf("error reading chain version json: %v", err.Error())
		}
		chainVersion, err = metadata.ParseVersion(cvRaw)
		if err != nil {
			return fmt.Errorf("error parsing chain version json: %v", err.Error())
		}
	}

	// structure:
	// (OPTIONAL) ./version/version.go
	// ./types/types.go
	// ./pallets/$PALLET/storage.go
	// ./pallets/$PALLET/calls.go

	if chainVersion != nil {
		versionPath := path.Join(extPkgPath, "/version")
		versDir := filepath.Join(".", "version")
		os.MkdirAll(versDir, os.ModePerm)
		versionGen := versiongen.NewVersionGenerator(versionPath, *chainVersion)
		versFp := filepath.Join(versDir, "version.go")
		typesGenerated, err := versionGen.Generate()
		if err != nil {
			return fmt.Errorf("error parsing version information: %v", err.Error())
		}
		ioutil.WriteFile(versFp, []byte(typesGenerated), 0644)
	}

	typesPath := path.Join(extPkgPath, "/types")
	tg := typegen.NewTypeGenerator(meta, encResp, typesPath)

	for _, pallet := range meta.Pallets {
		lowerName := strings.ToLower(string(pallet.Name))
		palletPath := path.Join(extPkgPath, "/"+lowerName)
		pg := palletgen.NewPalletGenerator(&pallet, &tg)

		fp := filepath.Join(".", lowerName)
		err = os.MkdirAll(fp, os.ModePerm)
		if err != nil {
			return fmt.Errorf("error creating pallet %v path: %v", pallet.Name, err)
		}

		storage, isSome, err := pg.GenerateStorage(palletPath)
		if err != nil {
			return fmt.Errorf("error generating storage for pallet %v: %v", pallet.Name, err)
		}
		if isSome {
			err = ioutil.WriteFile(filepath.Join(fp, "storage.go"), []byte(storage), 0644)
			if err != nil {
				return fmt.Errorf("error writing storage.go for pallet %v: %v", pallet.Name, err)
			}
		}

		calls, isSome, err := pg.GenerateCalls(palletPath)
		if err != nil {
			return fmt.Errorf("error generating calls for pallet %v: %v", pallet.Name, err)
		}
		if isSome {
			err = ioutil.WriteFile(filepath.Join(fp, "calls.go"), []byte(calls), 0644)
			if err != nil {
				return fmt.Errorf("error writing calls.go for pallet %v: %v", pallet.Name, err)
			}
		}
	}
	err = tg.GenerateCallHelpers()
	if err != nil {
		return err
	}

	typesDir := filepath.Join(".", "types")
	os.MkdirAll(typesDir, os.ModePerm)
	typesFp := filepath.Join(typesDir, "types.go")
	typesGenerated := tg.GetGenerated()
	ioutil.WriteFile(typesFp, []byte(typesGenerated), 0644)

	return nil
}
