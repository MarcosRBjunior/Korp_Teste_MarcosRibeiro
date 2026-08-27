import { ChangeDetectionStrategy, Component, computed, inject, OnInit, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { finalize } from 'rxjs';

import { mensagemDeErro } from '@/core/api-error';
import { ZardButtonComponent } from '@/shared/components/button';
import { ZardDialogRef } from '@/shared/components/dialog';
import { ZardInputComponent } from '@/shared/components/input';
import { ZardSelectImports } from '@/shared/components/select';
import { ZardSonnerService } from '@/shared/components/sonner';

import type { Produto } from '../../produtos/produto.model';
import { ProdutoService } from '../../produtos/produto.service';
import type { NotaFiscal } from '../nota.model';
import { NotaFiscalService } from '../nota.service';

interface LinhaItem {
  produtoId: string;
  quantidade: number;
}

let proximoIdLinha = 0;

@Component({
  selector: 'app-nota-form',
  imports: [FormsModule, ZardButtonComponent, ZardInputComponent, ...ZardSelectImports],
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `
    <form class="flex flex-col gap-4" (ngSubmit)="salvar()">
      <div class="flex flex-col gap-3">
        @for (linha of linhas(); track linha.chave) {
          <div class="flex items-end gap-2">
            <label class="flex flex-1 flex-col gap-1.5 text-sm font-medium">
              Produto
              <div z-select [(ngModel)]="linha.produtoId" [ngModelOptions]="{ standalone: true }" zPlaceholder="Selecione um produto">
                @for (produto of produtos(); track produto.id) {
                  <div z-select-item [zValue]="produto.id.toString()">
                    {{ produto.codigo }} — {{ produto.descricao }} (saldo: {{ produto.saldo }})
                  </div>
                }
              </div>
            </label>

            <label class="flex w-28 flex-col gap-1.5 text-sm font-medium">
              Qtd.
              <input
                z-input
                type="number"
                min="1"
                [(ngModel)]="linha.quantidade"
                [ngModelOptions]="{ standalone: true }"
              />
            </label>

            <button
              type="button"
              z-button
              zType="ghost"
              zSize="icon"
              [zDisabled]="linhas().length === 1"
              (click)="removerLinha(linha.chave)"
              aria-label="Remover item"
            >
              &times;
            </button>
          </div>
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
  readonly linhas = signal<(LinhaItem & { chave: number })[]>([{ chave: proximoIdLinha++, produtoId: '', quantidade: 1 }]);
  readonly salvando = signal(false);

  readonly formValido = computed(
    () => this.linhas().length > 0 && this.linhas().every(l => l.produtoId !== '' && l.quantidade > 0),
  );

  ngOnInit(): void {
    this.produtoService.listar().subscribe({
      next: produtos => this.produtos.set(produtos),
      error: err => this.sonner.error(mensagemDeErro(err, 'Erro ao carregar produtos.')),
    });
  }

  adicionarLinha(): void {
    this.linhas.update(atual => [...atual, { chave: proximoIdLinha++, produtoId: '', quantidade: 1 }]);
  }

  removerLinha(chave: number): void {
    this.linhas.update(atual => atual.filter(l => l.chave !== chave));
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
