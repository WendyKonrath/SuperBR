package estoque

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Repository encapsula o acesso ao banco de dados para ItemEstoque e Estoque.
type Repository struct {
	db *gorm.DB
}

// NewRepository cria um novo Repository com a conexão injetada.
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// DB expõe a conexão para uso em transações iniciadas no service.
func (r *Repository) DB() *gorm.DB {
	return r.db
}

// =====================
// ItemEstoque
// =====================

// CriarItem persiste um novo item de estoque no banco de dados.
func (r *Repository) CriarItem(item *ItemEstoque) error {
	return r.db.Create(item).Error
}

// BuscarItemPorID retorna um item de estoque pelo ID, carregando o Produto associado.
func (r *Repository) BuscarItemPorID(id uint) (*ItemEstoque, error) {
	var item ItemEstoque
	result := r.db.Preload("Produto").First(&item, id)
	return &item, result.Error
}

// ListarItens retorna todos os itens de estoque cadastrados.
func (r *Repository) ListarItens() ([]ItemEstoque, error) {
	var itens []ItemEstoque
	result := r.db.Preload("Produto").Find(&itens)
	return itens, result.Error
}

// ListarItensPorProduto retorna os itens de um produto específico.
func (r *Repository) ListarItensPorProduto(produtoID uint) ([]ItemEstoque, error) {
	var itens []ItemEstoque
	result := r.db.Preload("Produto").Where("produto_id = ?", produtoID).Find(&itens)
	return itens, result.Error
}

// ListarItensPorEstado filtra itens pelo estado (ex: "disponivel").
func (r *Repository) ListarItensPorEstado(estado string) ([]ItemEstoque, error) {
	var itens []ItemEstoque
	result := r.db.Preload("Produto").Where("estado = ?", estado).Find(&itens)
	return itens, result.Error
}

// AtualizarItem salva as alterações de um item de estoque existente.
func (r *Repository) AtualizarItem(item *ItemEstoque) error {
	return r.db.Save(item).Error
}

// BuscarItemDisponivel localiza o primeiro item disponível de um produto
// usando SELECT FOR UPDATE para evitar condição de corrida em vendas concorrentes.
// Deve ser chamado dentro de uma transação.
func (r *Repository) BuscarItemDisponivel(produtoID uint, tx *gorm.DB) (*ItemEstoque, error) {
	var item ItemEstoque
	result := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("produto_id = ? AND estado = ?", produtoID, "disponivel").
		First(&item)
	return &item, result.Error
}

// =====================
// Estoque (resumo por produto)
// =====================

// BuscarEstoquePorProduto retorna o resumo de estoque de um produto.
func (r *Repository) BuscarEstoquePorProduto(produtoID uint) (*Estoque, error) {
	var e Estoque
	result := r.db.Preload("Produto").Where("produto_id = ?", produtoID).First(&e)
	return &e, result.Error
}

// ListarEstoque retorna o resumo de estoque de todos os produtos.
func (r *Repository) ListarEstoque() ([]Estoque, error) {
	var estoques []Estoque
	result := r.db.Preload("Produto").Find(&estoques)
	return estoques, result.Error
}

// CriarEstoque persiste um novo resumo de estoque para um produto.
func (r *Repository) CriarEstoque(e *Estoque) error {
	return r.db.Create(e).Error
}

// AtualizarEstoque salva as alterações no resumo de estoque.
func (r *Repository) AtualizarEstoque(e *Estoque) error {
	return r.db.Save(e).Error
}