import { ChangeDetectionStrategy, Component, computed, DestroyRef, inject, Injector, OnInit, signal } from '@angular/core';
import { takeUntilDestroyed, toObservable } from '@angular/core/rxjs-interop';
import { FormsModule } from '@angular/forms';
import { filter, finalize } from 'rxjs';

import { mensagemDeErro } from '@/core/api-error';
import { ZardBadgeComponent } from '@/shared/components/badge';
import { ZardButtonComponent } from '@/shared/components/button';
import { ZardDialogService } from '@/shared/components/dialog';
import { ZardInputComponent } from '@/shared/components/input';
import { ZardSonnerService } from '@/shared/components/sonner';
import { ZardSpinnerComponent } from '@/shared/components/spinner';
import { ZardTableImports } from '@/shared/components/table';

import { ProdutoFormComponent } from '../produto-form/produto-form.component';
import type { Produto } from '../produto.model';
import { ProdutoService } from '../produto.service';

const SALDO_BAIXO = 5;

@Component({
  selector: 'app-produtos-page',
  imports: [FormsModule, ZardBadgeComponent, ZardButtonComponent, ZardInputComponent, ZardSpinnerComponent, ...ZardTableImports],
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `
    <div class="flex flex-col gap-4 p-6">
      <div class="flex items-center justify-between">
        <h1 class="text-xl font-semibold">Produtos</h1>
        <button z-button (click)="abrirCadastro()">Novo produto</button>
      </div>

      <input
        z-input
        type="text"
        placeholder="Buscar por código ou descrição..."
        class="max-w-sm"
        [(ngModel)]="busca"
        [ngModelOptions]="{ standalone: true }"
      />

      @if (carregando()) {
        <div class="flex justify-center py-10">
          <z-spinner />
        </div>
      } @else {
        <table z-table>
          <thead z-table-header>
            <tr z-table-row>
              <th z-table-head>Código</th>
              <th z-table-head>Descrição</th>
              <th z-table-head>Saldo</th>
            </tr>
          </thead>
          <tbody z-table-body>
            @for (produto of produtosFiltrados(); track produto.id) {
              <tr z-table-row>
                <td z-table-cell>{{ produto.codigo }}</td>
                <td z-table-cell>{{ produto.descricao }}</td>
                <td z-table-cell class="flex items-center gap-2">
                  {{ produto.saldo }}
                  @if (produto.saldo < SALDO_BAIXO) {
                    <z-badge zType="destructive">saldo baixo</z-badge>
                  }
                </td>
              </tr>
            } @empty {
              <tr z-table-row>
                <td z-table-cell colspan="3" class="text-center text-muted-foreground">
                  {{ busca() ? 'Nenhum produto encontrado.' : 'Nenhum produto cadastrado.' }}
                </td>
              </tr>
            }
          </tbody>
        </table>
      }
    </div>
  `,
})
export class ProdutosPageComponent implements OnInit {
  private readonly produtoService = inject(ProdutoService);
  private readonly dialogService = inject(ZardDialogService);
  private readonly sonner = inject(ZardSonnerService);
  private readonly injector = inject(Injector);
  private readonly destroyRef = inject(DestroyRef);

  readonly produtos = signal<Produto[]>([]);
  readonly carregando = signal(false);
  readonly busca = signal('');
  readonly SALDO_BAIXO = SALDO_BAIXO;

  readonly produtosFiltrados = computed(() => {
    const termo = this.busca().trim().toLowerCase();
    if (termo === '') {
      return this.produtos();
    }
    return this.produtos().filter(
      p => p.codigo.toLowerCase().includes(termo) || p.descricao.toLowerCase().includes(termo),
    );
  });

  ngOnInit(): void {
    this.carregar();
  }

  carregar(): void {
    this.carregando.set(true);
    this.produtoService
      .listar()
      .pipe(finalize(() => this.carregando.set(false)))
      .subscribe({
        next: produtos => this.produtos.set(produtos),
        error: err => this.sonner.error(mensagemDeErro(err, 'Erro ao carregar produtos.')),
      });
  }

  abrirCadastro(): void {
    const dialogRef = this.dialogService.create<ProdutoFormComponent, void>({
      zTitle: 'Novo produto',
      zContent: ProdutoFormComponent,
      zHideFooter: true,
      zWidth: '420px',
    });

    toObservable(dialogRef.result, { injector: this.injector })
      .pipe(
        filter((produto): produto is Produto => !!produto),
        takeUntilDestroyed(this.destroyRef),
      )
      .subscribe(() => this.carregar());
  }
}
