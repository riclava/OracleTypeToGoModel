package main

import (
	"database/sql"
	"encoding/json"

	"github.com/riclava/oracletypeconverter/models"
	"github.com/riclava/oracletypeconverter/pkg/config"
	"github.com/riclava/oracletypeconverter/pkg/logger"
	"github.com/riclava/oracletypeconverter/pkg/oracle"
)

func test(db *sql.DB) {
	svc := models.NewFieldExampleService(db)
	logger.Infof("=======================Create============================")
	// Create
	err := svc.Create(&models.FieldExample{
		F1:   sql.NullString{String: "test1"},
		F2:   sql.NullString{String: "test1"},
		F3X1: sql.NullFloat64{Float64: 0.1},
	})
	if err != nil {
		logger.Fatalf(err.Error())
	}
	err = svc.Create(&models.FieldExample{
		F1: sql.NullString{String: "test2"},
		F2: sql.NullString{String: "test2"},
	})
	if err != nil {
		logger.Fatalf(err.Error())
	}

	// Gets & Get
	logger.Infof("=======================Gets============================")
	rows, err := svc.GetPage(1, 10, "f1", "DESC", "", "")
	if err != nil {
		logger.Fatalf("%s", err.Error())
	}

	bits, err := json.MarshalIndent(rows, "", "  ")
	if err != nil {
		logger.Fatalf("%s", err.Error())
	}

	logger.Infof("%v", string(bits))
	logger.Infof("=======================Get 1============================")
	row, err := svc.GetOne("f1", "test2")
	if err != nil {
		logger.Fatalf("%s", err.Error())
	}
	bits, err = json.MarshalIndent(row, "", "  ")
	if err != nil {
		logger.Fatalf("%s", err.Error())
	}
	logger.Infof("%v", string(bits))

	// Update
	row.F2 = sql.NullString{String: "test2222"}
	err = svc.Update("f1", "test2", row)
	if err != nil {
		logger.Fatalf("%s", err.Error())
	}
	// get again
	logger.Infof("=======================Get 2============================")
	row, err = svc.GetOne("f1", "test2")
	if err != nil {
		logger.Fatalf("%s", err.Error())
	}
	bits, err = json.MarshalIndent(row, "", "  ")
	if err != nil {
		logger.Fatalf("%s", err.Error())
	}
	logger.Infof("%v", string(bits))

	// Delete
	err = svc.Delete("f1", "test2")
	if err != nil {
		logger.Fatalf("%s", err.Error())
	}
	// get again
	logger.Infof("=======================Get 3============================")
	_, err = svc.GetOne("f1", "test2")
	if err != nil {
		logger.Errorf("%s", err.Error())
	}

}

func main() {

	conf := config.AutoLoadConfig()
	db, err := oracle.NewOracle(conf)
	if err != nil {
		logger.Fatalf("%s", err.Error())
	}

	test(db)

}
