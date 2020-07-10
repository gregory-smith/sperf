
.PHONY: build
build:
	go build -o ./bin/sperf ./cmd/...

.PHONY: lint
lint:
	#get this tool from https://golangci-lint.run/usage/install/#local-installation
	#easy mode is:
	#curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.27.0
	golangci-lint run

.PHONY: test 
test:
	go test -v ./...

.PHONY: install
install:
	go install ./cmd/...

.PHONY: deploy
deploy:
	#this is only for local builds, to build a new release in github
	#just push a tag that starts with 'v' example v0.8.0 and the ci server will do the hard work
	#edit .github/workflow/goreleaser.yml to change the deployment to github
	GOOS=linux GOARCH=amd64 go build -o ./bin/sperf ./cmd/...
	go build -o ./bin/sperf ./cmd/...
	go run hack/zip.go ./bin/sperf-linux.zip ./bin/sperf
	GOOS=windows GOARCH=amd64 go build -o ./bin/sperf ./cmd/...
	go run hack/zip.go ./bin/sperf-windows.zip ./bin/sperf
	GOOS=darwin GOARCH=amd64 go build -o ./bin/sperf ./cmd/...
	go run hack/zip.go ./bin/sperf-macos.zip ./bin/sperf

.PHONY: all
all: lint test install deploy
