#TAG := $(shell date +%Y%m%d)
#GOPATH := $(CURDIR)/libraries
#GOROOT := /opt/go-1.9.2
#PATH := $(GOROOT)/bin:$(PATH)

#dep:
#	go get github.com/gin-gonic/gin

build:
# 	chrome.exe --pack-extension=$(CURRDIR) --pack-extension-key=$(CURRDIR)/../TodoChrome.pem
# 	google-chrome --headless --pack-extension=/home/martin/Projects/Todo/TodoChrome --pack-extension-key=/home/martin/Projects/Todo/TodoChrome.pem
	chromium-browser --headless --pack-extension=/home/martin/Projects/Todo/TodoChrome --pack-extension-key=/home/martin/Projects/Todo/TodoChrome.pem

# upx

# README.md: README.mds
#	mdpreproc < README.mds > README.md

install: build # README.md
	cp -v TodoServe ~/bin/

#release: install
#	git tag $(TAG)
#	git push origin --tags
#	cp -v -v BioRythm ~/Dropbox/Martin/Projects/BioRythm/BioRythm.Linux.x64.$(TAG)
