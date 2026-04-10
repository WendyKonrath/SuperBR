package usuario

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Handler agrupa os endpoints HTTP do domínio de usuários.
type Handler struct {
	service *Service
}

// NewHandler cria o handler com o service injetado.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// loginInput representa o corpo da requisição de login.
type loginInput struct {
	Login string `json:"login" binding:"required"`
	Senha string `json:"senha"`
}

// primeiroAcessoInput representa o corpo da requisição de definição de senha inicial.
type primeiroAcessoInput struct {
	Login     string `json:"login" binding:"required"`
	NovaSenha string `json:"nova_senha" binding:"required,min=8"`
}

// criarUsuarioInput representa o corpo da requisição de criação de usuário.
type criarUsuarioInput struct {
	Nome   string `json:"nome" binding:"required"`
	Login  string `json:"login" binding:"required"`
	Perfil string `json:"perfil" binding:"required,oneof=admin financeiro vendas"`
}

// atualizarUsuarioInput representa o corpo da requisição de atualização de usuário.
type atualizarUsuarioInput struct {
	Nome   string `json:"nome" binding:"required"`
	Perfil string `json:"perfil" binding:"required,oneof=admin financeiro vendas"`
}

// Login autentica o usuário com login e senha.
// POST /api/auth/login
func (h *Handler) Login(c *gin.Context) {
	var input loginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "informe login e senha"})
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
			"mensagem":        "Defina sua senha antes de continuar",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":           token,
		"primeiro_acesso": false,
	})
}

// PrimeiroAcesso define a senha inicial do usuário após o cadastro.
// POST /api/auth/primeiro-acesso
func (h *Handler) PrimeiroAcesso(c *gin.Context) {
	var input primeiroAcessoInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "nova_senha deve ter no mínimo 8 caracteres"})
		return
	}

	token, err := h.service.PrimeiroAcesso(input.Login, input.NovaSenha)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

// Me retorna os dados do usuário atualmente autenticado.
// GET /api/auth/me
func (h *Handler) Me(c *gin.Context) {
	usuarioID, _ := c.Get("usuario_id")

	u, err := h.service.Me(usuarioID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"erro": err.Error()})
		return
	}

	c.JSON(http.StatusOK, u)
}

// Criar cadastra um novo usuário (somente admin/superadmin).
// POST /api/usuarios
func (h *Handler) Criar(c *gin.Context) {
	var input criarUsuarioInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": "dados inválidos — verifique nome, login e perfil"})
		return
	}

	u, err := h.service.Criar(input.Nome, input.Login, input.Perfil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"erro": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, u)
}

// Atualizar altera nome e perfil de um usuário existente.
// PUT /api/usuarios/:id
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

// Desativar bloqueia o acesso de um usuário.
// PATCH /api/usuarios/:id/desativar
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

// Ativar reativa um usuário desativado.
// PATCH /api/usuarios/:id/ativar
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

// ResetarSenha força o usuário a redefinir a senha no próximo acesso.
// PATCH /api/usuarios/:id/resetar-senha
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

	c.JSON(http.StatusOK, gin.H{"mensagem": "senha resetada — usuário deverá definir nova senha no próximo acesso"})
}

// Listar retorna todos os usuários do sistema.
// GET /api/usuarios
func (h *Handler) Listar(c *gin.Context) {
	usuarios, err := h.service.Listar()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"erro": "erro ao listar usuários"})
		return
	}

	c.JSON(http.StatusOK, usuarios)
}