#!/bin/bash
go mod tidy -compat=1.17 && go mod vendor && make build && make install

