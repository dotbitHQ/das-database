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

func Initialize(db *gorm.DB, logMode, isUpdate bool) (*DbDao, error) {
	if logMode {
		db = db.Debug()
	}

	var err error
	if isUpdate {
		// AutoMigrate will create tables, missing foreign keys, constraints, columns and indexes.
		// It will change existing column’s type if its size, precision, nullable changed.
		// It WON’T delete unused columns to protect your data.
		err = db.AutoMigrate(
			&TableAccountInfo{},
			&TableBlockInfo{},
			&TableIncomeCellInfo{},
			&TableOfferInfo{},
			&TableRebateInfo{},
			&TableRecordsInfo{},
			&TableReverseInfo{},
			&TableSmtInfo{},
			&TableTokenPriceInfo{},
			&TableTradeDealInfo{},
			&TableTradeInfo{},
			&TableTransactionInfo{},
		)
	}
	return &DbDao{db: db}, err
}
