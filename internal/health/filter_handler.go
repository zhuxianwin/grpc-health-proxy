package health

import (
	"encoding/json"
	"net/http"
)

// NewFilterStatusHandler returns an HTTP handler that reports the filter allow list.
func NewFilterStatusHandler(c Checker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		type response struct {
			AllowAll bool     `json:"allow_all"`
			Services []string `json:"allowed_services"`
		}

		var resp response
		if fc, ok := c.(*FilterChecker); ok {
			allowed := fc.AllowedServices()
			if len(allowed) == 0 {
				resp.AllowAll = true
				resp.Services = []string{}
			} else {
				resp.AllowAll = false
				for svc := range allowed {
					resp.Services = append(resp.Services, svc)
				}
			}
		} else {
			resp.AllowAll = true
			resp.Services = []string{}
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})
}
