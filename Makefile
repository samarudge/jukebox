TARGET=output/
HOSTTYPE=$(shell uname)

ifeq ($(HOSTTYPE), Darwin)
	GLIDE_SOURCE="https://github.com/Masterminds/glide/releases/download/0.9.1/glide-0.9.1-darwin-amd64.tar.gz"
	GLIDE_OUTPUT=darwin-amd64
else
	GLIDE_SOURCE="https://github.com/Masterminds/glide/releases/download/0.9.1/glide-0.9.1-linux-amd64.tar.gz"
	GLIDE_OUTPUT=linux-amd64
endif

all: clean build

build: $(TARGET)/jukebox

$(TARGET)/jukebox: vendor/ views/bindata.go
	go build -v -o ./$(TARGET)/jukebox

views/bindata.go: views/**.html | .tools/go-bindata
	.tools/go-bindata -o ./views/bindata.go -pkg views -ignore "views/bindata.go" views/...

vendor/: glide.lock | .tools/glide
	.tools/glide install
	if [ "${UPDATE_DEPS}" == "1" ]; then .tools/glide up; git diff glide.lock; fi

clean:
	rm -f ./output/jukebox
	rm -Rf ./output/

clean-all: clean
	rm -Rf ./vendor/
	rm -Rf ./.tools/

.tools/:
	mkdir ./.tools/

.tools/glide: .tools/glide-0.9.1/$(GLIDE_OUTPUT)/glide
	cp .tools/glide-0.9.1/$(GLIDE_OUTPUT)/glide .tools/glide
	chmod +x .tools/glide

.tools/glide-0.9.1/$(GLIDE_OUTPUT)/glide: .tools/glide-0.9.1-$(GLIDE_OUTPUT).tar.gz | .tools/glide-0.9.1/
	tar -xvf .tools/glide-0.9.1-$(GLIDE_OUTPUT).tar.gz -C .tools/glide-0.9.1/
	touch .tools/glide-0.9.1/$(GLIDE_OUTPUT)/glide # Touch is due to tar preserving timestamps

.tools/glide-0.9.1-$(GLIDE_OUTPUT).tar.gz: | .tools/
	curl -L -o ./.tools/glide-0.9.1-$(GLIDE_OUTPUT).tar.gz $(GLIDE_SOURCE)

.tools/glide-0.9.1/: | .tools/
	mkdir .tools/glide-0.9.1/

.tools/go-bindata: | .tools/
	cd vendor/github.com/jteeuwen/go-bindata/go-bindata && go install -v
	cp ${GOPATH}/bin/go-bindata .tools/go-bindata
