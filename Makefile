.PHONY: build ci default fmt imports init run test vet

default: ci

build:
	GOBIN=$(CURDIR)/bin/ govendor install +local

init:
	go get -u github.com/kardianos/govendor
	go get -u github.com/golang/lint/golint
	go get -u golang.org/x/tools/cmd/goimports
	go get -u golang.org/x/tools/cmd/cover
	go get -u github.com/wadey/gocovmerge

check-missing-deps:
	bash -c 'diff -u <(echo -n) <(govendor list -no-status +missing)'

vet:
	govendor vet +local

test:
	govendor test +local -test.race -cover

fmt:
	govendor fmt +local

imports:
	goimports -w $$(govendor list -p -no-status +local)

cover:
	@rm -rf .cover/
	@mkdir -p .cover/
	@for MOD in $$(govendor list +l | awk '{print $$2}'); do \
		go test -coverpkg=$$MOD \
			-coverprofile=.cover/unit-`echo $$MOD|tr "/" "_"`.out \
			$$MOD 2>&1 | grep -v "no packages being tested depend on"; \
	done

	@gocovmerge .cover/*.out > .cover/all.merged
	@test -f .cover/all.merged && go tool cover -func .cover/all.merged

ci: check-missing-deps vet test build
