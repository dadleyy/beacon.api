GO=go

COMPILE=$(GO) build
VERSION_PACKAGE=github.com/dadleyy/beacon.api/beacon/version
LDFLAGS="-X $(VERSION_PACKAGE).Semver=$(VERSION)"
BUILD_FLAGS=-x -v -ldflags $(LDFLAGS)
PBCC=protoc

GLIDE=glide
VENDOR_DIR=vendor

LINT=golint
LINT_FLAGS=-set_exit_status
LINT_RESULT=.lint-results

EXE=beacon-api
MAIN=$(wildcard ./main.go)

SRC_DIR=./beacon
GO_SRC=$(wildcard $(MAIN) $(SRC_DIR)/**/*.go)

INTERCHANGE_DIR=$(SRC_DIR)/interchange
INTERCHANGE_SRC=$(wildcard $(INTERCHANGE_DIR)/*.proto)
INTERCHANGE_OBJ=$(patsubst %.proto,%.pb.go,$(INTERCHANGE_SRC))

all: $(EXE)

$(EXE): $(VENDOR_DIR) $(INTERCHANGE_OBJ) $(GO_SRC) $(LINT_RESULT)
	$(COMPILE) $(BUILD_FLAGS) -o $(EXE) $(MAIN)

$(INTERCHANGE_OBJ): $(INTERCHANGE_SRC)
	$(PBCC) -I$(INTERCHANGE_DIR) --go_out=$(INTERCHANGE_DIR) $(INTERCHANGE_SRC)

$(LINT_RESULT): $(GO_SRC)
	$(LINT) $(LINT_FLAGS) $(shell $(GO) list $(SRC_DIR)/... | grep -v 'interchange')
	touch $(LINT_RESULT)

test: $(GO_SRC)
	$(GO) vet $(shell go list ./... | grep -vi 'vendor\|testing')
	$(GO) test -v -p=1 -covermode=atomic ./beacon/...

$(VENDOR_DIR):
	go get -u github.com/Masterminds/glide
	go get -u github.com/golang/protobuf/protoc-gen-go
	go get -u github.com/golang/lint/golint
	$(GLIDE) install

clean:
	rm -rf $(LINT_RESULT)
	rm -rf $(VENDOR_DIR)
	rm -rf $(EXE)
	rm -rf $(INTERCHANGE_OBJ)
