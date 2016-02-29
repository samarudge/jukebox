TARGET="./output/"
HOSTTYPE=$(shell uname)

ifeq ($(HOSTTYPE), Darwin)
	GLIDE_SOURCE="https://github.com/Masterminds/glide/releases/download/0.9.1/glide-0.9.1-darwin-amd64.tar.gz"
	GLIDE_OUTPUT="darwin-amd64"
else
	GLIDE_SOURCE="https://github.com/Masterminds/glide/releases/download/0.9.1/glide-0.9.1-linux-amd64.tar.gz"
	GLIDE_OUTPUT="linux-amd64"
endif

all: build

build: vendor/
	go build -o ./$(TARGET)/jukebox

vendor/: .tools/glide
	.tools/glide install

clean:
	rm -f ./output/jukebox
	rm -Rf ./output/
	rm -Rf ./vendor/
	rm -Rf ./.tools/

.tools:
	mkdir ./.tools/

.tools/glide: .tools/glide-0.9.1/$(GLIDE_OUTPUT)/glide
	cp .tools/glide-0.9.1/$(GLIDE_OUTPUT)/glide .tools/glide
	chmod +x .tools/glide

.tools/glide-0.9.1/$(GLIDE_OUTPUT)/glide: .tools/glide-0.9.1-$(GLIDE_OUTPUT).tar.gz .tools/glide-0.9.1
	tar -xvf .tools/glide-0.9.1-$(GLIDE_OUTPUT).tar.gz -C .tools/glide-0.9.1/

.tools/glide-0.9.1-$(GLIDE_OUTPUT).tar.gz: .tools
	curl -L -o ./.tools/glide-0.9.1-$(GLIDE_OUTPUT).tar.gz $(GLIDE_SOURCE)

.tools/glide-0.9.1: .tools/
	mkdir ./.tools/glide-0.9.1/
