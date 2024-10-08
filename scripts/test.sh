# check if go.mod exists in current directory
if [ ! -f go.mod ]; then
    echo "go.mod not found"
    echo "Please make sure you are in the root sentinel directory"
    exit 1
fi

# check if test-env.sh exists in scripts directory
if [ ! -f scripts/test-env.sh ]; then
    echo "scripts/test-env.sh not found"
    echo "Please make sure you are in the root sentinel directory"
    exit 1
fi

. scripts/test-env.sh
go test ./... -race -covermode=atomic -coverprofile=coverage.out
go tool cover -html coverage.out -o coverage.html