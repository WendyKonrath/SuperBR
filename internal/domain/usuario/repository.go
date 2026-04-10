package usuario

import "gorm.io/gorm"

// Repository encapsula o acesso ao banco de dados para a entidade Usuario.
type Repository struct {
	db *gorm.DB
}

// NewRepository cria um novo Repository com a conexão injetada.
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// BuscarPorLogin retorna o usuário com o login informado.
// Retorna gorm.ErrRecordNotFound se não existir.
func (r *Repository) BuscarPorLogin(login string) (*Usuario, error) {
	var u Usuario
	result := r.db.Where("login = ?", login).First(&u)
	return &u, result.Error
}

// BuscarPorID retorna o usuário com o ID informado.
// Retorna gorm.ErrRecordNotFound se não existir.
func (r *Repository) BuscarPorID(id uint) (*Usuario, error) {
	var u Usuario
	result := r.db.First(&u, id)
	return &u, result.Error
}

// Criar persiste um novo usuário no banco de dados.
func (r *Repository) Criar(u *Usuario) error {
	return r.db.Create(u).Error
}

// Atualizar salva as alterações de um usuário existente.
func (r *Repository) Atualizar(u *Usuario) error {
	return r.db.Save(u).Error
}

// Listar retorna todos os usuários cadastrados, independente do status.
func (r *Repository) Listar() ([]Usuario, error) {
	var usuarios []Usuario
	result := r.db.Find(&usuarios)
	return usuarios, result.Error
}