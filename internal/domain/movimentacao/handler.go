package movimentacao

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// Handler agrupa os endpoints HTTP do domínio de movimentação.
type Handler struct {
	service *Service
}

// NewHandler cria o handler com o service injetado.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Listar retorna movimentações com filtros opcionais via query params.
// Sem filtro: retorna todas.
// ?tipo=entrada           → somente entradas
// ?tipo=saida             → somente saídas
// ?item_id=1              → movimentações de um item específico
// ?produto_id=1           → movimentações de todos os itens de um produto
// ?inicio=2025-01-01&fim=2025-01-31 → por período
// GET /api/movimentacoes
func (h *Handler) Listar(c *gin.Context) {
	// Filtro por item
	if itemIDStr := c.Query("item_id"); itemIDStr != "" {
		itemID, err := strconv.ParseUint(itemIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"erro": "item_id inválido"})
			return
		}
		movs, err := h.service.ListarPorItem(uint(itemID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"erro": "erro ao listar movimentações"})
			return
		}
		c.JSON(http.StatusOK, movs)
		return
	}

	// Filtro por produto
	if produtoIDStr := c.Query("produto_id"); produtoIDStr != "" {
		produtoID, err := strconv.ParseUint(produtoIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"erro": "produto_id inválido"})
			return
		}
		movs, err := h.service.ListarPorProduto(uint(produtoID))
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
		// Inclui o dia inteiro até 23:59:59
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
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "erro ao listar movimentações"})
		return
	}

	c.JSON(http.StatusOK, movs)
}