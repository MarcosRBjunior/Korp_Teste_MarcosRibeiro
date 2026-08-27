package routes

import (
	"net/http"

	"github.com/MarcosRBjunior/Korp_Teste_MarcosRibeiro/services/faturamento/internal/handlers"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Setup(router *gin.Engine, db *gorm.DB, notaHandler *handlers.NotaFiscalHandler) {
	router.GET("/health", func(c *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil || sqlDB.Ping() != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "erro", "database": "indisponível"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok", "database": "conectado"})
	})

	notas := router.Group("/notas")
	{
		notas.POST("", notaHandler.Criar)
		notas.GET("", notaHandler.Listar)
		notas.GET("/:id", notaHandler.BuscarPorID)
		notas.POST("/:id/itens", notaHandler.AdicionarItem)
		notas.POST("/:id/imprimir", notaHandler.Imprimir)
	}
}
