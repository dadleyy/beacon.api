GO=go

COMPILE=$(GO) build
BUILD_FLAGS=-x -v
PBCC=protoc

GLIDE=glide
VENDOR_DIR=vendor

LINT=golint
LINT_FLAGS=-set_exit_status
LINT_RESULT=.lint-results

EXE=beacon-api
MAIN=$(wildcard ./main.go)

COVERAGE=goverage
COVERAGE_REPORT=coverage.out

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

test: $(GO_SRC) $(COVERAGE_REPORT)
	$(GO) vet $(shell go list ./... | grep -vi 'vendor\|testing')

$(COVERAGE_REPORT):
	$(COVERAGE) -v -parallel=1 -covermode=atomic -coverprofile=$(COVERAGE_REPORT) $(shell $(GO) list $(SRC_DIR)/...)

$(VENDOR_DIR):
	go get -u github.com/Masterminds/glide
	go get -u github.com/golang/protobuf/protoc-gen-go
	go get -u github.com/golang/lint/golint
	go get -u github.com/haya14busa/goverage
	$(GLIDE) install

clean:
	rm -rf $(COVERAGE_REPORT)
	rm -rf $(LINT_RESULT)
	rm -rf $(VENDOR_DIR)
	rm -rf $(EXE)
	rm -rf $(INTERCHANGE_OBJ)
