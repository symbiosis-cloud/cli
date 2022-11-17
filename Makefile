BINARY_NAME=sym

build:
	GOARCH=amd64 GOOS=darwin go build -o dist/${BINARY_NAME}-darwin main.go
	GOARCH=amd64 GOOS=linux go build -o dist/${BINARY_NAME}-linux main.go
	GOARCH=amd64 GOOS=window go build -o dist/${BINARY_NAME}-windows main.go

run:
	./${BINARY_NAME}

build_and_run: build run

clean:
	go clean
	rm dist/${BINARY_NAME}-darwin
	rm dist/${BINARY_NAME}-linux
	rm dist/${BINARY_NAME}-windows