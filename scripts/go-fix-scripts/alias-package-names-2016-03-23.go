package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		println("usage: aliased-package-names path/to/package/to/fix")
		os.Exit(1)
	}

	packageToFix := os.Args[1]
	filepath.Walk(packageToFix, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		astFile, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.AllErrors)
		if err != nil {
			println(fmt.Sprintf("unexpected error: '%s'", err.Error()))
			os.Exit(2)
		}

		if astFile.Name.Name == "fakes" {
			migrateOldVersionFake(path)
		}

		return nil
	})
}

func migrateOldVersionFake(path string) {
	path, _ = filepath.Abs(path)
	fakesDir := filepath.Dir(path)
	_, packageBeingFaked := filepath.Split(filepath.Dir(fakesDir))

	newFakesDir := filepath.Join(filepath.Dir(fakesDir), packageBeingFaked+"fakes")

	_ = os.Mkdir(newFakesDir, os.ModePerm)

	newPath := filepath.Join(newFakesDir, filepath.Base(path))
	err := os.Rename(path, newPath)
	if err != nil {
		println(fmt.Sprintf("Unexpected error renaming file: %s", err.Error()))
		os.Exit(3)
	}

	bytes, err := ioutil.ReadFile(newPath)
	if err != nil {
		println(fmt.Sprintf("unexpected error reading migrated file: %s", err.Error()))
		os.Exit(4)
	}

	oldPackageDecl := "package fakes"
	newPackageDecl := "package " + packageBeingFaked + "fakes"
	newContents := strings.Replace(string(bytes), oldPackageDecl, newPackageDecl, 1)

	err = ioutil.WriteFile(newPath, []byte(newContents), 0)
	if err != nil {
		println(fmt.Sprintf("unexpected error writing new file: %s", err.Error()))
		os.Exit(5)
	}

	println(fmt.Sprintf("migrated %s to %s", filepath.Base(path), newPath))
}
