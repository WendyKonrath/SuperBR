package movimentacao

import (
	"time"

	"gorm.io/gorm"
)

// Repository encapsula o acesso ao banco de dados para Movimentacao.
type Repository struct {
	db *gorm.DB
}

// NewRepository cria um novo Repository com a conexão injetada.
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Registrar persiste uma nova movimentação dentro de uma transação existente.
// Deve sempre ser chamado com o tx da transação pai para garantir atomicidade.
func (r *Repository) Registrar(tx *gorm.DB, itemID, usuarioID uint, tipo string) error {
	mov := Movimentacao{
		ItemID:    itemID,
		UsuarioID: usuarioID,
		Tipo:      tipo,
		Data:      time.Now(),
	}
	return tx.Create(&mov).Error
}

// ListarPorItem retorna todas as movimentações de um item específico.
func (r *Repository) ListarPorItem(itemID uint) ([]Movimentacao, error) {
	var movs []Movimentacao
	result := r.db.Preload("Item.Produto").Preload("Usuario").
		Where("item_id = ?", itemID).
		Order("data DESC").
		Find(&movs)
	return movs, result.Error
}

// ListarPorPeriodo retorna movimentações dentro de um intervalo de datas.
// Útil para geração de relatórios mensais.
func (r *Repository) ListarPorPeriodo(inicio, fim time.Time) ([]Movimentacao, error) {
	var movs []Movimentacao
	result := r.db.Preload("Item.Produto").Preload("Usuario").
		Where("data BETWEEN ? AND ?", inicio, fim).
		Order("data DESC").
		Find(&movs)
	return movs, result.Error
}