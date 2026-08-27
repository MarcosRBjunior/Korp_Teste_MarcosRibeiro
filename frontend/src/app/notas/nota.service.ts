import { HttpClient, HttpContext } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { type Observable, tap } from 'rxjs';

import { environment } from '../../environments/environment';
import { IDEMPOTENCY_KEY } from '@/core/idempotency';
import type { ItemNotaInput, NotaFiscal } from './nota.model';

@Injectable({ providedIn: 'root' })
export class NotaFiscalService {
  private readonly http = inject(HttpClient);
  private readonly baseUrl = environment.faturamentoApiUrl;

  // Uma chave por nota, reaproveitada entre chamadas até a impressão dar
  // certo. Se a 1ª chamada tiver sucesso no servidor mas a resposta se
  // perder no caminho, o usuário clica de novo em "Imprimir" — precisa ser
  // a MESMA chave para que o backend devolva o resultado já salvo em vez de
  // tentar debitar o estoque de novo.
  private readonly chavesImpressao = new Map<number, string>();

  listar(): Observable<NotaFiscal[]> {
    return this.http.get<NotaFiscal[]>(`${this.baseUrl}/notas`);
  }

  criar(itens: ItemNotaInput[]): Observable<NotaFiscal> {
    return this.http.post<NotaFiscal>(`${this.baseUrl}/notas`, { itens });
  }

  imprimir(id: number): Observable<NotaFiscal> {
    let chave = this.chavesImpressao.get(id);
    if (!chave) {
      chave = crypto.randomUUID();
      this.chavesImpressao.set(id, chave);
    }

    const context = new HttpContext().set(IDEMPOTENCY_KEY, chave);
    return this.http
      .post<NotaFiscal>(`${this.baseUrl}/notas/${id}/imprimir`, {}, { context })
      .pipe(tap(() => this.chavesImpressao.delete(id)));
  }
}
