package estoque

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

type entradaEstoqueInput struct {
	ProdutoID uint   `json:"produto_id" binding:"required"`
	CodLote   string `json:"cod_lote" binding:"required"`
}

type saidaEstoqueInput struct {
	ItemID uint `json:"item_id" binding:"required"`
}

func (h *Handler) EntradaEstoque(c *gin.Context) {
	var input entradaEstoqueInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "dados inválidos"})
		return
	}

	// Pega o usuário logado
	usuarioID, _ := c.Get("usuario_id")

	item, err := h.service.EntradaEstoque(input.ProdutoID, input.CodLote, usuarioID.(uint))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, item)
}

func (h *Handler) SaidaEstoque(c *gin.Context) {
	var input saidaEstoqueInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "dados inválidos"})
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

func (h *Handler) ListarItens(c *gin.Context) {
	// Filtro por produto
	produtoIDStr := c.Query("produto_id")
	if produtoIDStr != "" {
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

	// Filtro por estado
	estado := c.Query("estado")
	if estado != "" {
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

func (h *Handler) ListarEstoque(c *gin.Context) {
	estoques, err := h.service.ListarEstoque()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "erro ao listar estoque"})
		return
	}

	c.JSON(http.StatusOK, estoques)
}

func (h *Handler) BuscarEstoquePorProduto(c *gin.Context) {
	produtoID, err := strconv.ParseUint(c.Param("produto_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "produto_id inválido"})
		return
	}

	estoque, err := h.service.BuscarEstoquePorProduto(uint(produtoID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"erro": err.Error()})
		return
	}

	c.JSON(http.StatusOK, estoque)
}

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