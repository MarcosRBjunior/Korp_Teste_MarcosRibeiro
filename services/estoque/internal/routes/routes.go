package routes

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/MarcosRBjunior/Korp_Teste_MarcosRibeiro/services/estoque/internal/handlers"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Setup(router *gin.Engine, db *gorm.DB, produtoHandler *handlers.ProdutoHandler) {
	router.Use(cors.New(cors.Config{
		AllowOrigins:     corsOrigins(),
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Idempotency-Key"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

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
		produtos.POST("/:id/creditar", produtoHandler.Creditar)
	}
}

// corsOrigins lê CORS_ALLOWED_ORIGINS (separado por vírgula); por padrão
// libera o dev server do Angular.
func corsOrigins() []string {
	if raw, ok := os.LookupEnv("CORS_ALLOWED_ORIGINS"); ok && raw != "" {
		return strings.Split(raw, ",")
	}
	return []string{"http://localhost:4200"}
}
