build:
	go mod download
	CGO_ENABLED=0 go build -ldflags "-X main.version=`git describe --tags`" -o sasha

default: build

upgrade:
	go get -u -v
	go mod download
	go mod tidy
	go mod verify

run:
	./sasha

clean:
	go clean
	go mod tidy
	rm -f sasha

install:
	mkdir -p $(DESTDIR)/usr/local/bin
	cp sasha $(DESTDIR)/usr/local/bin/sasha
	chmod 755 $(DESTDIR)/usr/local/bin/sasha

uninstall:
	rm -f $(DESTDIR)/usr/local/bin/sasha
