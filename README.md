Please use https://github.com/petoju/terraform-provider-mysql because it is not maintained.

---

**This repository is an unofficial fork**

---

Terraform Provider
==================

Usage
-----

```hcl
terraform {
  required_providers {
    mysql = {
      source  = "winebarrel/mysql"
      version = "~> 1.10.2"
    }
  }
  required_version = ">= 0.13"
}

provider "mysql" {
  endpoint = "localhost"
  username = "root"
}
```
