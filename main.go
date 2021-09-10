package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

const trackerURL = "https://tracker-api.toptal.com"

type TrackerSessionsRequest struct {
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
}

type TrackerSessionsResponse struct {
	AccessToken string `json:"access_token,omitempty"`
}

type TrackerFiltersResponse struct {
	Filters *TrackerFilters `json:"filters,omitempty"`
}

type TrackerFilters struct {
	Projects TrackerEntities `json:"projects,omitempty"`
	Workers  TrackerEntities `json:"workers,omitempty"`
}

type TrackerEntity struct {
	ID    int          `json:"id,omitempty"`
	Label string       `json:"label,omitempty"`
	Dates TrackerDates `json:"dates,omitempty"`
}

type TrackerEntities []TrackerEntity

type TrackerDate struct {
	Date    string `json:"date,omitempty"`
	Seconds int    `json:"seconds,omitempty"`
}

type TrackerDates []TrackerDate

type TrackerChartResponse struct {
	Reports *TrackerChartReports `json:"reports,omitempty"`
}

type TrackerChartReports struct {
	Projects *TrackerReportData `json:"projects,omitempty"`
}

type TrackerReportData struct {
	Data         TrackerEntities `json:"data,omitempty"`
	TotalSeconds int             `json:"total_seconds,omitempty"`
}

func main() {
	accessToken, err := getAccessToken()
	if err != nil {
		panic(err)
	}

	filters, err := getFilters(accessToken)
	if err != nil {
		panic(err)
	}

	date, err := time.Parse("2006-01-02", os.Getenv("DATE"))
	if err != nil {
		panic(err)
	}

	startDate, endDate := getDates(date)

	fmt.Println(startDate, endDate)

	report, err := getReport(filters, startDate, endDate, accessToken)
	if err != nil {
		panic(err)
	}

	projects := report.Reports.Projects
	fmt.Printf("%d %.f %.f %.2f %.2f\n", projects.TotalSeconds,
		math.Floor(float64(projects.TotalSeconds)/60/60),
		math.Floor(float64(projects.TotalSeconds)/60/60)*100,
		float64(projects.TotalSeconds)/60/60,
		float64(projects.TotalSeconds)/60/60*100)
	for _, data := range projects.Data {
		fmt.Println(data.Label)
		for _, date := range data.Dates {
			fmt.Println(date.Date, date.Seconds)
		}
	}
}

func getAccessToken() (string, error) {
	reqData := &TrackerSessionsRequest{
		Email:    os.Getenv("EMAIL"),
		Password: os.Getenv("PASSWORD"),
	}
	reqDataBuf, err := json.Marshal(reqData)
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal request data")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx,
		http.MethodPost,
		fmt.Sprintf("%s/sessions", trackerURL),
		bytes.NewReader(reqDataBuf))
	if err != nil {
		return "", errors.Wrap(err, "failed to create new request")
	}
	defer func() {
		if err := req.Body.Close(); err != nil {
			// TODO: Handle error here.
		}
	}()
	req.Header.Set("Content-Type", "application/json")

	res, err := (&http.Client{}).Do(req)
	if err != nil {
		return "", errors.Wrap(err, "failed to do request")
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			// TODO: Handle error here.
		}
	}()

	if res.StatusCode != http.StatusCreated {
		return "", errors.Errorf(`incorrect response status code: %d`, res.StatusCode)
	}

	resData := &TrackerSessionsResponse{}
	if err := json.NewDecoder(res.Body).Decode(resData); err != nil {
		return "", errors.Wrap(err, "failed to decode response data")
	}

	return resData.AccessToken, nil
}

func getFilters(accessToken string) (*TrackerFilters, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/reports/filters", trackerURL), nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new request")
	}
	reqQuery := req.URL.Query()
	reqQuery.Set("access_token", accessToken)
	req.URL.RawQuery = reqQuery.Encode()

	res, err := (&http.Client{}).Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to do request")
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			// TODO: Handle error here.
		}
	}()

	if res.StatusCode != http.StatusOK {
		return nil, errors.Errorf(`incorrect response status code: %d`, res.StatusCode)
	}

	resData := &TrackerFiltersResponse{}
	if err := json.NewDecoder(res.Body).Decode(resData); err != nil {
		return nil, errors.Wrap(err, "failed to decode response data")
	}

	return resData.Filters, nil
}

func getReport(filters *TrackerFilters, startDate, endDate, accessToken string) (*TrackerChartResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/reports/chart", trackerURL), nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new request")
	}
	reqQuery := req.URL.Query()
	reqQuery.Set("access_token", accessToken)
	reqQuery.Set("start_date", startDate)
	reqQuery.Set("end_date", endDate)
	for _, p := range filters.Projects {
		reqQuery.Add("project_ids[]", strconv.Itoa(p.ID))
	}
	for _, w := range filters.Workers {
		reqQuery.Add("worker_ids[]", strconv.Itoa(w.ID))
	}
	req.URL.RawQuery = reqQuery.Encode()

	res, err := (&http.Client{}).Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to do request")
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			// TODO: Handle error here.
		}
	}()

	if res.StatusCode != http.StatusOK {
		return nil, errors.Errorf(`incorrect response status code: %d`, res.StatusCode)
	}

	resData := &TrackerChartResponse{}
	if err := json.NewDecoder(res.Body).Decode(resData); err != nil {
		return nil, errors.Wrap(err, "failed to decode response data")
	}

	return resData, nil
}

func getDates(date time.Time) (string, string) {
	if date.Day() >= 10 && date.Day() <= 24 {
		return fmt.Sprintf("%d-%s-%d", date.Year(), prepareMonth(date.Month()), 10),
			fmt.Sprintf("%d-%s-%d", date.Year(), prepareMonth(date.Month()), 24)
	}
	if date.Day() <= 9 {
		year := date.Year()
		month := date.Month()
		if date.Month() == 1 {
			year -= 1
			month = time.December
		} else {
			month--
		}
		return fmt.Sprintf("%d-%s-%d", year, prepareMonth(month), 25),
			fmt.Sprintf("%d-%s-0%d", date.Year(), prepareMonth(date.Month()), 9)
	}
	if date.Day() >= 25 {
		year := date.Year()
		month := date.Month()
		if date.Month() == 12 {
			year++
			month = time.January
		} else {
			month++
		}
		return fmt.Sprintf("%d-%s-%d", date.Year(), prepareMonth(date.Month()), 25),
			fmt.Sprintf("%d-%s-0%d", year, prepareMonth(month), 9)
	}
	panic("not implemented")
}

func prepareMonth(month time.Month) string {
	if month < 10 {
		return fmt.Sprintf("0%d", month)
	}
	return fmt.Sprintf("%d", month)
}
