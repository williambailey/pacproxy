NAME = $(shell awk -F\" '/^const Name/ { print $$2 }' ./pacproxy.go)
VERSION = $(shell awk -F\" '/^const Version/ { print $$2 }' ./pacproxy.go)

all: test build

build:
	@mkdir -p bin/
	go build -race -o bin/$(NAME)

test:
	go test -race ./...

xcompile:
	@rm -rf build/
	#GOOS=android GOARCH=arm go build -o ./build/$(NAME)_$(VERSION)_android_arm/$(NAME)
	#GOOS=darwin GOARCH=386 go build -o ./build/$(NAME)_$(VERSION)_darwin_386/$(NAME)
	GOOS=darwin GOARCH=amd64 go build -o ./build/$(NAME)_$(VERSION)_darwin_amd64/$(NAME)
	#GOOS=darwin GOARCH=arm go build -o ./build/$(NAME)_$(VERSION)_darwin_arm/$(NAME)
	#GOOS=darwin GOARCH=arm64 go build -o ./build/$(NAME)_$(VERSION)_darwin_arm64/$(NAME)
	GOOS=dragonfly GOARCH=amd64 go build -o ./build/$(NAME)_$(VERSION)_dragonfly_amd64/$(NAME)
	GOOS=freebsd GOARCH=386 go build -o ./build/$(NAME)_$(VERSION)_freebsd_386/$(NAME)
	GOOS=freebsd GOARCH=amd64 go build -o ./build/$(NAME)_$(VERSION)_freebsd_amd64/$(NAME)
	GOOS=freebsd GOARCH=arm go build -o ./build/$(NAME)_$(VERSION)_freebsd_arm/$(NAME)
	GOOS=linux GOARCH=386 go build -o ./build/$(NAME)_$(VERSION)_linux_386/$(NAME)
	GOOS=linux GOARCH=amd64 go build -o ./build/$(NAME)_$(VERSION)_linux_amd64/$(NAME)
	GOOS=linux GOARCH=arm go build -o ./build/$(NAME)_$(VERSION)_linux_arm/$(NAME)
	GOOS=linux GOARCH=arm64 go build -o ./build/$(NAME)_$(VERSION)_linux_arm64/$(NAME)
	GOOS=linux GOARCH=ppc64 go build -o ./build/$(NAME)_$(VERSION)_linux_ppc64/$(NAME)
	GOOS=linux GOARCH=ppc64le go build -o ./build/$(NAME)_$(VERSION)_linux_ppc64le/$(NAME)
	GOOS=linux GOARCH=mips go build -o ./build/$(NAME)_$(VERSION)_linux_mips/$(NAME)
	GOOS=linux GOARCH=mipsle go build -o ./build/$(NAME)_$(VERSION)_linux_mipsle/$(NAME)
	GOOS=linux GOARCH=mips64 go build -o ./build/$(NAME)_$(VERSION)_linux_mips64/$(NAME)
	GOOS=linux GOARCH=mips64le go build -o ./build/$(NAME)_$(VERSION)_linux_mips64le/$(NAME)
	GOOS=netbsd GOARCH=386 go build -o ./build/$(NAME)_$(VERSION)_netbsd_386/$(NAME)
	GOOS=netbsd GOARCH=amd64 go build -o ./build/$(NAME)_$(VERSION)_netbsd_amd64/$(NAME)
	GOOS=netbsd GOARCH=arm go build -o ./build/$(NAME)_$(VERSION)_netbsd_arm/$(NAME)
	GOOS=openbsd GOARCH=386 go build -o ./build/$(NAME)_$(VERSION)_openbsd_386/$(NAME)
	GOOS=openbsd GOARCH=amd64 go build -o ./build/$(NAME)_$(VERSION)_openbsd_amd64/$(NAME)
	GOOS=openbsd GOARCH=arm go build -o ./build/$(NAME)_$(VERSION)_openbsd_arm/$(NAME)
	GOOS=plan9 GOARCH=386 go build -o ./build/$(NAME)_$(VERSION)_plan9_386/$(NAME)
	GOOS=plan9 GOARCH=amd64 go build -o ./build/$(NAME)_$(VERSION)_plan9_amd64/$(NAME)
	GOOS=solaris GOARCH=amd64 go build -o ./build/$(NAME)_$(VERSION)_solaris_amd64/$(NAME)
	GOOS=windows GOARCH=386 go build -o ./build/$(NAME)_$(VERSION)_windows_386/$(NAME).exe
	GOOS=windows GOARCH=amd64 go build -o ./build/$(NAME)_$(VERSION)_windows_amd64/$(NAME).exe

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
	ab -X 127.0.0.1:8083 -k -c 20 -n 30000 "http://127.0.0.1:8080/favicon.ico"
	killall pacproxy

.PHONY: all build test xcompile package clean favicon ab
