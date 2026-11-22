APP_ID = com.chameth.glauncher
PREFIX ?= $(HOME)/.local
BINDIR ?= $(PREFIX)/bin
DATADIR ?= $(PREFIX)/share

build:
	go build -ldflags="-X gioui.org/app.ID=$(APP_ID)" -o glauncher .

verify: build
	go vet ./...
	go fix ./...
	staticcheck ./...
	go fmt ./...

install: build
	install -Dm755 glauncher $(BINDIR)/glauncher
	install -Dm644 assets/$(APP_ID).svg $(DATADIR)/icons/hicolor/scalable/apps/$(APP_ID).svg
	install -d $(DATADIR)/applications
	sed 's|^Exec=.*|Exec=$(BINDIR)/glauncher|' assets/$(APP_ID).desktop > $(DATADIR)/applications/$(APP_ID).desktop

.PHONY: build verify install
