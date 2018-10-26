SHELL := /bin/bash
PACKAGE := "github.com/Laughs-In-Flowers/countfloyd"
CLIENTPATH := ${PWD}
CLIENTSRC := $(shell find . -type f -name '*.go' -not -path "./vendor/*")
CLIENT := countfloyd
SERVERPATH := ${PWD}/lib/cfs
SERVERSRC := $(shell find ${SERVERPATH} -type f -name '*.go' -not -path "./vendor/*") 
SERVER := cfs
BIN := ${PWD}/bin
VERSIONTAG := "0.0.2"
VERSIONHASH := `git rev-parse HEAD`
VERSIONDATE := `date -u +%d-%m-%Y.%H:%M:%S`
LDFLAGS = -ldflags "-X=main.versionTag=$(VERSIONTAG) -X=main.versionHash=$(VERSIONHASH) -X=main.versionDate=$(VERSIONDATE)"

.PHONY: all build clean install uninstall fmt simplify check run

all: check install

$(CLIENT): $(CLIENTSRC)
	@go build $(LDFLAGS) -o $(BIN)/$(CLIENT)

$(SERVER): $(SERVERSRC)
	@go build $(LDFLAGS) -o $(BIN)/$(SERVER) ${SERVERSRC}

build: $(CLIENT) $(SERVER) 
	@true

clean:
	@rm -f $(BIN)/$(CLIENT)
	@rm -f $(BIN)/$(SERVER)	

#install:
#	@go install $(LDFLAGS)

install-binary:
	@cp $(BIN)/$(CLIENT) $(GOPATH)/bin/$(CLIENT) 
	@cp $(BIN)/$(SERVER) $(GOPATH)/bin/$(SERVER)

uninstall: clean
	@rm -f $$(which ${CLIENT})
	@rm -f $$(which ${SERVER})

fmt:
	@gofmt -l -w $(CLIENTSRC)
	@gofmt -l -w $(SERVERSRC)

simplify:
	@gofmt -s -l -w $(CLIENTSRC)
	@gofmt -s -l -w $(SERVERSRC)	

check:
	@test -z $(shell gofmt -l main.go | tee /dev/stderr) || echo "[WARN] Fix formatting issues with 'make fmt'"
	@for d in $$(go list ./... | grep -v /vendor/); do golint $${d}; done
	@go tool vet ${CLIENTSRC}

#run: install
#	@$(CLIENT)
#	@$(SERVER)
