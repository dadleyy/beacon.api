GO=go

COMPILE=$(GO) build
VERSION_PACKAGE=github.com/dadleyy/beacon.api/beacon/version
LDFLAGS="-X $(VERSION_PACKAGE).Semver=$(VERSION) -s -w"
BUILD_FLAGS=-x -v -ldflags $(LDFLAGS)
PBCC=protoc

GLIDE=glide
VENDOR_DIR=vendor

LINT=golint
LINT_FLAGS=-set_exit_status

EXE=beacon-api
MAIN=$(wildcard ./main.go)

VET=$(GO) vet
VET_FLAGS=-x

SRC_DIR=./beacon
GO_SRC=$(wildcard $(MAIN) $(SRC_DIR)/**/*.go)

INTERCHANGE_DIR=$(SRC_DIR)/interchange
INTERCHANGE_SRC=$(wildcard $(INTERCHANGE_DIR)/*.proto)
INTERCHANGE_OBJ=$(patsubst %.proto,%.pb.go,$(INTERCHANGE_SRC))

MAX_TEST_CONCURRENCY=10
TEST_FLAGS=-covermode=atomic -coverprofile={{.Dir}}/.coverprofile -p=$(MAX_TEST_CONCURRENCY)
TEST_LIST_FMT='{{if len .TestGoFiles}}"go test $(SRC_DIR)/{{.Name}} $(TEST_FLAGS)"{{end}}'

all: $(EXE)

$(EXE): $(VENDOR_DIR) $(INTERCHANGE_OBJ) $(GO_SRC)
	$(COMPILE) $(BUILD_FLAGS) -o $(EXE) $(MAIN)

$(INTERCHANGE_OBJ): $(INTERCHANGE_SRC)
	$(GO) generate $(SRC_DIR)/...

lint: $(GO_SRC)
	$(LINT) $(LINT_FLAGS) $(shell $(GO) list $(SRC_DIR)/... | grep -v 'interchange')
	$(LINT) $(LINT_FLAGS) $(MAIN)

test: $(GO_SRC) $(VENDOR_DIR) $(INTERCHANGE_OBJ) lint
	$(VET) $(VET_FLAGS) $(SRC_DIR)/...
	$(VET) $(VET_FLAGS) $(MAIN)
	$(GO) list -f $(TEST_LIST_FMT) $(SRC_DIR)/... | xargs -L 1 sh -c

$(VENDOR_DIR):
	go get -u github.com/Masterminds/glide
	go get -u github.com/golang/protobuf/protoc-gen-go
	go get -u github.com/golang/lint/golint
	$(GLIDE) install

clean:
	rm -rf $(VENDOR_DIR)
	rm -rf $(EXE)
	rm -rf $(INTERCHANGE_OBJ)
	$(GO) list -f '"rm -f {{.Dir}}/.coverprofile"' $(SRC_DIR)/... | xargs -L 1 sh -c
