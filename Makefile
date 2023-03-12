default: all

version=$(shell ver=$$(git log -n 1 --pretty=oneline --format=%D | awk -F, '{print $$1}' | awk '{print $$3}'); \
	if [ "$$ver" = "master" ] ; then \
	ver="master($$(git log -n 1 --pretty=oneline --format=%h))" ; \
	fi ; \
	echo $$ver)

client: 
	mkdir -p build
	go build -ldflags "-X main.version=${version}" ./cmd/ck-client 
	mv ck-client* ./build

server: 
	mkdir -p build
	go build -ldflags "-X main.version=${version}" ./cmd/ck-server
	mv ck-server* ./build

ovpn-plugin: 
	mkdir -p build
	go build -buildmode=c-archive -ldflags "-X main.version=${version}" ./cmd/ck-ovpn-plugin
	mv libck-ovpn-plugin* ./build

install:
	mv build/libck-* /usr/local/bin

all: client server ovpn-plugin

clean:
	rm -rf ./build/libck-*
