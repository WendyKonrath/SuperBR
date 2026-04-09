package notificacao

import (
	"super-br/internal/domain/usuario"
	"time"
)

type Notificacao struct {
	ID        uint            `gorm:"primaryKey;autoIncrement" json:"id"`
	UsuarioID uint            `gorm:"not null" json:"usuario_id"`
	Usuario   usuario.Usuario `gorm:"foreignKey:UsuarioID" json:"usuario"`
	Tipo      string          `gorm:"type:varchar(50);not null" json:"tipo"`
	Mensagem  string          `gorm:"type:text;not null" json:"mensagem"`
	Lida      bool            `gorm:"default:false" json:"lida"`
	CreatedAt time.Time       `json:"created_at"`
}