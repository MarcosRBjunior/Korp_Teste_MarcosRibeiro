package models

// ProdutoID é uma referência lógica ao Produto do serviço de Estoque
// (não é FK de banco real, pois estoque_db e faturamento_db são bancos distintos).
type ItemNota struct {
	ID           uint `gorm:"primaryKey" json:"id"`
	NotaFiscalID uint `gorm:"not null;index" json:"nota_fiscal_id"`
	ProdutoID    uint `gorm:"not null" json:"produto_id"`
	Quantidade   int  `gorm:"not null;check:quantidade > 0" json:"quantidade"`
}

func (ItemNota) TableName() string {
	return "itens_nota"
}
