package produto

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

type produtoInput struct {
	Nome         string  `json:"nome" binding:"required"`
	Categoria    string  `json:"categoria" binding:"required"`
	ValorAtacado float64 `json:"valor_atacado" binding:"required,min=0"`
	ValorVarejo  float64 `json:"valor_varejo" binding:"required,min=0"`
}

func (h *Handler) Criar(c *gin.Context) {
	var input produtoInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "dados inválidos"})
		return
	}

	p, err := h.service.Criar(input.Nome, input.Categoria, input.ValorAtacado, input.ValorVarejo)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, p)
}

func (h *Handler) BuscarPorID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "id inválido"})
		return
	}

	p, err := h.service.BuscarPorID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"erro": err.Error()})
		return
	}

	c.JSON(http.StatusOK, p)
}

func (h *Handler) Listar(c *gin.Context) {
	categoria := c.Query("categoria")
	if categoria != "" {
		produtos, err := h.service.ListarPorCategoria(categoria)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"erro": "erro ao listar produtos"})
			return
		}
		c.JSON(http.StatusOK, produtos)
		return
	}

	produtos, err := h.service.Listar()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "erro ao listar produtos"})
		return
	}

	c.JSON(http.StatusOK, produtos)
}

func (h *Handler) Atualizar(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "id inválido"})
		return
	}

	var input produtoInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "dados inválidos"})
		return
	}

	p, err := h.service.Atualizar(uint(id), input.Nome, input.Categoria, input.ValorAtacado, input.ValorVarejo)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	c.JSON(http.StatusOK, p)
}

func (h *Handler) Deletar(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "id inválido"})
		return
	}

	if err := h.service.Deletar(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"mensagem": "produto deletado com sucesso"})
}