package railway

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const apiURL = "https://backboard.railway.com/graphql/v2"

type Client struct {
	httpClient *http.Client
}

func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	return &Client{httpClient: httpClient}
}

type graphQLRequest struct {
	Query string `json:"query"`
}

type workspacesResponse struct {
	Data struct {
		Me struct {
			Workspaces []Workspace `json:"workspaces"`
		} `json:"me"`
	} `json:"data"`
	Errors []graphQLError `json:"errors,omitempty"`
}

type Workspace struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type graphQLError struct {
	Message string `json:"message"`
}

type UserInfo struct {
	Sub     string `json:"sub"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

func (c *Client) FetchUserInfo(accessToken string) (*UserInfo, error) {
	req, err := http.NewRequest("GET", "https://backboard.railway.com/oauth/me", nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var userInfo UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &userInfo, nil
}

func (c *Client) FetchUserWorkspaces(accessToken string) ([]Workspace, error) {
	query := `query { me { workspaces { id name } } }`

	body := graphQLRequest{Query: query}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var result workspacesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if len(result.Errors) > 0 {
		return nil, fmt.Errorf("graphql error: %s", result.Errors[0].Message)
	}

	return result.Data.Me.Workspaces, nil
}

func (c *Client) UserHasWorkspaceAccess(accessToken, workspaceID string) (bool, error) {
	workspaces, err := c.FetchUserWorkspaces(accessToken)
	if err != nil {
		return false, fmt.Errorf("fetch workspaces: %w", err)
	}

	for _, ws := range workspaces {
		if ws.ID == workspaceID {
			return true, nil
		}
	}

	return false, nil
}
