package produto

import "gorm.io/gorm"

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Criar(p *Produto) error {
	return r.db.Create(p).Error
}

func (r *Repository) BuscarPorID(id uint) (*Produto, error) {
	var p Produto
	result := r.db.First(&p, id)
	return &p, result.Error
}

func (r *Repository) BuscarPorNomeECategoria(nome, categoria string) (*Produto, error) {
	var p Produto
	result := r.db.Where("nome = ? AND categoria = ?", nome, categoria).First(&p)
	return &p, result.Error
}

func (r *Repository) Listar() ([]Produto, error) {
	var produtos []Produto
	result := r.db.Find(&produtos)
	return produtos, result.Error
}

func (r *Repository) ListarPorCategoria(categoria string) ([]Produto, error) {
	var produtos []Produto
	result := r.db.Where("categoria = ?", categoria).Find(&produtos)
	return produtos, result.Error
}

func (r *Repository) Atualizar(p *Produto) error {
	return r.db.Save(p).Error
}

func (r *Repository) Deletar(id uint) error {
	return r.db.Delete(&Produto{}, id).Error
}

func (r *Repository) PossuiItensNoEstoque(id uint) (bool, error) {
	var count int64
	result := r.db.Table("item_estoques").Where("produto_id = ?", id).Count(&count)
	return count > 0, result.Error
}