.PHONY: help
help: # show list of all commands
	@grep -E '^[a-zA-Z_-]+:.*?# .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?# "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: todo
todo: # show list of all todos left in code
	@rg 'TODO' --glob '**/*.go' || echo 'All done!'

.PHONY: run
run: # Run emotes data collection pipeline
	go run main.go
