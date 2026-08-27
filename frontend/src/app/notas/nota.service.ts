import { HttpClient, HttpContext } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import type { Observable } from 'rxjs';

import { environment } from '../../environments/environment';
import { IDEMPOTENCY_KEY } from '@/core/idempotency';
import type { ItemNotaInput, NotaFiscal } from './nota.model';

@Injectable({ providedIn: 'root' })
export class NotaFiscalService {
  private readonly http = inject(HttpClient);
  private readonly baseUrl = environment.faturamentoApiUrl;

  listar(): Observable<NotaFiscal[]> {
    return this.http.get<NotaFiscal[]>(`${this.baseUrl}/notas`);
  }

  criar(itens: ItemNotaInput[]): Observable<NotaFiscal> {
    return this.http.post<NotaFiscal>(`${this.baseUrl}/notas`, { itens });
  }

  // Gera a Idempotency-Key uma vez por clique e a propaga via HttpContext.
  // O interceptor a anexa como header e reaproveita a chamada em andamento
  // caso o botão seja clicado de novo antes da resposta chegar.
  imprimir(id: number): Observable<NotaFiscal> {
    const chave = crypto.randomUUID();
    const context = new HttpContext().set(IDEMPOTENCY_KEY, chave);
    return this.http.post<NotaFiscal>(`${this.baseUrl}/notas/${id}/imprimir`, {}, { context });
  }
}
