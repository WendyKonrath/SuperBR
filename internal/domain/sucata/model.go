package sucata

import "time"

type EstoqueSucata struct {
	ID            uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	TipoBateria   string    `gorm:"type:varchar(50);not null" json:"tipo_bateria"`
	PesoUnitario  float64   `gorm:"type:decimal(10,2);not null" json:"peso_unitario"`
	Qtd           int       `gorm:"not null;default:0" json:"qtd"`
	PesoTotal     float64   `gorm:"type:decimal(10,2);not null;default:0" json:"peso_total"`
	ValorUnitario float64   `gorm:"type:decimal(10,2);not null" json:"valor_unitario"`
	ValorTotal    float64   `gorm:"type:decimal(10,2);not null;default:0" json:"valor_total"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}