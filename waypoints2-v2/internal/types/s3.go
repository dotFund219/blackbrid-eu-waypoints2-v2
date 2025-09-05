package types

type Parameter struct {
	Key         string
	Value       string
	Type        string
	CSProcessed bool
}

type URL struct {
	Method        string      `json:"method"`
	URL           string      `json:"url"`
	StatusCode    int         `json:"statusCode"`
	ContentLength int         `json:"contentLength"`
	ContentType   string      `json:"contentType"`
	Parameters    []Parameter `json:"parameters"`
}

type S3DeepScanData struct {
	CustomerId string `json:"customerId"`
	ScanId     string `json:"scanId"`
	Data       []URL  `json:"data"`
}

type SPIDERXAPIResponse struct {
	Data map[string]struct {
		Data []struct {
			URL string `json:"url"`
		} `json:"data"`
	} `json:"data"`
}

type APIResponse struct {
	Success bool     `json:"success"`
	Data    []string `json:"data"`
}
