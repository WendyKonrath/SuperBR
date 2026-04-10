// Package db também contém o seed inicial do banco de dados.
// O seed garante que o superadmin sempre existe ao iniciar a aplicação.
package db

import (
	"fmt"
	"super-br/internal/domain/usuario"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Seed garante a existência do usuário superadmin no banco de dados.
// Se o superadmin não existir, ele é criado com a senha padrão.
// Se já existir, seus campos críticos (perfil, ativo) são verificados e corrigidos.
//
// IMPORTANTE: a senha padrão "superadmin123" deve ser alterada imediatamente
// após o primeiro acesso em ambiente de produção.
func Seed(db *gorm.DB) {
	// Usa custo 12 para produção — mais resistente a brute-force que o DefaultCost (10).
	senhaHash, err := bcrypt.GenerateFromPassword([]byte("superadmin123"), 12)
	if err != nil {
		fmt.Println("Erro ao gerar hash da senha do superadmin:", err)
		return
	}

	var existente usuario.Usuario
	result := db.Where("login = ?", "superadmin").First(&existente)

	if result.Error != nil {
		// Superadmin não existe — cria pela primeira vez.
		superAdmin := usuario.Usuario{
			Nome:           "Super Admin",
			Login:          "superadmin",
			Senha:          string(senhaHash),
			Perfil:         "superadmin",
			PrimeiroAcesso: false,
			Ativo:          true,
		}

		if err := db.Create(&superAdmin).Error; err != nil {
			fmt.Println("Erro ao criar superadmin:", err)
			return
		}
		fmt.Println("Superadmin criado com sucesso.")
	} else {
		// Superadmin já existe — garante que perfil e status estão corretos.
		// Não altera a senha para não sobrescrever uma senha já personalizada.
		existente.Perfil = "superadmin"
		existente.Ativo = true
		existente.PrimeiroAcesso = false

		if err := db.Save(&existente).Error; err != nil {
			fmt.Println("Erro ao verificar superadmin:", err)
			return
		}
		fmt.Println("Superadmin verificado.")
	}
}