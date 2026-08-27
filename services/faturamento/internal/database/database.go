package database

import (
	"fmt"

	"github.com/MarcosRBjunior/Korp_Teste_MarcosRibeiro/services/faturamento/internal/config"
	"github.com/MarcosRBjunior/Korp_Teste_MarcosRibeiro/services/faturamento/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// SequenciaNumeroNota é a sequence nativa do Postgres usada para gerar o
// numero_sequencial da nota. Sequences do Postgres são atômicas por natureza,
// então múltiplas notas criadas ao mesmo tempo nunca recebem o mesmo número.
const SequenciaNumeroNota = "nota_fiscal_numero_seq"

func Connect(cfg config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("falha ao conectar ao banco: %w", err)
	}

	if err := db.AutoMigrate(&models.NotaFiscal{}, &models.ItemNota{}, &models.IdempotencyKey{}); err != nil {
		return nil, fmt.Errorf("falha ao migrar o schema: %w", err)
	}

	if err := db.Exec("CREATE SEQUENCE IF NOT EXISTS " + SequenciaNumeroNota + " START 1").Error; err != nil {
		return nil, fmt.Errorf("falha ao criar sequence de numeração: %w", err)
	}

	return db, nil
}
