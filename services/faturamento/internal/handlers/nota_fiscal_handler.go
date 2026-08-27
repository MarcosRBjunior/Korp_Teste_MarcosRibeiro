package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/MarcosRBjunior/Korp_Teste_MarcosRibeiro/services/faturamento/internal/database"
	"github.com/MarcosRBjunior/Korp_Teste_MarcosRibeiro/services/faturamento/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// errNotaFechada sinaliza, dentro de uma transação, que a nota não está mais
// Aberta — usado para distinguir esse caso do "não encontrada" no retorno HTTP.
var errNotaFechada = errors.New("nota fiscal não está Aberta")

type NotaFiscalHandler struct {
	DB *gorm.DB
}

func NewNotaFiscalHandler(db *gorm.DB) *NotaFiscalHandler {
	return &NotaFiscalHandler{DB: db}
}

type itemRequest struct {
	ProdutoID  uint `json:"produto_id" binding:"required"`
	Quantidade int  `json:"quantidade" binding:"required,gt=0"`
}

type criarNotaRequest struct {
	Itens []itemRequest `json:"itens" binding:"required,min=1,dive"`
}

func respondError(c *gin.Context, status int, mensagem string) {
	c.JSON(status, gin.H{"error": mensagem})
}

// Criar godoc
// POST /notas
func (h *NotaFiscalHandler) Criar(c *gin.Context) {
	var req criarNotaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "dados inválidos: "+err.Error())
		return
	}

	var nota models.NotaFiscal

	err := h.DB.Transaction(func(tx *gorm.DB) error {
		var numero int64
		if err := tx.Raw("SELECT nextval('" + database.SequenciaNumeroNota + "')").Scan(&numero).Error; err != nil {
			return err
		}

		nota = models.NotaFiscal{
			NumeroSequencial: numero,
			Status:           models.StatusAberta,
			CriadoEm:         time.Now(),
		}
		for _, item := range req.Itens {
			nota.Itens = append(nota.Itens, models.ItemNota{
				ProdutoID:  item.ProdutoID,
				Quantidade: item.Quantidade,
			})
		}

		return tx.Create(&nota).Error
	})

	if err != nil {
		respondError(c, http.StatusInternalServerError, "erro ao criar nota fiscal")
		return
	}

	c.JSON(http.StatusCreated, nota)
}

// AdicionarItem godoc
// POST /notas/:id/itens
func (h *NotaFiscalHandler) AdicionarItem(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, http.StatusBadRequest, "id inválido")
		return
	}

	var req itemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "dados inválidos: "+err.Error())
		return
	}

	var nota models.NotaFiscal
	var item models.ItemNota

	// A trava (FOR UPDATE) impede que um Imprimir concorrente feche a nota
	// entre a checagem de status e a inserção do item: o UPDATE de status
	// do Imprimir só consegue seguir depois que esta transação commitar.
	err = h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Preload("Itens").First(&nota, id).Error; err != nil {
			return err
		}

		if nota.Status != models.StatusAberta {
			return errNotaFechada
		}

		item = models.ItemNota{
			NotaFiscalID: nota.ID,
			ProdutoID:    req.ProdutoID,
			Quantidade:   req.Quantidade,
		}
		return tx.Create(&item).Error
	})

	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		respondError(c, http.StatusNotFound, "nota fiscal não encontrada")
		return
	case errors.Is(err, errNotaFechada):
		respondError(c, http.StatusConflict, "só é possível adicionar itens a uma nota Aberta")
		return
	case err != nil:
		respondError(c, http.StatusInternalServerError, "erro ao adicionar item")
		return
	}

	nota.Itens = append(nota.Itens, item)
	c.JSON(http.StatusCreated, nota)
}

// Listar godoc
// GET /notas
func (h *NotaFiscalHandler) Listar(c *gin.Context) {
	var notas []models.NotaFiscal
	if err := h.DB.Preload("Itens").Order("id").Find(&notas).Error; err != nil {
		respondError(c, http.StatusInternalServerError, "erro ao listar notas fiscais")
		return
	}
	c.JSON(http.StatusOK, notas)
}

// BuscarPorID godoc
// GET /notas/:id
func (h *NotaFiscalHandler) BuscarPorID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, http.StatusBadRequest, "id inválido")
		return
	}

	nota, status, err := h.buscarNotaPorID(id)
	if err != nil {
		respondError(c, status, err.Error())
		return
	}
	c.JSON(http.StatusOK, nota)
}

// Imprimir godoc
// POST /notas/:id/imprimir
//
// Nesta fase a impressão apenas fecha a nota (Aberta -> Fechada). A chamada
// ao serviço de Estoque para debitar o saldo, protegida por circuit breaker,
// e a idempotência via header Idempotency-Key entram nas Fases 4 e 5.
func (h *NotaFiscalHandler) Imprimir(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, http.StatusBadRequest, "id inválido")
		return
	}

	// Update condicional e atômico: só fecha se a nota ainda estiver Aberta.
	// Evita que duas impressões concorrentes fechem a nota mais de uma vez
	// (o que, a partir da Fase 4, poderia gerar um débito duplicado no Estoque).
	result := h.DB.Model(&models.NotaFiscal{}).
		Where("id = ? AND status = ?", id, models.StatusAberta).
		Update("status", models.StatusFechada)

	if result.Error != nil {
		respondError(c, http.StatusInternalServerError, "erro ao fechar nota fiscal")
		return
	}

	if result.RowsAffected == 0 {
		if _, status, err := h.buscarNotaPorID(id); err != nil {
			respondError(c, status, err.Error())
			return
		}
		respondError(c, http.StatusConflict, "somente uma nota Aberta pode ser impressa")
		return
	}

	nota, status, err := h.buscarNotaPorID(id)
	if err != nil {
		respondError(c, status, err.Error())
		return
	}
	c.JSON(http.StatusOK, nota)
}

func (h *NotaFiscalHandler) buscarNotaPorID(id uint64) (*models.NotaFiscal, int, error) {
	var nota models.NotaFiscal
	if err := h.DB.Preload("Itens").First(&nota, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, http.StatusNotFound, errors.New("nota fiscal não encontrada")
		}
		return nil, http.StatusInternalServerError, errors.New("erro ao buscar nota fiscal")
	}
	return &nota, http.StatusOK, nil
}
