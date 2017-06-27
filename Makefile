deps:
	go get github.com/google/go-github/github
	go get golang.org/x/oauth2

build:deps
	go build main.go

install:deps
	go install
