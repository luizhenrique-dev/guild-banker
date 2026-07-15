package importer

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

const (
	c6ExpectedColumns = 9
	dateLayoutBR      = "02/01/2006"
)

var c6Header = []string{
	"Data de Compra",
	"Nome no Cartão",
	"Final do Cartão",
	"Categoria",
	"Descrição",
	"Parcela",
	"Valor (em US$)",
	"Cotação (em R$)",
	"Valor (em R$)",
}

type ParsedRow struct {
	OccurredAt   time.Time
	CardLast4    string
	BankCategory string
	Description  string
	Installment  string
	Amount       decimal.Decimal
}

type Parser interface {
	Parse(r io.Reader) ([]ParsedRow, error)
}

type csvC6Parser struct{}

func NewC6CSVParser() Parser {
	return csvC6Parser{}
}

func (csvC6Parser) Parse(r io.Reader) ([]ParsedRow, error) {
	reader := csv.NewReader(r)
	reader.Comma = ';'
	reader.FieldsPerRecord = c6ExpectedColumns

	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("read csv header: %w", err)
	}

	for i, expected := range c6Header {
		if strings.TrimSpace(header[i]) != expected {
			return nil, fmt.Errorf("invalid csv header column %d: got %q expected %q", i, header[i], expected)
		}
	}

	rows := make([]ParsedRow, 0)
	line := 1
	for {
		record, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}

			return nil, fmt.Errorf("read csv line %d: %w", line+1, err)
		}

		line++
		if len(record) != c6ExpectedColumns {
			return nil, fmt.Errorf("invalid column count at line %d", line)
		}

		occurredAt, err := time.Parse(dateLayoutBR, strings.TrimSpace(record[0]))
		if err != nil {
			return nil, fmt.Errorf("parse date at line %d: %w", line, err)
		}

		amount, err := decimal.NewFromString(strings.TrimSpace(record[8]))
		if err != nil {
			return nil, fmt.Errorf("parse amount at line %d: %w", line, err)
		}

		if len(strings.TrimSpace(record[2])) != 4 {
			if _, convErr := strconv.Atoi(strings.TrimSpace(record[2])); convErr != nil {
				return nil, fmt.Errorf("parse card last4 at line %d: %w", line, convErr)
			}
		}

		rows = append(rows, ParsedRow{
			OccurredAt:   occurredAt,
			CardLast4:    strings.TrimSpace(record[2]),
			BankCategory: strings.TrimSpace(record[3]),
			Description:  strings.TrimSpace(record[4]),
			Installment:  strings.TrimSpace(record[5]),
			Amount:       amount,
		})
	}

	return rows, nil
}
