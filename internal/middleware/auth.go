package middleware

import (
	"net/http"
	"strings"
	"super-br/internal/auth"

	"github.com/gin-gonic/gin"
)

func Autenticar() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"erro": "token não informado"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"erro": "formato de token inválido"})
			c.Abort()
			return
		}

		claims, err := auth.ValidarToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"erro": "token inválido ou expirado"})
			c.Abort()
			return
		}

		c.Set("usuario_id", claims.UsuarioID)
		c.Set("login", claims.Login)
		c.Set("perfil", claims.Perfil)

		c.Next()
	}
}

func ExigirPerfil(perfis ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		perfil, _ := c.Get("perfil")

		// Superadmin tem acesso a tudo sempre
		if perfil == "superadmin" {
			c.Next()
			return
		}

		// Verifica se o perfil está na lista de permitidos
		for _, p := range perfis {
			if p == perfil {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"erro": "acesso negado"})
		c.Abort()
	}
}