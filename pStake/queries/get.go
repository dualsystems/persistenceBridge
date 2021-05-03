package queries

import (
	"encoding/json"
	"github.com/persistenceOne/persistenceCore/pStake/constants"
	"io"
	"net/http"
	"time"
)

var httpClient = &http.Client{Timeout: constants.Timeout * time.Second}

func get(url string, target interface{}) error {
	r, err := httpClient.Get(url)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(r.Body)

	if err := json.NewDecoder(r.Body).Decode(target); err == io.EOF {
		return nil
	} else if err != nil {
		return err
	}

	return err
}
