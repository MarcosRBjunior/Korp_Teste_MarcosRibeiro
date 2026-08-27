package database

import (
	"fmt"

	"github.com/MarcosRBjunior/Korp_Teste_MarcosRibeiro/services/estoque/internal/config"
	"github.com/MarcosRBjunior/Korp_Teste_MarcosRibeiro/services/estoque/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(cfg config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("falha ao conectar ao banco: %w", err)
	}

	if err := db.AutoMigrate(&models.Produto{}); err != nil {
		return nil, fmt.Errorf("falha ao migrar o schema: %w", err)
	}

	return db, nil
}
