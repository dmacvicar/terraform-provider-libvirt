## 1.0.0 (September 13, 2017)

IMPROVEMENTS: 

- Ignition config should be marshaled as compact JSON [\#2](https://github.com/terraform-providers/terraform-provider-ignition/issues/2)
- Allow users to omit the optional verification hashes [\#9](https://github.com/terraform-providers/terraform-provider-ignition/pull/9) ([MerlinDMC](https://github.com/MerlinDMC))
- vendor: github.com/hashicorp/terraform/...@v0.10.0 [\#11](https://github.com/terraform-providers/terraform-provider-ignition/pull/11)

BUG FIXES:

- Added nil check for empty lists in Ignition Config builders [\#7](https://github.com/terraform-providers/terraform-provider-ignition/pull/7)
- Fixed issue with ignition\_filesystem when empty options passed in [\#5](https://github.com/terraform-providers/terraform-provider-ignition/pull/5)
- ignition\_config: render to non-indented json [\#3](https://github.com/terraform-providers/terraform-provider-ignition/pull/3) ([mcuadros](https://github.com/mcuadros))



## 0.1.0 (June 20, 2017)

NOTES:

* Same functionality as that of Terraform 0.9.8. Repacked as part of [Provider Splitout](https://www.hashicorp.com/blog/upcoming-provider-changes-in-terraform-0-10/)
