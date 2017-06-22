GO=go

COMPILE=$(GO) build
LINT=golint

PBCC=protoc

BUILD_FLAGS=-x -v
LINT_FLAGS=-set_exit_status

EXE=beacon-api
MAIN=$(./main.go)

SRC_DIR=./beacon

GO_SRC=$(wildcard $(SRC_DIR)/**/*.go)

INTERCHANGE_DIR=$(SRC_DIR)/interchange
INTERCHANGE_SRC=$(wildcard $(INTERCHANGE_DIR)/*.proto)
INTERCHANGE_OBJ=$(patsubst %.proto,%.pb.go,$(INTERCHANGE_SRC))

all: lint $(EXE)

$(EXE): $(INTERCHANGE_OBJ) $(GO_SRC) $(MAIN)
	$(COMPILE) $(BUILD_FLAGS) -o $(EXE) $(MAIN)

$(INTERCHANGE_OBJ): $(INTERCHANGE_SRC)
	$(PBCC) -I$(INTERCHANGE_DIR) --go_out=$(INTERCHANGE_DIR) $(INTERCHANGE_SRC)

lint: $(GO_SRC) $(MAIN)
	$(LINT) $(LINT_FLAGS) $(SRC_DIR)/...

test:
	$(GO) test -v -p=1 -covermode=atomic $(SRC_DIR)/...

clean:
	rm -rf $(EXE)
	rm -rf $(INTERCHANGE_OBJ)
