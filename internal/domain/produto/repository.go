package produto

import "gorm.io/gorm"

// Repository encapsula o acesso ao banco de dados para a entidade Produto.
type Repository struct {
	db *gorm.DB
}

// NewRepository cria um novo Repository com a conexão injetada.
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Criar persiste um novo produto no banco de dados.
func (r *Repository) Criar(p *Produto) error {
	return r.db.Create(p).Error
}

// BuscarPorID retorna o produto com o ID informado.
func (r *Repository) BuscarPorID(id uint) (*Produto, error) {
	var p Produto
	result := r.db.First(&p, id)
	return &p, result.Error
}

// BuscarPorNomeECategoria verifica se já existe produto com a mesma combinação.
// Usado para evitar duplicatas no catálogo.
func (r *Repository) BuscarPorNomeECategoria(nome, categoria string) (*Produto, error) {
	var p Produto
	result := r.db.Where("nome = ? AND categoria = ?", nome, categoria).First(&p)
	return &p, result.Error
}

// Listar retorna todos os produtos do catálogo.
func (r *Repository) Listar() ([]Produto, error) {
	var produtos []Produto
	result := r.db.Find(&produtos)
	return produtos, result.Error
}

// ListarPorCategoria retorna os produtos filtrados pela categoria (ex: "60Ah").
func (r *Repository) ListarPorCategoria(categoria string) ([]Produto, error) {
	var produtos []Produto
	result := r.db.Where("categoria = ?", categoria).Find(&produtos)
	return produtos, result.Error
}

// Atualizar salva as alterações de um produto existente.
func (r *Repository) Atualizar(p *Produto) error {
	return r.db.Save(p).Error
}

// Deletar remove um produto do banco de dados pelo ID.
// Só deve ser chamado após verificar que não há itens no estoque.
func (r *Repository) Deletar(id uint) error {
	return r.db.Delete(&Produto{}, id).Error
}

// PossuiItensNoEstoque verifica se o produto tem itens cadastrados em item_estoques.
// Usado para bloquear a exclusão de produtos com estoque.
func (r *Repository) PossuiItensNoEstoque(id uint) (bool, error) {
	var count int64
	result := r.db.Table("item_estoques").Where("produto_id = ?", id).Count(&count)
	return count > 0, result.Error
}