package generate_defaults

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate -o fakes -fake-name-template "The{{.TargetName}}Imposter"

//counterfeiter:generate . Sing
type Sing interface {
	Sing() string
}

//counterfeiter:generate -o other-fakes . Sang
type Sang interface {
	Sang() string
}

//counterfeiter:generate -fake-name NotTheRealSong . Song
type Song interface {
	Song() string
}

//counterfeiter:generate -o other-fakes -fake-name Ponger . Pong
type Pong interface {
	Pong() string
}
