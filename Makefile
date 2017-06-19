BINARY=swift-backup

all:	swift-backup

swift-backup: main.go
	go build .	

clean:
	rm -f $(BINARY)
	rm -rf build/
	rm -rf archive/

xcompile: main.go
	$(GOPATH)/bin/gox -osarch="linux/amd64 linux/386 linux/arm windows/386 windows/amd64 darwin/386 darwin/amd64" -output="build/{{.OS}}/{{.Dir}}-{{.Arch}}"
	lipo -create build/darwin/$(BINARY)-386 build/darwin/$(BINARY)-amd64 -output build/darwin/${BINARY}

zip: clean xcompile
	mkdir archive
	zip -j archive/mac.zip build/darwin/${BINARY}
	mv build/windows/garl-backup-amd64.exe build/windows/garl-backup-64bit.exe
	mv build/windows/garl-backup-386.exe build/windows/garl-backup-32bit.exe
	zip -j archive/windows.zip build/windows/*.exe
	zip -j archive/linux.zip build/linux/${BINARY}-*
