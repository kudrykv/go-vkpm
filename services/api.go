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
	"runtime/trace"
	"strconv"
	"sync"
	"time"

	"github.com/antchfx/htmlquery"
	"github.com/kudrykv/vkpm/config"
	"github.com/kudrykv/vkpm/types"
	"golang.org/x/net/html"
)

type API struct {
	hc  *http.Client
	cfg config.Config
	c   config.Cookies

	blocksOn bool
	mux      *sync.Mutex
	sem      chan struct{}
}

var (
	ErrBadCreds  = errors.New("bad credentials")
	ErrNoNode    = errors.New("node not found")
	ErrNoID      = errors.New("no id found")
	ErrNonEmpty  = errors.New("expected empty response")
	ErrBadStatus = errors.New("bad status")
	ErrNoReport  = errors.New("no report found")
)

func NewAPI(hc *http.Client, cfg config.Config) *API {
	return &API{hc: hc, cfg: cfg, mux: &sync.Mutex{}, sem: make(chan struct{}, 4)}
}

func (a *API) WithCookies(c config.Cookies) *API {
	a.c = c
	return a
}

func (a *API) Login(ctx context.Context, username, password string) (config.Cookies, error) {
	ctx, task := trace.NewTask(ctx, "login")
	defer task.End()

	csrf, err := a.cookies(ctx)
	if err != nil {
		return config.Cookies{}, fmt.Errorf("cookies: %w", err)
	}

	values := url.Values{}
	values.Set("csrfmiddlewaretoken", csrf)
	values.Set("username", username)
	values.Set("password", password)
	values.Set("next", "/")

	buff := bytes.NewBufferString(values.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://"+a.cfg.Domain+"/login/", buff)
	if err != nil {
		return config.Cookies{}, fmt.Errorf("post /login/: %w", err)
	}

	req.Header.Set("Accept", "*/*")
	req.Header.Set("Referer", "https://"+a.cfg.Domain+"/login/")
	req.Header.Set("Cookie", "csrftoken="+csrf)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := a.hc.Do(req)
	if err != nil {
		return config.Cookies{}, fmt.Errorf("do: %w", err)
	}

	if _, err = ioutil.ReadAll(resp.Body); err != nil {
		return config.Cookies{}, fmt.Errorf("read all: %w", err)
	}

	if err = resp.Body.Close(); err != nil {
		return config.Cookies{}, fmt.Errorf("close: %w", err)
	}

	cc := config.Cookies{}
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

func (a *API) Salary(ctx context.Context, year int, month time.Month) (types.Salary, error) {
	ctx, task := trace.NewTask(ctx, "salary")
	defer task.End()

	var salary types.Salary

	if err := a.allBlocksOn(ctx); err != nil {
		return salary, fmt.Errorf("turn blocks on: %w", err)
	}

	body := url.Values{"year": {strconv.Itoa(year)}, "month": {strconv.Itoa(int(month))}}

	doc, err := a.doParse(ctx, http.MethodPost, "/dashboard/block/user_salary_block/", body)
	if err != nil {
		return salary, fmt.Errorf("do parse: %w", err)
	}

	if salary, err = types.NewSalaryFromHTMLNode(doc, year, month); err != nil {
		return salary, fmt.Errorf("new salary from html node: %w", err)
	}

	return salary, nil
}

func (a *API) Birthdays(ctx context.Context) (types.Persons, error) {
	ctx, task := trace.NewTask(ctx, "birthdays")
	defer task.End()

	if err := a.allBlocksOn(ctx); err != nil {
		return nil, fmt.Errorf("turn blocks on: %w", err)
	}

	uri := "/dashboard/block/birthdays_block/?time_range_field=birthdays_block_end_days&time_range_value=366"

	doc, err := a.doParse(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, fmt.Errorf("do parse: %w", err)
	}

	persons, err := types.NewPersonsFromHTMLNode(doc)
	if err != nil {
		return nil, fmt.Errorf("new persons from html node: %w", err)
	}

	return persons, nil
}

func (a *API) History(ctx context.Context, year int, month time.Month) (types.ReportEntries, error) {
	ctx, task := trace.NewTask(ctx, "history")
	defer task.End()

	body := url.Values{"year": {strconv.Itoa(year)}, "month": {strconv.Itoa(int(month))}}

	doc, err := a.doParse(ctx, http.MethodPost, "/history/", body)
	if err != nil {
		return nil, fmt.Errorf("do parse: %w", err)
	}

	entries, err := types.NewReportEntriesFromHTMLNode(doc)
	if err != nil {
		return nil, fmt.Errorf("new report entries from html node: %w", err)
	}

	return entries, nil
}

func (a *API) VacationsHolidays(ctx context.Context, year int) (types.Vacations, types.Holidays, error) {
	ctx, task := trace.NewTask(ctx, "vacations holidays")
	defer task.End()

	body := url.Values{"year": {strconv.Itoa(year)}, "year_changed": {"true"}}

	doc, err := a.doParse(ctx, http.MethodPost, "/vacations/", body)
	if err != nil {
		return nil, nil, fmt.Errorf("do parse: %w", err)
	}

	vacations, err := types.NewVacationsFromHTMLNode(doc)
	if err != nil {
		return nil, nil, fmt.Errorf("new vacations from html node: %w", err)
	}

	holidays, err := types.NewHolidaysFromHTMLNode(doc)
	if err != nil {
		return nil, nil, fmt.Errorf("new holidays from html node: %w", err)
	}

	return vacations, holidays, nil
}

func (a *API) Projects(ctx context.Context) (types.Projects, error) {
	ctx, task := trace.NewTask(ctx, "projects")
	defer task.End()

	doc, err := a.doParse(ctx, http.MethodGet, "/report/", nil)
	if err != nil {
		return nil, fmt.Errorf("do parse: %w", err)
	}

	projects, err := types.NewProjectsFromHTMLNode(doc)
	if err != nil {
		return nil, fmt.Errorf("new projects from html node: %w", err)
	}

	return projects, nil
}

func (a *API) Report(ctx context.Context, entry types.ReportEntry) (types.ReportEntry, error) {
	ctx, task := trace.NewTask(ctx, "report")
	defer task.End()

	body, err := entry.URLValues()
	if err != nil {
		return entry, fmt.Errorf("url values: %w", err)
	}

	bts, resp, err := a.do(ctx, http.MethodPost, "https://"+a.cfg.Domain+"/report/", body, a.h())
	if err != nil {
		return entry, fmt.Errorf("do: %w", err)
	}

	if len(bts) > 0 {
		return entry, fmt.Errorf(resp.Status+": %w", ErrBadStatus)
	}

	history, err := a.History(ctx, entry.ReportDate.Year(), entry.ReportDate.Month())
	if err != nil {
		return entry, fmt.Errorf("history: %w", err)
	}

	today := history.FindLatestForToday(entry.ReportDate)
	if today == nil {
		return entry, fmt.Errorf("did not find reported: %w", ErrNoReport)
	}

	if !today.IsSame(entry) {
		return entry, fmt.Errorf("report didn't work: %w", ErrNoReport)
	}

	return *today, nil
}

func (a *API) doParse(ctx context.Context, method, url string, body url.Values) (*html.Node, error) {
	defer trace.StartRegion(ctx, "do and parse").End()

	bts, resp, err := a.do(ctx, method, "https://"+a.cfg.Domain+url, body, a.h())
	if err != nil {
		return nil, fmt.Errorf("do: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(resp.Status+": %w", ErrBadStatus)
	}

	region := trace.StartRegion(ctx, "parse html")
	doc, err := htmlquery.Parse(bytes.NewReader(bts))
	region.End()

	if err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}

	return doc, nil
}

func (a *API) allBlocksOn(ctx context.Context) error {
	defer trace.StartRegion(ctx, "enable all blocks").End()

	a.mux.Lock()
	defer a.mux.Unlock()

	trace.Logf(ctx, "vars", "blocksOn=%v", a.blocksOn)
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
	defer trace.StartRegion(ctx, "cookies").End()

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
	defer trace.StartRegion(ctx, "http request").End()

	a.sem <- struct{}{}
	defer func() { <-a.sem }()

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
