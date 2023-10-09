module github.com/shadowsocks/go-shadowsocks2/systemtest

go 1.21

require (
	github.com/google/nftables v0.1.0
	github.com/stretchr/testify v1.8.4
	github.com/vishvananda/netlink v1.2.1-beta.2
	github.com/vishvananda/netns v0.0.4
	golang.org/x/net v0.16.0
	golang.org/x/sys v0.13.0
)

require (
	github.com/BurntSushi/toml v0.4.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/josharian/native v0.0.0-20200817173448-b6b71def0850 // indirect
	github.com/mdlayher/netlink v1.4.2 // indirect
	github.com/mdlayher/socket v0.0.0-20211102153432-57e3fa563ecb // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/mod v0.8.0 // indirect
	golang.org/x/tools v0.6.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	honnef.co/go/tools v0.2.2 // indirect
)

replace golang.org/x/net => github.com/marshall-lee/golang-net v0.15.1-0.20231006214855-d431cc52d155
