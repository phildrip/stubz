#!/bin/bash
# Script to regenerate golden test files and examples

set -e  # Exit on error

echo "Building toe..."
go build

echo "Regenerating golden test files..."
./toe -o testdata/golden/stubs/stub_myinterface.go testdata/input/simple MyInterface
./toe -o testdata/golden/stubs/stub_genericinterface.go testdata/input/generic GenericInterface
./toe --stub-dir customstubs -o testdata/golden/customstubs/stub_myinterface.go testdata/input/simple MyInterface
./toe -o testdata/golden/stubs/stub_aliasinterface.go testdata/input/alias AliasInterface
./toe -o testdata/golden/stubs/stub_externalinterface.go testdata/input/external ExternalInterface

echo "Regenerating example stubs..."
./toe -o examples/calculator/stubs/stub_calculator.go examples/calculator/lib Calculator

echo "Done! All golden files and examples regenerated."

