// GORSO is a Riot OAuth API wrapper written in pure Go. Provides idiomatic access to RSO API endpoints
// Available at https://github.com/lf-group/gorso
//
// Example:
//   var rso = gorso.Client{
//     ID:       "CLIENT_ID",
//   	 Secret:   "CLIENT_SECRET",
//   	 Redirect: "REDIRECT_URL",
//   }
//
//   func ExampleAuthUser() {
//   	 code := "CLIENT_CODE" // code is obtained on a client side
//
//   	 data, err := rso.GetToken(code)
//   	 if err != nil {
//   	   if errors.Is(err, gorso.ErrSystem) {
//   		   panic(err)
//    		}
//
//   	    return
//    	}
//
//   	 fmt.Println(data.AccessToken)
//   }
package gorso

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// CodeResponse contains tokens to access user private data
type CodeResponse struct {
	// A predefined data scope
	Scope Scope `json:"scope"`
	// Life span of the access token in ms
	ExpiresIn int `json:"expires_in"` // TODO: time.Duration
	// Method of authorization token provides
	TokenType TokenType `json:"token_type"`
	// Issued for the purpose of obtaining new access tokens when an older one expires
	// To reissue an access token, use client.RefreshToken method
	RefreshToken string `json:"refresh_token"`
	// Decryptable JWT Token. Provides information to authenticate a player’s identity
	IDToken string `json:"id_token"`
	// The identifier of an existing session (SID) for the subject (player)
	SubSID string `json:"sub_sid"`
	// Undecryptable JWT Token
	// Used for scoped authentication of a client and player to a resource
	AccessToken string `json:"access_token"`
}

// GetToken returns access&refresh tokens based on a provided code
func (c *Client) GetToken(code string) (*CodeResponse, error) {
	client := http.Client{Timeout: c.getTimeout()}

	formData := url.Values{}
	formData.Add("grant_type", "authorization_code")
	formData.Add("code", code)
	formData.Add("redirect_uri", c.Redirect)

	req, err := http.NewRequest(http.MethodPost, "https://auth.riotgames.com/token", strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, errorCreate(ErrSystem, err)
	}

	c.addAuthHeader(req)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := client.Do(req)
	if err != nil {
		return nil, errorCreate(ErrSystem, err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errorCreate(ErrSystem, err)
	}

	if res.StatusCode != http.StatusOK {
		// TODO: handle errors
		return nil, errorCreate(ErrUnhandled, errors.New("status code not 200"))
	}

	data := CodeResponse{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, errorCreate(ErrSystem, err)
	}

	return &data, nil
}

// RefreshResponse contains a token info to access user private data
type RefreshResponse struct {
	// A predefined data scope
	Scope Scope `json:"scope"`
	// Life span of the access token in ms
	ExpiresIn int `json:"expires_in"` // TODO: time.Duration
	// Method of authorization token provides
	TokenType TokenType `json:"token_type"`
	// Issued for the purpose of obtaining new access tokens when an older one expires
	// To reissue an access token, use client.RefreshToken method
	RefreshToken string `json:"refresh_token"`
	// Decryptable JWT Token. Provides information to authenticate a player’s identity
	IDToken string `json:"id_token"`
	// The identifier of an existing session (SID) for the subject (player)
	SubSID string `json:"sub_sid"`
	// Undecryptable JWT Token
	// Used for scoped authentication of a client and player to a resource
	AccessToken string `json:"access_token"`
}

// RefreshToken returns a new refresh token based on a provided refresh token
func (c *Client) RefreshToken(refreshToken string) (*CodeResponse, error) {
	client := http.Client{Timeout: c.getTimeout()}

	formData := url.Values{}
	formData.Add("grant_type", "refresh_token")
	formData.Add("refresh_token", refreshToken)

	req, err := http.NewRequest(http.MethodPost, "https://auth.riotgames.com/token", strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, errorCreate(ErrSystem, err)
	}

	c.addAuthHeader(req)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := client.Do(req)
	if err != nil {
		return nil, errorCreate(ErrSystem, err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errorCreate(ErrSystem, err)
	}

	if res.StatusCode != http.StatusOK {
		// TODO: handle errors
		return nil, errorCreate(ErrUnhandled, errors.New("status code not 200"))
	}

	data := CodeResponse{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, errorCreate(ErrSystem, err)
	}

	return &data, nil
}

type UserInfoResponse struct {
	Sub string `json:"sub"`
	JTI string `json:"cpid"`
}

// GetUserInfo returns user info based on a provided token
func (c *Client) GetUserInfo(token string) (*UserInfoResponse, error) {
	client := http.Client{Timeout: c.getTimeout()}

	req, err := http.NewRequest(http.MethodGet, "https://auth.riotgames.com/userinfo", nil)
	if err != nil {
		return nil, errorCreate(ErrSystem, err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := client.Do(req)
	if err != nil {
		return nil, errorCreate(ErrSystem, err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errorCreate(ErrSystem, err)
	}

	if res.StatusCode != http.StatusOK {
		// TODO: handle errors
		return nil, errorCreate(ErrUnhandled, errors.New("status code not 200"))
	}

	data := UserInfoResponse{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, errorCreate(ErrSystem, err)
	}

	return &data, nil
}

type AccountResponse struct {
	PUUID    string `json:"puuid"`
	GameName string `json:"gameName"`
	TagLine  string `json:"tagLine"`
}

// GetUserInfo returns user info based on a provided token
func (c *Client) GetAccount(token string) (*AccountResponse, error) {
	client := http.Client{Timeout: c.getTimeout()}

	req, err := http.NewRequest(http.MethodGet, "https://europe.api.riotgames.com/riot/account/v1/accounts/me", nil)
	if err != nil {
		return nil, errorCreate(ErrSystem, err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := client.Do(req)
	if err != nil {
		return nil, errorCreate(ErrSystem, err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errorCreate(ErrSystem, err)
	}

	if res.StatusCode != http.StatusOK {
		// TODO: handle errors
		return nil, errorCreate(ErrUnhandled, errors.New("status code not 200"))
	}

	data := AccountResponse{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, errorCreate(ErrSystem, err)
	}

	return &data, nil
}
