PROJECTNAME=$(shell basename "$(PWD)")

# Go переменные.
GOBASE=$(shell pwd)
GOFILES=$(wildcard *.go)

exec:
	go build .

test: