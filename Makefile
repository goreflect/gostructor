MODULES := . yaml toml hocon vault examples/multisource

.PHONY: build vet test tidy all

build:
	@for m in $(MODULES); do \
		echo "==> $$m"; \
		(cd $$m && go build ./...) || exit 1; \
	done

vet:
	@for m in $(MODULES); do \
		echo "==> $$m"; \
		(cd $$m && go vet ./...) || exit 1; \
	done

test:
	@for m in $(MODULES); do \
		echo "==> $$m"; \
		(cd $$m && go test ./... -race -cover) || exit 1; \
	done

tidy:
	@for m in $(MODULES); do \
		echo "==> $$m"; \
		(cd $$m && go mod tidy) || exit 1; \
	done

all: build vet test
