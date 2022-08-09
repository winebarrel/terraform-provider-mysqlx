**This repository is an unofficial fork**

---

# Terraform Provider

## Publish

Packages for version and different os_architecture are generated with `.goreleaser.yaml`, triggered by pushing a new `git tag`

### Manual build and publish to private terraform registry using [citizen](https://github.com/outsideris/citizen/)

Before publishing an executable is required, executable is found in the releases page on the specific version of the private tf registry that is needed(example: [v0.5.2](https://github.com/outsideris/citizen/releases/tag/v0.5.2))

```sh
env GOOS=target-OS GOARCH=target-architecture make build
```

A zip archive format is required by the private terraform registry, multiple os_architecture builds under the same name and version will be uploaded together:

```sh
zip stable-mysql_0.0.1_<target-OS>_<target-architecture>.zip terraform-provider-mysql
ex:
  stable-mysql_0.0.1_darwin_arm64.zip
  stable-mysql_0.0.1_linux_amd64.zip

CITIZEN_ADDR=http://localhost:3000  ./citizen-macos-amd64 provider stable mysql 0.0.1 -v
```

---

## Usage

```hcl
terraform {
  required_providers {
    mysql = {
      source  = "localhost/stable/mysql"
      version = "~> 0.0.1"
    }
  }
  required_version = ">= 0.13"
}

provider "mysql" {
  endpoint = "localhost"
  username = "root"
}

resource "mysql_grant" "test_db1" {
  user       = "test"
  host       = "%"
  database   = "db"
  table      = "table"
  privileges = [ "SELECT", "INSERT (`column1`)", "UPDATE (`column1`, `column2`, `column3`)"]
}
```
