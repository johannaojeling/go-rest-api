package database

import (
	"fmt"

	"database/sql"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func GetConnection(driver string, dsn string) (*gorm.DB, error) {
	sqlDB, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("error opening SQL DB: %v", err)
	}

	dialector := postgres.New(postgres.Config{
		Conn: sqlDB,
	})
	gormDB, err := gorm.Open(dialector)
	if err != nil {
		return nil, fmt.Errorf("error opening Gorm DB: %v", err)
	}
	return gormDB, nil
}
