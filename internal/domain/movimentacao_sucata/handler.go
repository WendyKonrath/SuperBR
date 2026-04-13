package movimentacao_sucata

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// Handler agrupa os endpoints HTTP do domínio de movimentação de sucata.
type Handler struct {
	service *Service
}

// NewHandler cria o handler com o service injetado.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Listar retorna movimentações de sucata com filtros opcionais via query params.
// Sem filtro: retorna todas.
// ?tipo=entrada_sucata       → somente entradas
// ?tipo=saida_sucata         → somente saídas
// ?sucata_id=1               → movimentações de um tipo de sucata específico
// ?inicio=2025-01-01&fim=2025-01-31 → por período
// GET /api/movimentacoes/sucata
func (h *Handler) Listar(c *gin.Context) {
	// Filtro por sucata
	if sucataIDStr := c.Query("sucata_id"); sucataIDStr != "" {
		sucataID, err := strconv.ParseUint(sucataIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"erro": "sucata_id inválido"})
			return
		}
		movs, err := h.service.ListarPorSucata(uint(sucataID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"erro": "erro ao listar movimentações"})
			return
		}
		c.JSON(http.StatusOK, movs)
		return
	}

	// Filtro por tipo
	if tipo := c.Query("tipo"); tipo != "" {
		movs, err := h.service.ListarPorTipo(tipo)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
			return
		}
		c.JSON(http.StatusOK, movs)
		return
	}

	// Filtro por período — inicio e fim devem vir juntos
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

		movs, err := h.service.ListarPorPeriodo(inicio, fim)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
			return
		}
		c.JSON(http.StatusOK, movs)
		return
	}

	// Sem filtro: retorna todas
	movs, err := h.service.ListarTodas()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "erro ao listar movimentações de sucata"})
		return
	}

	c.JSON(http.StatusOK, movs)
}