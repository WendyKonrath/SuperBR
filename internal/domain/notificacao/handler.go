package notificacao

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Handler agrupa os endpoints HTTP do domínio de notificações.
type Handler struct {
	service *Service
}

// NewHandler cria o handler com o service injetado.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Listar retorna as notificações do usuário autenticado.
// GET /api/notificacoes
// GET /api/notificacoes?apenas_nao_lidas=true
func (h *Handler) Listar(c *gin.Context) {
	usuarioID, _ := c.Get("usuario_id")

	apenasNaoLidas := c.Query("apenas_nao_lidas") == "true"

	notificacoes, err := h.service.Listar(usuarioID.(uint), apenasNaoLidas)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "erro ao listar notificações"})
		return
	}

	c.JSON(http.StatusOK, notificacoes)
}

// MarcarComoLida marca uma notificação específica como lida.
// PATCH /api/notificacoes/:id/ler
func (h *Handler) MarcarComoLida(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "id inválido"})
		return
	}

	usuarioID, _ := c.Get("usuario_id")

	if err := h.service.MarcarComoLida(uint(id), usuarioID.(uint)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"mensagem": "notificação marcada como lida"})
}

// MarcarTodasComoLidas marca todas as notificações do usuário como lidas.
// PATCH /api/notificacoes/ler-todas
func (h *Handler) MarcarTodasComoLidas(c *gin.Context) {
	usuarioID, _ := c.Get("usuario_id")

	if err := h.service.MarcarTodasComoLidas(usuarioID.(uint)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "erro ao marcar notificações"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"mensagem": "todas as notificações marcadas como lidas"})
}