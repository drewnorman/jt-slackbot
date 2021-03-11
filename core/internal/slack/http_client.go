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

// A slack.HttpClient provides alias
// methods for making Slack API requests.
type HttpClient struct {
	logger     *zap.Logger
	apiUrl     string
	appToken   string
	botToken   string
	httpClient *http.Client
}

// slack.HttpClientParameters describe how
// a new slack.HttpClient should be created.
type HttpClientParameters struct {
	Logger     *zap.Logger
	ApiUrl     string
	AppToken   string
	BotToken   string
	HttpClient *http.Client
}

// defaultTimeout specifies the timeout for
// requests in seconds
const defaultTimeout = time.Duration(10) * time.Second

// NewHttpClient returns a new slack.HttpClient
// according to the given parameters.
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

// RequestWssUrl returns a Slack WebSocket server
// URL or an error if the request failed. If
// debugWssReconnects is true, the URL is appended
// with a query parameter indicating to Slack
// that we intend to debug reconnecting to the
// WebSocket server.
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
		client.logger.Debug("debugging wss reconnects")
		wssUrl += "&debug_reconnects=true"
	}
	return wssUrl, nil
}

// JoinChannel makes a request to Slack to
// have the app to join the channel matching
// the given channel ID.
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

// PublicChannels returns an array of all public
// public channels for the workspace.
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

// SendMessageToChannel makes a request to Slack
// to send a given message on behalf of the app
// to the channel matching the given channelId.
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

// post makes a POST request to the Slack API
// with the Slack authorization token and
// returns the decoded response.
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

// get makes a GET request to the Slack API
// with the Slack authorization token and
// returns the decoded response.
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
