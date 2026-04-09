package produto

import "time"

type Produto struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Nome         string    `gorm:"type:varchar(100);not null" json:"nome"`
	Categoria    string    `gorm:"type:varchar(50);not null" json:"categoria"`
	ValorAtacado float64   `gorm:"type:decimal(10,2);not null" json:"valor_atacado"`
	ValorVarejo  float64   `gorm:"type:decimal(10,2);not null" json:"valor_varejo"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}