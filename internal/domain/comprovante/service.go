// Package comprovante gera o PDF de comprovante de venda para assinatura do cliente.
// O layout segue o modelo oficial "Comprovante de Pedidos" da Baterias Super BR LTDA.
// Biblioteca utilizada: github.com/jung-kurt/gofpdf — 100% Go, sem dependências C.
//
// ATENÇÃO: gofpdf com fontes built-in (Arial/Helvetica) usa encoding Latin-1 (ISO-8859-1).
// Strings com caracteres especiais do português devem ser convertidas via iso() antes de
// serem passadas para qualquer método de célula/texto do pdf.
package comprovante

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
	"unicode/utf8"

	"github.com/jung-kurt/gofpdf"
	"super-br/internal/domain/venda"
)

// Dados fixos da empresa — alterar aqui caso mudem no futuro.
const (
	empresaNome     = "Baterias Super BR LTDA."
	empresaEndereco = "R. Cel Matos Dourado 388"
	empresaCNPJ     = "CNPJ: 07.093.375/0001-92"
	empresaTelefone = "(85) 32356040"
)

// iso converte uma string UTF-8 para Latin-1 (ISO-8859-1).
// O gofpdf usa Latin-1 internamente nas fontes padrão (Arial, Helvetica, Times).
// Caracteres sem representação em Latin-1 são substituídos por '?'.
func iso(s string) string {
	result := make([]byte, 0, len(s))
	for _, r := range s {
		if r < utf8.RuneSelf {
			// ASCII puro — copia direto
			result = append(result, byte(r))
			continue
		}
		// Tenta mapear para Latin-1 (U+00A0 a U+00FF)
		if r >= 0x00A0 && r <= 0x00FF {
			result = append(result, byte(r))
		} else {
			result = append(result, '?')
		}
	}
	return string(result)
}

// Service encapsula a lógica de geração do comprovante em PDF.
type Service struct {
	pastaComprovantes string
}

// NewService cria o service com o diretório de destino dos PDFs.
// O diretório é criado automaticamente se não existir.
func NewService(pastaComprovantes string) (*Service, error) {
	if err := os.MkdirAll(pastaComprovantes, 0755); err != nil {
		return nil, fmt.Errorf("erro ao criar diretório de comprovantes: %w", err)
	}
	return &Service{pastaComprovantes: pastaComprovantes}, nil
}

// GerarComprovante gera o PDF do comprovante de uma venda e retorna o caminho do arquivo.
// O arquivo é nomeado como "comprovante_<id_venda>_<timestamp>.pdf".
func (s *Service) GerarComprovante(v *venda.Venda) (string, error) {
	nomeArquivo := fmt.Sprintf(
		"comprovante_%d_%s.pdf",
		v.ID,
		time.Now().Format("20060102_150405"),
	)
	caminhoCompleto := filepath.Join(s.pastaComprovantes, nomeArquivo)

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(12, 10, 12)

	// Informa ao gofpdf que vamos entregar strings em Latin-1
	pdf.SetFont("Arial", "", 10)
	pdf.AddPage()

	// Largura útil: 210 - 12 - 12 = 186mm
	const larg = 186.0

	// =========================================================
	// CABEÇALHO
	// =========================================================
	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(larg, 8, iso("Comprovante de Pedidos"), "", 1, "C", false, 0, "")

	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(larg, 5, iso(empresaNome), "", 1, "C", false, 0, "")

	pdf.Ln(3)

	pdf.SetFont("Arial", "", 9)
	pdf.CellFormat(larg, 5, iso(empresaEndereco), "", 1, "R", false, 0, "")

	pdf.SetFont("Arial", "B", 9)
	pdf.CellFormat(larg, 5, iso(empresaCNPJ), "", 1, "R", false, 0, "")

	pdf.SetFont("Arial", "", 9)
	pdf.CellFormat(larg, 5, iso(empresaTelefone), "", 1, "R", false, 0, "")

	pdf.Ln(4)

	// =========================================================
	// INFORMAÇÕES DO PEDIDO
	// =========================================================
	linhaInfo := func(label, valor string) {
		pdf.SetFont("Arial", "B", 10)
		pdf.CellFormat(22, 6, iso(label), "", 0, "L", false, 0, "")
		pdf.SetFont("Arial", "", 10)
		pdf.CellFormat(0, 6, iso(valor), "", 1, "L", false, 0, "")
	}

	linhaInfo("Cliente:", v.NomeCliente)
	linhaInfo("Empresa:", "")
	linhaInfo("Pedido:", fmt.Sprintf("#%d", v.ID))
	linhaInfo("Contato:", v.TelefoneCliente)
	linhaInfo("Data:", v.Data.Format("02/01/2006"))

	pdf.Ln(4)

	// =========================================================
	// TABELA DE ITENS
	// Colunas: Quantidade | Discriminação | Sucata | Valor Unitário | Total
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

	// Cabeçalho — fundo preto, texto branco
	pdf.SetFillColor(0, 0, 0)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 9)

	pdf.CellFormat(colQtd, altLinha, iso("Quantidade"), "1", 0, "C", true, 0, "")
	pdf.CellFormat(colDesc, altLinha, iso("Discriminação"), "1", 0, "C", true, 0, "")
	pdf.CellFormat(colSucata, altLinha, iso("Sucata"), "1", 0, "C", true, 0, "")
	pdf.CellFormat(colVUnit, altLinha, iso("Valor Unitário"), "1", 0, "C", true, 0, "")
	pdf.CellFormat(colTotal, altLinha, iso("Total"), "1", 1, "C", true, 0, "")

	// Linhas dos itens
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
		pdf.CellFormat(colDesc, altLinha, iso(discriminacao), "1", 0, "L", false, 0, "")
		pdf.CellFormat(colSucata, altLinha, "", "1", 0, "C", false, 0, "")
		pdf.CellFormat(colVUnit, altLinha, fmt.Sprintf("R$ %.2f", item.ValorUnitario), "1", 0, "R", false, 0, "")
		pdf.CellFormat(colTotal, altLinha, fmt.Sprintf("R$ %.2f", total), "1", 1, "R", false, 0, "")
		linhasUsadas++
	}

	// Linhas vazias para completar o mínimo do modelo
	for linhasUsadas < minLinhas {
		pdf.CellFormat(colQtd, altLinha, "", "1", 0, "C", false, 0, "")
		pdf.CellFormat(colDesc, altLinha, "", "1", 0, "L", false, 0, "")
		pdf.CellFormat(colSucata, altLinha, "", "1", 0, "C", false, 0, "")
		pdf.CellFormat(colVUnit, altLinha, "R$", "1", 0, "R", false, 0, "")
		pdf.CellFormat(colTotal, altLinha, "R$", "1", 1, "R", false, 0, "")
		linhasUsadas++
	}

	// =========================================================
	// SEÇÃO DE ASSINATURA + PAGAMENTOS
	// =========================================================
	const (
		colAssVend  = 60.0
		colAssCli   = 58.0
		colPagLabel = 34.0
		colPagValor = 34.0
		altHeader   = 7.0
		altAss      = 18.0
	)

	// Cabeçalhos: Ass. Vendedor | Ass. Cliente | Valor Total
	pdf.SetFillColor(0, 0, 0)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 9)

	pdf.CellFormat(colAssVend, altHeader, iso("Ass. Vendedor"), "1", 0, "C", true, 0, "")
	pdf.CellFormat(colAssCli, altHeader, iso("Ass. Cliente"), "1", 0, "C", true, 0, "")
	pdf.CellFormat(colPagLabel+colPagValor, altHeader, iso("Valor Total"), "1", 1, "C", true, 0, "")

	// Área de assinatura + pagamentos lado a lado
	xBase := pdf.GetX()
	yBase := pdf.GetY()

	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Arial", "", 9)

	// Caixas de assinatura
	pdf.SetXY(xBase, yBase)
	pdf.CellFormat(colAssVend, altAss, "", "1", 0, "C", false, 0, "")
	pdf.SetXY(xBase+colAssVend, yBase)
	pdf.CellFormat(colAssCli, altAss, "", "1", 0, "C", false, 0, "")

	// Valores por tipo de pagamento
	valores := map[string]float64{
		"dinheiro": 0,
		"pix":      0,
		"credito":  0,
		"debito":   0,
	}
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
		{iso("Crédito"), "credito"},
		{iso("Débito"), "debito"},
	}

	for _, lp := range linhasPag {
		pdf.SetXY(xPag, yPag)
		pdf.CellFormat(colPagLabel, altPagLinha, iso(lp.label), "1", 0, "C", false, 0, "")
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
	// =========================================================
	const altObs = 18.0

	// Cabeçalho "Observações" | "Total"
	pdf.SetFillColor(0, 0, 0)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 9)

	pdf.CellFormat(colAssVend+colAssCli, altHeader, iso("Observações"), "1", 0, "C", true, 0, "")
	pdf.CellFormat(colPagLabel, altHeader, iso("Total"), "1", 0, "C", true, 0, "")

	pdf.SetTextColor(0, 0, 0)
	pdf.SetFont("Arial", "B", 9)
	pdf.CellFormat(colPagValor, altHeader, fmt.Sprintf("R$ %.2f", v.ValorTotal), "1", 1, "C", false, 0, "")

	// Área de observações
	xObs := pdf.GetX()
	yObs := pdf.GetY()

	pdf.SetFont("Arial", "", 9)
	pdf.CellFormat(colAssVend+colAssCli, altObs, "", "1", 0, "L", false, 0, "")

	pdf.SetXY(xObs+colAssVend+colAssCli, yObs)
	pdf.CellFormat(colPagLabel+colPagValor, altObs, "", "1", 1, "L", false, 0, "")

	// =========================================================
	// RODAPÉ
	// =========================================================
	pdf.Ln(4)
	pdf.SetFont("Arial", "", 7)
	pdf.SetTextColor(150, 150, 150)
	pdf.CellFormat(larg, 5,
		iso(fmt.Sprintf("Documento gerado em %s", time.Now().Format("02/01/2006 às 15:04:05"))),
		"", 1, "C", false, 0, "")

	if err := pdf.OutputFileAndClose(caminhoCompleto); err != nil {
		return "", fmt.Errorf("erro ao salvar comprovante: %w", err)
	}

	return caminhoCompleto, nil
}