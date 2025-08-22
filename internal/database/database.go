package database

import (
	"fmt"
	"go-devops/internal/config"
	"go-devops/internal/logger"
	"go-devops/internal/models"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Init(cfg *config.Config) (*gorm.DB, error) {
	logger.Infof("初始化数据库连接: %s", cfg.Database.Type)
	
	var db *gorm.DB
	var err error
	
	switch cfg.Database.Type {
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s",
			cfg.Database.MySQL.Username,
			cfg.Database.MySQL.Password,
			cfg.Database.MySQL.Host,
			cfg.Database.MySQL.Port,
			cfg.Database.MySQL.Database,
			cfg.Database.MySQL.Charset,
			cfg.Database.MySQL.ParseTime,
			cfg.Database.MySQL.Loc,
		)
		logger.Infof("MySQL DSN: %s:***@tcp(%s:%d)/%s", 
			cfg.Database.MySQL.Username,
			cfg.Database.MySQL.Host,
			cfg.Database.MySQL.Port,
			cfg.Database.MySQL.Database)
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	case "sqlite":
		fallthrough
	default:
		logger.Infof("SQLite 文件: %s", cfg.Database.SQLite.File)
		db, err = gorm.Open(sqlite.Open(cfg.Database.SQLite.File), &gorm.Config{})
	}
	
	if err != nil {
		logger.Errorf("数据库连接失败: %v", err)
		logger.LogDBOperation("connect", "database", false, err.Error())
		return nil, err
	}
	
	logger.LogDBOperation("connect", "database", true, "")

	// 自动迁移数据库表
	logger.Info("开始数据库表迁移")
	err = db.AutoMigrate(
		&models.User{},
		&models.Host{},
		&models.Job{},
		&models.Script{},
		&models.JobExecution{},
		&models.Business{},
		&models.Environment{},
		&models.Cluster{},
		&models.HostTopology{},
		&models.UserActivity{},
	)
	if err != nil {
		logger.Errorf("数据库表迁移失败: %v", err)
		logger.LogDBOperation("migrate", "tables", false, err.Error())
		return nil, err
	}
	
	logger.Info("数据库表迁移完成")
	logger.LogDBOperation("migrate", "tables", true, "")

	// 初始化种子数据
	logger.Info("开始初始化种子数据")
	err = SeedData(db)
	if err != nil {
		logger.Errorf("种子数据初始化失败: %v", err)
		return nil, err
	}
	
	logger.Info("种子数据初始化完成")

	return db, nil
}
