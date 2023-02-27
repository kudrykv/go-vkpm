package services

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"runtime/trace"
	"strconv"
	"sync"
	"time"

	"github.com/antchfx/htmlquery"
	"github.com/kudrykv/go-vkpm/app/config"
	"github.com/kudrykv/go-vkpm/app/th"
	"github.com/kudrykv/go-vkpm/app/types"
	"github.com/kudrykv/littlehttp"
	"golang.org/x/net/html"
)

type API struct {
	cfg     config.Config
	cookies config.Cookies

	blocksOn   bool
	mutex      *sync.Mutex
	semaphore  chan struct{}
	littleHTTP *littlehttp.LittleHTTP
}

var (
	ErrBadCreds  = errors.New("bad credentials")
	ErrNoNode    = errors.New("node not found")
	ErrNoID      = errors.New("no id found")
	ErrNonEmpty  = errors.New("expected empty response")
	ErrBadStatus = errors.New("bad status")
	ErrNoReport  = errors.New("no report found")
)

func NewAPI(littleHTTP *littlehttp.LittleHTTP, cfg config.Config) *API {
	return &API{
		littleHTTP: littleHTTP,
		cfg:        cfg,
		mutex:      &sync.Mutex{},
		semaphore:  make(chan struct{}, 4),
	}
}

func (a *API) WithCookies(c config.Cookies) *API {
	a.cookies = c
	return a
}

func (a *API) Login(ctx context.Context, username, password string) (config.Cookies, error) {
	ctx, end := th.RegionTask(ctx, "login")
	defer end()

	csrf, err := a.getCookies(ctx)
	if err != nil {
		return config.Cookies{}, fmt.Errorf("cookies: %w", err)
	}

	values := url.Values{
		"csrfmiddlewaretoken": {csrf},
		"username":            {username},
		"password":            {password},
		"next":                {"/"},
	}

	h := http.Header{
		"Accept":       {"*/*"},
		"Referer":      {"https://" + a.cfg.Domain + "/login/"},
		"Cookie":       {"csrftoken=" + csrf},
		"Content-Type": {"application/x-www-form-urlencoded"},
	}

	_, resp, err := a.do(ctx, http.MethodPost, "/login/", values, h)
	if err != nil {
		return config.Cookies{}, fmt.Errorf("do: %w", err)
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
	ctx, end := th.RegionTask(ctx, "salary")
	defer end()

	var salary types.Salary

	if err := a.allBlocksOn(ctx); err != nil {
		return salary, fmt.Errorf("turn blocks on: %w", err)
	}

	body := url.Values{"year": {strconv.Itoa(year)}, "month": {strconv.Itoa(int(month))}}

	doc, err := a.doParse(ctx, http.MethodPost, "/dashboard/block/user_salary_block/", body)
	if err != nil {
		return salary, fmt.Errorf("do parse: %w", err)
	}

	if salary, err = types.NewSalaryFromHTMLNode(ctx, doc, year, month); err != nil {
		return salary, fmt.Errorf("new salary from html node: %w", err)
	}

	return salary, nil
}

func (a *API) Birthdays(ctx context.Context) (types.Persons, error) {
	ctx, end := th.RegionTask(ctx, "birthdays")
	defer end()

	if err := a.allBlocksOn(ctx); err != nil {
		return nil, fmt.Errorf("turn blocks on: %w", err)
	}

	uri := "/dashboard/block/birthdays_block/?time_range_field=birthdays_block_end_days&time_range_value=366"

	doc, err := a.doParse(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, fmt.Errorf("do parse: %w", err)
	}

	persons, err := types.NewPersonsFromHTMLNode(ctx, doc)
	if err != nil {
		return nil, fmt.Errorf("new persons from html node: %w", err)
	}

	return persons, nil
}

func (a *API) History(ctx context.Context, year int, month time.Month) (types.ReportEntries, error) {
	ctx, end := th.RegionTask(ctx, "history")
	defer end()

	body := url.Values{"year": {strconv.Itoa(year)}, "month": {strconv.Itoa(int(month))}}

	doc, err := a.doParse(ctx, http.MethodPost, "/history/", body)
	if err != nil {
		return nil, fmt.Errorf("do parse: %w", err)
	}

	entries, err := types.NewReportEntriesFromHTMLNode(ctx, doc)
	if err != nil {
		return nil, fmt.Errorf("new report entries from html node: %w", err)
	}

	return entries, nil
}

func (a *API) VacationsHolidays(ctx context.Context, year int) (int, types.Vacations, types.Holidays, error) {
	ctx, end := th.RegionTask(ctx, "vacations holidays")
	defer end()

	body := url.Values{"year": {strconv.Itoa(year)}, "year_changed": {"true"}}

	doc, err := a.doParse(ctx, http.MethodPost, "/breaks/", body)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("do parse: %w", err)
	}

	paidVacDays, vacations, err := types.NewVacationsFromHTMLNode(doc)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("new vacations from html node: %w", err)
	}

	holidays, err := types.NewHolidaysFromHTMLNode(ctx, doc)
	if err != nil {
		return 0, nil, nil, fmt.Errorf("new holidays from html node: %w", err)
	}

	return paidVacDays, vacations, holidays, nil
}

func (a *API) Projects(ctx context.Context) (types.Projects, error) {
	ctx, end := th.RegionTask(ctx, "projects")
	defer end()

	doc, err := a.doParse(ctx, http.MethodGet, "/report/", nil)
	if err != nil {
		return nil, fmt.Errorf("do parse: %w", err)
	}

	projects, err := types.NewProjectsFromHTMLNode(ctx, doc)
	if err != nil {
		return nil, fmt.Errorf("new projects from html node: %w", err)
	}

	return projects, nil
}

func (a *API) Report(ctx context.Context, entry types.ReportEntry) (types.ReportEntry, error) {
	ctx, end := th.RegionTask(ctx, "report")
	defer end()

	body, err := entry.URLValues()
	if err != nil {
		return entry, fmt.Errorf("url values: %w", err)
	}

	bts, resp, err := a.do(ctx, http.MethodPost, "/report/", body, a.h())
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

func (a *API) PersonInfo(ctx context.Context, id int) (types.Person, error) {
	ctx, end := th.RegionTask(ctx, "user info")
	defer end()

	node, err := a.doParse(ctx, http.MethodGet, "/dashboard/user_profile/"+strconv.Itoa(id)+"/", nil)
	if err != nil {
		return types.Person{}, fmt.Errorf("do parse: %w", err)
	}

	person, err := types.NewPersonUserProfileFromHTMLNode(ctx, node)
	if err != nil {
		return person, fmt.Errorf("new person user profile from html node: %w", err)
	}

	person.ID = id

	return person, nil
}

func (a *API) GetPicture(ctx context.Context, uri string) ([]byte, error) {
	bts, _, err := a.do(ctx, http.MethodGet, uri, nil, a.h())
	if err != nil {
		return nil, fmt.Errorf("do: %w", err)
	}

	return bts, nil
}

func (a *API) doParse(ctx context.Context, method, url string, body url.Values) (*html.Node, error) {
	defer trace.StartRegion(ctx, "do and parse").End()

	bts, resp, err := a.do(ctx, method, url, body, a.h())
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

	a.mutex.Lock()
	defer a.mutex.Unlock()

	trace.Logf(ctx, "vars", "blocksOn=%v", a.blocksOn)
	if a.blocksOn {
		return nil
	}

	h := a.h()

	bts, _, err := a.do(ctx, http.MethodGet, "/dashboard/", nil, h)
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

	bts, _, err = a.do(ctx, http.MethodPost, "/dashboard/update/", values, h)
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
		"Cookie":      {"csrftoken=" + a.cookies.CSRFToken + "; sessionid=" + a.cookies.SessionID},
		"Referer":     {"https://" + a.cfg.Domain + "/dashboard/"},
		"x-csrftoken": {a.cookies.CSRFToken},
	}
}

func (a *API) getCookies(ctx context.Context) (string, error) {
	defer trace.StartRegion(ctx, "cookies").End()

	_, resp, err := a.do(ctx, http.MethodGet, "/login/", nil, nil)
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

	trace.Logf(ctx, "httpreq", "method: %v, url: %v", method, url)

	a.semaphore <- struct{}{}
	defer func() { <-a.semaphore }()

	if h == nil {
		h = http.Header{}
	}

	if len(h.Get("Referer")) == 0 {
		h.Set("Referer", "https://"+a.cfg.Domain+"/login/")
	}

	resp, err := a.littleHTTP.Do(ctx, littlehttp.NewRequest(method, url, h, body))
	if err != nil {
		return nil, nil, fmt.Errorf("do: %w", err)
	}

	bts, err := resp.Bytes()
	if err != nil {
		return nil, nil, fmt.Errorf("bytes: %w", err)
	}

	return bts, resp.Raw(), nil
}
