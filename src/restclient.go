package src

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/DiscoOrg/disgo/src/endpoints"
)

// RestClient is the client used for HTTP requests to discord
type RestClient struct {
	Client    *http.Client
	Token     string
	UserAgent string
}

type ErrorResponse struct {
	Code    int
	Message string
}

// Request makes a new rest request to discords api with the specific route
// route the route to make a request to
// rqBody the request body if one should be sent
// v the struct which the response should be unmarshalled in
// args path params to compile the route
func (c RestClient) Request(route endpoints.Route, rqBody interface{}, v interface{}, args ...interface{}) error {

	var reader io.Reader
	if rqBody != nil {
		rqJSON, err := json.Marshal(rqBody)
		if err != nil {
			return err
		}
		reader = bytes.NewBuffer(rqJSON)
	} else {
		reader = nil
	}

	rq, err := http.NewRequest(route.Method.String(), route.Compile(args...), reader)
	if err != nil {
		return err
	}

	rq.Header.Set("User-Agent", c.UserAgent)
	rq.Header.Set("Authorization", "Bot "+c.Token)
	rq.Header.Set("Content-Type", "application/json")

	rs, err := c.Client.Do(rq)
	if err != nil {
		return err
	}

	defer func() {
		err := rs.Body.Close()
		if err != nil {
			log.Error("error closing response body", err.Error())
		}
	}()

	rsBody, err := ioutil.ReadAll(rs.Body)
	if err != nil {
		return err
	}

	switch rs.StatusCode {
	case http.StatusOK:
	case http.StatusCreated:
	case http.StatusNoContent:
		if err = json.Unmarshal(rsBody, v); err != nil {
			return err
		}

	case http.StatusTooManyRequests:
		limit := rs.Header.Get("X-RateLimit-Limit")
		remaining := rs.Header.Get("X-RateLimit-Limit")
		reset := rs.Header.Get("X-RateLimit-Limit")
		bucket := rs.Header.Get("X-RateLimit-Limit")
		return fmt.Errorf("too many requests. limit: %s, remaining: %s, reset: %s,bucket: %s", limit, remaining, reset, bucket)

	case http.StatusBadGateway:
		return errors.New("bad gateway could not reach discord")

	case http.StatusUnauthorized:
		return errors.New("the provided token is invalid")

	default:
		var errorRs ErrorResponse
		if err = json.Unmarshal(rsBody, &errorRs); err != nil {
			return err
		}
		return fmt.Errorf("request to %s failed. statuscode: %d, errorcode: %d, message: %s", rq.URL, rs.StatusCode, errorRs.Code, errorRs.Message)
	}

	return nil
}
