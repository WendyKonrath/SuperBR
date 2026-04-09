package auth

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UsuarioID uint   `json:"usuario_id"`
	Login     string `json:"login"`
	Perfil    string `json:"perfil"`
	jwt.RegisteredClaims
}

func GerarToken(usuarioID uint, login string, perfil string) (string, error) {
	claims := Claims{
		UsuarioID: usuarioID,
		Login:     login,
		Perfil:    perfil,
		RegisteredClaims: jwt.RegisteredClaims{
			// Token expira em 8 horas
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(8 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

func ValidarToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Garante que o método de assinatura é o correto
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("método de assinatura inválido")
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
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