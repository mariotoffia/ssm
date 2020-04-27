GOOS=linux 
GOARCH=amd64
GO111MODULE=on

build:
	@go build ./...
test:
	@go test -cover ./...
testverbose:
	@go test -v -cover ./...
dep:
	@go get
tidy:
	go mod tidy