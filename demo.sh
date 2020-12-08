#!/bin/sh
go build && ./ifcmp ./demodata/README.md ./demodata/gocloak.go GoCloak
