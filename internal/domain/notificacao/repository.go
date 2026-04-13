package notificacao

import "gorm.io/gorm"

// Repository encapsula o acesso ao banco de dados para Notificacao.
type Repository struct {
	db *gorm.DB
}

// NewRepository cria um novo Repository com a conexão injetada.
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Criar persiste uma nova notificação dentro de uma transação existente.
// Deve ser chamado com o tx da transação pai para garantir atomicidade.
func (r *Repository) Criar(tx *gorm.DB, n *Notificacao) error {
	return tx.Create(n).Error
}

// ListarPorUsuario retorna todas as notificações de um usuário, da mais recente para a mais antiga.
func (r *Repository) ListarPorUsuario(usuarioID uint) ([]Notificacao, error) {
	var notificacoes []Notificacao
	result := r.db.
		Preload("Usuario").
		Where("usuario_id = ?", usuarioID).
		Order("created_at DESC").
		Find(&notificacoes)
	return notificacoes, result.Error
}

// ListarNaoLidasPorUsuario retorna somente as notificações não lidas de um usuário.
func (r *Repository) ListarNaoLidasPorUsuario(usuarioID uint) ([]Notificacao, error) {
	var notificacoes []Notificacao
	result := r.db.
		Preload("Usuario").
		Where("usuario_id = ? AND lida = false", usuarioID).
		Order("created_at DESC").
		Find(&notificacoes)
	return notificacoes, result.Error
}

// MarcarComoLida marca uma notificação específica como lida.
func (r *Repository) MarcarComoLida(id, usuarioID uint) error {
	result := r.db.Model(&Notificacao{}).
		Where("id = ? AND usuario_id = ?", id, usuarioID).
		Update("lida", true)
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}

// MarcarTodasComoLidas marca todas as notificações de um usuário como lidas.
func (r *Repository) MarcarTodasComoLidas(usuarioID uint) error {
	return r.db.Model(&Notificacao{}).
		Where("usuario_id = ? AND lida = false", usuarioID).
		Update("lida", true).Error
}

// BuscarAdmins retorna os IDs de todos os usuários com perfil admin ou superadmin.
// Usado para enviar notificações para todos os administradores do sistema.
func (r *Repository) BuscarAdmins() ([]uint, error) {
	var ids []uint
	result := r.db.
		Table("usuarios").
		Where("perfil IN ? AND ativo = true", []string{"admin", "superadmin"}).
		Pluck("id", &ids)
	return ids, result.Error
}