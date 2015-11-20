github-orgs-repos-build:
	go build -o github-orgs-repos github-orgs-repos.go

docker-github-orgs-repos-build:
	docker run --rm -v "$$PWD":/usr/src/myapp -v "$$GOPATH":/go -w /usr/src/myapp -e GOOS=darwin -e GOARCH=amd64 golang:1.3-cross make github-orgs-repos-build
