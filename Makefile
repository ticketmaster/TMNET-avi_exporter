OUT = aviexporter

default:
	$(MAKE) clean
	$(MAKE) build-all
clean:
	rm -rf target
build-darwin:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -a -tags netgo -ldflags '-w' -o target/darwin/$(OUT)
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo -ldflags '-w' -o target/linux/$(OUT)
build-all:
	$(MAKE) build-darwin
	$(MAKE) build-linux
all:
	$(MAKE) clean
	$(MAKE) build-all
