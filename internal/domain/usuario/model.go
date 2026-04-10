// Package usuario gerencia os usuários do sistema, autenticação e controle de acesso.
package usuario

import "time"

// Usuario representa um usuário do sistema com suas credenciais e permissões.
// O campo Senha nunca é serializado para JSON (tag json:"-").
// Perfis válidos: "superadmin", "admin", "financeiro", "vendas".
type Usuario struct {
	ID             uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Nome           string    `gorm:"type:varchar(100);not null" json:"nome"`
	Login          string    `gorm:"type:varchar(50);not null;unique" json:"login"`
	Senha          string    `gorm:"type:varchar(255)" json:"-"`
	Perfil         string    `gorm:"type:varchar(20);not null" json:"perfil"`
	PrimeiroAcesso bool      `gorm:"default:true" json:"primeiro_acesso"`
	Ativo          bool      `gorm:"default:true" json:"ativo"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}