// Package sucata gerencia o estoque de baterias sucateadas (descartadas).
// Sucata é contabilizada por tipo, peso e valor — diferente do estoque principal
// que rastreia cada bateria individualmente por ID.
package sucata

import "time"

// EstoqueSucata representa o estoque consolidado de sucata de um tipo de bateria.
// Cada tipo de bateria (ex: "60Ah", "45Ah") tem seu próprio registro.
// Peso e valor total são sempre recalculados a partir das quantidades unitárias.
type EstoqueSucata struct {
	ID            uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	TipoBateria   string    `gorm:"type:varchar(50);not null;unique" json:"tipo_bateria"`
	PesoUnitario  float64   `gorm:"type:decimal(10,2);not null" json:"peso_unitario"`
	Qtd           int       `gorm:"not null;default:0" json:"qtd"`
	PesoTotal     float64   `gorm:"type:decimal(10,2);not null;default:0" json:"peso_total"`
	ValorUnitario float64   `gorm:"type:decimal(10,2);not null" json:"valor_unitario"`
	ValorTotal    float64   `gorm:"type:decimal(10,2);not null;default:0" json:"valor_total"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}