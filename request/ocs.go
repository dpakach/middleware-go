package request

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"io"
)

var createdUsers = []string{}

type OcsMeta struct {
	Status string `json:"status"`
	StatusCode string `json:"statuscode"`
	Message string `json:"message"`
}

type OcsResponse struct{
	Ocs struct{
		Meta OcsMeta `json:"meta"`
		Data interface{} `json:"data"`
	} `json:"ocs"`
}

type Ocs struct {
	Client http.Client
	Base string
}

type RequestOpts struct {
	method string
	path string
	headers map[string]string
	body string
}


func (o *Ocs) BuildUrl(reqPath string) string {
	u, _ := url.Parse(o.Base)
	u.Path = path.Join(u.Path, "ocs", "v2.php", reqPath)
	return u.String()
}

func (o *Ocs) Request(r RequestOpts) (*http.Response, error) {
	url := o.BuildUrl(r.path)

	fmt.Println(url)
	body := strings.NewReader(r.body)
	req, err := http.NewRequest(r.method, url, body)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
    q.Add("format", "json")
    req.URL.RawQuery = q.Encode()

	for key, val := range r.headers {
		req.Header.Add(key, val)
	}
	res, err := o.Client.Do(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (o *Ocs) CreateUser(username, password, email, displayname string) error {
	data := url.Values{}
    data.Set("userid", username)
    data.Add("password", password)
    data.Add("email", email)
	data.Add("displayname", displayname)

	fmt.Println(data.Encode())

	options := RequestOpts{
		method: "POST",
		path: "/cloud/users",
		headers: map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
			"Authorization": "basic " +  base64.StdEncoding.EncodeToString([]byte("admin:admin")),
		},
		body: data.Encode(),
	}
	res, err := o.Request(options)
	// fmt.Println(res, err)

	if err != nil {
		return errors.New("Failed on http request")
	}

	var body OcsResponse
	if res.StatusCode != 200 {
		defer res.Body.Close()

		buf := new(strings.Builder)
		_, err := io.Copy(buf, res.Body)
		if err != nil {
			return err
		}

		json.Unmarshal([]byte(buf.String()), &body)
		fmt.Println(body)
		return fmt.Errorf("Failed to create User: %s", body.Ocs.Meta.Message)
	}
	createdUsers = append(createdUsers, username)
	return nil
}

func (o *Ocs) Cleanup() error {

	for _, user := range createdUsers {
		options := RequestOpts{
			method: "DELETE",
			path: "/cloud/users/" + user,
			headers: map[string]string{
				"Authorization": "basic " +  base64.StdEncoding.EncodeToString([]byte("admin:admin")),
			},
		}
		res, err := o.Request(options)
	
		if err != nil {
			return errors.New("Failed on http request")
		}
	
		var body OcsResponse
		if res.StatusCode != 200 {
			defer res.Body.Close()
	
			buf := new(strings.Builder)
			_, err := io.Copy(buf, res.Body)
			if err != nil {
				return err
			}
	
			json.Unmarshal([]byte(buf.String()), &body)
			fmt.Println(body)
			return fmt.Errorf("Failed to create User: %s", body.Ocs.Meta.Message)
		}
	}
	createdUsers = []string{}
	return nil
}