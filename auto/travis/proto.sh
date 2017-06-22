#!/bin/bash

protoc -I./beacon/interchange --go_out=./beacon/interchange ./beacon/interchange/*.proto
