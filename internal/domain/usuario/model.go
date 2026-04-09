package usuario

import "time"

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