import type { HttpErrorResponse } from '@angular/common/http';

// Os handlers Go respondem erros como { "error": "mensagem" }.
export function mensagemDeErro(err: unknown, fallback: string): string {
  if (isHttpErrorResponse(err)) {
    if (err.status === 0) {
      return 'Não foi possível conectar ao servidor.';
    }
    const corpo = err.error as { error?: string } | null;
    return corpo?.error ?? fallback;
  }
  return fallback;
}

function isHttpErrorResponse(err: unknown): err is HttpErrorResponse {
  return typeof err === 'object' && err !== null && 'status' in err;
}
