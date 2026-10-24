.PHONY: help
help: # show list of all commands
	@grep -E '^[a-zA-Z_-]+:.*?# .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?# "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: todo
todo: # show list of all todos left in code
	@rg 'TODO' --glob '**/*.go' || echo 'All done!'

.PHONY: run
run: # Run emotes data collection pipeline
	go run main.go

.PHONY: collect_raw
collect_raw: # Crawl BTTV search API for all 3-char queries into `_<query>.json` files
	go run ./collect_raw

.PHONY: status
status: # See how many queries have completed
	@find . -maxdepth 1 -name '_???.json' | wc -l

.PHONY: save_empty
save_empty: # Save list of empty query results as `empty_list.txt`
	grep '{}' . -rl | rg "^\./_['\w]{3}\.json$$" | awk '{print substr($$0,3)}' > empty_list.txt

.PHONY: clear
clear: # Clear query results
	for line in $$(cat empty_list.txt); do rm -f "$$line"; done

.PHONY: union
union: # Unite all query results in *.json files into `data.json`
	go run ./union

.PHONY: collect
collect: # Run emotes data collection pipeline
	$(MAKE) collect_raw
	$(MAKE) save_empty
	$(MAKE) union
	$(MAKE) clear
