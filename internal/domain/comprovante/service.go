// Package comprovante gera o PDF de comprovante de venda para assinatura do cliente.
// Layout baseado no modelo oficial "Comprovante de Pedidos" da Baterias Super BR LTDA.
// Biblioteca: github.com/jung-kurt/gofpdf
// Encoding: UnicodeTranslatorFromDescriptor("cp1252") resolve acentuação em português.
package comprovante

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jung-kurt/gofpdf"
	"super-br/internal/domain/venda"
)

const (
	empresaNome     = "Baterias Super BR LTDA."
	empresaEndereco = "R. Cel Matos Dourado 388"
	empresaCNPJ     = "CNPJ: 07.093.375/0001-92"
	empresaTelefone = "(85) 32356040"
)

// Service encapsula a lógica de geração do comprovante em PDF.
type Service struct {
	pastaComprovantes string
}

// NewService cria o service com o diretório de destino dos PDFs.
func NewService(pastaComprovantes string) (*Service, error) {
	if err := os.MkdirAll(pastaComprovantes, 0755); err != nil {
		return nil, fmt.Errorf("erro ao criar diretório de comprovantes: %w", err)
	}
	return &Service{pastaComprovantes: pastaComprovantes}, nil
}

// GerarComprovante gera o PDF do comprovante de uma venda e retorna o caminho do arquivo.
func (s *Service) GerarComprovante(v *venda.Venda) (string, error) {
	nomeArquivo := fmt.Sprintf("comprovante_%d_%s.pdf", v.ID, time.Now().Format("20060102_150405"))
	caminhoCompleto := filepath.Join(s.pastaComprovantes, nomeArquivo)

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(12, 10, 12)

	// tr converte UTF-8 para cp1252 (Latin-1 estendido) — resolve ã, ç, é, etc.
	tr := pdf.UnicodeTranslatorFromDescriptor("cp1252")

	pdf.AddPage()
	const larg = 186.0

	// =========================================================
	// CABEÇALHO
	// =========================================================
	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(larg, 8, tr("Comprovante de Pedidos"), "", 1, "C", false, 0, "")

	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(larg, 5, tr(empresaNome), "", 1, "C", false, 0, "")

	pdf.Ln(3)

	pdf.SetFont("Arial", "", 9)
	pdf.CellFormat(larg, 5, tr(empresaEndereco), "", 1, "R", false, 0, "")

	pdf.SetFont("Arial", "B", 9)
	pdf.CellFormat(larg, 5, tr(empresaCNPJ), "", 1, "R", false, 0, "")

	pdf.SetFont("Arial", "", 9)
	pdf.CellFormat(larg, 5, tr(empresaTelefone), "", 1, "R", false, 0, "")

	pdf.Ln(4)

	// =========================================================
	// INFORMAÇÕES DO PEDIDO
	// =========================================================
	linhaInfo := func(label, valor string) {
		pdf.SetFont("Arial", "B", 10)
		pdf.CellFormat(22, 6, tr(label), "", 0, "L", false, 0, "")
		pdf.SetFont("Arial", "", 10)
		pdf.CellFormat(0, 6, tr(valor), "", 1, "L", false, 0, "")
	}

	linhaInfo("Cliente:", v.NomeCliente)
	linhaInfo("Empresa:", "")
	linhaInfo("Pedido:", fmt.Sprintf("#%d", v.ID))
	linhaInfo("Contato:", v.TelefoneCliente)
	linhaInfo("Data:", v.Data.Format("02/01/2006"))

	pdf.Ln(4)

	// =========================================================
	// TABELA DE ITENS
	// =========================================================
	const (
		colQtd    = 25.0
		colDesc   = 65.0
		colSucata = 28.0
		colVUnit  = 34.0
		colTotal  = 34.0
		altLinha  = 7.0
		minLinhas = 5
	)

	pdf.SetFillColor(0, 0, 0)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 9)

	pdf.CellFormat(colQtd, altLinha, tr("Quantidade"), "1", 0, "C", true, 0, "")
	pdf.CellFormat(colDesc, altLinha, tr("Discriminação"), "1", 0, "C", true, 0, "")
	pdf.CellFormat(colSucata, altLinha, tr("Sucata"), "1", 0, "C", true, 0, "")
	pdf.CellFormat(colVUnit, altLinha, tr("Valor Unitário"), "1", 0, "C", true, 0, "")
	pdf.CellFormat(colTotal, altLinha, tr("Total"), "1", 1, "C", true, 0, "")

	pdf.SetFillColor(255, 255, 255)
	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Arial", "", 9)

	linhasUsadas := 0
	for _, item := range v.Itens {
		discriminacao := fmt.Sprintf("%s %s (ID: %d / Lote: %s)",
			item.ItemEstoque.Produto.Nome,
			item.ItemEstoque.Produto.Categoria,
			item.ItemEstoque.ID,
			item.ItemEstoque.CodLote,
		)
		total := item.ValorUnitario * float64(item.Quantidade)

		pdf.CellFormat(colQtd, altLinha, fmt.Sprintf("%d", item.Quantidade), "1", 0, "C", false, 0, "")
		pdf.CellFormat(colDesc, altLinha, tr(discriminacao), "1", 0, "L", false, 0, "")
		pdf.CellFormat(colSucata, altLinha, "", "1", 0, "C", false, 0, "")
		pdf.CellFormat(colVUnit, altLinha, fmt.Sprintf("R$ %.2f", item.ValorUnitario), "1", 0, "R", false, 0, "")
		pdf.CellFormat(colTotal, altLinha, fmt.Sprintf("R$ %.2f", total), "1", 1, "R", false, 0, "")
		linhasUsadas++
	}

	for linhasUsadas < minLinhas {
		pdf.CellFormat(colQtd, altLinha, "", "1", 0, "C", false, 0, "")
		pdf.CellFormat(colDesc, altLinha, "", "1", 0, "L", false, 0, "")
		pdf.CellFormat(colSucata, altLinha, "", "1", 0, "C", false, 0, "")
		pdf.CellFormat(colVUnit, altLinha, "R$", "1", 0, "R", false, 0, "")
		pdf.CellFormat(colTotal, altLinha, "R$", "1", 1, "R", false, 0, "")
		linhasUsadas++
	}

	// =========================================================
	// ASSINATURA + PAGAMENTOS
	// =========================================================
	const (
		colAssVend  = 60.0
		colAssCli   = 58.0
		colPagLabel = 34.0
		colPagValor = 34.0
		altHeader   = 7.0
		altAss      = 18.0
	)

	pdf.SetFillColor(0, 0, 0)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 9)

	pdf.CellFormat(colAssVend, altHeader, tr("Ass. Vendedor"), "1", 0, "C", true, 0, "")
	pdf.CellFormat(colAssCli, altHeader, tr("Ass. Cliente"), "1", 0, "C", true, 0, "")
	pdf.CellFormat(colPagLabel+colPagValor, altHeader, tr("Valor Total"), "1", 1, "C", true, 0, "")

	xBase := pdf.GetX()
	yBase := pdf.GetY()

	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Arial", "", 9)

	// Caixas de assinatura — sem conteúdo interno
	pdf.SetXY(xBase, yBase)
	pdf.CellFormat(colAssVend, altAss, "", "1", 0, "C", false, 0, "")
	pdf.SetXY(xBase+colAssVend, yBase)
	pdf.CellFormat(colAssCli, altAss, "", "1", 0, "C", false, 0, "")

	// Valores por tipo de pagamento
	valores := map[string]float64{"dinheiro": 0, "pix": 0, "credito": 0, "debito": 0}
	for _, pg := range v.Pagamentos {
		if _, ok := valores[pg.Tipo]; ok {
			valores[pg.Tipo] += pg.Valor
		}
	}

	xPag := xBase + colAssVend + colAssCli
	yPag := yBase
	altPagLinha := altAss / 4.0

	linhasPag := []struct{ label, chave string }{
		{"Dinheiro", "dinheiro"},
		{"Pix", "pix"},
		{"Crédito", "credito"},
		{"Débito", "debito"},
	}

	for _, lp := range linhasPag {
		pdf.SetXY(xPag, yPag)
		pdf.CellFormat(colPagLabel, altPagLinha, tr(lp.label), "1", 0, "C", false, 0, "")
		valStr := "R$"
		if valores[lp.chave] > 0 {
			valStr = fmt.Sprintf("R$ %.2f", valores[lp.chave])
		}
		pdf.CellFormat(colPagValor, altPagLinha, valStr, "1", 0, "R", false, 0, "")
		yPag += altPagLinha
	}

	pdf.SetXY(xBase, yBase+altAss)

	// =========================================================
	// OBSERVAÇÕES + TOTAL
	// Correção: a área de observações é um bloco único sem divisória interna.
	// Desenhamos a borda manualmente com Rect para evitar o traço do CellFormat.
	// =========================================================
	const altObs = 18.0

	// Cabeçalho "Observações" | "Total"
	pdf.SetFillColor(0, 0, 0)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 9)

	pdf.CellFormat(colAssVend+colAssCli, altHeader, tr("Observações"), "1", 0, "C", true, 0, "")
	pdf.CellFormat(colPagLabel, altHeader, tr("Total"), "1", 0, "C", true, 0, "")

	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Arial", "B", 9)
	pdf.CellFormat(colPagValor, altHeader, fmt.Sprintf("R$ %.2f", v.ValorTotal), "1", 1, "C", false, 0, "")

	// Posição atual após os cabeçalhos
	xObs := pdf.GetX()
	yObs := pdf.GetY()

	largObs := colAssVend + colAssCli
	largDir := colPagLabel + colPagValor

	// Bloco de observações — borda externa apenas, sem linhas internas
	// Usa Rect diretamente para ter controle total do traço
	pdf.SetDrawColor(0, 0, 0)
	pdf.Rect(xObs, yObs, largObs, altObs, "D")

	// Texto das observações dentro do bloco (com margem interna de 2mm)
	if v.Observacoes != "" {
		pdf.SetFont("Arial", "", 8)
		pdf.SetXY(xObs+2, yObs+2)
		pdf.MultiCell(largObs-4, 4, tr(v.Observacoes), "", "L", false)
	}

	// Bloco direito (vazio, apenas borda)
	pdf.Rect(xObs+largObs, yObs, largDir, altObs, "D")

	// Move cursor para após os blocos
	pdf.SetXY(xObs, yObs+altObs)

	// =========================================================
	// RODAPÉ
	// =========================================================
	pdf.Ln(4)
	pdf.SetFont("Arial", "", 7)
	pdf.SetTextColor(150, 150, 150)
	pdf.CellFormat(larg, 5,
		tr(fmt.Sprintf("Documento gerado em %s", time.Now().Format("02/01/2006 às 15:04:05"))),
		"", 1, "C", false, 0, "")

	if err := pdf.OutputFileAndClose(caminhoCompleto); err != nil {
		return "", fmt.Errorf("erro ao salvar comprovante: %w", err)
	}

	return caminhoCompleto, nil
}