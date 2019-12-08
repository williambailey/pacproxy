NAME = $(shell awk -F\" '/^const Name/ { print $$2 }' ./pacproxy.go)
VERSION = $(shell awk -F\" '/^const Version/ { print $$2 }' ./pacproxy.go)

all: test build

build:
	@mkdir -p bin/
	go build -trimpath -race -o bin/$(NAME)

test:
	go test -race ./...

xcompile:
	@rm -rf build/
	for dist in $$(go tool dist list); do \
		a=($$(echo "$${dist}" | tr "/" "\n")); \
		export GOOS="$${a[0]}"; \
		export GOARCH="$${a[1]}"; \
		export EXT=$$([ "$${GOOS}" == "windows" ] && echo ".exe" || echo ""); \
		go build -v -x -trimpath -o "./build/$(NAME)_$(VERSION)_$${GOOS}_$${GOARCH}/$(NAME)$${EXT}"; \
	done

package: xcompile
	$(eval FILES := $(shell ls build))
	@mkdir -p build/tgz
	for f in $(FILES); do \
		(cd $(shell pwd)/build && tar -zcvf tgz/$$f.tar.gz $$f); \
		echo $$f; \
	done

clean:
	@rm -rf bin/
	@rm -rf build/

favicon:
	echo 'package main' > favicon.go
	xxd -p favicon.ico | tr -d '\n' | sed 's/.\{2\}/\\x&/g' | sed 's/.*/var faviconIco = []byte("&")/g' >> favicon.go

ab: build
	killall pacproxy || true
	./bin/pacproxy -v -c "function FindProxyForURL(url, host){ return 'DIRECT'; }" -l 127.0.0.1:8080 2>proxy.8080.log &
	./bin/pacproxy -v -c "function FindProxyForURL(url, host){ return 'PROXY 127.0.0.1:8080'; }" -l 127.0.0.1:8081 2>proxy.8081.log &
	./bin/pacproxy -v -c "function FindProxyForURL(url, host){ return 'PROXY 127.0.0.1:8081'; }" -l 127.0.0.1:8082 2>proxy.8082.log &
	./bin/pacproxy -v -c "function FindProxyForURL(url, host){ return 'PROXY 127.0.0.1:8082'; }" -l 127.0.0.1:8083 2>proxy.8083.log &
	ps auxww | grep "./bin/pacproxy"
	sleep 2
	ab -X 127.0.0.1:8083 -k -c 100 -n 30000 "http://127.0.0.1:8080/favicon.ico"
	killall pacproxy

.PHONY: all build test xcompile package clean favicon ab
