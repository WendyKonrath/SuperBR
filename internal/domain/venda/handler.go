package venda

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// Handler agrupa os endpoints HTTP do domínio de vendas.
type Handler struct {
	service *Service
}

// NewHandler cria o handler com o service injetado.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// ── Structs de input ──────────────────────────────────────────────────────────

// itemVendaInput representa um produto a ser vendido com o tipo de preço desejado.
type itemVendaInput struct {
	ProdutoID uint   `json:"produto_id" binding:"required"`
	TipoPreco string `json:"tipo_preco" binding:"required,oneof=atacado varejo"`
}

// pagamentoInput representa uma forma de pagamento para a venda.
type pagamentoVendaInput struct {
	Tipo  string  `json:"tipo" binding:"required,oneof=pix credito debito dinheiro sucata"`
	Valor float64 `json:"valor" binding:"required,gt=0"`
}

// criarVendaInput representa o corpo completo da requisição de criação de venda.
type criarVendaInput struct {
	NomeCliente      string               `json:"nome_cliente" binding:"required"`
	DocumentoCliente string               `json:"documento_cliente"`
	TelefoneCliente  string               `json:"telefone_cliente"`
	Itens            []itemVendaInput     `json:"itens" binding:"required,min=1,dive"`
	Pagamentos       []pagamentoVendaInput `json:"pagamentos" binding:"required,min=1,dive"`
}

// ── Handlers ──────────────────────────────────────────────────────────────────

// CriarVenda inicia uma nova venda com status "pendente".
// Reserva os itens do estoque imediatamente para evitar venda dupla.
// POST /api/vendas
func (h *Handler) CriarVenda(c *gin.Context) {
	var input criarVendaInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	usuarioID, _ := c.Get("usuario_id")

	// Converte os inputs para os tipos internos do service.
	itens := make([]itemInput, len(input.Itens))
	for i, it := range input.Itens {
		itens[i] = itemInput{
			ProdutoID: it.ProdutoID,
			TipoPreco: it.TipoPreco,
		}
	}

	pags := make([]pagamentoInput, len(input.Pagamentos))
	for i, pg := range input.Pagamentos {
		pags[i] = pagamentoInput{
			Tipo:  pg.Tipo,
			Valor: pg.Valor,
		}
	}

	v, err := h.service.CriarVenda(
		input.NomeCliente,
		input.DocumentoCliente,
		input.TelefoneCliente,
		itens,
		pags,
		usuarioID.(uint),
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, v)
}

// ConfirmarVenda finaliza uma venda pendente e dá baixa definitiva no estoque.
// PATCH /api/vendas/:id/confirmar
func (h *Handler) ConfirmarVenda(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "id inválido"})
		return
	}

	usuarioID, _ := c.Get("usuario_id")

	v, err := h.service.ConfirmarVenda(uint(id), usuarioID.(uint))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	c.JSON(http.StatusOK, v)
}

// CancelarVenda cancela uma venda pendente e devolve os itens ao estoque.
// PATCH /api/vendas/:id/cancelar
func (h *Handler) CancelarVenda(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "id inválido"})
		return
	}

	usuarioID, _ := c.Get("usuario_id")

	v, err := h.service.CancelarVenda(uint(id), usuarioID.(uint))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	c.JSON(http.StatusOK, v)
}

// BuscarPorID retorna os detalhes de uma venda específica.
// GET /api/vendas/:id
func (h *Handler) BuscarPorID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "id inválido"})
		return
	}

	v, err := h.service.BuscarPorID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"erro": err.Error()})
		return
	}

	c.JSON(http.StatusOK, v)
}

// Listar retorna vendas com filtros opcionais por status ou período.
// GET /api/vendas
// GET /api/vendas?status=pendente
// GET /api/vendas?inicio=2025-01-01&fim=2025-01-31
func (h *Handler) Listar(c *gin.Context) {
	// Filtro por status.
	if status := c.Query("status"); status != "" {
		vendas, err := h.service.ListarPorStatus(status)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
			return
		}
		c.JSON(http.StatusOK, vendas)
		return
	}

	// Filtro por período (ambos os parâmetros são obrigatórios juntos).
	inicioStr := c.Query("inicio")
	fimStr := c.Query("fim")
	if inicioStr != "" || fimStr != "" {
		if inicioStr == "" || fimStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"erro": "informe 'inicio' e 'fim' juntos (formato: 2006-01-02)"})
			return
		}

		inicio, err := time.Parse("2006-01-02", inicioStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"erro": "formato de 'inicio' inválido — use 2006-01-02"})
			return
		}
		// Ajusta 'fim' para incluir o dia inteiro (23:59:59).
		fim, err := time.Parse("2006-01-02", fimStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"erro": "formato de 'fim' inválido — use 2006-01-02"})
			return
		}
		fim = fim.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

		vendas, err := h.service.ListarPorPeriodo(inicio, fim)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
			return
		}
		c.JSON(http.StatusOK, vendas)
		return
	}

	// Sem filtro: retorna todas as vendas do dia atual.
	hoje := time.Now().Truncate(24 * time.Hour)
	fim := hoje.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	vendas, err := h.service.ListarPorPeriodo(hoje, fim)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "erro ao listar vendas"})
		return
	}

	c.JSON(http.StatusOK, vendas)
}