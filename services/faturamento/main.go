package main

import (
	"log"

	"github.com/MarcosRBjunior/Korp_Teste_MarcosRibeiro/services/faturamento/internal/config"
	"github.com/MarcosRBjunior/Korp_Teste_MarcosRibeiro/services/faturamento/internal/database"
	"github.com/MarcosRBjunior/Korp_Teste_MarcosRibeiro/services/faturamento/internal/estoqueclient"
	"github.com/MarcosRBjunior/Korp_Teste_MarcosRibeiro/services/faturamento/internal/handlers"
	"github.com/MarcosRBjunior/Korp_Teste_MarcosRibeiro/services/faturamento/internal/routes"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("erro ao conectar ao banco: %v", err)
	}

	estoqueClient := estoqueclient.New(cfg.EstoqueURL, cfg.EstoqueTimeout)

	router := gin.Default()
	notaHandler := handlers.NewNotaFiscalHandler(db, estoqueClient)
	routes.Setup(router, db, notaHandler)

	log.Printf("serviço de faturamento rodando na porta %s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("erro ao iniciar servidor: %v", err)
	}
}
