package routes

import (
	"github.com/MarcosRBjunior/Korp_Teste_MarcosRibeiro/services/estoque/internal/handlers"
	"github.com/gin-gonic/gin"
)

func Setup(router *gin.Engine, produtoHandler *handlers.ProdutoHandler) {
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	produtos := router.Group("/produtos")
	{
		produtos.POST("", produtoHandler.Criar)
		produtos.GET("", produtoHandler.Listar)
		produtos.GET("/:id", produtoHandler.BuscarPorID)
		produtos.POST("/:id/debitar", produtoHandler.Debitar)
	}
}
