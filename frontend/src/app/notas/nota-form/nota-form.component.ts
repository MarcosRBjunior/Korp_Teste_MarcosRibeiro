import { ChangeDetectionStrategy, Component, computed, inject, OnInit, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { finalize } from 'rxjs';

import { mensagemDeErro } from '@/core/api-error';
import { ZardButtonComponent } from '@/shared/components/button';
import { ZardDialogRef } from '@/shared/components/dialog';
import { ZardInputComponent } from '@/shared/components/input';
import { ZardSonnerService } from '@/shared/components/sonner';

import type { Produto } from '../../produtos/produto.model';
import { ProdutoService } from '../../produtos/produto.service';
import type { NotaFiscal } from '../nota.model';
import { NotaFiscalService } from '../nota.service';

interface LinhaItem {
  chave: number;
  produtoId: string;
  quantidade: number;
  busca: string;
}

let proximoIdLinha = 0;

function rotuloProduto(produto: Produto): string {
  return `${produto.codigo} — ${produto.descricao} (saldo: ${produto.saldo})`;
}

@Component({
  selector: 'app-nota-form',
  imports: [FormsModule, ZardButtonComponent, ZardInputComponent],
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `
    <form class="flex flex-col gap-4" (ngSubmit)="salvar()">
      <div class="flex flex-col gap-3">
        @for (linha of linhas(); track linha.chave) {
          <div class="flex items-end gap-2">
            <label class="relative flex min-w-0 flex-1 flex-col gap-1.5 text-sm font-medium">
              Produto
              <input
                z-input
                type="text"
                autocomplete="off"
                placeholder="Digite para buscar um produto..."
                [ngModel]="textoExibicao(linha)"
                (ngModelChange)="onBuscaChange(linha.chave, $event)"
                (focus)="abrirSugestoes(linha.chave)"
                (blur)="fecharSugestoes(linha.chave)"
                (keydown.enter)="$event.preventDefault(); selecionarPrimeiraSugestao(linha.chave)"
                [ngModelOptions]="{ standalone: true }"
              />

              @if (linhaAberta() === linha.chave) {
                <div
                  class="absolute top-full left-0 z-50 mt-1 max-h-56 w-full overflow-auto rounded-lg border bg-popover py-1 text-sm shadow-md"
                >
                  @for (produto of produtosFiltrados(linha); track produto.id) {
                    <button
                      type="button"
                      class="block w-full truncate px-3 py-1.5 text-left hover:bg-muted"
                      (mousedown)="$event.preventDefault(); selecionarProduto(linha.chave, produto)"
                    >
                      {{ rotulo(produto) }}
                    </button>
                  } @empty {
                    <p class="px-3 py-1.5 text-muted-foreground">Nenhum produto encontrado.</p>
                  }
                </div>
              }
            </label>

            <label class="flex w-28 shrink-0 flex-col gap-1.5 text-sm font-medium">
              Qtd.
              <input
                z-input
                type="number"
                min="1"
                [max]="saldoDoProduto(linha.produtoId) ?? null"
                [attr.aria-invalid]="excedeSaldo(linha) ? 'true' : null"
                [ngModel]="linha.quantidade"
                (ngModelChange)="atualizarQuantidade(linha.chave, $event)"
                [ngModelOptions]="{ standalone: true }"
              />
            </label>

            <button
              type="button"
              z-button
              zType="ghost"
              zSize="icon"
              class="shrink-0"
              [zDisabled]="linhas().length === 1"
              (click)="removerLinha(linha.chave)"
              aria-label="Remover item"
            >
              &times;
            </button>
          </div>
          @if (excedeSaldo(linha)) {
            <p class="-mt-2 text-xs text-destructive">Quantidade maior que o saldo disponível.</p>
          }
        }
      </div>

      <button type="button" z-button zType="outline" zSize="sm" class="self-start" (click)="adicionarLinha()">
        + Adicionar produto
      </button>

      <div class="flex justify-end gap-2 pt-2">
        <button type="button" z-button zType="outline" [zDisabled]="salvando()" (click)="dialogRef.close()">Cancelar</button>
        <button type="submit" z-button [zLoading]="salvando()" [zDisabled]="!formValido()">Criar nota</button>
      </div>
    </form>
  `,
})
export class NotaFormComponent implements OnInit {
  readonly dialogRef = inject(ZardDialogRef<NotaFormComponent, NotaFiscal>);
  private readonly produtoService = inject(ProdutoService);
  private readonly notaService = inject(NotaFiscalService);
  private readonly sonner = inject(ZardSonnerService);

  readonly produtos = signal<Produto[]>([]);
  readonly linhas = signal<LinhaItem[]>([{ chave: proximoIdLinha++, produtoId: '', quantidade: 1, busca: '' }]);
  readonly salvando = signal(false);
  readonly linhaAberta = signal<number | null>(null);

  private readonly quantidadeSolicitadaPorProduto = computed(() => {
    const totais = new Map<string, number>();
    for (const linha of this.linhas()) {
      if (linha.produtoId === '') {
        continue;
      }
      totais.set(linha.produtoId, (totais.get(linha.produtoId) ?? 0) + linha.quantidade);
    }
    return totais;
  });

  readonly formValido = computed(() => {
    if (this.linhas().length === 0 || !this.linhas().every(l => l.produtoId !== '' && l.quantidade > 0)) {
      return false;
    }

    const totais = this.quantidadeSolicitadaPorProduto();
    return this.produtos().every(produto => (totais.get(produto.id.toString()) ?? 0) <= produto.saldo);
  });

  ngOnInit(): void {
    this.produtoService.listar().subscribe({
      next: produtos => this.produtos.set(produtos),
      error: err => this.sonner.error(mensagemDeErro(err, 'Erro ao carregar produtos.')),
    });
  }

  rotulo(produto: Produto): string {
    return rotuloProduto(produto);
  }

  textoExibicao(linha: LinhaItem): string {
    if (this.linhaAberta() === linha.chave) {
      return linha.busca;
    }

    const produto = this.produtos().find(p => p.id.toString() === linha.produtoId);
    return produto ? rotuloProduto(produto) : '';
  }

  produtosFiltrados(linha: LinhaItem): Produto[] {
    const termo = linha.busca.trim().toLowerCase();
    if (termo === '') {
      return this.produtos();
    }

    return this.produtos().filter(
      p => p.codigo.toLowerCase().includes(termo) || p.descricao.toLowerCase().includes(termo),
    );
  }

  abrirSugestoes(chave: number): void {
    this.linhaAberta.set(chave);
    this.linhas.update(atual => atual.map(l => (l.chave === chave ? { ...l, busca: '' } : l)));
  }

  fecharSugestoes(chave: number): void {
    if (this.linhaAberta() === chave) {
      this.linhaAberta.set(null);
    }
  }

  onBuscaChange(chave: number, valor: string): void {
    this.linhas.update(atual => atual.map(l => (l.chave === chave ? { ...l, busca: valor, produtoId: '' } : l)));
  }

  selecionarProduto(chave: number, produto: Produto): void {
    this.linhas.update(atual =>
      atual.map(l => (l.chave === chave ? { ...l, produtoId: produto.id.toString(), busca: '' } : l)),
    );
    this.linhaAberta.set(null);
  }

  selecionarPrimeiraSugestao(chave: number): void {
    const linha = this.linhas().find(l => l.chave === chave);
    if (!linha) {
      return;
    }

    const [primeiro] = this.produtosFiltrados(linha);
    if (primeiro) {
      this.selecionarProduto(chave, primeiro);
    }
  }

  adicionarLinha(): void {
    this.linhas.update(atual => [...atual, { chave: proximoIdLinha++, produtoId: '', quantidade: 1, busca: '' }]);
  }

  removerLinha(chave: number): void {
    this.linhas.update(atual => atual.filter(l => l.chave !== chave));
  }

  atualizarQuantidade(chave: number, quantidade: number): void {
    this.linhas.update(atual => atual.map(l => (l.chave === chave ? { ...l, quantidade } : l)));
  }

  saldoDoProduto(produtoId: string): number | undefined {
    return this.produtos().find(p => p.id.toString() === produtoId)?.saldo;
  }

  excedeSaldo(linha: LinhaItem): boolean {
    if (linha.produtoId === '') {
      return false;
    }

    const saldo = this.saldoDoProduto(linha.produtoId);
    if (saldo === undefined) {
      return false;
    }

    return (this.quantidadeSolicitadaPorProduto().get(linha.produtoId) ?? 0) > saldo;
  }

  salvar(): void {
    if (!this.formValido() || this.salvando()) {
      return;
    }

    this.salvando.set(true);
    const itens = this.linhas().map(l => ({ produto_id: Number(l.produtoId), quantidade: l.quantidade }));

    this.notaService
      .criar(itens)
      .pipe(finalize(() => this.salvando.set(false)))
      .subscribe({
        next: nota => {
          this.sonner.success(`Nota fiscal #${nota.numero_sequencial} criada.`);
          this.dialogRef.close(nota);
        },
        error: err => this.sonner.error(mensagemDeErro(err, 'Erro ao criar nota fiscal.')),
      });
  }
}
