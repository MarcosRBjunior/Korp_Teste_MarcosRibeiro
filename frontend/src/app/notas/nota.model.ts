export type StatusNota = 'Aberta' | 'Processando' | 'Fechada';

export interface ItemNota {
  id: number;
  nota_fiscal_id: number;
  produto_id: number;
  quantidade: number;
}

export interface NotaFiscal {
  id: number;
  numero_sequencial: number;
  status: StatusNota;
  criado_em: string;
  itens: ItemNota[];
}

export interface ItemNotaInput {
  produto_id: number;
  quantidade: number;
}
