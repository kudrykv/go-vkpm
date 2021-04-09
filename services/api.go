package services

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/kudrykv/go-vkpm/types"
)

type API struct {
	hc  *http.Client
	cfg types.Config
}

func NewAPI(hc *http.Client, cfg types.Config) API {
	return API{hc: hc, cfg: cfg}
}

func (a API) Login(ctx context.Context, username, password string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, a.cfg.Domain+"/login", nil)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	resp, err := a.hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do: %w", err)
	}

	bts, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read all: %w", err)
	}

	if err = resp.Body.Close(); err != nil {
		return nil, fmt.Errorf("close: %w", err)
	}

	return bts, nil
}
