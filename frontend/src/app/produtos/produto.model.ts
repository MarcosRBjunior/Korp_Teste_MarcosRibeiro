export interface Produto {
  id: number;
  codigo: string;
  descricao: string;
  saldo: number;
  criado_em: string;
}

export interface NovoProduto {
  codigo: string;
  descricao: string;
  saldo: number;
}
