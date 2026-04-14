// Package comprovante fornece o endpoint HTTP para geração e download
// do comprovante de venda em PDF.
package comprovante

import (
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"super-br/internal/domain/venda"
)

// Handler agrupa os endpoints HTTP do domínio de comprovante.
type Handler struct {
	service      *Service
	vendaService *venda.Service
}

// NewHandler cria o handler com os services injetados.
func NewHandler(service *Service, vendaService *venda.Service) *Handler {
	return &Handler{
		service:      service,
		vendaService: vendaService,
	}
}

// Gerar gera o comprovante PDF de uma venda e retorna o arquivo para download.
// A venda precisa estar com status "concluida" para gerar o comprovante.
// GET /api/vendas/:id/comprovante
func (h *Handler) Gerar(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "id inválido"})
		return
	}

	// Busca a venda com todos os relacionamentos carregados.
	v, err := h.vendaService.BuscarPorID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"erro": "venda não encontrada"})
		return
	}

	// Apenas vendas concluídas geram comprovante.
	// Vendas pendentes ainda podem ser canceladas — não faz sentido assinar.
	if v.Status != "concluida" {
		c.JSON(http.StatusBadRequest, gin.H{
			"erro": "comprovante só pode ser gerado para vendas com status 'concluida'",
		})
		return
	}

	// Gera o PDF em disco.
	caminho, err := h.service.GerarComprovante(v)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "erro ao gerar comprovante"})
		return
	}

	// Retorna o arquivo como download direto no navegador/Postman.
	nomeArquivo := filepath.Base(caminho)
	c.Header("Content-Disposition", "attachment; filename="+nomeArquivo)
	c.Header("Content-Type", "application/pdf")
	c.File(caminho)
}