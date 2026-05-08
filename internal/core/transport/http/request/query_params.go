package core_http_request

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	core_errors "github.com/Kosench/golang-todoapp/internal/core/errors"
)

func GetIntQueryParam(r *http.Request, key string) (*int, error) {
	param := r.URL.Query().Get(key)
	if param == "" {
		return nil, nil
	}

	val, err := strconv.Atoi(param)
	if err != nil {
		return nil, fmt.Errorf(
			"param='%s' by key='%s' not a valid integer: %v: %w",
			param,
			key,
			err,
			core_errors.ErrInvalidArgument,
		)
	}

	return &val, nil
}

func GetPositiveIntQueryParam(r *http.Request, key string) (*int, error) {
	val, err := GetIntQueryParam(r, key)
	if err != nil {
		return nil, err
	}
	if val == nil {
		return nil, nil
	}

	if *val <= 0 {
		return nil, fmt.Errorf(
			"param by key='%s' must be positive: %w",
			key,
			core_errors.ErrInvalidArgument,
		)
	}

	return val, nil
}

func GetDateQueryParams(r *http.Request, key string) (*time.Time, error) {
	param := r.URL.Query().Get(key)
	if param == "" {
		return nil, nil
	}

	layout := "2006-01-02"
	date, err := time.Parse(layout, param)
	if err != nil {
		return nil, fmt.Errorf("params=`%s` by key=`%s` not a valid date: %v:%w",
			param,
			key,
			err,
			core_errors.ErrInvalidArgument,
		)
	}

	return &date, err
}
