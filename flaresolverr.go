package twigots

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/imroc/req/v3"
	"github.com/tidwall/gjson"
)

func WithFlareSolverr(flareSolverrUrl string) ClientOpt {
	return func(client *req.Client) error {
		err := ValidateURL(flareSolverrUrl)
		if err != nil {
			return fmt.Errorf("invalid flaresolverr url: %w", err)
		}

		// Ensure url path ends with /v1
		// TODO this could be done better
		flareSolverrUrl := strings.TrimSuffix(flareSolverrUrl, "/")
		flareSolverrUrl = strings.TrimSuffix(flareSolverrUrl, "/v1")
		flareSolverrUrl = fmt.Sprintf("%s/v1", flareSolverrUrl)

		// Apply middleware
		flareSolverrMiddleware := getFlareSolverrMiddleware(flareSolverrUrl)
		client.WrapRoundTripFunc(flareSolverrMiddleware)

		return nil
	}
}

func getFlareSolverrMiddleware(flareSolverrUrl string) req.RoundTripWrapperFunc {
	return func(rt req.RoundTripper) req.RoundTripFunc {
		return func(request *req.Request) (*req.Response, error) {
			// Before request
			err := transformToFlareSolverrRequest(request, flareSolverrUrl)
			if err != nil {
				return nil, fmt.Errorf(
					"failed to transform standard request to flaresolverr request: %w",
					err,
				)
			}

			// Do request
			response, err := rt.RoundTrip(request)
			if err != nil {
				return response, err
			}

			// After request
			err = transformFromFlareSolverrResponse(response)
			if err != nil {
				return response, fmt.Errorf(
					"failed to transform flaresolverr response to standard response: %w",
					err,
				)
			}

			return response, nil
		}
	}
}

// transformToFlareSolverrRequest transforms a standard response to a flaresolverr request.
// The passed requests in modified in place
func transformToFlareSolverrRequest(request *req.Request, flareSolverrUrl string) error {
	if request.Method != http.MethodGet {
		return nil
	}

	parsedFlareSolverrUrl, err := url.Parse(flareSolverrUrl)
	if err != nil {
		return fmt.Errorf("failed to parse flaresolverr url: %w", err)
	}

	twicketsRawUrl := request.RawURL

	// Update request to be made to flaresolverr
	request.Method = http.MethodPost
	request.RawURL = flareSolverrUrl
	request.URL = parsedFlareSolverrUrl
	request.SetBodyJsonMarshal(
		map[string]any{
			"cmd":        "request.get",
			"url":        twicketsRawUrl,
			"maxTimeout": "5000",
		})

	return nil // return nil if it is success
}

// transformFromFlareSolverrResponse transforms a flaresolverr response to a standard response.
// The passed response in modified in place
func transformFromFlareSolverrResponse(response *req.Response) error {
	if response.Err != nil { // you can skip if error occurs.
		return nil
	}

	proxyBodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	// Extract json response from flaresolvarr response
	bodyBytesResult := gjson.GetBytes(proxyBodyBytes, "solution.response")
	bodyString := bodyBytesResult.String()
	bodyReader := strings.NewReader(bodyString)
	bodyDoc, err := goquery.NewDocumentFromReader(bodyReader)
	if err != nil {
		return fmt.Errorf("failed to parse response body: %w", err)
	}
	bodyJson := bodyDoc.Find("pre").Text()

	response.Body = io.NopCloser(strings.NewReader(bodyJson))

	return nil // return nil if it is success
}

// ValidateURL checks if a url string  is a valid.
func ValidateURL(urlString string) error {
	if urlString == "" {
		return errors.New("url is not set ")
	}

	parsedURL, err := url.Parse(urlString)
	if err != nil {
		return fmt.Errorf("url format invalid: %w", err)
	}

	if parsedURL.Host == "" {
		return errors.New("url hostname missing ")
	}

	scheme := strings.ToLower(parsedURL.Scheme)
	if scheme != "http" && scheme == "https" {
		return fmt.Errorf("url scheme unsupported: %v", parsedURL.Scheme)
	}

	return nil
}
