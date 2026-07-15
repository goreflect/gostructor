# Library modules: the core plus each opt-in source/adapter and the shared
# snapshot store.
LIB_MODULES := . yaml toml hocon vault snapshot git consul etcd springcloud watch

# Runnable examples, each its own module (they pull in the heavier adapter deps
# and so are kept out of the library builds consumers care about).
EXAMPLE_MODULES := examples/multisource examples/hotreload-file examples/consul \
	examples/etcd examples/vault examples/git examples/springcloud

MODULES := $(LIB_MODULES) $(EXAMPLE_MODULES)

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
