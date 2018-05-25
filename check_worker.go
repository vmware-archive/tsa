package tsa

import (
	"encoding/json"
	"errors"
	"net/http"

	"net/http/httputil"

	"fmt"

	"code.cloudfoundry.org/lager"
	"github.com/concourse/atc"
	"github.com/tedsuo/rata"
)

const ()

type CheckWorker struct {
	ATCEndpoint    *rata.RequestGenerator
	TokenGenerator TokenGenerator
}

func (l *CheckWorker) CheckStatus(logger lager.Logger, worker atc.Worker) error {
	logger.Info("start")
	defer logger.Info("end")

	request, err := l.ATCEndpoint.CreateRequest(atc.ListWorkers, nil, nil)
	if err != nil {
		logger.Error("failed-to-construct-request", err)
		return err
	}

	jwtToken, err := l.TokenGenerator.GenerateSystemToken()
	if err != nil {
		logger.Error("failed-to-construct-request", err)
		return err
	}

	request.Header.Add("Authorization", "Bearer "+jwtToken)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		logger.Error("failed-to-delete", err)
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		logger.Error("bad-response", nil, lager.Data{
			"status-code": response.StatusCode,
		})

		b, _ := httputil.DumpResponse(response, true)
		return fmt.Errorf("bad-response (%d): %s", response.StatusCode, string(b))
	}
	var workersList []atc.Worker
	err = json.NewDecoder(response.Body).Decode(workersList)

	if err != nil {
		logger.Error("failed-to-read-response-body", err)
		return fmt.Errorf("bad-repsonse-body (%d): %s", response.StatusCode, err.Error())
	}

	for _, worker := range workersList {
		if worker.Name == worker.Name {
			return errors.New("worker present in the list")
		}
	}
	return nil
}
