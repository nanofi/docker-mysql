.SILENT :
.PHONY : docker-mysql clean fmt

TAG:=`git describe --abbrev=0 --tags`
LDFLAGS:=-X main.buildVersion $(TAG)

all: docker-mysql

deps:
	go get github.com/mitchellh/gox
	go get github.com/fsouza/go-dockerclient

tool:
	gox -build-toolchain -os "darwin linux"

test:
	go test -v

docker-mysql:
	echo "Building docker-mysql"
	go build -ldflags "$(LDFLAGS)"

dist-clean:
	rm -rf dist

dist: dist-clean
	gox -os "darwin linux" -output "dist/{{.Dir}}-{{.OS}}-{{.Arch}}"

release: dist
	ls dist | xargs -I {} tar -cvzf {}-$(TAG).tar.gz dist/{}
