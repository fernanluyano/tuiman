package ui

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
	method  string
	name    string
	url     string
	headers []header
	params  []param
	body    string
	auth    requestAuth
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
}
