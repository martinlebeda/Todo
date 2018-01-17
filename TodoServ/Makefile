TAG := $(shell date +%Y%m%d)
GOPATH := $(CURDIR)/libraries
GOROOT := /opt/go-1.9.2
PATH := $(GOROOT)/bin:$(PATH)


build:
	go build -ldflags "-s -w" -o TodoServe ./

dep:
	go get github.com/gin-gonic/gin
	go get github.com/satori/go.uuid
	go get gopkg.in/russross/blackfriday.v2

test:
	go test
# upx

# README.md: README.mds
#	mdpreproc < README.mds > README.md

install: build test # README.md
	cp -v TodoServe ~/bin/

#release: install
#	git tag $(TAG)
#	git push origin --tags
#	cp -v -v BioRythm ~/Dropbox/Martin/Projects/BioRythm/BioRythm.Linux.x64.$(TAG)
