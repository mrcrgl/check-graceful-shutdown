package transport

import "net/http"

type UserAgent struct {
	Transport http.RoundTripper
	UserAgent string
}

func (ua *UserAgent) RoundTrip(req *http.Request) (*http.Response, error) {
	if len(ua.UserAgent) != 0 {
		req.Header.Add("User-Agent", ua.UserAgent)
	}

	return ua.Transport.RoundTrip(req)
}
