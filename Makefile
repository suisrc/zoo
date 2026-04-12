.PHONY: start build

NOW = $(shell date -u '+%Y%m%d%I%M%S')

# 初始化mod
init:
	go mod init "github.com/suisrc/zoo"

tidy:
	go mod tidy

git:
	@if [ "$(m)" ]; then \
		git add -A && git commit -am "$(m)" && git push; \
	fi
	@if [ "$(t)" ]; then \
		git tag -a $(t) -m "${t}" && git push origin $(t); \
	fi
