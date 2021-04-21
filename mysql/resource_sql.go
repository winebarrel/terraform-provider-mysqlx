package mysql

import (
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceSql() *schema.Resource {
	return &schema.Resource{
		Create: CreateSql,
		Read:   ReadSql,
		Delete: DeleteSql,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"create_sql": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"delete_sql": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func CreateSql(d *schema.ResourceData, meta interface{}) error {
	db, err := connectToMySQL(meta.(*MySQLConfiguration))

	if err != nil {
		return err
	}

	name := d.Get("name").(string)
	create_sql := d.Get("create_sql").(string)

	log.Println("Executing SQL", create_sql)

	_, err = db.Exec(create_sql)

	if err != nil {
		return err
	}

	d.SetId(name)

	return nil
}

func ReadSql(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func DeleteSql(d *schema.ResourceData, meta interface{}) error {
	db, err := connectToMySQL(meta.(*MySQLConfiguration))

	if err != nil {
		return err
	}

	delete_sql := d.Get("delete_sql").(string)

	log.Println("Executing SQL:", delete_sql)

	_, err = db.Exec(delete_sql)

	if err == nil {
		d.SetId("")
	}

	return err
}
