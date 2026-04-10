// Package movimentacao registra todo histórico de entradas e saídas do estoque.
// Cada operação (entrada, saída, empréstimo, devolução) gera uma Movimentacao,
// garantindo rastreabilidade completa conforme exigido pelo documento de escopo.
package movimentacao

import (
	"super-br/internal/domain/usuario"
	"time"
)

// Movimentacao representa um evento de movimentação de um item no estoque.
// Tipos válidos: "entrada", "saida".
type Movimentacao struct {
	ID        uint                `gorm:"primaryKey;autoIncrement" json:"id"`
	ItemID    uint                `gorm:"not null" json:"item_id"`
	Tipo      string              `gorm:"type:varchar(10);not null" json:"tipo"`
	Data      time.Time           `gorm:"not null" json:"data"`
	UsuarioID uint                `gorm:"not null" json:"usuario_id"`
	Usuario   usuario.Usuario     `gorm:"foreignKey:UsuarioID" json:"usuario"`
	CreatedAt time.Time           `json:"created_at"`
}