package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/hashicorp/go-version"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	gocloudmysql "gocloud.dev/mysql"
	_ "gocloud.dev/mysql/awsmysql"
	_ "gocloud.dev/mysql/gcpmysql"
	"golang.org/x/net/proxy"
)

const (
	cleartextPasswords = "cleartext"
	nativePasswords    = "native"
)

type MySQLConfiguration struct {
	Config                 *mysql.Config
	Scheme                 string
	MaxConnLifetime        time.Duration
	MaxOpenConns           int
	ConnectRetryTimeoutSec time.Duration
	db                     *sql.DB
}

func (c *MySQLConfiguration) GetDbConn() (*sql.DB, error) {
	if c.db == nil {
		db, err := connectToMySQL(c)
		if err != nil {
			return nil, err
		}
		c.db = db
	}
	return c.db, nil
}

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"scheme": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "mysql",
				ValidateFunc: validation.StringInSlice([]string{
					"mysql",
					"awsmysql",
					"gcpmysql",
				}, false),
			},
			"endpoint": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("MYSQL_ENDPOINT", nil),
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					if value == "" {
						errors = append(errors, fmt.Errorf("Endpoint must not be an empty string"))
					}

					return
				},
			},

			"username": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("MYSQL_USERNAME", nil),
			},

			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("MYSQL_PASSWORD", nil),
			},

			"proxy": {
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"ALL_PROXY",
					"all_proxy",
				}, nil),
				ValidateFunc: validation.StringMatch(regexp.MustCompile("^socks5h?://.*:\\d+$"), "The proxy URL is not a valid socks url."),
			},

			"tls": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("MYSQL_TLS_CONFIG", "false"),
				ValidateFunc: validation.StringInSlice([]string{
					"true",
					"false",
					"skip-verify",
				}, false),
			},

			"max_conn_lifetime_sec": {
				Type:     schema.TypeInt,
				Optional: true,
			},

			"max_open_conns": {
				Type:     schema.TypeInt,
				Optional: true,
			},

			"authentication_plugin": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      nativePasswords,
				ValidateFunc: validation.StringInSlice([]string{cleartextPasswords, nativePasswords}, true),
			},

			"connect_retry_timeout_sec": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  300,
			},
		},

		DataSourcesMap: map[string]*schema.Resource{
			"mysql_tables": dataSourceTables(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"mysql_database":      resourceDatabase(),
			"mysql_grant":         resourceGrant(),
			"mysql_role":          resourceRole(),
			"mysql_user":          resourceUser(),
			"mysql_user_password": resourceUserPassword(),
			"mysql_sql":           resourceSql(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {

	var endpoint = d.Get("endpoint").(string)

	proto := "tcp"
	if len(endpoint) > 0 && endpoint[0] == '/' {
		proto = "unix"
	}

	conf := mysql.Config{
		User:                    d.Get("username").(string),
		Passwd:                  d.Get("password").(string),
		Net:                     proto,
		Addr:                    endpoint,
		TLSConfig:               d.Get("tls").(string),
		AllowNativePasswords:    d.Get("authentication_plugin").(string) == nativePasswords,
		AllowCleartextPasswords: d.Get("authentication_plugin").(string) == cleartextPasswords,
	}

	dialer, err := makeDialer(d)
	if err != nil {
		return nil, err
	}

	mysql.RegisterDial("tcp", func(network string) (net.Conn, error) {
		return dialer.Dial("tcp", network)
	})

	mysqlConf := &MySQLConfiguration{
		Scheme:                 d.Get("scheme").(string),
		Config:                 &conf,
		MaxConnLifetime:        time.Duration(d.Get("max_conn_lifetime_sec").(int)) * time.Second,
		MaxOpenConns:           d.Get("max_open_conns").(int),
		ConnectRetryTimeoutSec: time.Duration(d.Get("connect_retry_timeout_sec").(int)) * time.Second,
		db:                     nil,
	}

	return mysqlConf, nil
}

var identQuoteReplacer = strings.NewReplacer("`", "``")

func makeDialer(d *schema.ResourceData) (proxy.Dialer, error) {
	proxyFromEnv := proxy.FromEnvironment()
	proxyArg := d.Get("proxy").(string)

	if len(proxyArg) > 0 {
		proxyURL, err := url.Parse(proxyArg)
		if err != nil {
			return nil, err
		}
		proxy, err := proxy.FromURL(proxyURL, proxy.Direct)
		if err != nil {
			return nil, err
		}

		return proxy, nil
	}

	return proxyFromEnv, nil
}

func quoteIdentifier(in string) string {
	return fmt.Sprintf("`%s`", identQuoteReplacer.Replace(in))
}

func serverVersion(db *sql.DB) (*version.Version, error) {
	var versionString string
	err := db.QueryRow("SELECT @@GLOBAL.innodb_version").Scan(&versionString)
	if err != nil {
		return nil, err
	}

	return version.NewVersion(versionString)
}

func serverVersionString(db *sql.DB) (string, error) {
	var versionString string
	err := db.QueryRow("SELECT @@GLOBAL.version").Scan(&versionString)
	if err != nil {
		return "", err
	}

	return versionString, nil
}

func (conf *MySQLConfiguration) connStr(params string) string {
	host := conf.Config.Addr
	re := regexp.MustCompile(`(.+):([0-9]+)$`)
	matches := re.FindStringSubmatch(host)
	port := "3306"
	if matches != nil {
		host = matches[1]
		port = matches[2]
	}
	// For GCP, support both project/region/instance and project:region:instance
	// (The second one allows to use the output of google_sql_database_instance as host
	if conf.Scheme == "gcpmysql" {
		host = strings.ReplaceAll(host, ":", "/")
	}

	connStr := fmt.Sprintf(
		"%s://%s:%s@%s:%s/mysql?%s",
		conf.Scheme,
		url.QueryEscape(conf.Config.User),
		url.QueryEscape(conf.Config.Passwd),
		host,
		port,
		params,
	)
	return connStr
}

func connectToMySQL(conf *MySQLConfiguration) (*sql.DB, error) {

	dsn := conf.Config.FormatDSN()
	log.Println("DSN is", dsn)
	var db *sql.DB
	var err error
	if conf.Scheme != "mysql" {
		dsn = conf.connStr(dsn[strings.LastIndex(dsn, "?")+1:])
		log.Println("DSN is now", dsn)
	}

	// When provisioning a database server there can often be a lag between
	// when Terraform thinks it's available and when it is actually available.
	// This is particularly acute when provisioning a server and then immediately
	// trying to provision a database on it.
	retryError := resource.Retry(conf.ConnectRetryTimeoutSec, func() *resource.RetryError {
		if conf.Scheme == "mysql" {
			db, err = sql.Open("mysql", dsn)
		} else {
			db, err = gocloudmysql.Open(context.Background(), dsn)
		}

		if err != nil {
			return resource.RetryableError(err)
		}

		err = db.Ping()
		if err != nil {
			return resource.RetryableError(err)
		}

		return nil
	})

	if retryError != nil {
		return nil, fmt.Errorf("Could not connect to server: %s", retryError)
	}
	db.SetConnMaxLifetime(conf.MaxConnLifetime)
	db.SetMaxOpenConns(conf.MaxOpenConns)
	return db, nil
}
