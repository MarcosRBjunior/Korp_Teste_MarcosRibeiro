package main

import (
	"log"

	"github.com/MarcosRBjunior/Korp_Teste_MarcosRibeiro/services/estoque/internal/config"
	"github.com/MarcosRBjunior/Korp_Teste_MarcosRibeiro/services/estoque/internal/database"
	"github.com/MarcosRBjunior/Korp_Teste_MarcosRibeiro/services/estoque/internal/handlers"
	"github.com/MarcosRBjunior/Korp_Teste_MarcosRibeiro/services/estoque/internal/routes"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("erro ao conectar ao banco: %v", err)
	}

	router := gin.Default()
	produtoHandler := handlers.NewProdutoHandler(db)
	routes.Setup(router, produtoHandler)

	log.Printf("serviço de estoque rodando na porta %s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("erro ao iniciar servidor: %v", err)
	}
}
