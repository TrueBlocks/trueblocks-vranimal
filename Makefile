INSTALL_DIR = $(HOME)/source
BINARIES = vrml-fmt vrml-validate vrml-serialize
CGO_BINARIES = viewer
CGO_CFLAGS = -I/opt/homebrew/include
CGO_LDFLAGS = -L/opt/homebrew/lib
MSG ?= update

.PHONY: build clean clobber test add commit push

build:
	@for bin in $(BINARIES); do \
		echo "Building $$bin..."; \
		go build -o $(INSTALL_DIR)/$$bin ./cmd/$$bin/.; \
	done
	@for bin in $(CGO_BINARIES); do \
		echo "Building $$bin (CGO)..."; \
		CGO_CFLAGS="$(CGO_CFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS)" \
			go build -o $(INSTALL_DIR)/$$bin ./cmd/$$bin/.; \
	done

clean:
	@for bin in $(BINARIES) $(CGO_BINARIES); do \
		rm -f $(INSTALL_DIR)/$$bin; \
	done

clobber: clean

test:
	go test ./...

add:
	@git add -A

commit: build
	@git add -A
	@git commit -m "$(MSG)" || true

push: build
	@git add -A
	@git commit -m "$(MSG)" || true
	@git push
