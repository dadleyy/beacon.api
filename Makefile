GO=go
CC=$(GO) build
GOLINT=golint
PBCC=protoc
BUILD_FLAGS=-x -v
LINT_FLAGS=-set_exit_status
EXE=beacon-api
SRC_DIR=./beacon
INTERCHANGE_DIR=$(SRC_DIR)/interchange
INTERCHANGE_SRC=$(wildcard $(INTERCHANGE_DIR)/*.proto)
INTERCHANGE_OBJ=$(patsubst %.proto,%.pb.go,$(INTERCHANGE_SRC))

all: lint $(EXE)

$(EXE): $(INTERCHANGE_OBJ)
	$(CC) $(BUILD_FLAGS) -o $(EXE) main.go

$(INTERCHANGE_OBJ): $(INTERCHANGE_SRC)
	$(PBCC) -I$(INTERCHANGE_DIR) --go_out=$(INTERCHANGE_DIR) $(INTERCHANGE_SRC)

lint: $(INTERCHANGE_SRC)
	$(GOLINT) $(LINT_FLAGS) $(SRC_DIR)/...

test:
	$(GO) test -v -p=1 -covermode=atomic $(SRC_DIR)/...

clean:
	rm -rf $(EXE)
	rm -rf $(INTERCHANGE_OBJ)
