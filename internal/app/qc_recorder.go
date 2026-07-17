package app

import (
	"encoding/json"
	"net/url"

	"grocerics-backend/internal/domain"
	"grocerics-backend/internal/integration/quickcommerce"
	"grocerics-backend/internal/repository"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func qcRecorder(db *gorm.DB) func(quickcommerce.RawCall) {
	repo := repository.NewQCRawResponseRepository(db)
	return func(call quickcommerce.RawCall) {
		row := &domain.QCRawResponse{
			Endpoint:   call.Endpoint,
			Params:     paramsJSON(call.Params),
			StatusCode: call.StatusCode,
			DurationMs: call.DurationMs,
		}
		if call.Err != "" {
			row.Error = &call.Err
		}
		if len(call.Body) > 0 {
			body := string(call.Body)
			if json.Valid(call.Body) {
				row.Response = &body
			} else {
				row.ResponseText = &body
			}
		}
		if err := repo.Create(row); err != nil {
			zap.S().Warnw("qc: raw response insert failed", "endpoint", call.Endpoint, "error", err)
		}
	}
}

func paramsJSON(q url.Values) string {
	m := make(map[string]string, len(q))
	for k := range q {
		m[k] = q.Get(k)
	}
	b, err := json.Marshal(m)
	if err != nil {
		return "{}"
	}
	return string(b)
}
