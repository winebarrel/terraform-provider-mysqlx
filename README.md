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
      version = "1.9.0-p8"
    }
  }
  required_version = ">= 0.13"
}

provider "mysql" {
  endpoint = "localhost"
  username = "root"
}
```
