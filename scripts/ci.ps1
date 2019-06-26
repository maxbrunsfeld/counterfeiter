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

echo "installing counterfeiter..."
echo "---------------------------"
echo ""
go install .
if ($LASTEXITCODE -ne 0) {
  ExitWithCode -exitcode $LASTEXITCODE
}
set-alias counterfeiter counterfeiter.exe

echo "generating fakes..."
echo "-------------------"
echo ""

go generate ./...
if ($LASTEXITCODE -ne 0) {
  ExitWithCode -exitcode $LASTEXITCODE
}

echo "ensuring generated fakes compile..."
echo "-----------------------------------"
echo ""
go build -v ./...
if ($LASTEXITCODE -ne 0) {
  ExitWithCode -exitcode $LASTEXITCODE
}

echo "running tests..."
echo "----------------"
echo ""
go test -v -race ./...
if ($LASTEXITCODE -ne 0) {
  ExitWithCode -exitcode $LASTEXITCODE
}

echo "Windows test suite was a 'sweet' success"
