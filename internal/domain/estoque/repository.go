package estoque

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// =====================
// ItemEstoque
// =====================

func (r *Repository) CriarItem(item *ItemEstoque) error {
	return r.db.Create(item).Error
}

func (r *Repository) BuscarItemPorID(id uint) (*ItemEstoque, error) {
	var item ItemEstoque
	// Preload traz os dados do Produto junto
	result := r.db.Preload("Produto").First(&item, id)
	return &item, result.Error
}

func (r *Repository) ListarItens() ([]ItemEstoque, error) {
	var itens []ItemEstoque
	result := r.db.Preload("Produto").Find(&itens)
	return itens, result.Error
}

func (r *Repository) ListarItensPorProduto(produtoID uint) ([]ItemEstoque, error) {
	var itens []ItemEstoque
	result := r.db.Preload("Produto").Where("produto_id = ?", produtoID).Find(&itens)
	return itens, result.Error
}

func (r *Repository) ListarItensPorEstado(estado string) ([]ItemEstoque, error) {
	var itens []ItemEstoque
	result := r.db.Preload("Produto").Where("estado = ?", estado).Find(&itens)
	return itens, result.Error
}

func (r *Repository) AtualizarItem(item *ItemEstoque) error {
	return r.db.Save(item).Error
}

// BuscarItemDisponivel busca um item disponível de um produto
// usando FOR UPDATE para evitar condição de corrida na venda
func (r *Repository) BuscarItemDisponivel(produtoID uint, tx *gorm.DB) (*ItemEstoque, error) {
	var item ItemEstoque
	result := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("produto_id = ? AND estado = ?", produtoID, "disponivel").
		First(&item)
	return &item, result.Error
}

// =====================
// Estoque (resumo)
// =====================

func (r *Repository) BuscarEstoquePorProduto(produtoID uint) (*Estoque, error) {
	var estoque Estoque
	result := r.db.Preload("Produto").Where("produto_id = ?", produtoID).First(&estoque)
	return &estoque, result.Error
}

func (r *Repository) ListarEstoque() ([]Estoque, error) {
	var estoques []Estoque
	result := r.db.Preload("Produto").Find(&estoques)
	return estoques, result.Error
}

func (r *Repository) CriarEstoque(e *Estoque) error {
	return r.db.Create(e).Error
}

func (r *Repository) AtualizarEstoque(e *Estoque) error {
	return r.db.Save(e).Error
}

func (r *Repository) DB() *gorm.DB {
	return r.db
}