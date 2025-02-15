package services

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type result struct {
	output string
	state  CheckState
}

func (s *Service) validate() error { //nolint:cyclop
	s.state = StateUnknown

	if s.Name == "" {
		return fmt.Errorf("%s: %w", s.Value, ErrNoName)
	} else if s.Value == "" {
		return fmt.Errorf("%s: %w", s.Name, ErrNoCheck)
	}

	switch s.Type {
	case CheckHTTP:
		if s.Expect == "" {
			s.Expect = "200"
		}
	case CheckTCP:
		if !strings.Contains(s.Value, ":") {
			return ErrBadTCP
		}
	case CheckPROC:
		if err := s.checkProcValues(); err != nil {
			return err
		}
	case CheckPING:
	default:
		return ErrInvalidType
	}

	if s.Timeout.Duration == 0 {
		s.Timeout.Duration = DefaultTimeout
	} else if s.Timeout.Duration < MinimumTimeout {
		s.Timeout.Duration = MinimumTimeout
	}

	if s.Interval.Duration == 0 {
		s.Interval.Duration = DefaultCheckInterval
	} else if s.Interval.Duration < MinimumCheckInterval {
		s.Interval.Duration = MinimumCheckInterval
	}

	return nil
}

func (s *Service) check() bool {
	// check this service.
	switch s.Type {
	case CheckHTTP:
		return s.update(s.checkHTTP())
	case CheckTCP:
		return s.update(s.checkTCP())
	case CheckPING:
		return s.update(s.checkPING())
	case CheckPROC:
		return s.update(s.checkProccess())
	default:
		return false
	}
}

// Return true if the service state changed.
func (s *Service) update(res *result) bool {
	if s.lastCheck = time.Now().Round(time.Microsecond); s.since.IsZero() {
		s.since = s.lastCheck
	}

	s.output = res.output

	if s.state == res.state {
		s.log.Printf("Service Checked: %s, state: %s for %v, output: %s",
			s.Name, s.state, time.Since(s.since).Round(time.Second), s.output)
		return false
	}

	s.log.Printf("Service Checked: %s, state: %s ~> %s, output: %s", s.Name, s.state, res.state, s.output)
	s.since = s.lastCheck
	s.state = res.state

	return true
}

const maxBody = 150

func (s *Service) checkHTTP() *result {
	res := &result{
		state:  StateUnknown,
		output: "unknown",
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.Timeout.Duration)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.Value, nil)
	if err != nil {
		res.output = "creating request: " + RemoveSecrets(s.Value, err.Error())
		return res
	}

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		res.output = "making request: " + RemoveSecrets(s.Value, err.Error())
		return res
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		res.output = "reading body: " + RemoveSecrets(s.Value, err.Error())
		return res
	}

	if strconv.Itoa(resp.StatusCode) == s.Expect {
		res.state = StateOK
		res.output = resp.Status

		return res
	}

	bodyStr := string(body)
	if len(bodyStr) > maxBody {
		bodyStr = bodyStr[:maxBody]
	}

	res.state = StateCritical
	res.output = resp.Status + ": " + strings.TrimSpace(RemoveSecrets(s.Value, bodyStr))

	return res
}

// RemoveSecrets removes secret token values in a message parsed from a url.
func RemoveSecrets(appURL, message string) string {
	url, err := url.Parse(appURL)
	if err != nil {
		return message
	}

	for _, keyName := range []string{"apikey", "token", "pass", "password", "secret"} {
		if secret := url.Query().Get(keyName); secret != "" {
			message = strings.ReplaceAll(message, secret, "********")
		}
	}

	return message
}

func (s *Service) checkTCP() *result {
	res := &result{
		state:  StateUnknown,
		output: "unknown",
	}

	switch conn, err := net.DialTimeout("tcp", s.Value, s.Timeout.Duration); {
	case err != nil:
		res.state = StateCritical
		res.output = "connection error: " + err.Error()
	case conn == nil:
		res.state = StateUnknown
		res.output = "connection failed, no specific error"
	default:
		conn.Close()

		res.state = StateOK
		res.output = "connected to port " + strings.Split(s.Value, ":")[1] + " OK"
	}

	return res
}

func (s *Service) checkPING() *result {
	return &result{
		state:  StateUnknown,
		output: "ping does not work yet",
	}
}
