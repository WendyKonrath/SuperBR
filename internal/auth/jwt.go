// Package auth fornece funções de geração e validação de tokens JWT
// utilizados para autenticar os usuários da API.
package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims representa o payload do JWT com os dados do usuário autenticado.
// Embeds jwt.RegisteredClaims para incluir campos padrão como ExpiresAt.
type Claims struct {
	UsuarioID uint   `json:"usuario_id"`
	Login     string `json:"login"`
	Perfil    string `json:"perfil"`
	jwt.RegisteredClaims
}

// duracaoToken define por quanto tempo um token permanece válido após a emissão.
const duracaoToken = 8 * time.Hour

// GerarToken cria e assina um JWT com os dados do usuário informado.
// O secret deve ser lido da configuração central e injetado aqui —
// nunca use os.Getenv diretamente nesta função.
func GerarToken(usuarioID uint, login, perfil, secret string) (string, error) {
	claims := Claims{
		UsuarioID: usuarioID,
		Login:     login,
		Perfil:    perfil,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duracaoToken)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidarToken verifica a assinatura e a validade do tokenString.
// Retorna os Claims extraídos caso o token seja legítimo e não expirado.
func ValidarToken(tokenString, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Garante que somente HMAC é aceito — rejeita algoritmos como "none" ou RSA inesperado.
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("método de assinatura inválido")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("token inválido")
	}

	return claims, nil
}