package models

import "time"

type Produto struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Codigo    string    `gorm:"uniqueIndex;not null" json:"codigo"`
	Descricao string    `gorm:"not null" json:"descricao"`
	Saldo     int       `gorm:"not null;check:saldo >= 0" json:"saldo"`
	CriadoEm  time.Time `json:"criado_em"`
}
