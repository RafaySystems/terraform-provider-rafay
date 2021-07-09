module github.com/RafaySystems/terraform-provider-rafay

go 1.16


require (
	//github.com/RafaySystems/rctl v0.0.0-00010101000000-000000000000
	github.com/RafaySystems/rctl latest
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.7.0
)


replace (
	//github.com/RafaySystems/terraform-provider-rafay/rafay => ../rafay
	//github.com/RafaySystems/rctl => ../rctl
	k8s.io/api => k8s.io/api v0.18.4
	k8s.io/apimachinery => k8s.io/apimachinery v0.18.4
    github.com/RafaySystems/rctl =>  github.com/RafaySystems/rctl v1.5.14
)
