module github.com/jeremmfr/terraform-provider-junos

go 1.16

require (
	github.com/hashicorp/go-cty v1.4.1-0.20200414143053-d3edf31b6320
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.10.0
	github.com/jeremmfr/go-netconf v0.4.2
	github.com/jeremmfr/go-utils v0.4.1
	github.com/jeremmfr/junosdecode v1.1.0
	golang.org/x/crypto v0.0.0-20220112180741-5e0467b6c7ce
)

replace github.com/hashicorp/terraform-plugin-sdk/v2 => github.com/jeremmfr/terraform-plugin-sdk/v2 v2.10.1-0.20211216113247-43f5422548b6
