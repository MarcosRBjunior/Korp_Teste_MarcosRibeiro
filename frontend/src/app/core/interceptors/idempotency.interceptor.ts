import type { HttpEvent, HttpInterceptorFn } from '@angular/common/http';
import { type Observable, shareReplay } from 'rxjs';

import { IDEMPOTENCY_KEY } from '@/core/idempotency';

// Requisições em andamento, por chave de idempotência. Um clique duplicado
// no botão "Imprimir" enquanto a primeira chamada ainda não respondeu
// reaproveita essa mesma requisição em vez de disparar uma nova.
const emAndamento = new Map<string, Observable<HttpEvent<unknown>>>();

export const idempotencyInterceptor: HttpInterceptorFn = (req, next) => {
  const chave = req.context.get(IDEMPOTENCY_KEY);
  if (!chave) {
    return next(req);
  }

  const existente = emAndamento.get(chave);
  if (existente) {
    return existente;
  }

  const requisicaoComHeader = req.clone({ setHeaders: { 'Idempotency-Key': chave } });
  const compartilhada = next(requisicaoComHeader).pipe(shareReplay({ bufferSize: 1, refCount: false }));

  emAndamento.set(chave, compartilhada);
  compartilhada.subscribe({
    complete: () => emAndamento.delete(chave),
    error: () => emAndamento.delete(chave),
  });

  return compartilhada;
};
