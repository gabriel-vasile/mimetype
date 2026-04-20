BINARY_NAME := mimetype-detect


PREFIX := /usr/local

.PHONY: all build test clean install uninstall

all: build


build:
	go build -o $(BINARY_NAME) ./cmd/mimetype-detect


test:
	go test -v ./...


install: build
	install -d $(DESTDIR)$(PREFIX)/bin
	install -m 0755 $(BINARY_NAME) $(DESTDIR)$(PREFIX)/bin/$(BINARY_NAME)
	@echo "Installed to $(DESTDIR)$(PREFIX)/bin/$(BINARY_NAME)"


uninstall:
	rm -f $(DESTDIR)$(PREFIX)/bin/$(BINARY_NAME)
	@echo "Removed: $(DESTDIR)$(PREFIX)/bin/$(BINARY_NAME)"


clean:
	rm -f $(BINARY_NAME)
	@echo "Cleanup done."