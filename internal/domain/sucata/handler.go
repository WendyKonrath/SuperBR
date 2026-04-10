package sucata

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

type cadastrarTipoInput struct {
	TipoBateria   string  `json:"tipo_bateria" binding:"required"`
	PesoUnitario  float64 `json:"peso_unitario" binding:"required,min=0"`
	ValorUnitario float64 `json:"valor_unitario" binding:"required,min=0"`
}

type atualizarTipoInput struct {
	PesoUnitario  float64 `json:"peso_unitario" binding:"required,min=0"`
	ValorUnitario float64 `json:"valor_unitario" binding:"required,min=0"`
}

type movimentacaoSucataInput struct {
	TipoBateria string `json:"tipo_bateria" binding:"required"`
	Qtd         int    `json:"qtd" binding:"required,min=1"`
}

func (h *Handler) CadastrarTipo(c *gin.Context) {
	var input cadastrarTipoInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "dados inválidos"})
		return
	}

	sucata, err := h.service.CadastrarTipo(input.TipoBateria, input.PesoUnitario, input.ValorUnitario)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, sucata)
}

func (h *Handler) AtualizarTipo(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "id inválido"})
		return
	}

	var input atualizarTipoInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "dados inválidos"})
		return
	}

	sucata, err := h.service.AtualizarTipo(uint(id), input.PesoUnitario, input.ValorUnitario)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	c.JSON(http.StatusOK, sucata)
}

func (h *Handler) EntradaSucata(c *gin.Context) {
	var input movimentacaoSucataInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "dados inválidos"})
		return
	}

	sucata, err := h.service.EntradaSucata(input.TipoBateria, input.Qtd)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	c.JSON(http.StatusOK, sucata)
}

func (h *Handler) SaidaSucata(c *gin.Context) {
	var input movimentacaoSucataInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "dados inválidos"})
		return
	}

	sucata, err := h.service.SaidaSucata(input.TipoBateria, input.Qtd)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	c.JSON(http.StatusOK, sucata)
}

func (h *Handler) Listar(c *gin.Context) {
	sucatas, err := h.service.Listar()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "erro ao listar sucata"})
		return
	}

	c.JSON(http.StatusOK, sucatas)
}

func (h *Handler) BuscarPorTipo(c *gin.Context) {
	tipo := c.Param("tipo")

	sucata, err := h.service.BuscarPorTipo(tipo)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"erro": err.Error()})
		return
	}

	c.JSON(http.StatusOK, sucata)
}