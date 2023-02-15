
VERSION := $(shell git describe --tags --dirty --always)

DATE := $(shell date -u +%d/%m/%Y)

main:
	go install -ldflags="-X 'main.gitversion=$(VERSION)' -X 'main.builddate=$(DATE)'" github.com/dalzilio/hue/cmd/hsimplify
	go install -ldflags="-X 'main.gitversion=$(VERSION)' -X 'main.builddate=$(DATE)'" github.com/dalzilio/hue/cmd/hwalk

