package locator

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func GetInterface(interfaceName, importPath string) (*ast.InterfaceType, error) {
	importPaths, err := expandPackagePath(importPath)
	if err != nil {
		fmt.Println("Error expanding package paths: ", err)
		os.Exit(1)
	}

	dirPath, err := findDirInGoPath(importPaths[0])
	if err != nil {
		return nil, err
	}
	return findInterface(interfaceName, dirPath)
}

func expandPackagePath(path string) ([]string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return []string{}, err
	}

	cmd := exec.Command("go", "list", path)
	cmd.Dir = cwd
	output, err := cmd.StdoutPipe()
	if err != nil {
		return []string{}, err
	}

	err = cmd.Start()
	if err != nil {
		return []string{}, err
	}

	bytes, err := ioutil.ReadAll(output)
	if err != nil {
		return []string{}, err
	}

	err = cmd.Wait()
	if err != nil {
		return []string{}, err
	}

	return strings.Split(string(bytes), "\n"), nil
}

func findInterface(name, dir string) (*ast.InterfaceType, error) {
	fileSet := token.NewFileSet()
	packages, err := parser.ParseDir(fileSet, dir, nil, parser.AllErrors)
	if err != nil {
		return nil, err
	}

	basename := filepath.Base(dir)
	pkg := packages[basename]
	if pkg == nil {
		return nil, fmt.Errorf("Couldn't find package '%s' in directory", basename)
	}

	var result *ast.InterfaceType
	ast.Inspect(pkg, func(node ast.Node) bool {
		if typeSpec, ok := node.(*ast.TypeSpec); ok {
			if typeSpec.Name.Name == name {
				if interfaceType, ok := typeSpec.Type.(*ast.InterfaceType); ok {
					result = interfaceType
				} else {
					err = fmt.Errorf("Name '%s' does not refer to an interface", name)
				}
				return false
			}
		}
		return true
	})

	if result == nil {
		return nil, fmt.Errorf("Could not find interface '%s'", name)
	}

	return result, err
}

func findDirInGoPath(packageName string) (string, error) {
	gopaths := strings.Split(os.Getenv("GOPATH"), ":")
	for _, gopath := range gopaths {
		path := filepath.Join(gopath, "src", packageName)
		stat, err := os.Stat(path)
		if err == nil && stat.IsDir() {
			return path, nil
		}
	}

	return "", fmt.Errorf("Could not find package '%s'", packageName)
}
