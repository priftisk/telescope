.PHONY: run build

run: build
	./bin/telescope.o

build:
	CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o ./bin/telescope.o