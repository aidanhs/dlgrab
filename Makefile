GIT_COMMIT = $(shell git rev-parse --short HEAD)
GIT_STATUS = $(shell test -n "`git status --porcelain`" && echo "+CHANGES")

.PHONY: build binary

build: docker
	docker run --rm dlgrab cat /dlgrab/bin/dlgrab > dlgrab
	chmod +x dlgrab

docker:
	docker build --force-rm -t dlgrab .

binary:
	go build -a -ldflags "-X main.GITCOMMIT $(GIT_COMMIT)$(GIT_STATUS)" -o ./bin/dlgrab
