package routes

import (
	"net/http"

	"github.com/MarcosRBjunior/Korp_Teste_MarcosRibeiro/services/estoque/internal/handlers"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Setup(router *gin.Engine, db *gorm.DB, produtoHandler *handlers.ProdutoHandler) {
	router.GET("/health", func(c *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil || sqlDB.Ping() != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "erro", "database": "indisponível"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok", "database": "conectado"})
	})

	produtos := router.Group("/produtos")
	{
		produtos.POST("", produtoHandler.Criar)
		produtos.GET("", produtoHandler.Listar)
		produtos.GET("/:id", produtoHandler.BuscarPorID)
		produtos.POST("/:id/debitar", produtoHandler.Debitar)
	}
}
