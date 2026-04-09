package usuario

import "gorm.io/gorm"

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) BuscarPorLogin(login string) (*Usuario, error) {
	var u Usuario
	result := r.db.Where("login = ?", login).First(&u)
	return &u, result.Error
}

func (r *Repository) BuscarPorID(id uint) (*Usuario, error) {
	var u Usuario
	result := r.db.First(&u, id)
	return &u, result.Error
}

func (r *Repository) Criar(u *Usuario) error {
	return r.db.Create(u).Error
}

func (r *Repository) Atualizar(u *Usuario) error {
	return r.db.Save(u).Error
}

func (r *Repository) Listar() ([]Usuario, error) {
	var usuarios []Usuario
	result := r.db.Find(&usuarios)
	return usuarios, result.Error
}