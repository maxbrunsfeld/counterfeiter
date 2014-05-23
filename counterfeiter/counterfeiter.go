package counterfeiter

import (
	"github.com/maxbrunsfeld/counterfeiter/generator"
	"github.com/maxbrunsfeld/counterfeiter/locator"
)

func Generate(sourceDir, interfaceName, fakePackageName, fakeName string) (string, error) {
	interfaceNode, err := locator.GetInterface(interfaceName, sourceDir)
	if err != nil {
		return "", err
	}

	return generator.GenerateFake(fakeName, fakePackageName, interfaceNode)
}
