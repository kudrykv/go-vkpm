package services

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/kudrykv/go-vkpm/types"
)

type API struct {
	hc  *http.Client
	cfg types.Config
}

var (
	ErrBadCreds = errors.New("bad credentials")
)

func NewAPI(hc *http.Client, cfg types.Config) API {
	return API{hc: hc, cfg: cfg}
}

func (a API) Login(ctx context.Context, username, password string) (types.Cookies, error) {
	csrf, err := a.cookies(ctx)
	if err != nil {
		return types.Cookies{}, fmt.Errorf("cookies: %w", err)
	}

	values := url.Values{}
	values.Set("csrfmiddlewaretoken", csrf)
	values.Set("username", username)
	values.Set("password", password)
	values.Set("next", "/")

	buff := bytes.NewBufferString(values.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://"+a.cfg.Domain+"/login/", buff)
	if err != nil {
		return types.Cookies{}, fmt.Errorf("post /login/: %w", err)
	}

	req.Header.Set("Accept", "*/*")
	req.Header.Set("Referer", "https://"+a.cfg.Domain+"/login/")
	req.Header.Set("Cookie", "csrftoken="+csrf)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := a.hc.Do(req)
	if err != nil {
		return types.Cookies{}, fmt.Errorf("do: %w", err)
	}

	if _, err = ioutil.ReadAll(resp.Body); err != nil {
		return types.Cookies{}, fmt.Errorf("read all: %w", err)
	}

	if err = resp.Body.Close(); err != nil {
		return types.Cookies{}, fmt.Errorf("close: %w", err)
	}

	cc := types.Cookies{}
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "csrftoken" {
			cc.CSRFToken = cookie.Value

			continue
		}

		if cookie.Name == "sessionid" {
			cc.SessionID = cookie.Value

			continue
		}
	}

	if len(cc.CSRFToken) == 0 || len(cc.SessionID) == 0 {
		return cc, ErrBadCreds
	}

	return cc, nil
}

func (a API) cookies(ctx context.Context) (string, error) {
	// resp body is already closed
	// nolint: bodyclose
	_, resp, err := a.do(ctx, http.MethodGet, "https://"+a.cfg.Domain+"/login/", nil, nil)
	if err != nil {
		return "", fmt.Errorf("do: %w", err)
	}

	for _, cookie := range resp.Cookies() {
		if cookie.Name == "csrftoken" {
			return cookie.Value, nil
		}
	}

	return "", nil
}

func (a API) do(
	ctx context.Context, method, url string, body io.Reader, h http.Header,
) ([]byte, *http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, nil, fmt.Errorf("new request: %w", err)
	}

	req.Header.Set("Referer", "https://"+a.cfg.Domain+"/login/")

	for k, v := range h {
		for _, s := range v {
			req.Header.Add(k, s)
		}
	}

	resp, err := a.hc.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("do: %w", err)
	}

	bts, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("read all: %w", err)
	}

	if err = resp.Body.Close(); err != nil {
		return nil, nil, fmt.Errorf("close: %w", err)
	}

	return bts, resp, nil
}
