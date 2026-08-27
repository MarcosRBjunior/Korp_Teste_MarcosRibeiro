package models

import "time"

// IdempotencyKey guarda a resposta já dada para uma chave de idempotência
// usada em POST /notas/{id}/imprimir, para que um retry com a mesma chave
// devolva o resultado salvo em vez de repetir o débito no Estoque.
type IdempotencyKey struct {
	Chave      string    `gorm:"primaryKey" json:"chave"`
	StatusHTTP int       `gorm:"not null" json:"status_http"`
	Resultado  string    `gorm:"type:text;not null" json:"resultado"`
	CriadoEm   time.Time `json:"criado_em"`
}

func (IdempotencyKey) TableName() string {
	return "idempotency_keys"
}
