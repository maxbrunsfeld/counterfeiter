# counterfeiter template for mockery

[`counterfeiter.templ`](counterfeiter.templ) is a [mockery](https://vektra.github.io/mockery/) template that generates counterfeiter fakes. It produces the same API (`<Method>Stub`, `<Method>CallCount`, `<Method>Calls`, `<Method>ArgsForCall`, `<Method>Args`, `<Method>Returns`, `<Method>ReturnsOnCall`, `Invocations`, `CallOrder`) and, with the configuration below, the same file layout. Tests written against counterfeiter fakes keep working when the fakes come from mockery instead. The generated code itself differs in ways that don't affect the API, listed at the end.

It needs mockery v3.

## Configuration

Reference the template from your `.mockery.yml`, pinned to a counterfeiter release that contains it:

```yaml
template: "https://raw.githubusercontent.com/maxbrunsfeld/counterfeiter/refs/tags/vX.Y.Z/mockery/counterfeiter.templ"
# Lay fakes out the way counterfeiter does by default: one file per
# interface in a sibling <package>fakes package.
dir: "{{.InterfaceDir}}/{{.SrcPackageName}}fakes"
filename: "fake_{{.InterfaceName | snakecase}}.go"
pkgname: "{{.SrcPackageName}}fakes"
structname: "Fake{{.InterfaceName | firstUpper}}"
packages:
  github.com/you/yourmodule/yourpackage:
    interfaces:
      YourInterface: {}
```

mockery fetches the JSON schema for `template-data` from next to the template (`counterfeiter.templ.schema.json`). To use a local copy instead, copy both files into your repository and set `template: "file://path/to/counterfeiter.templ"`.

Then run `mockery`, or add `//go:generate mockery` to a file in your module.

### Fakes in the interface's own package

counterfeiter's `-o .` writes the fake into the package that declares the interface. The equivalent is a package-level override. The `structname` keeps the fake of an unexported interface unexported, as counterfeiter does:

```yaml
packages:
  github.com/you/yourmodule/yourpackage:
    config:
      dir: "{{.InterfaceDir}}"
      pkgname: "{{.SrcPackageName}}"
      structname: "{{if firstIsLower .InterfaceName}}fake{{else}}Fake{{end}}{{.InterfaceName | firstUpper}}"
```

### A header on every fake

`boilerplate-file` replaces counterfeiter's `-header`. It can be set at any level of the config. When several interfaces share one file, the last interface's value wins.

```yaml
template-data:
  boilerplate-file: ./hack/header.go.txt
```

## Differences from counterfeiter

- Only interfaces can be faked. mockery does not generate fakes for function types or whole packages (counterfeiter's `-p`).
- Parameter and result names inside the generated methods are mockery's (the interface's own names, or ones derived from the type when it has none) rather than counterfeiter's `arg1…argN` and `result1…resultN`: `DoThings(s string, v uint64)` instead of `DoThings(arg1 string, arg2 uint64)`. The API is unchanged; the fields of the `<Method>Args` structs are still `Arg1…ArgN`.
- Import statements come from mockery, which writes an alias only when one is needed. counterfeiter also names a package whose name differs from its directory (`hyphenpackage "example.com/go-hyphenpackage"`), and when two imported packages share a name the aliases differ (`sync0` rather than `synca`).
- mockery puts every interface of a package into one file unless `filename` includes the interface name, as in the configuration above. Both layouts work.
