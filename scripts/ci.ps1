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

echo "running go vet..."
echo "-------------------"
echo ""

go vet ./...
if ($LASTEXITCODE -ne 0) {
  ExitWithCode -exitcode $LASTEXITCODE
}

echo "generating fakes..."
echo "-------------------"
echo ""

go generate ./...
if ($LASTEXITCODE -ne 0) {
  ExitWithCode -exitcode $LASTEXITCODE
}

echo "running tests..."
echo "----------------"
echo ""
go test -v -race . ./fixtures/...
if ($LASTEXITCODE -ne 0) {
  ExitWithCode -exitcode $LASTEXITCODE
}
go test -v ./arguments/ ./command/ ./generator/ ./integration/
if ($LASTEXITCODE -ne 0) {
  ExitWithCode -exitcode $LASTEXITCODE
}

echo "Windows test suite was a 'sweet' success"
