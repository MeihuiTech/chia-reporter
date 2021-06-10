package main

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

func GetDb(config *Config) (*gorm.DB, error) {

	gormLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second,               // Slow SQL threshold
			LogLevel:                  logger.Warn,               // Log level
			IgnoreRecordNotFoundError: config.IgnoreGormNotFoundError, // Ignore ErrRecordNotFound error for logger
			Colorful:                  false,                     // Disable color
		},
	)
	db, err := gorm.Open(mysql.Open(config.Dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open db connection: %v \n", err)
	}

	err = db.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8").AutoMigrate(&ChiaBlockRecord{})
	if err != nil {
		return nil, fmt.Errorf("error migrate db: %v", err)
	}

	err = db.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='矿工累计出块数量'").AutoMigrate(&ChiaTotalFarmerBlocks{})
	if err != nil {
		return nil, fmt.Errorf("error migrate db: %v", err)
	}

	err = db.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='矿工每天出块数量'").AutoMigrate(&ChiaDailyFarmerBlocks{})
	if err != nil {
		return nil, fmt.Errorf("error migrate db: %v", err)
	}

	err = db.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='区块同步高度'").AutoMigrate(&ChiaBlockSyncHeight{})
	if err != nil {
		return nil, fmt.Errorf("error migrate db: %v", err)
	}
	return db, nil
}
