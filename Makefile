GOOS=linux 
GOARCH=amd64
GO111MODULE=on

build:
	@go build -v ./...
buildforce:
	@go build -a -v ./...
test:
	@go test -cover ./...
testverbose:
	@go test -v -cover ./...
dep:
	@go get
tidy:
	go mod tidy
play:
	go test -v -run TestLek