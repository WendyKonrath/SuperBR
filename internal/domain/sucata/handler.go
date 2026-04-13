package sucata

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Handler agrupa os endpoints HTTP do domínio de sucata.
type Handler struct {
	service *Service
}

// NewHandler cria o handler com o service injetado.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

type cadastrarTipoInput struct {
	TipoBateria   string  `json:"tipo_bateria" binding:"required"`
	PesoUnitario  float64 `json:"peso_unitario" binding:"required,gt=0"`
	ValorUnitario float64 `json:"valor_unitario" binding:"required,gt=0"`
}

type atualizarTipoInput struct {
	PesoUnitario  float64 `json:"peso_unitario" binding:"required,gt=0"`
	ValorUnitario float64 `json:"valor_unitario" binding:"required,gt=0"`
}

type movimentacaoSucataInput struct {
	TipoBateria string `json:"tipo_bateria" binding:"required"`
	Qtd         int    `json:"qtd" binding:"required,min=1"`
}

// CadastrarTipo registra um novo tipo de sucata com peso e valor unitário.
// POST /api/sucata/tipos
func (h *Handler) CadastrarTipo(c *gin.Context) {
	var input cadastrarTipoInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "informe tipo_bateria, peso_unitario e valor_unitario (ambos > 0)"})
		return
	}

	sucata, err := h.service.CadastrarTipo(input.TipoBateria, input.PesoUnitario, input.ValorUnitario)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, sucata)
}

// AtualizarTipo altera peso e valor unitário de um tipo de sucata existente.
// PUT /api/sucata/tipos/:id
func (h *Handler) AtualizarTipo(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "id inválido"})
		return
	}

	var input atualizarTipoInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "informe peso_unitario e valor_unitario (ambos > 0)"})
		return
	}

	sucata, err := h.service.AtualizarTipo(uint(id), input.PesoUnitario, input.ValorUnitario)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	c.JSON(http.StatusOK, sucata)
}

// EntradaSucata registra a chegada de unidades de sucata.
// POST /api/sucata/entrada
func (h *Handler) EntradaSucata(c *gin.Context) {
	var input movimentacaoSucataInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "informe tipo_bateria e qtd (mínimo 1)"})
		return
	}

	usuarioID, _ := c.Get("usuario_id")

	sucata, err := h.service.EntradaSucata(input.TipoBateria, input.Qtd, usuarioID.(uint))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	c.JSON(http.StatusOK, sucata)
}

// SaidaSucata registra a saída de unidades de sucata.
// POST /api/sucata/saida
func (h *Handler) SaidaSucata(c *gin.Context) {
	var input movimentacaoSucataInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "informe tipo_bateria e qtd (mínimo 1)"})
		return
	}

	usuarioID, _ := c.Get("usuario_id")

	sucata, err := h.service.SaidaSucata(input.TipoBateria, input.Qtd, usuarioID.(uint))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	c.JSON(http.StatusOK, sucata)
}

// Listar retorna todos os tipos de sucata com seus estoques atuais.
// GET /api/sucata
func (h *Handler) Listar(c *gin.Context) {
	sucatas, err := h.service.Listar()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "erro ao listar sucata"})
		return
	}

	c.JSON(http.StatusOK, sucatas)
}

// BuscarPorTipo retorna o estoque de sucata de um tipo específico.
// GET /api/sucata/:tipo
func (h *Handler) BuscarPorTipo(c *gin.Context) {
	tipo := c.Param("tipo")

	sucata, err := h.service.BuscarPorTipo(tipo)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"erro": err.Error()})
		return
	}

	c.JSON(http.StatusOK, sucata)
}