package mysql

import (
	"fmt"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceSSMParameterUserPassword() *schema.Resource {
	return &schema.Resource{
		Create: SetUserPasswordFromParameterStore,
		Read:   ReadParameterStoreUserPassword,
		Delete: DeleteParameterStoreUserPassword,
		Schema: map[string]*schema.Schema{
			"user": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"host": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "localhost",
			},
			"parameter_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"triggers": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func SetUserPasswordFromParameterStore(d *schema.ResourceData, meta interface{}) error {
	conf := meta.(*MySQLConfiguration)
	db, err := connectToMySQL(conf)

	if err != nil {
		return err
	}

	name := d.Get("parameter_name").(string)
	password, err := getParameter(conf.SSM, name)

	if err != nil {
		return err
	}

	requiredVersion, _ := version.NewVersion("8.0.0")
	currentVersion, err := serverVersion(db)

	if err != nil {
		return err
	}

	passSQL := fmt.Sprintf("'%s'", password)

	if currentVersion.LessThan(requiredVersion) {
		passSQL = fmt.Sprintf("PASSWORD(%s)", passSQL)
	}

	sql := fmt.Sprintf(
		"SET PASSWORD FOR '%s'@'%s' = %s",
		d.Get("user").(string),
		d.Get("host").(string),
		passSQL,
	)

	_, err = db.Exec(sql)

	if err != nil {
		return err
	}

	user := fmt.Sprintf("%s@%s", d.Get("user").(string), d.Get("host").(string))
	d.SetId(user)

	return nil
}

func ReadParameterStoreUserPassword(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func DeleteParameterStoreUserPassword(d *schema.ResourceData, meta interface{}) error {
	return nil
}
