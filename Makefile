GIT_COMMIT = $(shell git rev-parse --short HEAD)
GIT_STATUS = $(shell test -n "`git status --porcelain`" && echo "+CHANGES")

.PHONY: build binary

build: docker
	docker run dlgrab cat /dlgrab/bin/dlgrab > dlgrab
	chmod +x dlgrab

docker:
	docker build -t dlgrab .

binary:
	go build -a -ldflags "-X main.GITCOMMIT $(GIT_COMMIT)$(GIT_STATUS)" -o ./bin/dlgrab
