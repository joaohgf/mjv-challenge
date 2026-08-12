package mapper

import (
	"github.com/google/uuid"

	"github.com/example/mjv-challenge/api/dto"
	"github.com/example/mjv-challenge/core"
)

func ToJob(request dto.CreateJobRequest) core.Job {
	return core.Job{
		ID:      uuid.NewString(),
		Type:    request.Type,
		Payload: request.Payload,
	}
}
