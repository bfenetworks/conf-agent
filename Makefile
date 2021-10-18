# init project path
HOMEDIR := $(shell pwd)
OUTDIR  := $(HOMEDIR)/output

# init command params
GO      := $(GO_1_16_BIN)go
GOPATH  := $(shell $(GO) env GOPATH)
GOMOD   := $(GO) mod
GOBUILD := $(GO) build
GOTEST  := $(GO) test -gcflags="-N -l"
GOPKGS  := $$($(GO) list ./...| grep -vE "vendor")

# test cover files
COVPROF := $(HOMEDIR)/covprof.out  # coverage profile
COVFUNC := $(HOMEDIR)/covfunc.txt  # coverage profile information for each function
COVHTML := $(HOMEDIR)/covhtml.html # HTML representation of coverage profile

# make, make all
all: prepare compile test package


#make prepare, download dependencies
prepare: gomod

gomod: 
	$(GOMOD) download

#make compile
compile: build

build: prepare
	$(GOBUILD) -o $(HOMEDIR)/conf-agent

# make test, test your code
test: prepare test-case
test-case:
	$(GOTEST) -v -cover $(GOPKGS)

# make package
package: package-bin
package-bin:
	mkdir -p 		$(OUTDIR)
	cp -rf  conf 	$(OUTDIR)/
	cp -rf  docs 	$(OUTDIR)/
	mv conf-agent  	$(OUTDIR)/

# make clean
clean:
	$(GO) clean
	rm -rf $(OUTDIR)
	rm -rf $(HOMEDIR)/conf-agent

# avoid filename conflict and speed up build 
.PHONY: all prepare compile test package clean build
