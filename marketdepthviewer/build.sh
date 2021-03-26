#!/bin/bash
CGO_LDFLAGS_ALLOW="-Wl,-Bsymbolic-functions" go build ./marketviewer.go

