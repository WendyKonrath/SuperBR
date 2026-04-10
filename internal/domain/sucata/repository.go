package sucata

import "gorm.io/gorm"

// Repository encapsula o acesso ao banco de dados para EstoqueSucata.
type Repository struct {
	db *gorm.DB
}

// NewRepository cria um novo Repository com a conexão injetada.
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// BuscarPorTipo retorna o registro de sucata de um tipo de bateria específico.
// Retorna gorm.ErrRecordNotFound se o tipo ainda não foi cadastrado.
func (r *Repository) BuscarPorTipo(tipoBateria string) (*EstoqueSucata, error) {
	var s EstoqueSucata
	result := r.db.Where("tipo_bateria = ?", tipoBateria).First(&s)
	return &s, result.Error
}

// BuscarPorID retorna o registro de sucata pelo ID primário.
func (r *Repository) BuscarPorID(id uint) (*EstoqueSucata, error) {
	var s EstoqueSucata
	result := r.db.First(&s, id)
	return &s, result.Error
}

// Listar retorna todos os tipos de sucata cadastrados.
func (r *Repository) Listar() ([]EstoqueSucata, error) {
	var sucatas []EstoqueSucata
	result := r.db.Find(&sucatas)
	return sucatas, result.Error
}

// Criar persiste um novo tipo de sucata no banco de dados.
func (r *Repository) Criar(s *EstoqueSucata) error {
	return r.db.Create(s).Error
}

// Atualizar salva as alterações de um registro de sucata existente.
func (r *Repository) Atualizar(s *EstoqueSucata) error {
	return r.db.Save(s).Error
}