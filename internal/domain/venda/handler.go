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

// itemVendaInput representa um produto a ser vendido com o tipo de preço desejado.
type itemVendaInput struct {
	ProdutoID uint   `json:"produto_id" binding:"required"`
	TipoPreco string `json:"tipo_preco" binding:"required,oneof=atacado varejo"`
}

// pagamentoVendaInput representa uma forma de pagamento para a venda.
type pagamentoVendaInput struct {
	Tipo  string  `json:"tipo" binding:"required,oneof=pix credito debito dinheiro sucata"`
	Valor float64 `json:"valor" binding:"required,gt=0"`
}

// criarVendaInput representa o corpo completo da requisição de criação de venda.
// Observacoes é opcional — aparece no campo "Observações" do comprovante PDF.
type criarVendaInput struct {
	NomeCliente      string                `json:"nome_cliente" binding:"required"`
	DocumentoCliente string                `json:"documento_cliente"`
	TelefoneCliente  string                `json:"telefone_cliente"`
	Observacoes      string                `json:"observacoes"`
	Itens            []itemVendaInput      `json:"itens" binding:"required,min=1,dive"`
	Pagamentos       []pagamentoVendaInput `json:"pagamentos" binding:"omitempty,dive"`
}

// traduzirErroBinding converte mensagens técnicas do validator em mensagens amigáveis.
func traduzirErroBinding(err error) string {
	msg := err.Error()

	if contains(msg, "Itens") && contains(msg, "min") {
		return "a venda deve conter ao menos um item"
	}
	if contains(msg, "Itens") && contains(msg, "required") {
		return "informe ao menos um item na venda"
	}
	if contains(msg, "ProdutoID") && contains(msg, "required") {
		return "produto_id é obrigatório em cada item"
	}
	if contains(msg, "TipoPreco") && contains(msg, "oneof") {
		return "tipo_preco inválido — use 'atacado' ou 'varejo'"
	}
	if contains(msg, "Pagamentos") && contains(msg, "Valor") {
		return "valor de pagamento deve ser maior que zero"
	}
	if contains(msg, "Pagamentos") && contains(msg, "Tipo") && contains(msg, "oneof") {
		return "tipo de pagamento inválido — use: pix, credito, debito, dinheiro ou sucata"
	}
	if contains(msg, "NomeCliente") && contains(msg, "required") {
		return "nome_cliente é obrigatório"
	}

	return "dados inválidos — verifique os campos enviados"
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// CriarVenda inicia uma nova venda com status "pendente".
// POST /api/vendas
func (h *Handler) CriarVenda(c *gin.Context) {
	var input criarVendaInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": traduzirErroBinding(err)})
		return
	}

	usuarioID, _ := c.Get("usuario_id")

	itens := make([]itemInput, len(input.Itens))
	for i, it := range input.Itens {
		itens[i] = itemInput{ProdutoID: it.ProdutoID, TipoPreco: it.TipoPreco}
	}

	pags := make([]pagamentoInput, len(input.Pagamentos))
	for i, pg := range input.Pagamentos {
		pags[i] = pagamentoInput{Tipo: pg.Tipo, Valor: pg.Valor}
	}

	v, err := h.service.CriarVenda(
		input.NomeCliente,
		input.DocumentoCliente,
		input.TelefoneCliente,
		input.Observacoes,
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

// BuscarPorID retorna os detalhes de uma venda pelo ID.
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
// Sem filtro: retorna as vendas do dia atual.
// GET /api/vendas
// GET /api/vendas?status=pendente
// GET /api/vendas?inicio=2025-01-01&fim=2025-01-31
func (h *Handler) Listar(c *gin.Context) {
	if status := c.Query("status"); status != "" {
		vendas, err := h.service.ListarPorStatus(status)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
			return
		}
		c.JSON(http.StatusOK, vendas)
		return
	}

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

	hoje := time.Now().Truncate(24 * time.Hour)
	fim := hoje.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	vendas, err := h.service.ListarPorPeriodo(hoje, fim)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "erro ao listar vendas"})
		return
	}

	c.JSON(http.StatusOK, vendas)
}