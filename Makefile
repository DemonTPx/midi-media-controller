VERSION=0.1

.PHONY: all clean

all: build/midi-media-controller

clean:
	rm build/midi-media-controller

build/midi-media-controller:
	go build -o build/midi-media-controller .
	strip build/midi-media-controller
	upx build/midi-media-controller

build-deb: build/midi-media-controller
	mkdir -p build/deb/DEBIAN build/deb/usr/local/bin
	sed 's/%VERSION%/0.1/' debian/control > build/deb/DEBIAN/control
	cp build/midi-media-controller build/deb/usr/local/bin/
	dpkg-deb --build build/deb build/midi-media-controller_${VERSION}_amd64.deb
