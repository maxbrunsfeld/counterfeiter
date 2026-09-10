# `counterfeiter` [![GitHub Actions](https://github.com/maxbrunsfeld/counterfeiter/actions/workflows/go.yml/badge.svg)](https://github.com/maxbrunsfeld/counterfeiter/actions/workflows/go.yml) [![Go Report Card](https://goreportcard.com/badge/github.com/maxbrunsfeld/counterfeiter/v6)](https://goreportcard.com/report/github.com/maxbrunsfeld/counterfeiter/v6) [![GoDoc](https://godoc.org/github.com/maxbrunsfeld/counterfeiter/v6?status.svg)](https://godoc.org/github.com/maxbrunsfeld/counterfeiter/v6)

When writing unit-tests for an object, it is often useful to have fake implementations
of the object's collaborators. In go, such fake implementations cannot be generated
automatically at runtime, and writing them by hand can be quite arduous.

`counterfeiter` allows you to simply generate test doubles for a given interface.

### Supported Versions Of `go`

`counterfeiter` follows the [support policy of `go` itself](https://golang.org/doc/devel/release.html#policy):

> Each major Go release is supported until there are two newer major releases. For example, Go 1.5 was supported until the Go 1.7 release, and Go 1.6 was supported until the Go 1.8 release. We fix critical problems, including [critical security problems](https://golang.org/security), in supported releases as needed by issuing minor revisions (for example, Go 1.6.1, Go 1.6.2, and so on).

If you are having problems with `counterfeiter` and are not using a supported version of go, please update to use a supported version of go before opening an issue.

### Using `counterfeiter`

⚠️ Please use [`go modules`](https://blog.golang.org/using-go-modules) when working with counterfeiter.

Typically, `counterfeiter` is used in `go generate` directives. It can be frustrating when you change your interface declaration and suddenly all of your generated code is suddenly out-of-date. The best practice here is to use the [`go generate` command](https://blog.golang.org/generate) to make it easier to keep your test doubles up to date.

⚠️ If you are working with go 1.23 or earlier, please refer to an [older version of this README](https://github.com/maxbrunsfeld/counterfeiter/blob/e39cbe6aaa94a0b6718cf3d413cd5319c3a1f6fa/README.md#using-counterfeiter), as the instructions below assume go 1.24 (which added `go tool` support) and later.

#### Step 1 - Add `counterfeiter` as a tool dependency

Establish a tool dependency on counterfeiter by running the following command:

```shell
go get -tool github.com/maxbrunsfeld/counterfeiter/v6
```

#### Step 2a - Add `go:generate` Directives

You can add directives right next to your interface definitions (or not), in any `.go` file in your module.

```shell
$ cat myinterface.go
```

```go
package foo

//go:generate go tool counterfeiter . MySpecialInterface

type MySpecialInterface interface {
	DoThings(string, uint64) (int, error)
}
```

```shell
$ go generate ./...
Writing `FakeMySpecialInterface` to `foofakes/fake_my_special_interface.go`... Done
```

#### Step 2b - Add `counterfeiter:generate` Directives

If you plan to have many directives in a single package, consider using this
option, as it will speed things up considerably. You can add directives right
next to your interface definitions (or not), in any `.go` file in your module.

```shell
$ cat myinterface.go
```

```go
package foo

// You only need **one** of these per package!
//go:generate go tool counterfeiter -generate

// You will add lots of directives like these in the same package...
//counterfeiter:generate . MySpecialInterface
type MySpecialInterface interface {
	DoThings(string, uint64) (int, error)
}

// Like this...
//counterfeiter:generate . MyOtherInterface
type MyOtherInterface interface {
	DoOtherThings(string, uint64) (int, error)
}
```

```shell
$ go generate ./...
Writing `FakeMySpecialInterface` to `foofakes/fake_my_special_interface.go`... Done
Writing `FakeMyOtherInterface` to `foofakes/fake_my_other_interface.go`... Done
```

#### Step 3 - Run `go generate`

You can run `go generate` in the directory with your directive, or in the root of your module (to ensure you generate for all packages in your module):

```shell
$ go generate ./...
```

#### Invoking `counterfeiter` from the shell

You can use the following command to invoke `counterfeiter` from within a go module:

```shell
$ go tool counterfeiter

USAGE
	counterfeiter
		[-generate>] [-o <output-path>] [-p] [--fake-name <fake-name>]
		[-fake-name-template <template>] [-header <header-file>] [-q] [-test]
		[<source-path>] <interface> [-]
```

#### Installing `counterfeiter` to `$GOPATH/bin`

This is unnecessary if you're using the approach described above, but does allow you to invoke `counterfeiter` in your shell _outside_ of a module:

```shell
$ go install github.com/maxbrunsfeld/counterfeiter/v6
$ ~/go/bin/counterfeiter

USAGE
	counterfeiter
		[-generate>] [-o <output-path>] [-p] [--fake-name <fake-name>]
		[-fake-name-template <template>] [-header <header-file>] [-q] [-test]
		[<source-path>] <interface> [-]
```

### Generating Test Doubles

Given a path to a package and an interface name, you can generate a test double.

```shell
$ cat path/to/foo/file.go
```

```go
package foo

type MySpecialInterface interface {
		DoThings(string, uint64) (int, error)
}
```

```shell
$ go tool counterfeiter path/to/foo MySpecialInterface
Wrote `FakeMySpecialInterface` to `path/to/foo/foofakes/fake_my_special_interface.go`
```

#### Naming the fakes package

The package clause of a fake follows the package that already lives in the output directory, and is named after the directory only when there is none yet. So to name the package differently from its directory, say `impl_fakes` in `fakes/`, add a file declaring that package first:

```go
// fakes/doc.go
package impl_fakes
```

```go
//counterfeiter:generate -o fakes . MyInterface
```

#### Generating a test double into the interface's own package

By default the fake lives in a sibling `<package>fakes` package, which cannot be imported by tests inside `<package>` itself (it would be an import cycle). If you want to use a fake from a white-box test in the same package, point `-o` at the interface's own directory:

```go
//counterfeiter:generate -o . . MySpecialInterface
```

When the output directory is the directory of the package that declares the interface, `counterfeiter` generates the fake as a member of that package: it does not import the package, refers to its types unqualified, and can fake unexported interfaces too. `-o` may also name a file in that directory, for example `-o fake_my_special_interface_test.go` to keep the fake out of the non-test build.

#### Generating a test double into the external test package

If your tests are black-box tests in `<package>_test`, `-test` generates the fake into that external test package instead, as a `_test.go` file in the current directory, next to the tests that use it:

```go
//counterfeiter:generate -test . MySpecialInterface
//counterfeiter:generate -test ../otherpackage OtherInterface
//counterfeiter:generate -test io.WriteCloser
```

The fake is then only compiled for tests, and the tests use it unqualified (`&FakeMySpecialInterface{}`). The interface's package is imported as usual, so the interface must be exported. This works for interfaces from the package itself, from other packages in your module, from the standard library and from third-party modules, and the fakes sit with the tests rather than in a fakes package next to each interface. With `-o <dir>` the fake goes into the external test package of that directory.

#### Setting defaults for every directive

Flags given alongside `-generate` on the `//go:generate` line are the defaults for every `//counterfeiter:generate` directive in the package: `-o`, `-header`, `-q`, `-test` and `-fake-name-template`. A directive's own flags take precedence. `-fake-name-template` is a Go `text/template` in which `{{.TargetName}}` is the name of the interface being faked (first letter upper-cased); `-fake-name` on a directive still wins over it. So if you would rather keep all of a package's fakes in a `fake` package, named after their interfaces, you can write that once:

```go
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate -o fake -fake-name-template '{{.TargetName}}'

//counterfeiter:generate . MyRepository
//counterfeiter:generate . MyPresenter
```

```shell
$ go generate ./...
Writing `MyRepository` to `fake/my_repository.go`... Done
Writing `MyPresenter` to `fake/my_presenter.go`... Done
```

### Using Test Doubles In Your Tests

Instantiate fakes:

```go
import "my-repo/path/to/foo/foofakes"

var fake = &foofakes.FakeMySpecialInterface{}
```

Fakes record the arguments they were called with:

```go
fake.DoThings("stuff", 5)

Expect(fake.DoThingsCallCount()).To(Equal(1))

str, num := fake.DoThingsArgsForCall(0)
Expect(str).To(Equal("stuff"))
Expect(num).To(Equal(uint64(5)))
```

You can stub their return values:

```go
fake.DoThingsReturns(3, errors.New("the-error"))

num, err := fake.DoThings("stuff", 5)
Expect(num).To(Equal(3))
Expect(err).To(Equal(errors.New("the-error")))
```

For more examples of using the `counterfeiter` API, look at [some of the provided examples](https://github.com/maxbrunsfeld/counterfeiter/blob/master/generated_fakes_test.go).

### Generating Test Doubles For Third Party Interfaces

For third party interfaces, you can specify the interface using the alternative syntax `<package>.<interface>`, for example:

```shell
$ go tool counterfeiter github.com/go-redis/redis.Pipeliner
```

### Running The Tests For `counterfeiter`

If you want to run the tests for `counterfeiter` (perhaps, because you want to contribute a PR), all you have to do is run `scripts/ci.sh`.

### Contributions

So you want to contribute to `counterfeiter`! That's great, here's exactly what you should do:

- open a new github issue, describing your problem, or use case
- help us understand how you want to fix or extend `counterfeiter`
- write one or more unit tests for the behavior you want
- write the simplest code you can for the feature you're working on
- try to find any opportunities to refactor
- avoid writing code that isn't covered by unit tests

`counterfeiter` has a few high level goals for contributors to keep in mind

- keep unit-level test coverage as high as possible
- keep `main.go` as simple as possible
- avoid making the command line options any more complicated
- avoid making the internals of `counterfeiter` any more complicated

If you have any questions about how to contribute, rest assured that @tjarratt and other maintainers will work with you to ensure we make `counterfeiter` better, together. This project has largely been maintained by the community, and we greatly appreciate any PR (whether big or small).

### License

`counterfeiter` is MIT-licensed.
