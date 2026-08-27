// Package estoqueclient é o cliente HTTP do Faturamento para o serviço de
// Estoque, protegido por circuit breaker (sony/gobreaker). Ele isola o
// restante do serviço de detalhes de transporte e do estado do breaker.
package estoqueclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/sony/gobreaker/v2"
)

var (
	// ErrIndisponivel cobre timeout, conexão recusada, erro 5xx do Estoque
	// e o circuit breaker aberto — tudo que deve virar um 503 amigável.
	ErrIndisponivel         = errors.New("serviço de estoque indisponível")
	ErrSaldoInsuficiente    = errors.New("saldo insuficiente")
	ErrProdutoNaoEncontrado = errors.New("produto não encontrado")
)

type Client struct {
	baseURL string
	http    *http.Client
	breaker *gobreaker.CircuitBreaker[*http.Response]
}

func New(baseURL string, timeout time.Duration) *Client {
	settings := gobreaker.Settings{
		Name:        "estoque",
		MaxRequests: 1,
		Interval:    0,
		Timeout:     10 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 3
		},
	}

	return &Client{
		baseURL: baseURL,
		http:    &http.Client{Timeout: timeout},
		breaker: gobreaker.NewCircuitBreaker[*http.Response](settings),
	}
}

type debitarRequest struct {
	Quantidade int `json:"quantidade"`
}

// Debitar chama POST /produtos/{id}/debitar no Estoque através do circuit
// breaker. Erros de rede, timeout e 5xx contam como falha para o breaker;
// 409 (saldo insuficiente) e 404 (produto inexistente) são erros de negócio
// e não abrem o breaker.
func (c *Client) Debitar(ctx context.Context, produtoID uint, quantidade int) error {
	body, err := json.Marshal(debitarRequest{Quantidade: quantidade})
	if err != nil {
		return fmt.Errorf("erro ao montar requisição de débito: %w", err)
	}

	resp, err := c.breaker.Execute(func() (*http.Response, error) {
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodPost,
			fmt.Sprintf("%s/produtos/%d/debitar", c.baseURL, produtoID),
			bytes.NewReader(body),
		)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.http.Do(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode >= 500 {
			resp.Body.Close()
			return nil, fmt.Errorf("estoque respondeu %d", resp.StatusCode)
		}
		return resp, nil
	})

	if err != nil {
		if errors.Is(err, gobreaker.ErrOpenState) || errors.Is(err, gobreaker.ErrTooManyRequests) {
			return ErrIndisponivel
		}
		return ErrIndisponivel
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusConflict:
		return ErrSaldoInsuficiente
	case http.StatusNotFound:
		return ErrProdutoNaoEncontrado
	default:
		return fmt.Errorf("estoque respondeu status inesperado %d", resp.StatusCode)
	}
}
