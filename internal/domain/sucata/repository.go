package sucata

import "gorm.io/gorm"

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) BuscarPorTipo(tipoBateria string) (*EstoqueSucata, error) {
	var sucata EstoqueSucata
	result := r.db.Where("tipo_bateria = ?", tipoBateria).First(&sucata)
	return &sucata, result.Error
}

func (r *Repository) BuscarPorID(id uint) (*EstoqueSucata, error) {
	var sucata EstoqueSucata
	result := r.db.First(&sucata, id)
	return &sucata, result.Error
}

func (r *Repository) Listar() ([]EstoqueSucata, error) {
	var sucatas []EstoqueSucata
	result := r.db.Find(&sucatas)
	return sucatas, result.Error
}

func (r *Repository) Criar(s *EstoqueSucata) error {
	return r.db.Create(s).Error
}

func (r *Repository) Atualizar(s *EstoqueSucata) error {
	return r.db.Save(s).Error
}

func (r *Repository) DB() *gorm.DB {
	return r.db
}