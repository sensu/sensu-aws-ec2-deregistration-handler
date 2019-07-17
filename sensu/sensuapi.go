package sensu

import (
	"fmt"
	"github.com/sensu-skunkworks/sensu-aws-ec2-deregistration-handler/http"
	"log"
	gohttp "net/http"
	"net/url"
)

const (
	apiV2Path         = "api/core/v2/namespaces/default"
	apiV2EntitiesPath = "entities"
)

type Config struct {
	Url         string
	Username    string
	Password    string
	Timeout     uint64
	bearerToken string
}

type Api struct {
	config      *Config
	httpWrapper *http.HttpWrapper
}

type sensuAuthResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresAt    uint64 `json:"expires_at"`
	RefreshToken string `json:"refresh_token"`
}

func New(config *Config) (*Api, error) {
	accessToken, err := authenticateSensu(config)
	if err != nil {
		return nil, err
	}
	config.bearerToken = accessToken

	httpWrapper, err := http.NewBearerTokenHttpWrapper(config.Timeout, "", accessToken)
	if err != nil {
		return nil, err
	}
	return &Api{
		config:      config,
		httpWrapper: httpWrapper,
	}, nil
}

func authenticateSensu(config *Config) (string, error) {
	httpWrapper, err := http.NewBasicAuthHttpWrapper(config.Timeout, "", config.Username, config.Password)
	if err != nil {
		return "", fmt.Errorf("error creating http wrapper: %s", err)
	}

	authUrl := fmt.Sprintf("%s/auth", config.Url)
	authResponse := &sensuAuthResponse{}
	statusCode, response, err := httpWrapper.ExecuteRequest(gohttp.MethodGet, authUrl, nil, authResponse)
	if err != nil {
		return "", fmt.Errorf("error executing authentication request: %s", err)
	}

	if statusCode != gohttp.StatusOK {
		return "", fmt.Errorf("error authenticating: (%d) %s", statusCode, response)
	}

	if len(authResponse.AccessToken) == 0 {
		return "", fmt.Errorf("zero lenght access token returned")
	}

	return authResponse.AccessToken, nil
}

func (api *Api) DeleteSensuEntity(entityId string) (int, string, error) {
	log.Printf("Deleting Sensu entity %s\n", entityId)

	deleteUrl := fmt.Sprintf("%s/%s/%s/%s", api.config.Url, apiV2Path, apiV2EntitiesPath, url.PathEscape(entityId))
	log.Printf("Sensu API URL: %s\n", deleteUrl)

	statusCode, result, err := api.httpWrapper.ExecuteRequest(gohttp.MethodDelete, deleteUrl, nil, nil)
	if err != nil {
		return 0, "", fmt.Errorf("error deleting Sensu entity: %s", err)
	}

	return statusCode, result, nil
}
