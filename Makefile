build:
	GOOS=linux CGO_ENABLED=0 go build -ldflags="-s -w -extldflags -static" -o jt cmd/jt/main.go