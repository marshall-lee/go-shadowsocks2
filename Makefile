NAME=shadowsocks2
BINDIR=bin
GOBUILD=CGO_ENABLED=0 go build -ldflags '-w -s -buildid='
# The -w and -s flags reduce binary sizes by excluding unnecessary symbols and debug info
# The -buildid= flag makes builds reproducible

all: linux-amd64 linux-arm64 macos-amd64 macos-arm64 win64 win32

linux-amd64:
	GOARCH=amd64 GOOS=linux $(GOBUILD) -o $(BINDIR)/$(NAME)-$@

linux-arm64:
	GOARCH=arm64 GOOS=linux $(GOBUILD) -o $(BINDIR)/$(NAME)-$@

macos-amd64:
	GOARCH=amd64 GOOS=darwin $(GOBUILD) -o $(BINDIR)/$(NAME)-$@

macos-arm64:
	GOARCH=arm64 GOOS=darwin $(GOBUILD) -o $(BINDIR)/$(NAME)-$@

win64:
	GOARCH=amd64 GOOS=windows $(GOBUILD) -o $(BINDIR)/$(NAME)-$@.exe

win32:
	GOARCH=386 GOOS=windows $(GOBUILD) -o $(BINDIR)/$(NAME)-$@.exe


test:
	go test

.PHONY: systemtest-linux
systemtest-linux:
	$(GOBUILD) -o ./systemtest/$(NAME)
	go test -C ./systemtest -c -o $(NAME)-systemtest
	sudo setcap "cap_net_admin+ep" ./systemtest/$(NAME)
	sudo setcap "cap_net_admin+ep cap_sys_admin+ep" ./systemtest/$(NAME)-systemtest
	./systemtest/$(NAME)-systemtest -test.v -shadowsocks-path=./systemtest/$(NAME)

.PHONY: systemtest
systemtest:
	$(GOBUILD) -o ./systemtest/$(NAME)
	go test -C ./systemtest -v -shadowsocks-path=./$(NAME)

systemtest-linux-amd64: linux-amd64
systemtest-linux-arm64: linux-arm64
systemtest-macos-amd64: macos-amd64
systemtest-macos-arm64: macos-arm64

releases: linux-amd64 linux-arm64 macos-amd64 macos-arm64 win64 win32
	chmod +x $(BINDIR)/$(NAME)-*
	tar czf $(BINDIR)/$(NAME)-linux-amd64.tgz -C $(BINDIR) $(NAME)-linux-amd64
	tar czf $(BINDIR)/$(NAME)-linux-arm64.tgz -C $(BINDIR) $(NAME)-linux-arm64
	gzip $(BINDIR)/$(NAME)-linux-amd64
	gzip $(BINDIR)/$(NAME)-linux-arm64
	gzip $(BINDIR)/$(NAME)-macos-amd64
	gzip $(BINDIR)/$(NAME)-macos-arm64
	zip -m -j $(BINDIR)/$(NAME)-win32.zip $(BINDIR)/$(NAME)-win32.exe
	zip -m -j $(BINDIR)/$(NAME)-win64.zip $(BINDIR)/$(NAME)-win64.exe

clean:
	rm $(BINDIR)/*

# Remove trailing {} from the release upload url
GITHUB_UPLOAD_URL=$(shell echo $${GITHUB_RELEASE_UPLOAD_URL%\{*})

upload: releases
	curl -H "Authorization: token $(GITHUB_TOKEN)" -H "Content-Type: application/gzip" --data-binary @$(BINDIR)/$(NAME)-linux-amd64.tgz  "$(GITHUB_UPLOAD_URL)?name=$(NAME)-linux-amd64.tgz"
	curl -H "Authorization: token $(GITHUB_TOKEN)" -H "Content-Type: application/gzip" --data-binary @$(BINDIR)/$(NAME)-linux-arm64.tgz  "$(GITHUB_UPLOAD_URL)?name=$(NAME)-linux-arm64.tgz"
	curl -H "Authorization: token $(GITHUB_TOKEN)" -H "Content-Type: application/gzip" --data-binary @$(BINDIR)/$(NAME)-linux-amd64.gz  "$(GITHUB_UPLOAD_URL)?name=$(NAME)-linux-amd64.gz"
	curl -H "Authorization: token $(GITHUB_TOKEN)" -H "Content-Type: application/gzip" --data-binary @$(BINDIR)/$(NAME)-linux-arm64.gz  "$(GITHUB_UPLOAD_URL)?name=$(NAME)-linux-arm64.gz"
	curl -H "Authorization: token $(GITHUB_TOKEN)" -H "Content-Type: application/gzip" --data-binary @$(BINDIR)/$(NAME)-macos-amd64.gz  "$(GITHUB_UPLOAD_URL)?name=$(NAME)-macos-amd64.gz"
	curl -H "Authorization: token $(GITHUB_TOKEN)" -H "Content-Type: application/gzip" --data-binary @$(BINDIR)/$(NAME)-macos-arm64.gz  "$(GITHUB_UPLOAD_URL)?name=$(NAME)-macos-arm64.gz"
	curl -H "Authorization: token $(GITHUB_TOKEN)" -H "Content-Type: application/zip"  --data-binary @$(BINDIR)/$(NAME)-win64.zip "$(GITHUB_UPLOAD_URL)?name=$(NAME)-win64.zip"
	curl -H "Authorization: token $(GITHUB_TOKEN)" -H "Content-Type: application/zip"  --data-binary @$(BINDIR)/$(NAME)-win32.zip "$(GITHUB_UPLOAD_URL)?name=$(NAME)-win32.zip"
