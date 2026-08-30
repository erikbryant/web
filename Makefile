.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: vuln
vuln:
	govulncheck ./...

.PHONY: test
test:
	go test -race -cover ./...

.PHONY: verify
verify:
	go mod verify

.PHONY: check
check: fmt vet vuln test verify
