package slack

import (
	"encoding/json"
	"errors"
	"go.uber.org/zap"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type HttpClient struct {
	logger     *zap.Logger
	apiUrl     string
	appToken   string
	botToken   string
	httpClient *http.Client
}

type HttpClientParameters struct {
	Logger     *zap.Logger
	ApiUrl     string
	AppToken   string
	BotToken   string
	HttpClient *http.Client
}

const defaultTimeout = time.Duration(10) * time.Second

func NewHttpClient(
	params *HttpClientParameters,
) (*HttpClient, error) {
	if params.Logger == nil {
		return nil, errors.New("missing logger")
	}
	if params.ApiUrl == "" {
		return nil, errors.New("missing api url")
	}
	if params.AppToken == "" {
		return nil, errors.New("missing app token")
	}
	if params.BotToken == "" {
		return nil, errors.New("missing bot token")
	}

	var httpClient *http.Client
	if params.HttpClient != nil {
		httpClient = params.HttpClient
	} else {
		httpClient = &http.Client{
			Timeout: defaultTimeout,
		}
	}

	return &HttpClient{
		logger:     params.Logger,
		apiUrl:     params.ApiUrl,
		appToken:   params.AppToken,
		botToken:   params.BotToken,
		httpClient: httpClient,
	}, nil
}

func (client *HttpClient) RequestWssUrl(
	debugWssReconnects bool,
) (string, error) {
	data, err := client.post(
		client.appToken,
		"apps.connections.open",
		map[string]string{},
	)
	if err != nil {
		return "", err
	}
	wssUrl, ok := data["url"].(string)
	if !ok {
		return "", errors.New("no url in response")
	}
	if debugWssReconnects {
		wssUrl += "&debug_reconnects=true"
	}
	return wssUrl, nil
}

func (client *HttpClient) JoinChannel(
	channelId string,
) error {
	if channelId == "" {
		return errors.New("missing channel id")
	}
	data, err := client.post(
		client.botToken,
		"conversations.join",
		map[string]string{
			"channel": channelId,
		},
	)
	if err != nil {
		return err
	}
	success, ok := data["ok"].(bool)
	if !ok || !success {
		return errors.New(data["error"].(string))
	}
	return nil
}

func (client *HttpClient) PublicChannels() ([]interface{}, error) {
	data, err := client.get(
		client.botToken,
		"conversations.list",
		map[string]string{
			"exclude_archived": "true",
			"types":            "public_channel",
		},
	)
	if err != nil {
		return nil, err
	}

	channels, ok := data["channels"].([]interface{})
	if !ok {
		return nil, errors.New("error requesting channels")
	}
	return channels, nil
}

func (client *HttpClient) SendMessageToChannel(
	message string,
	channelId string,
) error {
	if message == "" {
		return errors.New("missing message")
	}
	if channelId == "" {
		return errors.New("missing channel id")
	}
	data, err := client.post(
		client.botToken,
		"chat.postMessage",
		map[string]string{
			"text":    message,
			"channel": channelId,
		},
	)
	if err != nil {
		return err
	}
	success, ok := data["ok"].(bool)
	if !ok || !success {
		if data["error"] != nil {
			return errors.New(data["error"].(string))
		}
		return errors.New("failed to send message")
	}
	return nil
}

func (client *HttpClient) post(
	token string,
	endpoint string,
	params map[string]string,
) (map[string]interface{}, error) {
	values := url.Values{}
	for key, value := range params {
		values.Add(key, value)
	}
	req, err := http.NewRequest(
		"POST",
		client.apiUrl+endpoint,
		strings.NewReader(values.Encode()),
	)
	if err != nil {
		return nil, errors.New("failed to init request")
	}
	req.Header.Add("Authorization", "Bearer "+token)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return nil, errors.New("failed to make request")
	}
	defer resp.Body.Close()

	decoded := new(map[string]interface{})
	err = json.NewDecoder(resp.Body).Decode(decoded)
	if err != nil {
		return nil, errors.New("failed to decode response")
	}
	return *decoded, nil
}

func (client *HttpClient) get(
	token string,
	endpoint string,
	params map[string]string,
) (map[string]interface{}, error) {
	req, err := http.NewRequest(
		"GET",
		client.apiUrl+endpoint,
		nil,
	)
	if err != nil {
		return nil, errors.New("failed to init request")
	}
	req.Header.Add("Authorization", "Bearer "+token)
	query := req.URL.Query()
	for key, value := range params {
		query.Add(key, value)
	}
	req.URL.RawQuery = query.Encode()

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return nil, errors.New("failed to make request")
	}
	defer resp.Body.Close()

	decoded := new(map[string]interface{})
	err = json.NewDecoder(resp.Body).Decode(decoded)
	if err != nil {
		return nil, errors.New("failed to decode response")
	}
	return *decoded, nil
}
