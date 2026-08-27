import { HttpContextToken } from '@angular/common/http';

// Chave de idempotência de uma ação do usuário (ex.: um clique em "Imprimir").
// Gerada uma vez pelo chamador e propagada pelo HttpContext — não pelo
// interceptor — para que retries da mesma requisição HTTP reaproveitem a
// mesma chave em vez de gerar uma nova a cada tentativa.
export const IDEMPOTENCY_KEY = new HttpContextToken<string | undefined>(() => undefined);
