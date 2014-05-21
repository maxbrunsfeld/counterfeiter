package counterfeiter

import (
	"github.com/maxbrunsfeld/counterfeiter/generator"
	"github.com/maxbrunsfeld/counterfeiter/locator"
)

func Generate(packageName, interfaceName string, fakePackageName string) (string, error) {
	interfaceNode, err := locator.GetInterface(interfaceName, packageName)
	if err != nil {
		return "", err
	}

	return generator.GenerateFake("Fake"+interfaceName, fakePackageName, interfaceNode)
}
