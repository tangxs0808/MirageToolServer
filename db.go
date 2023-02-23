package main

import (
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func (t *MirageTool) initDB() error {
	db, err := t.openDB()
	if err != nil {
		return err
	}
	t.db = db

	err = db.AutoMigrate(&User{})
	if err != nil {
		return err
	}
	return err
}

func (t *MirageTool) openDB() (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	var log logger.Interface
	log = logger.Default.LogMode(logger.Silent)

	db, err = gorm.Open(
		sqlite.Open(t.cfg.DB.Path+"?_synchronous=1&_journal_mode=WAL"),
		&gorm.Config{
			DisableForeignKeyConstraintWhenMigrating: true,
			Logger:                                   log,
		},
	)

	db.Exec("PRAGMA foreign_keys=ON")

	// The pure Go SQLite library does not handle locking in
	// the same way as the C based one and we cant use the gorm
	// connection pool as of 2022/02/23.
	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetConnMaxIdleTime(time.Hour)

	if err != nil {
		return nil, err
	}

	return db, nil
}
