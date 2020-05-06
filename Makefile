GOOS=linux 
GOARCH=amd64
GO111MODULE=on

build:
	@go build -v ./...
buildforce:
	@go build -a -v ./...
test:
	@go test -cover ./...
testverbose: testclean
	@go test -v -cover ./...
testclean:
	@go clean -testcache
testremoteclean:
	@go test -v -run TestCleanAll -scope=clean
dep:
	@go get -v
	@go mod graph
tidy:
	@go mod tidy
play:
	@go test -v -run TestLek