PROJECTNAME=$(shell basename "$(PWD)")

# Go переменные.
GOBASE=$(shell pwd)
GOFILES=$(wildcard *.go)

#CGO_ENABLED=1 
#CGO_CFLAGS=-D_WIN32_WINNT=0x0400 
#CGO_CXXFLAGS=-D_WIN32_WINNT=0x0400 

exec:
	go get github.com/webview/webview
	GOOS=windows GOARCH=amd64 go build -o /tmp/parking.exe parking.go

test:
	go version