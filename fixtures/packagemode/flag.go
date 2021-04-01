package packagemode

import "flag"

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate -p github.com/maxbrunsfeld/counterfeiter/v6/fixtures/packagemode
//counterfeiter:generate -o flagcustomfakesdir -p github.com/maxbrunsfeld/counterfeiter/v6/fixtures/packagemode

func Arg(arg1 int) string {
	return flag.Arg(arg1)
}

func Args() []string {
	return flag.Args()
}

func Bool(arg1 string, arg2 bool, arg3 string) *bool {
	return flag.Bool(arg1, arg2, arg3)
}

func BoolVar(arg1 *bool, arg2 string, arg3 bool, arg4 string) {
	flag.BoolVar(arg1, arg2, arg3, arg4)
}
