package usuario

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

type loginInput struct {
	Login string `json:"login" binding:"required"`
	Senha string `json:"senha"`
}

type primeiroAcessoInput struct {
	Login     string `json:"login" binding:"required"`
	NovaSenha string `json:"nova_senha" binding:"required,min=6"`
}

type criarUsuarioInput struct {
	Nome   string `json:"nome" binding:"required"`
	Login  string `json:"login" binding:"required"`
	Perfil string `json:"perfil" binding:"required,oneof=admin financeiro vendas"`
}

type atualizarUsuarioInput struct {
	Nome   string `json:"nome" binding:"required"`
	Perfil string `json:"perfil" binding:"required,oneof=admin financeiro vendas"`
}

func (h *Handler) Login(c *gin.Context) {
	var input loginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "dados inválidos"})
		return
	}

	token, primeiroAcesso, err := h.service.Login(input.Login, input.Senha)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"erro": err.Error()})
		return
	}

	if primeiroAcesso {
		c.JSON(http.StatusOK, gin.H{
			"primeiro_acesso": true,
			"mensagem":        "cadastre sua senha para continuar",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":           token,
		"primeiro_acesso": false,
	})
}

func (h *Handler) PrimeiroAcesso(c *gin.Context) {
	var input primeiroAcessoInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "dados inválidos"})
		return
	}

	token, err := h.service.PrimeiroAcesso(input.Login, input.NovaSenha)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (h *Handler) Me(c *gin.Context) {
	// Pega o id do usuário logado que foi salvo pelo middleware
	usuarioID, _ := c.Get("usuario_id")

	u, err := h.service.Me(usuarioID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"erro": err.Error()})
		return
	}

	c.JSON(http.StatusOK, u)
}

func (h *Handler) Criar(c *gin.Context) {
	var input criarUsuarioInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "dados inválidos"})
		return
	}

	u, err := h.service.Criar(input.Nome, input.Login, input.Perfil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, u)
}

func (h *Handler) Atualizar(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "id inválido"})
		return
	}

	var input atualizarUsuarioInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "dados inválidos"})
		return
	}

	u, err := h.service.Atualizar(uint(id), input.Nome, input.Perfil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	c.JSON(http.StatusOK, u)
}

func (h *Handler) Desativar(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "id inválido"})
		return
	}

	if err := h.service.Desativar(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"mensagem": "usuário desativado com sucesso"})
}

func (h *Handler) ResetarSenha(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "id inválido"})
		return
	}

	if err := h.service.ResetarSenha(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"mensagem": "senha resetada com sucesso — usuário deverá cadastrar nova senha no próximo acesso"})
}

func (h *Handler) Listar(c *gin.Context) {
	usuarios, err := h.service.Listar()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "erro ao listar usuários"})
		return
	}

	c.JSON(http.StatusOK, usuarios)
}

func (h *Handler) Ativar(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "id inválido"})
		return
	}

	if err := h.service.Ativar(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"mensagem": "usuário ativado com sucesso"})
}