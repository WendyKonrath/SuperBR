// Package movimentacao_sucata registra o histórico de entradas e saídas
// do estoque de sucata, separado do histórico de baterias individuais.
package movimentacao_sucata

import (
	"super-br/internal/domain/usuario"
	"time"
)

// MovimentacaoSucata representa um evento de movimentação no estoque de sucata.
// Tipos válidos: "entrada_sucata", "saida_sucata".
//
// O campo SucataID referencia estoque_sucatas mas não carrega o struct diretamente
// para evitar importação circular com o pacote sucata.
// Os dados da sucata são retornados via SucataInfo, preenchido pelo repository
// com uma query separada ao listar movimentações.
type MovimentacaoSucata struct {
	ID        uint            `gorm:"primaryKey;autoIncrement" json:"id"`
	SucataID  uint            `gorm:"not null" json:"sucata_id"`
	Tipo      string          `gorm:"type:varchar(20);not null" json:"tipo"`
	Qtd       int             `gorm:"not null" json:"qtd"`
	Data      time.Time       `gorm:"not null" json:"data"`
	UsuarioID uint            `gorm:"not null" json:"usuario_id"`
	Usuario   usuario.Usuario `gorm:"foreignKey:UsuarioID" json:"usuario"`
	CreatedAt time.Time       `json:"created_at"`

	// SucataInfo é preenchido pelo repository ao listar movimentações.
	// Não é uma coluna no banco — é montado via JOIN para exibir
	// o tipo de bateria associado.
	SucataInfo *SucataInfo `gorm:"-" json:"sucata,omitempty"`
}

// TableName força o nome da tabela no banco para evitar pluralização incorreta do GORM.
func (MovimentacaoSucata) TableName() string {
	return "movimentacao_sucatas"
}

// SucataInfo agrupa as informações da sucata para exibição nas movimentações.
// É preenchido manualmente pelo repository via JOIN, sem depender do pacote sucata.
type SucataInfo struct {
	SucataID      uint    `json:"sucata_id"`
	TipoBateria   string  `json:"tipo_bateria"`
	PesoUnitario  float64 `json:"peso_unitario"`
	ValorUnitario float64 `json:"valor_unitario"`
}