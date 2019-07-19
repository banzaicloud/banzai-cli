# A Self-Documenting Makefile: http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html

NODE_VERSION = 11.10.0

.PHONY: build
build: node_modules/.auto ## Build form client
ifneq (${IGNORE_NODE_VERSION_REQ}, 1)
	@printf "${NODE_VERSION}\n$$(node --version | awk '{sub(/^v/, "", $$1);print $$1}')" | sort -t '.' -k 1,1 -k 2,2 -k 3,3 -g | head -1 | grep -q -E "^${NODE_VERSION}$$" || (printf "Required Node version is v${NODE_VERSION}\nInstalled: `node --version`\n" && exit 1)
endif

	npm run build -- --prod

node_modules/.auto: package-lock.json # install/update dependencies if needed
	npm ci
	@touch $@

.PHONY: list
list: ## List all make targets
	@${MAKE} -pRrn : -f $(MAKEFILE_LIST) 2>/dev/null | awk -v RS= -F: '/^# File/,/^# Finished Make data base/ {if ($$1 !~ "^[#.]") {print $$1}}' | egrep -v -e '^[^[:alnum:]]' -e '^$@$$' | sort

.PHONY: help
.DEFAULT_GOAL := help
help:
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

# Variable outputting/exporting rules
var-%: ; @echo $($*)
varexport-%: ; @echo $*=$($*)
