package db

import (
	"fmt"
	"super-br/internal/domain/usuario"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func Seed(db *gorm.DB) {
	// Senha padrão do super admin
	senhaHash, err := bcrypt.GenerateFromPassword([]byte("superadmin123"), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println("Erro ao gerar senha do super admin:", err)
		return
	}

	superAdmin := usuario.Usuario{
		Nome:           "Super Admin",
		Login:          "superadmin",
		Senha:          string(senhaHash),
		Perfil:         "superadmin",
		PrimeiroAcesso: false,
		Ativo:          true,
	}

	// Tenta buscar o super admin no banco
	var existente usuario.Usuario
	result := db.Where("login = ?", "superadmin").First(&existente)

	if result.Error != nil {
		// Não existe ainda — cria
		if err := db.Create(&superAdmin).Error; err != nil {
			fmt.Println("Erro ao criar super admin:", err)
			return
		}
		fmt.Println("Super Admin criado com sucesso!")
	} else {
		// Já existe — garante que está sempre ativo e com perfil correto
		existente.Perfil = "superadmin"
		existente.Ativo = true
		existente.PrimeiroAcesso = false
		if err := db.Save(&existente).Error; err != nil {
			fmt.Println("Erro ao atualizar super admin:", err)
			return
		}
		fmt.Println("Super Admin verificado e atualizado!")
	}
}