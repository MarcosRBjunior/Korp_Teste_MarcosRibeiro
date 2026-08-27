package models

import "time"

type StatusNota string

const (
	StatusAberta  StatusNota = "Aberta"
	StatusFechada StatusNota = "Fechada"
)

type NotaFiscal struct {
	ID               uint       `gorm:"primaryKey" json:"id"`
	NumeroSequencial int64      `gorm:"uniqueIndex;not null" json:"numero_sequencial"`
	Status           StatusNota `gorm:"not null;default:Aberta" json:"status"`
	CriadoEm         time.Time  `json:"criado_em"`
	Itens            []ItemNota `gorm:"foreignKey:NotaFiscalID" json:"itens"`
}

func (NotaFiscal) TableName() string {
	return "notas_fiscais"
}
