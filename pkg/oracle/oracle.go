package oracle

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/godror/godror"
	"github.com/riclava/oracletypeconverter/pkg/config"
)

func NewOracle(conf *config.Config) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s/%s@%s:%d/%s",
		conf.Username,
		conf.Password,
		conf.Host,
		conf.Port,
		conf.DbName,
	)
	db, err := sql.Open("godror", dsn)
	if err != nil {
		log.Fatal(err)
	}

	return db, nil
}
