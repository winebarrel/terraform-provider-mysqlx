**This repository is an unofficial fork**

---

Terraform Provider
==================

Usage
-----

For Terraform 0.13+ compatibility, the configuration should specify version 1.6 or higher:

```hcl
terraform {
  required_providers {
    mysql = {
      source  = "winebarrel/mysql"
      version = "1.9.0-p6"
    }
  }
  required_version = ">= 0.13"
}

provider "mysql" {
}
```
