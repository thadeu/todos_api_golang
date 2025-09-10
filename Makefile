.PHONY: test

test:
	gotest || bin/gotest

test-watch:
	bin/gotest && gotest --watch || bin/gotest --watch

test-cover:
	gocover || bin/gocover

dev:
	air -c .air.toml