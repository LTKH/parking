PROJECTNAME=$(shell basename "$(PWD)")

# Go переменные.
GOBASE=$(shell pwd)
GOFILES=$(wildcard *.go)

GOOS=windows
GOARCH=amd64
CGO_ENABLED=1
CGO_CFLAGS=-D_WIN32_WINNT=0x0400
CGO_CXXFLAGS=-D_WIN32_WINNT=0x0400

exec:
	go build -v --mod=vendor '--ldflags=-v -s -w' -o build/parking-windows-amd64.exe ./

test: