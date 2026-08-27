package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/MarcosRBjunior/Korp_Teste_MarcosRibeiro/services/estoque/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type ProdutoHandler struct {
	DB *gorm.DB
}

func NewProdutoHandler(db *gorm.DB) *ProdutoHandler {
	return &ProdutoHandler{DB: db}
}

type criarProdutoRequest struct {
	Codigo    string `json:"codigo" binding:"required"`
	Descricao string `json:"descricao" binding:"required"`
	Saldo     int    `json:"saldo" binding:"gte=0"`
}

type debitarRequest struct {
	Quantidade int `json:"quantidade" binding:"required,gt=0"`
}

func respondError(c *gin.Context, status int, mensagem string) {
	c.JSON(status, gin.H{"error": mensagem})
}

// Criar godoc
// POST /produtos
func (h *ProdutoHandler) Criar(c *gin.Context) {
	var req criarProdutoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "dados inválidos: "+err.Error())
		return
	}

	produto := models.Produto{
		Codigo:    req.Codigo,
		Descricao: req.Descricao,
		Saldo:     req.Saldo,
		CriadoEm:  time.Now(),
	}

	if err := h.DB.Create(&produto).Error; err != nil {
		if isUniqueViolation(err) {
			respondError(c, http.StatusConflict, "já existe um produto com este código")
			return
		}
		respondError(c, http.StatusInternalServerError, "erro ao criar produto")
		return
	}

	c.JSON(http.StatusCreated, produto)
}

// Listar godoc
// GET /produtos
func (h *ProdutoHandler) Listar(c *gin.Context) {
	var produtos []models.Produto
	if err := h.DB.Order("id").Find(&produtos).Error; err != nil {
		respondError(c, http.StatusInternalServerError, "erro ao listar produtos")
		return
	}
	c.JSON(http.StatusOK, produtos)
}

// BuscarPorID godoc
// GET /produtos/:id
func (h *ProdutoHandler) BuscarPorID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, http.StatusBadRequest, "id inválido")
		return
	}

	produto, status, err := h.buscarProdutoPorID(id)
	if err != nil {
		respondError(c, status, err.Error())
		return
	}
	c.JSON(http.StatusOK, produto)
}

// Debitar godoc
// POST /produtos/:id/debitar
// Endpoint interno, usado pelo serviço de Faturamento ao imprimir uma nota.
func (h *ProdutoHandler) Debitar(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, http.StatusBadRequest, "id inválido")
		return
	}

	var req debitarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "dados inválidos: "+err.Error())
		return
	}

	// Update condicional e atômico: só debita se houver saldo suficiente.
	// Evita saldo negativo mesmo com debitos concorrentes, sem precisar de lock explícito.
	result := h.DB.Model(&models.Produto{}).
		Where("id = ? AND saldo >= ?", id, req.Quantidade).
		Update("saldo", gorm.Expr("saldo - ?", req.Quantidade))

	if result.Error != nil {
		respondError(c, http.StatusInternalServerError, "erro ao debitar saldo")
		return
	}

	if result.RowsAffected == 0 {
		produto, status, err := h.buscarProdutoPorID(id)
		if err != nil {
			respondError(c, status, err.Error())
			return
		}
		respondError(c, http.StatusConflict, "saldo insuficiente: disponível "+strconv.Itoa(produto.Saldo))
		return
	}

	produto, status, err := h.buscarProdutoPorID(id)
	if err != nil {
		respondError(c, status, err.Error())
		return
	}
	c.JSON(http.StatusOK, produto)
}

func (h *ProdutoHandler) buscarProdutoPorID(id uint64) (*models.Produto, int, error) {
	var produto models.Produto
	if err := h.DB.First(&produto, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, http.StatusNotFound, errors.New("produto não encontrado")
		}
		return nil, http.StatusInternalServerError, errors.New("erro ao buscar produto")
	}

	return &produto, http.StatusOK, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
