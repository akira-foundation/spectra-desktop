package billing

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
	"sync"
	"time"

	billingsdk "github.com/akira-io/billing-sdk-go"

	"spectra-desktop/internal/domain"
)

const (
	oauthCallbackTimeout = 5 * time.Minute
	oauthReadTimeout     = 15 * time.Second
)

var oauthAbort = &abortRegistry{}

type abortRegistry struct {
	mu     sync.Mutex
	cancel context.CancelFunc
}

func (a *abortRegistry) replace(cancel context.CancelFunc) {
	a.mu.Lock()
	prev := a.cancel
	a.cancel = cancel
	a.mu.Unlock()
	if prev != nil {
		prev()
	}
}

func (a *abortRegistry) clear() {
	a.mu.Lock()
	a.cancel = nil
	a.mu.Unlock()
}

type OauthResult struct {
	AccessToken           string
	Customer              billingsdk.OauthExchangeCustomer
	Entitlement           *billingsdk.OauthExchangeEntitlement
	RequiresPlanSelection bool
}

type BrowserOpener func(url string) error

func (c *Client) StartOauthLogin(ctx context.Context, provider string, openBrowser BrowserOpener) (*OauthResult, error) {
	if provider == "" {
		return nil, errors.New("billing: provider required")
	}
	if openBrowser == nil {
		return nil, errors.New("billing: browser opener required")
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("billing: bind callback: %w", err)
	}
	defer listener.Close()

	addr, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		return nil, errors.New("billing: unexpected listener addr")
	}
	redirectURI := fmt.Sprintf("http://127.0.0.1:%d/cb", addr.Port)

	pkce, err := billingsdk.GeneratePkceChallenge()
	if err != nil {
		return nil, fmt.Errorf("billing: pkce: %w", err)
	}
	state, err := billingsdk.GenerateOauthState()
	if err != nil {
		return nil, fmt.Errorf("billing: state: %w", err)
	}

	authURL := billingsdk.BuildOauthInitURL(billingsdk.BuildOauthInitUrlOptions{
		BaseURL:             BillingURL,
		Provider:            provider,
		Product:             ProductSlug,
		RedirectURI:         redirectURI,
		CodeChallenge:       pkce.Challenge,
		CodeChallengeMethod: pkce.Method,
		State:               state,
	})

	if err := openBrowser(authURL); err != nil {
		return nil, fmt.Errorf("billing: open browser: %w", err)
	}

	callbackCtx, cancel := context.WithTimeout(ctx, oauthCallbackTimeout)
	defer cancel()
	oauthAbort.replace(cancel)
	defer oauthAbort.clear()

	code, returnedState, err := acceptCallback(callbackCtx, listener)
	if err != nil {
		return nil, err
	}
	if returnedState != state {
		return nil, errors.New("billing: oauth state mismatch")
	}

	exchange, err := c.SDK().ExchangeOauthCode(ctx, billingsdk.OauthExchangePayload{
		Code:         code,
		CodeVerifier: pkce.Verifier,
	})
	if err != nil {
		return nil, fmt.Errorf("billing: exchange code: %w", err)
	}

	if err := c.PersistAccessToken(ctx, exchange.AccessToken); err != nil {
		return nil, err
	}

	if err := c.persistCustomerFromOauth(ctx, exchange); err != nil {
		return nil, err
	}

	return &OauthResult{
		AccessToken:           exchange.AccessToken,
		Customer:              exchange.Customer,
		Entitlement:           exchange.Entitlement,
		RequiresPlanSelection: exchange.RequiresPlanSelection,
	}, nil
}

func CancelPendingOauth() {
	oauthAbort.replace(nil)
}

func acceptCallback(ctx context.Context, listener net.Listener) (string, string, error) {
	type acceptResult struct {
		conn net.Conn
		err  error
	}
	resultCh := make(chan acceptResult, 1)
	go func() {
		conn, err := listener.Accept()
		resultCh <- acceptResult{conn: conn, err: err}
	}()

	var conn net.Conn
	select {
	case <-ctx.Done():
		_ = listener.Close()
		return "", "", fmt.Errorf("billing: oauth callback: %w", ctx.Err())
	case r := <-resultCh:
		if r.err != nil {
			return "", "", fmt.Errorf("billing: accept callback: %w", r.err)
		}
		conn = r.conn
	}
	defer conn.Close()
	_ = conn.SetReadDeadline(time.Now().Add(oauthReadTimeout))

	reader := bufio.NewReader(conn)
	requestLine, err := reader.ReadString('\n')
	if err != nil {
		return "", "", fmt.Errorf("billing: read request line: %w", err)
	}
	parts := strings.Fields(requestLine)
	if len(parts) < 2 {
		return "", "", errors.New("billing: malformed request line")
	}

	requestURL, err := url.Parse(parts[1])
	if err != nil {
		return "", "", fmt.Errorf("billing: parse url: %w", err)
	}
	code := requestURL.Query().Get("code")
	state := requestURL.Query().Get("state")
	errCode := requestURL.Query().Get("error")

	writeCallbackResponse(conn, code != "" && errCode == "")

	if errCode != "" {
		return "", "", fmt.Errorf("billing: oauth provider error: %s", errCode)
	}
	if code == "" {
		return "", "", errors.New("billing: oauth callback missing code")
	}
	return code, state, nil
}

func writeCallbackResponse(conn net.Conn, success bool) {
	title := "Sign-in failed"
	body := "Something went wrong. Return to Spectra and try again."
	if success {
		title = "Signed in"
		body = "You can close this tab and return to Spectra."
	}
	html := fmt.Sprintf(`<!doctype html>
<html><head><meta charset="utf-8"><title>%s · Spectra</title>
<style>body{font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif;background:#0f0f12;color:#e8e8ec;display:flex;align-items:center;justify-content:center;height:100vh;margin:0}main{text-align:center;max-width:420px;padding:32px}h1{font-size:20px;margin:0 0 8px}p{font-size:13px;color:#9a9aa6;margin:0}</style>
</head><body><main><h1>%s</h1><p>%s</p></main></body></html>`, title, title, body)
	resp := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/html; charset=utf-8\r\nContent-Length: %d\r\nConnection: close\r\n\r\n%s", len(html), html)
	_ = conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	_, _ = conn.Write([]byte(resp))
}

func (c *Client) persistCustomerFromOauth(ctx context.Context, response *billingsdk.OauthExchangeResponse) error {
	license, err := c.repo.Get(ctx)
	if err != nil {
		return err
	}
	if license == nil {
		license = &domain.License{ID: "local", Status: "inactive", FeaturesJSON: "{}"}
	}
	license.CustomerID = response.Customer.ID
	license.CustomerEmail = response.Customer.Email
	if response.Customer.Name != nil {
		license.CustomerName = *response.Customer.Name
	}
	if response.Entitlement != nil && response.Entitlement.PlanKey != nil {
		license.Plan = *response.Entitlement.PlanKey
	}
	return c.repo.Save(ctx, *license)
}
