echo "========================="
echo "windows build is starting"
echo "========================="

$ErrorActionPreference = "Stop"

echo "getting test dependencies..."
echo "----------------------------"
echo ""

go get -t .

echo "building counterfeiter..."
echo "-------------------------"
echo ""
go build github.com/maxbrunsfeld/counterfeiter
if (-Not ($LastExitCode -eq 0)) {
    echo "FAILED"
    exit 1
}

remove-item ./counterfeiter.exe

echo "generating fakes..."
echo "-------------------"
echo ""
set-alias counterfeiter counterfeiter.exe
go generate ./...

echo "running tests..."
echo "----------------"
echo ""
go test -v -race ./...
if (-Not ($LastExitCode -eq 0)) {
    echo "FAILED"
    exit 2
}

echo "Windows test suite was a 'sweet' success"
