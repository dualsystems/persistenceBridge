package casp

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/persistenceOne/persistenceBridge/application/constants"
	"github.com/persistenceOne/persistenceBridge/application/rest/responses/casp"
	"io/ioutil"
	"net/http"
)

func GetUncompressedPublicKeys() (casp.UncompressedPublicKeysResponse, error) {
	var response casp.UncompressedPublicKeysResponse
	client := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}}

	request, err := http.NewRequest("GET", fmt.Sprintf("%s/casp/api/v1.0/mng/vaults/%s/coins/%s/accounts/0/chains/all/addresses?encoding=uncompressed", constants.CASP_URL, constants.CASP_VAULT_ID, constants.CASP_COIN), nil)

	if err != nil {
		return response, err
	}

	request.Header.Set("authorization", constants.CASP_API_TOKEN)
	resp, err := client.Do(request)
	if err != nil {
		return response, err
	}
	defer resp.Body.Close()
	//Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return response, err
	}
	err = json.Unmarshal(body, &response)
	return response, err
}
