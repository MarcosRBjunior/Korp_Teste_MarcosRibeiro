package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/MarcosRBjunior/Korp_Teste_MarcosRibeiro/services/faturamento/internal/database"
	"github.com/MarcosRBjunior/Korp_Teste_MarcosRibeiro/services/faturamento/internal/estoqueclient"
	"github.com/MarcosRBjunior/Korp_Teste_MarcosRibeiro/services/faturamento/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// errNotaFechada sinaliza, dentro de uma transação, que a nota não está mais
// Aberta — usado para distinguir esse caso do "não encontrada" no retorno HTTP.
var errNotaFechada = errors.New("nota fiscal não está Aberta")

type NotaFiscalHandler struct {
	DB      *gorm.DB
	Estoque *estoqueclient.Client
}

func NewNotaFiscalHandler(db *gorm.DB, estoque *estoqueclient.Client) *NotaFiscalHandler {
	return &NotaFiscalHandler{DB: db, Estoque: estoque}
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
// Debita o saldo de cada item no Estoque (via circuit breaker) antes de
// fechar a nota — nunca o contrário, para não fechar uma nota cujo saldo
// não foi confirmado. A idempotência via header Idempotency-Key entra na
// Fase 5.
func (h *NotaFiscalHandler) Imprimir(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		respondError(c, http.StatusBadRequest, "id inválido")
		return
	}

	// Claim atômico: Aberta -> Processando. Reserva a nota para esta
	// impressão e impede que uma segunda impressão concorrente também
	// passe a chamar o Estoque para os mesmos itens.
	claim := h.DB.Model(&models.NotaFiscal{}).
		Where("id = ? AND status = ?", id, models.StatusAberta).
		Update("status", models.StatusProcessando)

	if claim.Error != nil {
		respondError(c, http.StatusInternalServerError, "erro ao iniciar impressão da nota fiscal")
		return
	}

	if claim.RowsAffected == 0 {
		if _, status, err := h.buscarNotaPorID(id); err != nil {
			respondError(c, status, err.Error())
			return
		}
		respondError(c, http.StatusConflict, "somente uma nota Aberta pode ser impressa")
		return
	}

	nota, status, err := h.buscarNotaPorID(id)
	if err != nil {
		h.reabrirNota(id)
		respondError(c, status, err.Error())
		return
	}

	for _, item := range nota.Itens {
		if err := h.Estoque.Debitar(c.Request.Context(), item.ProdutoID, item.Quantidade); err != nil {
			h.reabrirNota(id)
			respondErroDebito(c, err)
			return
		}
	}

	if err := h.DB.Model(&models.NotaFiscal{}).
		Where("id = ?", id).
		Update("status", models.StatusFechada).Error; err != nil {
		respondError(c, http.StatusInternalServerError, "erro ao fechar nota fiscal")
		return
	}

	nota.Status = models.StatusFechada
	c.JSON(http.StatusOK, nota)
}

// reabrirNota volta a nota para Aberta após uma falha no débito, para que
// ela possa ser reimpressa depois. Erro aqui só é logado: já estamos no
// caminho de erro do Imprimir e a resposta ao cliente não deve mudar por
// causa disso.
func (h *NotaFiscalHandler) reabrirNota(id uint64) {
	h.DB.Model(&models.NotaFiscal{}).
		Where("id = ? AND status = ?", id, models.StatusProcessando).
		Update("status", models.StatusAberta)
}

func respondErroDebito(c *gin.Context, err error) {
	switch {
	case errors.Is(err, estoqueclient.ErrIndisponivel):
		respondError(c, http.StatusServiceUnavailable, "serviço de estoque indisponível, tente novamente em instantes")
	case errors.Is(err, estoqueclient.ErrSaldoInsuficiente):
		respondError(c, http.StatusConflict, "saldo insuficiente no estoque para um dos produtos da nota")
	case errors.Is(err, estoqueclient.ErrProdutoNaoEncontrado):
		respondError(c, http.StatusUnprocessableEntity, "produto da nota não encontrado no estoque")
	default:
		respondError(c, http.StatusInternalServerError, "erro ao debitar saldo no estoque")
	}
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
