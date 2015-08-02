GIT_COMMIT = $(shell git rev-parse --short HEAD)
GIT_STATUS = $(shell test -n "`git status --porcelain`" && echo "+CHANGES")

.PHONY: build binary

build: docker
	docker run --rm dlgrab cat /dlgrab/bin/dlgrab > dlgrab
	chmod +x dlgrab

docker:
	docker build --force-rm -t dlgrab .

CHECKDIFF = [ $$(cat diff | wc -l) -gt 0 ]

check:
	gofmt -d -e -s . >diff 2>&1 || true
	$(CHECKDIFF) && echo "go fmt failed" && cat diff && exit 1 || true
	go tool vet -all -printf=false . >diff 2>&1 || true
	$(CHECKDIFF) && echo "go vet failed" && cat diff && exit 1 || true
	go tool fix -force -diff . >diff 2>&1 || true
	$(CHECKDIFF) && echo "go fix failed" && cat diff && exit 1 || true
	rm diff

binary:
	go build -a -ldflags "-X main.GITCOMMIT $(GIT_COMMIT)$(GIT_STATUS)" -o ./bin/dlgrab
