package venda

import (
	"gorm.io/gorm"
	"time"
)

// Repository encapsula o acesso ao banco de dados para Venda, ItemVenda e Pagamento.
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

// Criar persiste uma nova venda no banco dentro de uma transação existente.
func (r *Repository) Criar(tx *gorm.DB, v *Venda) error {
	return tx.Create(v).Error
}

// BuscarPorID retorna uma venda pelo ID, carregando itens e pagamentos.
func (r *Repository) BuscarPorID(id uint) (*Venda, error) {
	var v Venda
	result := r.db.
		Preload("Usuario").
		Preload("Itens.ItemEstoque.Produto").
		Preload("Pagamentos").
		First(&v, id)
	return &v, result.Error
}

// Atualizar salva as alterações de uma venda existente dentro de uma transação.
func (r *Repository) Atualizar(tx *gorm.DB, v *Venda) error {
	return tx.Save(v).Error
}

// ListarPorPeriodo retorna vendas dentro de um intervalo de datas,
// ordenadas da mais recente para a mais antiga.
func (r *Repository) ListarPorPeriodo(inicio, fim time.Time) ([]Venda, error) {
	var vendas []Venda
	result := r.db.
		Preload("Usuario").
		Preload("Itens.ItemEstoque.Produto").
		Preload("Pagamentos").
		Where("data BETWEEN ? AND ?", inicio, fim).
		Order("data DESC").
		Find(&vendas)
	return vendas, result.Error
}

// ListarPorStatus retorna todas as vendas com um status específico.
func (r *Repository) ListarPorStatus(status string) ([]Venda, error) {
	var vendas []Venda
	result := r.db.
		Preload("Usuario").
		Preload("Itens.ItemEstoque.Produto").
		Preload("Pagamentos").
		Where("status = ?", status).
		Order("data DESC").
		Find(&vendas)
	return vendas, result.Error
}

// CriarItemVenda persiste um item de venda dentro de uma transação existente.
func (r *Repository) CriarItemVenda(tx *gorm.DB, item *ItemVenda) error {
	return tx.Create(item).Error
}

// CriarPagamento persiste um pagamento dentro de uma transação existente.
func (r *Repository) CriarPagamento(tx *gorm.DB, p *Pagamento) error {
	return tx.Create(p).Error
}

// ListarItensDaVenda retorna todos os itens de uma venda específica.
func (r *Repository) ListarItensDaVenda(vendaID uint) ([]ItemVenda, error) {
	var itens []ItemVenda
	result := r.db.
		Preload("ItemEstoque.Produto").
		Where("venda_id = ?", vendaID).
		Find(&itens)
	return itens, result.Error
}

// ListarItensDaVendaTx retorna os itens de uma venda dentro de uma transação existente.
// Usado no cancelamento para garantir que a reversão de estoque seja atômica.
func (r *Repository) ListarItensDaVendaTx(tx *gorm.DB, vendaID uint) ([]ItemVenda, error) {
	var itens []ItemVenda
	result := tx.Where("venda_id = ?", vendaID).Find(&itens)
	return itens, result.Error
}