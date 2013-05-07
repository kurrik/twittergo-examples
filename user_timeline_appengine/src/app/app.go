// Copyright 2013 Arne Roomann-Kurrik
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package app

import (
	"appengine"
	"appengine/datastore"
	"appengine/urlfetch"
	"fmt"
	"github.com/kurrik/oauth1a"
	"github.com/kurrik/twittergo"
	"html"
	"html/template"
	"net/http"
	"net/url"
)

const (
	COUNT       int = 100
	SCREEN_NAME     = "kurrik"
)

const ADMIN_TEMPLATE = `<!DOCTYPE html>
<html>
  <head>
    <style>
    label { display: inline-block; width: 150px; text-align: right; }
    button { margin-left: 155px; }
    </style>
  </head>
  <body>
    <form action="/admin" method="POST">
      <label>Consumer Key</label>
      <input type="text" name="consumer_key" value="{{.ConsumerKey}}"/><br>
      <label>Consumer Secret</label>
      <input type="password" name="consumer_secret" value="{{.ConsumerSecret}}" /><br>
      <label>Access Token</label>
      <input type="text" name="access_token" value="{{.AccessToken}}" /><br>
      <label>Access Secret</label>
      <input type="password" name="access_secret" value="{{.AccessSecret}}" /><br>
      <button>Save</button>
    </form>
    <a href="/">Back to main view</a>
  </body>
</html>`

const ADMIN_SAVED_TEMPLATE = `<!DOCTYPE html>
<html>
  <head>
    <meta http-equiv="refresh" content="2; url=/admin">
  </head>
  <body>
    <p>Saved, redirecting to <code>/admin</code> in 2 seconds.</p>
    <a href="/">Back to main view</a>
  </body>
</html>`

const NOT_CONFIGURED_TEMPLATE = `<!DOCTYPE html>
<html>
  <body>
    <p>
      This example is not configured.  Sign in as an admin
      <a href="/admin">here</a> and input your Twitter app
      credentials.
    </p>
  </body>
</html>`

const TIMELINE_TEMPLATE = `<!DOCTYPE html>
<html>
  <head>
    <style>
      .name { width: 150px; display: inline-block; }
    </style>
  </head>
  <body>
    {{range .}}
      <p><span class="name">@{{.User.ScreenName}}</span> {{.Text | unhtml}}</p>
    {{else}}
      <p>No entries in timeline!</p>
    {{end}}
  </body>
</html>`

type Credentials struct {
	ConsumerKey    string
	ConsumerSecret string
	AccessToken    string
	AccessSecret   string
}

func StoreCredentials(cred *Credentials, ctx appengine.Context) (err error) {
	key := datastore.NewKey(ctx, "Credentials", "main", 0, nil)
	_, err = datastore.Put(ctx, key, cred)
	return
}

func LoadCredentials(ctx appengine.Context) (cred *Credentials, err error) {
	key := datastore.NewKey(ctx, "Credentials", "main", 0, nil)
	cred = &Credentials{}
	err = datastore.Get(ctx, key, cred)
	return
}

func GetTwitterClient(ctx appengine.Context) (c *twittergo.Client, err error) {
	var (
		cred   *Credentials
		config *oauth1a.ClientConfig
		user   *oauth1a.UserConfig
	)
	if cred, err = LoadCredentials(ctx); err != nil {
		return
	}
	if cred.ConsumerKey == "" || cred.ConsumerSecret == "" {
		err = fmt.Errorf("Blank consumer secret and/or key")
		return
	}
	config = &oauth1a.ClientConfig{
		ConsumerKey:    cred.ConsumerKey,
		ConsumerSecret: cred.ConsumerSecret,
	}
	if cred.AccessToken != "" {
		user = oauth1a.NewAuthorizedConfig(cred.AccessToken, cred.AccessSecret)
	}
	c = twittergo.NewClient(config, user)
	c.HttpClient = urlfetch.Client(ctx)
	return
}

func GetTimeline(client *twittergo.Client) (t *twittergo.Timeline, err error) {
	var (
		req   *http.Request
		resp  *twittergo.APIResponse
		rle   twittergo.RateLimitError
		ok    bool
		query url.Values
		endpt string
	)
	query = url.Values{}
	query.Set("count", fmt.Sprintf("%v", COUNT))
	if client.User == nil {
		// With a user token, the user_timeline.json method
		// returns the current user.  Without, you need to specify
		// an explicit ID.
		query.Set("screen_name", SCREEN_NAME)
	}
	endpt = fmt.Sprintf("/1.1/statuses/user_timeline.json?%v", query.Encode())
	if req, err = http.NewRequest("GET", endpt, nil); err != nil {
		return
	}
	if resp, err = client.SendRequest(req); err != nil {
		return
	}
	t = &twittergo.Timeline{}
	if err = resp.Parse(t); err != nil {
		if rle, ok = err.(twittergo.RateLimitError); ok {
			err = fmt.Errorf("Rate limited. Reset at %v", rle.Reset)
		}
	}
	return
}

func RenderTemplate(w http.ResponseWriter, text string, data interface{}) {
	var (
		err   error
		tmpl  *template.Template
		funcs map[string]interface{}
	)
	funcs = map[string]interface{}{
		"unhtml": html.UnescapeString,
	}
	w.Header().Set("Content-Type", "text/html;charset=utf-8")
	tmpl = template.Must(template.New("root").Funcs(funcs).Parse(text))
	if err = tmpl.Execute(w, data); err != nil {
		http.Error(w, "Problem rendering template", 500)
	}
}

func AdminHandler(w http.ResponseWriter, r *http.Request) {
	var (
		cred *Credentials
		ctx  appengine.Context
		err  error
	)
	ctx = appengine.NewContext(r)
	if cred, err = LoadCredentials(ctx); err != nil {
		ctx.Errorf("Couldn't load credentials: %v", err)
		cred = &Credentials{}
	}
	if r.Method == "POST" {
		cred.ConsumerKey = r.FormValue("consumer_key")
		cred.ConsumerSecret = r.FormValue("consumer_secret")
		cred.AccessToken = r.FormValue("access_token")
		cred.AccessSecret = r.FormValue("access_secret")
		if err = StoreCredentials(cred, ctx); err != nil {
			http.Error(w, "Problem storing credentials", 500)
		}
		RenderTemplate(w, ADMIN_SAVED_TEMPLATE, nil)
		return
	}
	RenderTemplate(w, ADMIN_TEMPLATE, cred)
}

func RequestHandler(w http.ResponseWriter, r *http.Request) {
	var (
		ctx    appengine.Context
		client *twittergo.Client
		err    error
		tl     *twittergo.Timeline
	)
	ctx = appengine.NewContext(r)
	if client, err = GetTwitterClient(ctx); err != nil {
		RenderTemplate(w, NOT_CONFIGURED_TEMPLATE, nil)
		return
	}
	if tl, err = GetTimeline(client); err != nil {
		http.Error(w, fmt.Sprintf("Couldn't fetch timeline: %v", err), 500)
		return
	}
	RenderTemplate(w, TIMELINE_TEMPLATE, tl)
}

func init() {
	http.HandleFunc("/admin", AdminHandler)
	http.HandleFunc("/", RequestHandler)
}
