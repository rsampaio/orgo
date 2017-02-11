.PHONY: build ci default init test vet

PKGS=$(shell go list ./... | grep -v vendor)
CMDS=$(shell go list ./... | grep -v vendor | grep cmd)

default: ci

build:
	GOBIN=${CURDIR}/bin/ go get -v ${CMDS}

init:
	go get -u github.com/golang/dep/cmd/dep
	go get -u github.com/golang/lint/golint
	go get -u golang.org/x/tools/cmd/cover
	go get -u github.com/wadey/gocovmerge

vet:
	go vet ${PKGS}

test: vet
	go test ${PKGS} -test.race -cover

cover:
	@rm -rf .cover/
	@mkdir -p .cover/
	@for MOD in ${PKGS}; do \
		go test -coverpkg=$$MOD \
			-coverprofile=.cover/unit-`echo $$MOD|tr "/" "_"`.out \
			$$MOD 2>&1 | grep -v "no packages being tested depend on"; \
	done

	@gocovmerge .cover/*.out > .cover/all.merged
	@test -f .cover/all.merged && go tool cover -func .cover/all.merged

ci: vet test build
