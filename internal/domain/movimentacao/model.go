package movimentacao

import (
	"super-br/internal/domain/estoque"
	"super-br/internal/domain/usuario"
	"time"
)

type Movimentacao struct {
	ID          uint                `gorm:"primaryKey;autoIncrement" json:"id"`
	ItemID      uint                `gorm:"not null" json:"item_id"`
	Item        estoque.ItemEstoque `gorm:"foreignKey:ItemID" json:"item"`
	Tipo        string              `gorm:"type:varchar(10);not null" json:"tipo"`
	Data        time.Time           `gorm:"not null" json:"data"`
	UsuarioID   uint                `gorm:"not null" json:"usuario_id"`
	Usuario     usuario.Usuario     `gorm:"foreignKey:UsuarioID" json:"usuario"`
	CreatedAt   time.Time           `json:"created_at"`
}