echo "========================="
echo "windows build is starting"
echo "========================="

function ExitWithCode
{
    param
    (
        $exitcode
    )

    $host.SetShouldExit($exitcode)
    exit
}

echo "getting test dependencies..."
echo "----------------------------"
echo ""

go get -t .
if ($LASTEXITCODE -ne 0) {
  ExitWithCode -exitcode $LASTEXITCODE
}

echo "building counterfeiter..."
echo "-------------------------"
echo ""
go build github.com/maxbrunsfeld/counterfeiter
if ($LASTEXITCODE -ne 0) {
  ExitWithCode -exitcode $LASTEXITCODE
}

remove-item ./counterfeiter.exe

echo "generating fakes..."
echo "-------------------"
echo ""
set-alias counterfeiter counterfeiter.exe
go generate ./...
if ($LASTEXITCODE -ne 0) {
  ExitWithCode -exitcode $LASTEXITCODE
}

echo "running tests..."
echo "----------------"
echo ""
$env:CGO_ENABLED = "0"
go test -v -race ./...
if ($LASTEXITCODE -ne 0) {
  ExitWithCode -exitcode $LASTEXITCODE
}

echo "Windows test suite was a 'sweet' success"
