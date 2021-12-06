package dao

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type DbDao struct {
	db *gorm.DB
}

func NewGormDataBase(addr, user, password, dbName string, maxOpenConn, maxIdleConn int) (*gorm.DB, error) {
	conn := "%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local"
	dsn := fmt.Sprintf(conn, user, password, addr, dbName)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("gorm open :%v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("gorm db :%v", err)
	}

	sqlDB.SetMaxOpenConns(maxOpenConn)
	sqlDB.SetMaxIdleConns(maxIdleConn)

	return db, nil
}

func Initialize(db *gorm.DB, logMode bool) *DbDao {
	if logMode {
		db = db.Debug()
	}
	return &DbDao{db: db}
}
