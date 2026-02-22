package ui

import "strings"

type folder struct {
	name     string
	requests []request
}

type header struct {
	key   string
	value string
}

type param struct {
	key   string
	value string
}

type authKind string

const (
	authNone   authKind = "none"
	authBearer authKind = "bearer"
	authBasic  authKind = "basic"
	authAPIKey authKind = "apikey"
)

type requestAuth struct {
	kind     authKind
	token    string // bearer
	username string // basic
	password string // basic
	apiKey   string // apikey
	apiValue string // apikey
}

type request struct {
	method     string
	name       string
	url        string
	headers    []header
	params     []param
	body       string
	auth       requestAuth
	searchable string
}

func (r request) searchText() string {
	parts := []string{
		r.name, r.method, r.url, r.body,
		string(r.auth.kind), r.auth.token, r.auth.username, r.auth.apiKey, r.auth.apiValue,
	}
	for _, h := range r.headers {
		parts = append(parts, h.key, h.value)
	}
	for _, p := range r.params {
		parts = append(parts, p.key, p.value)
	}
	return strings.Join(parts, " ")
}

var httpMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE"}

var mockFolders = []folder{
	{
		name: "Examples",
		requests: []request{
			{
				method: "GET",
				name:   "Httpbin GET",
				url:    "https://httpbin.org/get",
				headers: []header{
					{key: "Accept", value: "application/json"},
				},
				params: []param{
					{key: "foo", value: "bar"},
					{key: "page", value: "1"},
				},
				auth: requestAuth{kind: authNone},
			},
			{
				method: "POST",
				name:   "Httpbin POST",
				url:    "https://httpbin.org/post",
				headers: []header{
					{key: "Content-Type", value: "application/json"},
					{key: "Accept", value: "application/json"},
				},
				body: `{
  "name": "example",
  "value": 42
}`,
				auth: requestAuth{
					kind:  authBearer,
					token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
				},
			},
		},
	},
	{
		name: "GitHub",
		requests: []request{
			{
				method: "GET",
				name:   "Get User",
				url:    "https://api.github.com/users/octocat",
				headers: []header{
					{key: "Accept", value: "application/vnd.github+json"},
					{key: "X-GitHub-Api-Version", value: "2022-11-28"},
				},
				auth: requestAuth{kind: authNone},
			},
			{
				method: "GET",
				name:   "List Repos",
				url:    "https://api.github.com/users/octocat/repos",
				headers: []header{
					{key: "Accept", value: "application/vnd.github+json"},
				},
				params: []param{
					{key: "per_page", value: "30"},
					{key: "sort", value: "updated"},
				},
				auth: requestAuth{kind: authBearer, token: "ghp_xxxxxxxxxxxxxxxxxxxx"},
			},
			{
				method: "POST",
				name:   "Create Issue",
				url:    "https://api.github.com/repos/octocat/Hello-World/issues",
				headers: []header{
					{key: "Accept", value: "application/vnd.github+json"},
					{key: "Content-Type", value: "application/json"},
				},
				body: `{
  "title": "Found a bug",
  "body": "Something is broken.",
  "labels": ["bug"]
}`,
				auth: requestAuth{kind: authBearer, token: "ghp_xxxxxxxxxxxxxxxxxxxx"},
			},
		},
	},
	{
		name: "Stripe",
		requests: []request{
			{
				method: "GET",
				name:   "List Customers",
				url:    "https://api.stripe.com/v1/customers",
				headers: []header{
					{key: "Content-Type", value: "application/x-www-form-urlencoded"},
				},
				params: []param{
					{key: "limit", value: "10"},
				},
				auth: requestAuth{kind: authBearer, token: "sk_test_xxxxxxxxxxxxxxxxxxxx"},
			},
			{
				method: "POST",
				name:   "Create Payment Intent",
				url:    "https://api.stripe.com/v1/payment_intents",
				headers: []header{
					{key: "Content-Type", value: "application/x-www-form-urlencoded"},
				},
				body: `amount=2000&currency=usd&payment_method_types[]=card`,
				auth: requestAuth{kind: authBearer, token: "sk_test_xxxxxxxxxxxxxxxxxxxx"},
			},
			{
				method: "DELETE",
				name:   "Cancel Payment Intent",
				url:    "https://api.stripe.com/v1/payment_intents/pi_xxx/cancel",
				headers: []header{
					{key: "Content-Type", value: "application/x-www-form-urlencoded"},
				},
				auth: requestAuth{kind: authBearer, token: "sk_test_xxxxxxxxxxxxxxxxxxxx"},
			},
		},
	},
}
