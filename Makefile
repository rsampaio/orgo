GOFILES=conf.go db.go web.go work.go dropbox.go google.go main.go
GOTESTFILES=*.go

orgo: $(GOFILES)
	go build -v -o $@ $^

test: $(GOTESTFILES)
	go test -v -cover -race $^
