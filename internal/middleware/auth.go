// Package middleware fornece os middlewares HTTP da API.
package middleware

import (
	"net/http"
	"strings"
	"super-br/internal/auth"

	"github.com/gin-gonic/gin"
)

// Autenticar retorna um middleware que valida o token JWT presente
// no header Authorization (formato "Bearer <token>").
// O jwtSecret deve ser injetado via closure a partir da Config central.
func Autenticar(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"erro": "token não informado"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"erro": "formato de token inválido — use: Bearer <token>"})
			c.Abort()
			return
		}

		claims, err := auth.ValidarToken(parts[1], jwtSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"erro": "token inválido ou expirado"})
			c.Abort()
			return
		}

		// Armazena os dados do usuário no contexto para uso nos handlers.
		c.Set("usuario_id", claims.UsuarioID)
		c.Set("login", claims.Login)
		c.Set("perfil", claims.Perfil)

		c.Next()
	}
}

// ExigirPerfil retorna um middleware que verifica se o usuário autenticado
// possui um dos perfis aceitos. Superadmin sempre tem acesso liberado.
func ExigirPerfil(perfis ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		perfil, _ := c.Get("perfil")

		// Superadmin tem acesso irrestrito a todos os endpoints.
		if perfil == "superadmin" {
			c.Next()
			return
		}

		for _, p := range perfis {
			if p == perfil {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"erro": "acesso negado — seu perfil não tem permissão para esta ação"})
		c.Abort()
	}
}