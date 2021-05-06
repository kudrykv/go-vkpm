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
	"strconv"
	"sync"
	"time"

	"github.com/antchfx/htmlquery"
	"github.com/kudrykv/go-vkpm/types"
)

type API struct {
	hc  *http.Client
	cfg types.Config
	c   types.Cookies

	blocksOn bool
	mux      *sync.Mutex
}

var (
	ErrBadCreds  = errors.New("bad credentials")
	ErrNoNode    = errors.New("node not found")
	ErrNoID      = errors.New("no id found")
	ErrNonEmpty  = errors.New("expected empty response")
	ErrBadStatus = errors.New("bad status")
)

func NewAPI(hc *http.Client, cfg types.Config) *API {
	return &API{hc: hc, cfg: cfg, mux: &sync.Mutex{}}
}

func (a *API) WithCookies(c types.Cookies) *API {
	a.c = c
	return a
}

func (a *API) Login(ctx context.Context, username, password string) (types.Cookies, error) {
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

func (a *API) Salary(ctx context.Context, year, month int) (types.Salary, error) {
	var salary types.Salary

	if err := a.allBlocksOn(ctx); err != nil {
		return salary, fmt.Errorf("turn blocks on: %w", err)
	}

	body := url.Values{"year": {strconv.Itoa(year)}, "month": {strconv.Itoa(month)}}
	bts, resp, err := a.do(ctx, http.MethodPost, "https://"+a.cfg.Domain+"/dashboard/block/user_salary_block/", body, a.h())
	if err != nil {
		return salary, fmt.Errorf("get salary block: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return salary, fmt.Errorf(resp.Status+": %w", ErrBadStatus)
	}

	doc, err := htmlquery.Parse(bytes.NewReader(bts))
	if err != nil {
		return salary, fmt.Errorf("parse salary block: %w", err)
	}

	if salary, err = types.NewSalaryFromHTMLNode(doc, year, month); err != nil {
		return salary, fmt.Errorf("new salary from html node: %w", err)
	}

	return salary, nil
}

func (a *API) Birthdays(ctx context.Context) (types.Persons, error) {
	if err := a.allBlocksOn(ctx); err != nil {
		return nil, fmt.Errorf("turn blocks on: %w", err)
	}

	uri := "https://" + a.cfg.Domain + "/dashboard/block/birthdays_block/?time_range_field=birthdays_block_end_days&time_range_value=366"
	bts, resp, err := a.do(ctx, http.MethodGet, uri, nil, a.h())
	if err != nil {
		return nil, fmt.Errorf("get birthdays: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(resp.Status+": %w", ErrBadStatus)
	}

	doc, err := htmlquery.Parse(bytes.NewReader(bts))
	if err != nil {
		return nil, fmt.Errorf("parse birthdays block: %w", err)
	}

	persons, err := types.NewPersonsFromHTMLNode(doc)
	if err != nil {
		return nil, fmt.Errorf("new persons from html node: %w", err)
	}

	return persons, nil
}

func (a *API) History(ctx context.Context, year int, month time.Month) (types.ReportEntries, error) {
	body := url.Values{"year": {strconv.Itoa(year)}, "month": {strconv.Itoa(int(month))}}

	bts, resp, err := a.do(ctx, http.MethodGet, "https://"+a.cfg.Domain+"/history/", body, a.h())
	if err != nil {
		return nil, fmt.Errorf("get history: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(resp.Status+": %w", ErrBadStatus)
	}

	doc, err := htmlquery.Parse(bytes.NewReader(bts))
	if err != nil {
		return nil, fmt.Errorf("parse history: %w", err)
	}

	entries, err := types.NewReportEntriesFromHTMLNode(doc)
	if err != nil {
		return nil, fmt.Errorf("new report entries from html node: %w", err)
	}

	return entries, nil
}

func (a *API) allBlocksOn(ctx context.Context) error {
	a.mux.Lock()
	defer a.mux.Unlock()

	if a.blocksOn {
		return nil
	}

	h := a.h()

	bts, _, err := a.do(ctx, http.MethodGet, "https://"+a.cfg.Domain+"/dashboard/", nil, h)
	if err != nil {
		return fmt.Errorf("do dashboard: %w", err)
	}

	doc, err := htmlquery.Parse(bytes.NewReader(bts))
	if err != nil {
		return fmt.Errorf("parse: %w", err)
	}

	nodes, err := htmlquery.QueryAll(doc, `//input[@name="id"]`)
	if err != nil {
		return fmt.Errorf("query all: %w", err)
	}

	var id string

	if len(nodes) == 0 {
		return fmt.Errorf(`looking for //input[@name="id"]: %w`, ErrNoNode)
	}

	for _, attribute := range nodes[0].Attr {
		if attribute.Key == "value" {
			id = attribute.Val
		}
	}

	if len(id) == 0 {
		return ErrNoID
	}

	values := url.Values{
		"id":                {id},
		"birthdays_block":   {"on"},
		"user_salary_block": {"on"},
		"users_block":       {"on"},
	}

	bts, _, err = a.do(ctx, http.MethodPost, "https://"+a.cfg.Domain+"/dashboard/update/", values, h)
	if err != nil {
		return fmt.Errorf("do dashboard update: %w", err)
	}

	if len(bts) > 0 {
		return fmt.Errorf(string(bts)+": %w", ErrNonEmpty) // nolint: goerr113
	}

	a.blocksOn = true

	return nil
}

func (a *API) h() http.Header {
	return http.Header{
		"Cookie":      {"csrftoken=" + a.c.CSRFToken + "; sessionid=" + a.c.SessionID},
		"Referer":     {"https://" + a.cfg.Domain + "/dashboard/"},
		"x-csrftoken": {a.c.CSRFToken},
	}
}

func (a *API) cookies(ctx context.Context) (string, error) {
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

func (a *API) do(
	ctx context.Context, method, url string, body url.Values, h http.Header,
) ([]byte, *http.Response, error) {
	var (
		req *http.Request
		err error
		rdr io.Reader
	)

	if body != nil {
		rdr = bytes.NewReader([]byte(body.Encode()))
	}

	req, err = http.NewRequestWithContext(ctx, method, url, rdr)
	if err != nil {
		return nil, nil, fmt.Errorf("new request: %w", err)
	}

	for k, v := range h {
		for _, s := range v {
			req.Header.Add(k, s)
		}
	}

	if len(req.Header.Get("Referer")) == 0 {
		req.Header.Set("Referer", "https://"+a.cfg.Domain+"/login/")
	}

	if rdr != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
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
