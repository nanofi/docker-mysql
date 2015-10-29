.SILENT :
.PHONY : docker-mysql clean fmt

OS:=linux
GLIDE_VERSION:=0.6.1
GLIDE_ZIP:=glide-$(OS)-$(GLIDE_VERSION).zip
TAG:=`git describe --abbrev=0 --tags`
LDFLAGS:=-X main.buildVersion=$(TAG)

all: docker-mysql

$(GLIDE_ZIP):
	wget https://github.com/Masterminds/glide/releases/download/$(GLIDE_VERSION)/glide-$(OS)-amd64.zip -O $(GLIDE_ZIP)

glide: $(GLIDE_ZIP)
	unzip $(GLIDE_ZIP) $(OS)-amd64/glide
	mv $(OS)-amd64/glide .
	rm -rf $(OS)-amd64

deps: glide
	go get github.com/mitchellh/gox
	./glide install

test:
	go test -v

docker-mysql:
	echo "Building docker-mysql"
	go build -ldflags "$(LDFLAGS)"

dist-clean:
	rm -rf dist

dist: dist-clean
	gox -ldflags "$(LDFLAGS)" -os "darwin linux" -output "dist/{{.Dir}}-{{.OS}}-{{.Arch}}"

release: deps dist
	ls dist | xargs -I {} tar -cvzf {}-$(TAG).tar.gz dist/{}
