package estoque

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Handler agrupa os endpoints HTTP do domínio de estoque.
type Handler struct {
	service *Service
}

// NewHandler cria o handler com o service injetado.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// entradaEstoqueInput representa o corpo da requisição de entrada de item.
type entradaEstoqueInput struct {
	ProdutoID uint   `json:"produto_id" binding:"required"`
	CodLote   string `json:"cod_lote" binding:"required"`
}

// saidaEstoqueInput representa o corpo da requisição de saída de item.
type saidaEstoqueInput struct {
	ItemID uint `json:"item_id" binding:"required"`
}

// EntradaEstoque registra a chegada de uma nova bateria no estoque.
// POST /api/estoque/entrada
func (h *Handler) EntradaEstoque(c *gin.Context) {
	var input entradaEstoqueInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "informe produto_id e cod_lote"})
		return
	}

	usuarioID, _ := c.Get("usuario_id")

	item, err := h.service.EntradaEstoque(input.ProdutoID, input.CodLote, usuarioID.(uint))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, item)
}

// SaidaEstoque registra a saída manual de um item do estoque pelo seu ID.
// POST /api/estoque/saida
func (h *Handler) SaidaEstoque(c *gin.Context) {
	var input saidaEstoqueInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "informe item_id"})
		return
	}

	usuarioID, _ := c.Get("usuario_id")

	item, err := h.service.SaidaEstoque(input.ItemID, usuarioID.(uint))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	c.JSON(http.StatusOK, item)
}

// ListarItens retorna itens de estoque com filtros opcionais por produto_id ou estado.
// GET /api/estoque/itens?produto_id=1
// GET /api/estoque/itens?estado=disponivel
func (h *Handler) ListarItens(c *gin.Context) {
	if produtoIDStr := c.Query("produto_id"); produtoIDStr != "" {
		produtoID, err := strconv.ParseUint(produtoIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"erro": "produto_id inválido"})
			return
		}
		itens, err := h.service.ListarItensPorProduto(uint(produtoID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"erro": "erro ao listar itens"})
			return
		}
		c.JSON(http.StatusOK, itens)
		return
	}

	if estado := c.Query("estado"); estado != "" {
		itens, err := h.service.ListarItensPorEstado(estado)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"erro": "erro ao listar itens"})
			return
		}
		c.JSON(http.StatusOK, itens)
		return
	}

	itens, err := h.service.ListarItens()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "erro ao listar itens"})
		return
	}

	c.JSON(http.StatusOK, itens)
}

// BuscarItemPorID retorna um item de estoque pelo seu ID único.
// GET /api/estoque/itens/:id
func (h *Handler) BuscarItemPorID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "id inválido"})
		return
	}

	item, err := h.service.BuscarItemPorID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"erro": err.Error()})
		return
	}

	c.JSON(http.StatusOK, item)
}

// ListarEstoque retorna o resumo consolidado de estoque por produto.
// GET /api/estoque
func (h *Handler) ListarEstoque(c *gin.Context) {
	estoques, err := h.service.ListarEstoque()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "erro ao listar estoque"})
		return
	}

	c.JSON(http.StatusOK, estoques)
}

// BuscarEstoquePorProduto retorna o resumo de estoque de um produto específico.
// GET /api/estoque/produto/:produto_id
func (h *Handler) BuscarEstoquePorProduto(c *gin.Context) {
	produtoID, err := strconv.ParseUint(c.Param("produto_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "produto_id inválido"})
		return
	}

	e, err := h.service.BuscarEstoquePorProduto(uint(produtoID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"erro": err.Error()})
		return
	}

	c.JSON(http.StatusOK, e)
}

// DevolverItem retorna ao estoque um item que havia saído manualmente.
// PATCH /api/estoque/itens/:id/devolver
func (h *Handler) DevolverItem(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "id inválido"})
		return
	}

	usuarioID, _ := c.Get("usuario_id")

	item, err := h.service.DevolverItem(uint(id), usuarioID.(uint))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	c.JSON(http.StatusOK, item)
}

// EmprestarItem marca uma bateria como emprestada, removendo-a do estoque disponível.
// PATCH /api/estoque/itens/:id/emprestar
func (h *Handler) EmprestarItem(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "id inválido"})
		return
	}

	usuarioID, _ := c.Get("usuario_id")

	item, err := h.service.EmprestarItem(uint(id), usuarioID.(uint))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	c.JSON(http.StatusOK, item)
}

// DevolverEmprestimo retorna ao estoque disponível um item que estava emprestado.
// PATCH /api/estoque/itens/:id/devolver-emprestimo
func (h *Handler) DevolverEmprestimo(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "id inválido"})
		return
	}

	usuarioID, _ := c.Get("usuario_id")

	item, err := h.service.DevolverEmprestimo(uint(id), usuarioID.(uint))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	c.JSON(http.StatusOK, item)
}