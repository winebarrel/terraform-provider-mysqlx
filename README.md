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


 ## GoCloud

 By default, the provider uses the [database/sql][databasesql] library to directly connect to MySQL host instance. For connections to AWS/GCP hosted instances, the provider can connect through the [GoCloud](gocloud) library. GoCloud simplifies connecting to AWS/GCP hosted databases, managing any proxy or custom authentication details.

 ### AWS

 To enable GoCloud based connections to AWS RDS instances, set `scheme` to `awsmysql` and `host` to the RDS database's endpoint value.
 (e.g.: `instance.xxxxxx.region.rds.amazonaws.com`)

 ```hcl
 provider "mysql" {
   scheme   = "awsmysql"
   endpoint = "test-instance.cvvrsv6scpgd.eu-central-1.rds.amazonaws.com"
   username = "mysql"
   port     = 5432
   password = "test1234"
 }
 ```

 ### GCP

 To enable GoCloud for GCP SQL, set `scheme` to `gcpmysql` and `host` to the connection name of the instance in following format: `project/region/instance` (or `project:region:instance`).

 For GCP, GoCloud also requires the `GOOGLE_APPLICATION_CREDENTIALS` environment variable to be set to the service account credentials file.
 These credentials can be created here: https://console.cloud.google.com/iam-admin/serviceaccounts

 See also: https://cloud.google.com/docs/authentication/production

 ---
 **Note**

 [Cloud SQL API](https://console.developers.google.com/apis/api/sqladmin.googleapis.com/overview) needs to be enabled for GoCloud to connect to your instance.

 ---

 ```hcl
 provider "mysql" {
   scheme   = "gcpmysql"
   endpoint = "test-project/europe-west3/test-instance"
   username = "admin"
   password = "test1234"
 }
 ```

 Example with GCP resources:

 ```hcl
 resource "google_sql_database_instance" "test" {
   project          = "test-project"
   name             = "test-instance"
   database_version = "MYSQL_8_0"
   region           = "europe-west3"
   settings {
     tier = "db-f1-micro"
   }
 }

 resource "google_sql_user" "admin" {
   project  = "test-project"
   name     = "admin"
   instance = google_sql_database_instance.test.name
   password = "xxxxxxxx"
 }

 provider "mysql" {
   scheme   = "gcpmysql"
   endpoint = google_sql_database_instance.test.connection_name
   username = google_sql_user.admin.name
   password = google_sql_user.admin.password
 }

 resource mysql_database "test" {
   name = "test-db"
 }
 ```

 [databasesql]: https://pkg.go.dev/github.com/lib/pq
 [gocloud]: https://gocloud.dev/howto/sql/