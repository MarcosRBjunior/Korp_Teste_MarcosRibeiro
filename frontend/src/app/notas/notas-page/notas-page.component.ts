import { ChangeDetectionStrategy, Component, computed, DestroyRef, inject, Injector, OnInit, signal } from '@angular/core';
import { takeUntilDestroyed, toObservable } from '@angular/core/rxjs-interop';
import { FormsModule } from '@angular/forms';
import { filter, finalize } from 'rxjs';

import { mensagemDeErro } from '@/core/api-error';
import { ZardBadgeComponent, type ZardBadgeTypeVariants } from '@/shared/components/badge';
import { ZardButtonComponent } from '@/shared/components/button';
import { ZardDialogService } from '@/shared/components/dialog';
import { ZardInputComponent } from '@/shared/components/input';
import { ZardSonnerService } from '@/shared/components/sonner';
import { ZardSpinnerComponent } from '@/shared/components/spinner';
import { ZardTableImports } from '@/shared/components/table';

import { NotaFormComponent } from '../nota-form/nota-form.component';
import type { NotaFiscal, StatusNota } from '../nota.model';
import { NotaFiscalService } from '../nota.service';

const BADGE_POR_STATUS: Record<StatusNota, ZardBadgeTypeVariants> = {
  Aberta: 'default',
  Processando: 'secondary',
  Fechada: 'outline',
};

@Component({
  selector: 'app-notas-page',
  imports: [FormsModule, ZardBadgeComponent, ZardButtonComponent, ZardInputComponent, ZardSpinnerComponent, ...ZardTableImports],
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `
    <div class="flex flex-col gap-4 p-6">
      <div class="flex items-center justify-between">
        <h1 class="text-xl font-semibold">Notas Fiscais</h1>
        <button z-button (click)="abrirCadastro()">Nova nota</button>
      </div>

      <input
        z-input
        type="text"
        placeholder="Buscar por número ou status..."
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
              <th z-table-head>Número</th>
              <th z-table-head>Status</th>
              <th z-table-head>Itens</th>
              <th z-table-head></th>
            </tr>
          </thead>
          <tbody z-table-body>
            @for (nota of notasFiltradas(); track nota.id) {
              <tr z-table-row>
                <td z-table-cell>#{{ nota.numero_sequencial }}</td>
                <td z-table-cell>
                  <z-badge [zType]="badgeDoStatus(nota.status)">{{ nota.status }}</z-badge>
                </td>
                <td z-table-cell>{{ nota.itens.length }} produto(s)</td>
                <td z-table-cell class="text-right">
                  <button
                    z-button
                    zSize="sm"
                    [zLoading]="imprimindo().has(nota.id)"
                    [zDisabled]="nota.status !== 'Aberta' || imprimindo().has(nota.id)"
                    (click)="imprimir(nota)"
                  >
                    Imprimir
                  </button>
                </td>
              </tr>
            } @empty {
              <tr z-table-row>
                <td z-table-cell colspan="4" class="text-center text-muted-foreground">
                  {{ busca() ? 'Nenhuma nota encontrada.' : 'Nenhuma nota fiscal cadastrada.' }}
                </td>
              </tr>
            }
          </tbody>
        </table>
      }
    </div>
  `,
})
export class NotasPageComponent implements OnInit {
  private readonly notaService = inject(NotaFiscalService);
  private readonly dialogService = inject(ZardDialogService);
  private readonly sonner = inject(ZardSonnerService);
  private readonly injector = inject(Injector);
  private readonly destroyRef = inject(DestroyRef);

  readonly notas = signal<NotaFiscal[]>([]);
  readonly carregando = signal(false);
  readonly imprimindo = signal<Set<number>>(new Set());
  readonly busca = signal('');

  readonly notasFiltradas = computed(() => {
    const termo = this.busca().trim().toLowerCase();
    if (termo === '') {
      return this.notas();
    }
    return this.notas().filter(
      n => n.numero_sequencial.toString().includes(termo) || n.status.toLowerCase().includes(termo),
    );
  });

  ngOnInit(): void {
    this.carregar();
  }

  badgeDoStatus(status: StatusNota): ZardBadgeTypeVariants {
    return BADGE_POR_STATUS[status];
  }

  carregar(): void {
    this.carregando.set(true);
    this.notaService
      .listar()
      .pipe(finalize(() => this.carregando.set(false)))
      .subscribe({
        next: notas => this.notas.set(notas),
        error: err => this.sonner.error(mensagemDeErro(err, 'Erro ao carregar notas fiscais.')),
      });
  }

  abrirCadastro(): void {
    const dialogRef = this.dialogService.create<NotaFormComponent, void>({
      zTitle: 'Nova nota fiscal',
      zContent: NotaFormComponent,
      zHideFooter: true,
      zWidth: '600px',
    });

    toObservable(dialogRef.result, { injector: this.injector })
      .pipe(
        filter((nota): nota is NotaFiscal => !!nota),
        takeUntilDestroyed(this.destroyRef),
      )
      .subscribe(() => this.carregar());
  }

  imprimir(nota: NotaFiscal): void {
    if (this.imprimindo().has(nota.id)) {
      return;
    }

    this.imprimindo.update(atual => new Set(atual).add(nota.id));

    this.notaService
      .imprimir(nota.id)
      .pipe(
        finalize(() =>
          this.imprimindo.update(atual => {
            const proximo = new Set(atual);
            proximo.delete(nota.id);
            return proximo;
          }),
        ),
      )
      .subscribe({
        next: notaImpressa => {
          this.notas.update(atual => atual.map(n => (n.id === notaImpressa.id ? notaImpressa : n)));
          this.sonner.success(`Nota #${notaImpressa.numero_sequencial} impressa, saldo debitado.`);
        },
        error: err => {
          const status = (err as { status?: number })?.status;

          if (status === 409) {
            this.sonner.error(mensagemDeErro(err, 'Esta nota não está mais aberta e não pode ser impressa de novo.'));
            this.carregar();
            return;
          }

          const mensagem =
            status === 503
              ? 'Estoque indisponível no momento, tente novamente em instantes.'
              : mensagemDeErro(err, 'Erro ao imprimir nota fiscal.');
          this.sonner.error(mensagem);
        },
      });
  }
}
