
lint:
	golint .
test:
	go test -v
bench:
	go test -bench=.

fmt:
	@echo "fmt code"
	find . -name "*.go" | xargs goimports -w
