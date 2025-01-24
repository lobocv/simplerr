

test:
	go test ./... -v -count=1 -race -coverprofile coverage_tmp.out && cat coverage_tmp.out | grep -v ping > coverage.out

lint:
	golangci-lint run

view-coverage: test
	go tool cover -html=coverage.out