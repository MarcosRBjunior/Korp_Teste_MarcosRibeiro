import { ChangeDetectionStrategy, Component, inject, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { finalize } from 'rxjs';

import { mensagemDeErro } from '@/core/api-error';
import { ZardButtonComponent } from '@/shared/components/button';
import { ZardInputComponent } from '@/shared/components/input';
import { ZardSonnerService } from '@/shared/components/sonner';
import { ZardDialogRef } from '@/shared/components/dialog';

import { ProdutoService } from '../produto.service';
import type { Produto } from '../produto.model';

@Component({
  selector: 'app-produto-form',
  imports: [FormsModule, ZardButtonComponent, ZardInputComponent],
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `
    <form class="flex flex-col gap-4" (ngSubmit)="salvar()">
      <label class="flex flex-col gap-1.5 text-sm font-medium">
        Código
        <input z-input name="codigo" [(ngModel)]="codigo" required maxlength="40" placeholder="Ex: PROD-001" />
      </label>

      <label class="flex flex-col gap-1.5 text-sm font-medium">
        Descrição
        <input z-input name="descricao" [(ngModel)]="descricao" required maxlength="120" placeholder="Descrição do produto" />
      </label>

      <label class="flex flex-col gap-1.5 text-sm font-medium">
        Saldo inicial
        <input z-input type="number" name="saldo" [(ngModel)]="saldo" required min="0" />
      </label>

      <div class="flex justify-end gap-2 pt-2">
        <button type="button" z-button zType="outline" [zDisabled]="salvando()" (click)="dialogRef.close()">Cancelar</button>
        <button type="submit" z-button [zLoading]="salvando()" [zDisabled]="!formValido()">Salvar</button>
      </div>
    </form>
  `,
})
export class ProdutoFormComponent {
  readonly dialogRef = inject(ZardDialogRef<ProdutoFormComponent, Produto>);
  private readonly produtoService = inject(ProdutoService);
  private readonly sonner = inject(ZardSonnerService);

  codigo = '';
  descricao = '';
  saldo: number | null = 0;

  readonly salvando = signal(false);

  formValido(): boolean {
    return this.codigo.trim().length > 0 && this.descricao.trim().length > 0 && this.saldo !== null && this.saldo >= 0;
  }

  salvar(): void {
    if (!this.formValido() || this.salvando()) {
      return;
    }

    this.salvando.set(true);
    this.produtoService
      .criar({ codigo: this.codigo.trim(), descricao: this.descricao.trim(), saldo: this.saldo ?? 0 })
      .pipe(finalize(() => this.salvando.set(false)))
      .subscribe({
        next: produto => {
          this.sonner.success('Produto cadastrado.');
          this.dialogRef.close(produto);
        },
        error: err => this.sonner.error(mensagemDeErro(err, 'Erro ao cadastrar produto.')),
      });
  }
}
