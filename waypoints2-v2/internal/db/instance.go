package db

import (
	"blackbird-eu/waypoints2-v2/internal/types"
	"blackbird-eu/waypoints2-v2/pkg/logger"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type DBHandle struct {
	cfg *Config
}

func NewDBHandler(cfg *Config) (*DBHandle, error) {
	return &DBHandle{
		cfg: cfg,
	}, nil
}

func (h *DBHandle) InitializeScan(APIKey string, targetId any, scanId string, startTime int64) bool {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(60000)*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(
		ctx,
		"PUT",
		fmt.Sprintf(`%v/api/%v/scan/%v/initialize`, h.cfg.BaseURL, h.cfg.Scanner, scanId),
		bytes.NewBuffer([]byte(fmt.Sprintf(`{"startTime": %v}`, startTime))),
	)
	if err != nil {
		logger.Log.Errorf("🔴 Failed to send request %v (%v)", req, err)
		return false
	}

	req.Header = http.Header{
		"X-API-Key":    []string{APIKey},
		"Content-Type": []string{"application/json; charset=utf-8"},
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		logger.Log.Errorf("🔴 Failed to read response %v (%v)", res, err)
		return false
	}

	if res != nil {
		defer res.Body.Close()
		// Read response body
		body, err := io.ReadAll(res.Body)
		if err != nil {
			logger.Log.Errorf("🔴 Failed to read response body (%v)", err)
			return false
		}

		if res.StatusCode == http.StatusOK {
			logger.Log.Debug("🟢 Scan initialized successfully!", string(body))
			return true
		}
	}

	return false
}

func (h *DBHandle) GetTargetURLs(APIKey string, customerId string, targetId any, scanId any, SPIDERXScanId any, vulnerabilityScanId any, chunkId any, IsVulnerabilityScan bool) []string {
	var targetURLs []string = []string{}

	if vulnerabilityScanId != nil {
		retries := 5

		for i := 0; i < retries; i++ {
			if i > 0 {
				time.Sleep(5 * time.Second)
				logger.Log.Errorf("🟠 Failed to retrieve URLsToScan from S3 bucket... Retrying (attempt: %v)", i+1)
			}

			sess, err := session.NewSession(&aws.Config{
				Region: aws.String("eu-west-2"), // Replace with your desired AWS region
			})
			if err != nil {
				logger.Log.Trace("🔴 Error creating AWS session: ", err)
				continue
			}

			svc := s3.New(sess)

			var fileName string = fmt.Sprintf("customers/%v/deep-scans/%v.json", customerId, vulnerabilityScanId)
			if chunkId != nil {
				fileName = fmt.Sprintf("customers/%v/deep-scans/%v.%v.json", customerId, vulnerabilityScanId, chunkId)
			}

			res, err := svc.GetObject(&s3.GetObjectInput{
				Bucket: aws.String("storage-bm92yxnly2lv-lnd"),
				Key:    aws.String(fileName),
			})

			if err != nil {
				continue
			}

			if res != nil {
				defer res.Body.Close()

				body, err := io.ReadAll(res.Body)
				if err != nil {
					logger.Log.Trace("🔴 Failed to read body! err: ", err)
					continue
				}

				var data types.S3DeepScanData
				err = json.Unmarshal(body, &data)
				if err != nil {
					logger.Log.Tracef("🔴 Failed to unmarshal JSON (%v)", err)
					continue
				}

				if data.CustomerId == customerId {
					for _, e := range data.Data {
						parsedURL, _ := url.Parse(e.URL)
						query := parsedURL.Query()

						for _, p := range e.Parameters {
							if !query.Has(p.Key) && strings.ToLower(p.Type) == "query" {
								query.Add(p.Key, p.Value)
							}
						}

						parsedURL.RawQuery = query.Encode()

						targetURLs = append(targetURLs, parsedURL.String())
					}

					return targetURLs
				}
			}

			logger.Log.Tracef("🔴 Failed to upload JSON object to S3. err: %v (%v)", err, res)
		}
	} else {
		if SPIDERXScanId != nil {
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(7500)*time.Millisecond)
			defer cancel()

			req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf(`%v/api/spiderx/results/%v`, h.cfg.BaseURL, SPIDERXScanId), nil)
			if err != nil {
				logger.Log.Errorf("🔴 Failed to send request %v (%v)", req, err)
				return targetURLs
			}

			req.Header = http.Header{
				"X-API-Key":    []string{APIKey},
				"Content-Type": []string{"application/json; charset=utf-8"},
			}

			client := &http.Client{}
			res, err := client.Do(req)
			if err != nil {
				logger.Log.Errorf("🔴 Failed to read response %v (%v)", res, err)
				return targetURLs
			}

			if res == nil {
				return targetURLs
			}

			defer res.Body.Close()
			// Read response body
			body, err := io.ReadAll(res.Body)
			if err != nil {
				logger.Log.Errorf("Failed to read response body (%v)", err)
				return targetURLs
			}

			// Unmarshal JSON
			var data types.SPIDERXAPIResponse
			err = json.Unmarshal(body, &data)
			if err != nil {
				logger.Log.Errorf("Failed to unmarshal JSON (%v)", err)
				return targetURLs
			}

			if res.StatusCode == http.StatusOK {
				if item, ok := data.Data[fmt.Sprintf("%v", SPIDERXScanId)]; ok {
					for _, data := range item.Data {
						if len(data.URL) > 0 {
							targetURLs = append(targetURLs, data.URL)
						}
					}
				}
			}
		} else if (chunkId == nil) && (targetId != nil) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(7500)*time.Millisecond)
			defer cancel()

			req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf(`%v/api/hosts/%v/live`, h.cfg.BaseURL, targetId), nil)
			if err != nil {
				logger.Log.Errorf("🔴 Failed to send request %v (%v)", req, err)
				return targetURLs
			}

			req.Header = http.Header{
				"X-API-Key":    []string{APIKey},
				"Content-Type": []string{"application/json; charset=utf-8"},
				"User-Agent":   []string{"Birdwatch/WAYPOINTS (+https://blackbirdsec.eu/)"},
			}

			client := &http.Client{}
			res, err := client.Do(req)
			if err != nil {
				logger.Log.Errorf("🔴 Failed to read response %v (%v)", res, err)
				return targetURLs
			}

			if res == nil {
				return targetURLs
			}

			defer res.Body.Close()
			// Read response body
			body, err := io.ReadAll(res.Body)
			if err != nil {
				logger.Log.Errorf("🔴 Failed to read response body (%v)", err)
				return targetURLs
			}

			// Unmarshal JSON
			var data types.APIResponse
			err = json.Unmarshal(body, &data)
			if err != nil {
				logger.Log.Errorf("🔴 Failed to unmarshal JSON (%v)", err)
				return targetURLs
			}

			if res.StatusCode == http.StatusOK {
				for _, value := range data.Data {
					if len(value) > 0 {
						targetURLs = append(targetURLs, value)
					}
				}
			}
		} else {
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(7500)*time.Millisecond)
			defer cancel()

			req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf(`%v/api/%v/scan/%v/chunks/%v`, h.cfg.BaseURL, h.cfg.Scanner, scanId, chunkId), nil)
			if err != nil {
				logger.Log.Errorf("🔴 Failed to send request %v (%v)", req, err)
				return targetURLs
			}

			req.Header = http.Header{
				"X-API-Key":    []string{APIKey},
				"Content-Type": []string{"application/json; charset=utf-8"},
			}

			client := &http.Client{}

			res, err := client.Do(req)
			if err != nil {
				logger.Log.Errorf("🔴 Failed to read response %v (%v)", res, err)
				return targetURLs
			}

			if res == nil {
				return targetURLs
			}

			defer res.Body.Close()
			// Read response body
			body, err := io.ReadAll(res.Body)
			if err != nil {
				logger.Log.Errorf("🔴 Failed to read response body (%v)", err)
				return targetURLs
			}

			// Unmarshal JSON
			var data types.APIResponse
			err = json.Unmarshal(body, &data)
			if err != nil {
				logger.Log.Errorf("🔴 Failed to unmarshal JSON (%v)", err)
				return targetURLs
			}

			if res.StatusCode == http.StatusOK {
				for _, value := range data.Data {
					if len(value) > 0 {
						targetURLs = append(targetURLs, value)
					}
				}
			}
		}
	}

	return targetURLs
}
