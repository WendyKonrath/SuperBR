package movimentacao

import (
	"time"

	"gorm.io/gorm"
)

// Repository encapsula o acesso ao banco de dados para Movimentacao.
type Repository struct {
	db *gorm.DB
}

// NewRepository cria um novo Repository com a conexão injetada.
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Registrar persiste uma nova movimentação dentro de uma transação existente.
// Deve sempre ser chamado com o tx da transação pai para garantir atomicidade.
func (r *Repository) Registrar(tx *gorm.DB, itemID, usuarioID uint, tipo string) error {
	mov := Movimentacao{
		ItemID:    itemID,
		UsuarioID: usuarioID,
		Tipo:      tipo,
		Data:      time.Now(),
	}
	return tx.Create(&mov).Error
}

// itemInfoRow é usado internamente para receber os dados do JOIN.
type itemInfoRow struct {
	MovID       uint      `gorm:"column:mov_id"`
	ItemID      uint      `gorm:"column:item_id"`
	CodLote     string    `gorm:"column:cod_lote"`
	Estado      string    `gorm:"column:estado"`
	ProdutoID   uint      `gorm:"column:produto_id"`
	NomeProduto string    `gorm:"column:nome_produto"`
	Categoria   string    `gorm:"column:categoria"`
	Tipo        string    `gorm:"column:tipo"`
	Data        time.Time `gorm:"column:data"`
	UsuarioID   uint      `gorm:"column:usuario_id"`
	CreatedAt   time.Time `gorm:"column:created_at"`
}

// popularItemInfo busca as movimentações com informações do item e produto via JOIN
// e preenche o campo ItemInfo de cada movimentação.
func (r *Repository) popularItemInfo(movs []Movimentacao) ([]Movimentacao, error) {
	if len(movs) == 0 {
		return movs, nil
	}

	// Coleta os IDs das movimentações para buscar os dados de item em lote.
	ids := make([]uint, len(movs))
	for i, m := range movs {
		ids[i] = m.ID
	}

	var rows []itemInfoRow
	err := r.db.Raw(`
		SELECT
			m.id          AS mov_id,
			ie.id         AS item_id,
			ie.cod_lote   AS cod_lote,
			ie.estado     AS estado,
			p.id          AS produto_id,
			p.nome        AS nome_produto,
			p.categoria   AS categoria
		FROM movimentacaos m
		JOIN item_estoques ie ON ie.id = m.item_id
		JOIN produtos p ON p.id = ie.produto_id
		WHERE m.id IN ?
	`, ids).Scan(&rows).Error
	if err != nil {
		return movs, err
	}

	// Indexa os resultados pelo ID da movimentação.
	infoMap := make(map[uint]*ItemInfo, len(rows))
	for _, row := range rows {
		r := row
		infoMap[r.MovID] = &ItemInfo{
			ItemID:      r.ItemID,
			CodLote:     r.CodLote,
			Estado:      r.Estado,
			ProdutoID:   r.ProdutoID,
			NomeProduto: r.NomeProduto,
			Categoria:   r.Categoria,
		}
	}

	// Preenche o ItemInfo em cada movimentação.
	for i := range movs {
		movs[i].ItemInfo = infoMap[movs[i].ID]
	}

	return movs, nil
}

// ListarTodas retorna todas as movimentações com usuário e dados do item carregados.
func (r *Repository) ListarTodas() ([]Movimentacao, error) {
	var movs []Movimentacao
	if err := r.db.Preload("Usuario").Order("data DESC").Find(&movs).Error; err != nil {
		return nil, err
	}
	return r.popularItemInfo(movs)
}

// ListarPorItem retorna todas as movimentações de um item específico.
func (r *Repository) ListarPorItem(itemID uint) ([]Movimentacao, error) {
	var movs []Movimentacao
	if err := r.db.Preload("Usuario").
		Where("item_id = ?", itemID).
		Order("data DESC").
		Find(&movs).Error; err != nil {
		return nil, err
	}
	return r.popularItemInfo(movs)
}

// ListarPorProduto retorna todas as movimentações de itens de um produto específico.
func (r *Repository) ListarPorProduto(produtoID uint) ([]Movimentacao, error) {
	var movs []Movimentacao
	if err := r.db.Preload("Usuario").
		Joins("JOIN item_estoques ON item_estoques.id = movimentacaos.item_id").
		Where("item_estoques.produto_id = ?", produtoID).
		Order("movimentacaos.data DESC").
		Find(&movs).Error; err != nil {
		return nil, err
	}
	return r.popularItemInfo(movs)
}

// ListarPorTipo retorna movimentações filtradas por tipo ("entrada" ou "saida").
func (r *Repository) ListarPorTipo(tipo string) ([]Movimentacao, error) {
	var movs []Movimentacao
	if err := r.db.Preload("Usuario").
		Where("tipo = ?", tipo).
		Order("data DESC").
		Find(&movs).Error; err != nil {
		return nil, err
	}
	return r.popularItemInfo(movs)
}

// ListarPorPeriodo retorna movimentações dentro de um intervalo de datas.
func (r *Repository) ListarPorPeriodo(inicio, fim time.Time) ([]Movimentacao, error) {
	var movs []Movimentacao
	if err := r.db.Preload("Usuario").
		Where("data BETWEEN ? AND ?", inicio, fim).
		Order("data DESC").
		Find(&movs).Error; err != nil {
		return nil, err
	}
	return r.popularItemInfo(movs)
}