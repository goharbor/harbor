package sling

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

type FakeParams struct {
	KindName string `url:"kind_name"`
	Count    int    `url:"count"`
}

// Url-tagged query struct
var paramsA = struct {
	Limit int `url:"limit"`
}{
	30,
}
var paramsB = FakeParams{KindName: "recent", Count: 25}

// Json-tagged model struct
type FakeModel struct {
	Text          string  `json:"text,omitempty"`
	FavoriteCount int64   `json:"favorite_count,omitempty"`
	Temperature   float64 `json:"temperature,omitempty"`
}

var modelA = FakeModel{Text: "note", FavoriteCount: 12}

func TestNew(t *testing.T) {
	sling := New()
	if sling.httpClient != http.DefaultClient {
		t.Errorf("expected %v, got %v", http.DefaultClient, sling.httpClient)
	}
	if sling.header == nil {
		t.Errorf("Header map not initialized with make")
	}
	if sling.queryStructs == nil {
		t.Errorf("queryStructs not initialized with make")
	}
}

func TestSlingNew(t *testing.T) {
	fakeBodyProvider := jsonBodyProvider{FakeModel{}}

	cases := []*Sling{
		&Sling{httpClient: &http.Client{}, method: "GET", rawURL: "http://example.com"},
		&Sling{httpClient: nil, method: "", rawURL: "http://example.com"},
		&Sling{queryStructs: make([]interface{}, 0)},
		&Sling{queryStructs: []interface{}{paramsA}},
		&Sling{queryStructs: []interface{}{paramsA, paramsB}},
		&Sling{bodyProvider: fakeBodyProvider},
		&Sling{bodyProvider: fakeBodyProvider},
		&Sling{bodyProvider: nil},
		New().Add("Content-Type", "application/json"),
		New().Add("A", "B").Add("a", "c").New(),
		New().Add("A", "B").New().Add("a", "c"),
		New().BodyForm(paramsB),
		New().BodyForm(paramsB).New(),
	}
	for _, sling := range cases {
		child := sling.New()
		if child.httpClient != sling.httpClient {
			t.Errorf("expected %v, got %v", sling.httpClient, child.httpClient)
		}
		if child.method != sling.method {
			t.Errorf("expected %s, got %s", sling.method, child.method)
		}
		if child.rawURL != sling.rawURL {
			t.Errorf("expected %s, got %s", sling.rawURL, child.rawURL)
		}
		// Header should be a copy of parent Sling header. For example, calling
		// baseSling.Add("k","v") should not mutate previously created child Slings
		if sling.header != nil {
			// struct literal cases don't init Header in usual way, skip header check
			if !reflect.DeepEqual(sling.header, child.header) {
				t.Errorf("not DeepEqual: expected %v, got %v", sling.header, child.header)
			}
			sling.header.Add("K", "V")
			if child.header.Get("K") != "" {
				t.Errorf("child.header was a reference to original map, should be copy")
			}
		}
		// queryStruct slice should be a new slice with a copy of the contents
		if len(sling.queryStructs) > 0 {
			// mutating one slice should not mutate the other
			child.queryStructs[0] = nil
			if sling.queryStructs[0] == nil {
				t.Errorf("child.queryStructs was a re-slice, expected slice with copied contents")
			}
		}
		// body should be copied
		if child.bodyProvider != sling.bodyProvider {
			t.Errorf("expected %v, got %v", sling.bodyProvider, child.bodyProvider)
		}
	}
}

func TestClientSetter(t *testing.T) {
	developerClient := &http.Client{}
	cases := []struct {
		input    *http.Client
		expected *http.Client
	}{
		{nil, http.DefaultClient},
		{developerClient, developerClient},
	}
	for _, c := range cases {
		sling := New()
		sling.Client(c.input)
		if sling.httpClient != c.expected {
			t.Errorf("input %v, expected %v, got %v", c.input, c.expected, sling.httpClient)
		}
	}
}

func TestDoerSetter(t *testing.T) {
	developerClient := &http.Client{}
	cases := []struct {
		input    Doer
		expected Doer
	}{
		{nil, http.DefaultClient},
		{developerClient, developerClient},
	}
	for _, c := range cases {
		sling := New()
		sling.Doer(c.input)
		if sling.httpClient != c.expected {
			t.Errorf("input %v, expected %v, got %v", c.input, c.expected, sling.httpClient)
		}
	}
}

func TestBaseSetter(t *testing.T) {
	cases := []string{"http://a.io/", "http://b.io", "/path", "path", ""}
	for _, base := range cases {
		sling := New().Base(base)
		if sling.rawURL != base {
			t.Errorf("expected %s, got %s", base, sling.rawURL)
		}
	}
}

func TestPathSetter(t *testing.T) {
	cases := []struct {
		rawURL         string
		path           string
		expectedRawURL string
	}{
		{"http://a.io/", "foo", "http://a.io/foo"},
		{"http://a.io/", "/foo", "http://a.io/foo"},
		{"http://a.io", "foo", "http://a.io/foo"},
		{"http://a.io", "/foo", "http://a.io/foo"},
		{"http://a.io/foo/", "bar", "http://a.io/foo/bar"},
		// rawURL should end in trailing slash if it is to be Path extended
		{"http://a.io/foo", "bar", "http://a.io/bar"},
		{"http://a.io/foo", "/bar", "http://a.io/bar"},
		// path extension is absolute
		{"http://a.io", "http://b.io/", "http://b.io/"},
		{"http://a.io/", "http://b.io/", "http://b.io/"},
		{"http://a.io", "http://b.io", "http://b.io"},
		{"http://a.io/", "http://b.io", "http://b.io"},
		// empty base, empty path
		{"", "http://b.io", "http://b.io"},
		{"http://a.io", "", "http://a.io"},
		{"", "", ""},
	}
	for _, c := range cases {
		sling := New().Base(c.rawURL).Path(c.path)
		if sling.rawURL != c.expectedRawURL {
			t.Errorf("expected %s, got %s", c.expectedRawURL, sling.rawURL)
		}
	}
}

func TestMethodSetters(t *testing.T) {
	cases := []struct {
		sling          *Sling
		expectedMethod string
	}{
		{New().Path("http://a.io"), "GET"},
		{New().Head("http://a.io"), "HEAD"},
		{New().Get("http://a.io"), "GET"},
		{New().Post("http://a.io"), "POST"},
		{New().Put("http://a.io"), "PUT"},
		{New().Patch("http://a.io"), "PATCH"},
		{New().Delete("http://a.io"), "DELETE"},
	}
	for _, c := range cases {
		if c.sling.method != c.expectedMethod {
			t.Errorf("expected method %s, got %s", c.expectedMethod, c.sling.method)
		}
	}
}

func TestAddHeader(t *testing.T) {
	cases := []struct {
		sling          *Sling
		expectedHeader map[string][]string
	}{
		{New().Add("authorization", "OAuth key=\"value\""), map[string][]string{"Authorization": []string{"OAuth key=\"value\""}}},
		// header keys should be canonicalized
		{New().Add("content-tYPE", "application/json").Add("User-AGENT", "sling"), map[string][]string{"Content-Type": []string{"application/json"}, "User-Agent": []string{"sling"}}},
		// values for existing keys should be appended
		{New().Add("A", "B").Add("a", "c"), map[string][]string{"A": []string{"B", "c"}}},
		// Add should add to values for keys added by parent Slings
		{New().Add("A", "B").Add("a", "c").New(), map[string][]string{"A": []string{"B", "c"}}},
		{New().Add("A", "B").New().Add("a", "c"), map[string][]string{"A": []string{"B", "c"}}},
	}
	for _, c := range cases {
		// type conversion from header to alias'd map for deep equality comparison
		headerMap := map[string][]string(c.sling.header)
		if !reflect.DeepEqual(c.expectedHeader, headerMap) {
			t.Errorf("not DeepEqual: expected %v, got %v", c.expectedHeader, headerMap)
		}
	}
}

func TestSetHeader(t *testing.T) {
	cases := []struct {
		sling          *Sling
		expectedHeader map[string][]string
	}{
		// should replace existing values associated with key
		{New().Add("A", "B").Set("a", "c"), map[string][]string{"A": []string{"c"}}},
		{New().Set("content-type", "A").Set("Content-Type", "B"), map[string][]string{"Content-Type": []string{"B"}}},
		// Set should replace values received by copying parent Slings
		{New().Set("A", "B").Add("a", "c").New(), map[string][]string{"A": []string{"B", "c"}}},
		{New().Add("A", "B").New().Set("a", "c"), map[string][]string{"A": []string{"c"}}},
	}
	for _, c := range cases {
		// type conversion from Header to alias'd map for deep equality comparison
		headerMap := map[string][]string(c.sling.header)
		if !reflect.DeepEqual(c.expectedHeader, headerMap) {
			t.Errorf("not DeepEqual: expected %v, got %v", c.expectedHeader, headerMap)
		}
	}
}

func TestBasicAuth(t *testing.T) {
	cases := []struct {
		sling        *Sling
		expectedAuth []string
	}{
		// basic auth: username & password
		{New().SetBasicAuth("Aladdin", "open sesame"), []string{"Aladdin", "open sesame"}},
		// empty username
		{New().SetBasicAuth("", "secret"), []string{"", "secret"}},
		// empty password
		{New().SetBasicAuth("admin", ""), []string{"admin", ""}},
	}
	for _, c := range cases {
		req, err := c.sling.Request()
		if err != nil {
			t.Errorf("unexpected error when building Request with .SetBasicAuth()")
		}
		username, password, ok := req.BasicAuth()
		if !ok {
			t.Errorf("basic auth missing when expected")
		}
		auth := []string{username, password}
		if !reflect.DeepEqual(c.expectedAuth, auth) {
			t.Errorf("not DeepEqual: expected %v, got %v", c.expectedAuth, auth)
		}
	}
}

func TestQueryStructSetter(t *testing.T) {
	cases := []struct {
		sling           *Sling
		expectedStructs []interface{}
	}{
		{New(), []interface{}{}},
		{New().QueryStruct(nil), []interface{}{}},
		{New().QueryStruct(paramsA), []interface{}{paramsA}},
		{New().QueryStruct(paramsA).QueryStruct(paramsA), []interface{}{paramsA, paramsA}},
		{New().QueryStruct(paramsA).QueryStruct(paramsB), []interface{}{paramsA, paramsB}},
		{New().QueryStruct(paramsA).New(), []interface{}{paramsA}},
		{New().QueryStruct(paramsA).New().QueryStruct(paramsB), []interface{}{paramsA, paramsB}},
	}

	for _, c := range cases {
		if count := len(c.sling.queryStructs); count != len(c.expectedStructs) {
			t.Errorf("expected length %d, got %d", len(c.expectedStructs), count)
		}
	check:
		for _, expected := range c.expectedStructs {
			for _, param := range c.sling.queryStructs {
				if param == expected {
					continue check
				}
			}
			t.Errorf("expected to find %v in %v", expected, c.sling.queryStructs)
		}
	}
}

func TestBodyJSONSetter(t *testing.T) {
	fakeModel := &FakeModel{}
	fakeBodyProvider := jsonBodyProvider{payload: fakeModel}

	cases := []struct {
		initial  BodyProvider
		input    interface{}
		expected BodyProvider
	}{
		// json tagged struct is set as bodyJSON
		{nil, fakeModel, fakeBodyProvider},
		// nil argument to bodyJSON does not replace existing bodyJSON
		{fakeBodyProvider, nil, fakeBodyProvider},
		// nil bodyJSON remains nil
		{nil, nil, nil},
	}
	for _, c := range cases {
		sling := New()
		sling.bodyProvider = c.initial
		sling.BodyJSON(c.input)
		if sling.bodyProvider != c.expected {
			t.Errorf("expected %v, got %v", c.expected, sling.bodyProvider)
		}
		// Header Content-Type should be application/json if bodyJSON arg was non-nil
		if c.input != nil && sling.header.Get(contentType) != jsonContentType {
			t.Errorf("Incorrect or missing header, expected %s, got %s", jsonContentType, sling.header.Get(contentType))
		} else if c.input == nil && sling.header.Get(contentType) != "" {
			t.Errorf("did not expect a Content-Type header, got %s", sling.header.Get(contentType))
		}
	}
}

func TestBodyFormSetter(t *testing.T) {
	fakeParams := FakeParams{KindName: "recent", Count: 25}
	fakeBodyProvider := formBodyProvider{payload: fakeParams}

	cases := []struct {
		initial  BodyProvider
		input    interface{}
		expected BodyProvider
	}{
		// url tagged struct is set as bodyStruct
		{nil, paramsB, fakeBodyProvider},
		// nil argument to bodyStruct does not replace existing bodyStruct
		{fakeBodyProvider, nil, fakeBodyProvider},
		// nil bodyStruct remains nil
		{nil, nil, nil},
	}
	for _, c := range cases {
		sling := New()
		sling.bodyProvider = c.initial
		sling.BodyForm(c.input)
		if sling.bodyProvider != c.expected {
			t.Errorf("expected %v, got %v", c.expected, sling.bodyProvider)
		}
		// Content-Type should be application/x-www-form-urlencoded if bodyStruct was non-nil
		if c.input != nil && sling.header.Get(contentType) != formContentType {
			t.Errorf("Incorrect or missing header, expected %s, got %s", formContentType, sling.header.Get(contentType))
		} else if c.input == nil && sling.header.Get(contentType) != "" {
			t.Errorf("did not expect a Content-Type header, got %s", sling.header.Get(contentType))
		}
	}
}

func TestBodySetter(t *testing.T) {
	fakeInput := ioutil.NopCloser(strings.NewReader("test"))
	fakeBodyProvider := bodyProvider{body: fakeInput}

	cases := []struct {
		initial  BodyProvider
		input    io.Reader
		expected BodyProvider
	}{
		// nil body is overriden by a set body
		{nil, fakeInput, fakeBodyProvider},
		// initial body is not overriden by nil body
		{fakeBodyProvider, nil, fakeBodyProvider},
		// nil body is returned unaltered
		{nil, nil, nil},
	}
	for _, c := range cases {
		sling := New()
		sling.bodyProvider = c.initial
		sling.Body(c.input)
		if sling.bodyProvider != c.expected {
			t.Errorf("expected %v, got %v", c.expected, sling.bodyProvider)
		}
	}
}

func TestRequest_urlAndMethod(t *testing.T) {
	cases := []struct {
		sling          *Sling
		expectedMethod string
		expectedURL    string
		expectedErr    error
	}{
		{New().Base("http://a.io"), "GET", "http://a.io", nil},
		{New().Path("http://a.io"), "GET", "http://a.io", nil},
		{New().Get("http://a.io"), "GET", "http://a.io", nil},
		{New().Put("http://a.io"), "PUT", "http://a.io", nil},
		{New().Base("http://a.io/").Path("foo"), "GET", "http://a.io/foo", nil},
		{New().Base("http://a.io/").Post("foo"), "POST", "http://a.io/foo", nil},
		// if relative path is an absolute url, base is ignored
		{New().Base("http://a.io").Path("http://b.io"), "GET", "http://b.io", nil},
		{New().Path("http://a.io").Path("http://b.io"), "GET", "http://b.io", nil},
		// last method setter takes priority
		{New().Get("http://b.io").Post("http://a.io"), "POST", "http://a.io", nil},
		{New().Post("http://a.io/").Put("foo/").Delete("bar"), "DELETE", "http://a.io/foo/bar", nil},
		// last Base setter takes priority
		{New().Base("http://a.io").Base("http://b.io"), "GET", "http://b.io", nil},
		// Path setters are additive
		{New().Base("http://a.io/").Path("foo/").Path("bar"), "GET", "http://a.io/foo/bar", nil},
		{New().Path("http://a.io/").Path("foo/").Path("bar"), "GET", "http://a.io/foo/bar", nil},
		// removes extra '/' between base and ref url
		{New().Base("http://a.io/").Get("/foo"), "GET", "http://a.io/foo", nil},
	}
	for _, c := range cases {
		req, err := c.sling.Request()
		if err != c.expectedErr {
			t.Errorf("expected error %v, got %v for %+v", c.expectedErr, err, c.sling)
		}
		if req.URL.String() != c.expectedURL {
			t.Errorf("expected url %s, got %s for %+v", c.expectedURL, req.URL.String(), c.sling)
		}
		if req.Method != c.expectedMethod {
			t.Errorf("expected method %s, got %s for %+v", c.expectedMethod, req.Method, c.sling)
		}
	}
}

func TestRequest_queryStructs(t *testing.T) {
	cases := []struct {
		sling       *Sling
		expectedURL string
	}{
		{New().Base("http://a.io").QueryStruct(paramsA), "http://a.io?limit=30"},
		{New().Base("http://a.io").QueryStruct(paramsA).QueryStruct(paramsB), "http://a.io?count=25&kind_name=recent&limit=30"},
		{New().Base("http://a.io/").Path("foo?path=yes").QueryStruct(paramsA), "http://a.io/foo?limit=30&path=yes"},
		{New().Base("http://a.io").QueryStruct(paramsA).New(), "http://a.io?limit=30"},
		{New().Base("http://a.io").QueryStruct(paramsA).New().QueryStruct(paramsB), "http://a.io?count=25&kind_name=recent&limit=30"},
	}
	for _, c := range cases {
		req, _ := c.sling.Request()
		if req.URL.String() != c.expectedURL {
			t.Errorf("expected url %s, got %s for %+v", c.expectedURL, req.URL.String(), c.sling)
		}
	}
}

func TestRequest_body(t *testing.T) {
	cases := []struct {
		sling               *Sling
		expectedBody        string // expected Body io.Reader as a string
		expectedContentType string
	}{
		// BodyJSON
		{New().BodyJSON(modelA), "{\"text\":\"note\",\"favorite_count\":12}\n", jsonContentType},
		{New().BodyJSON(&modelA), "{\"text\":\"note\",\"favorite_count\":12}\n", jsonContentType},
		{New().BodyJSON(&FakeModel{}), "{}\n", jsonContentType},
		{New().BodyJSON(FakeModel{}), "{}\n", jsonContentType},
		// BodyJSON overrides existing values
		{New().BodyJSON(&FakeModel{}).BodyJSON(&FakeModel{Text: "msg"}), "{\"text\":\"msg\"}\n", jsonContentType},
		// BodyForm
		{New().BodyForm(paramsA), "limit=30", formContentType},
		{New().BodyForm(paramsB), "count=25&kind_name=recent", formContentType},
		{New().BodyForm(&paramsB), "count=25&kind_name=recent", formContentType},
		// BodyForm overrides existing values
		{New().BodyForm(paramsA).New().BodyForm(paramsB), "count=25&kind_name=recent", formContentType},
		// Mixture of BodyJSON and BodyForm prefers body setter called last with a non-nil argument
		{New().BodyForm(paramsB).New().BodyJSON(modelA), "{\"text\":\"note\",\"favorite_count\":12}\n", jsonContentType},
		{New().BodyJSON(modelA).New().BodyForm(paramsB), "count=25&kind_name=recent", formContentType},
		{New().BodyForm(paramsB).New().BodyJSON(nil), "count=25&kind_name=recent", formContentType},
		{New().BodyJSON(modelA).New().BodyForm(nil), "{\"text\":\"note\",\"favorite_count\":12}\n", jsonContentType},
		// Body
		{New().Body(strings.NewReader("this-is-a-test")), "this-is-a-test", ""},
		{New().Body(strings.NewReader("a")).Body(strings.NewReader("b")), "b", ""},
	}
	for _, c := range cases {
		req, _ := c.sling.Request()
		buf := new(bytes.Buffer)
		buf.ReadFrom(req.Body)
		// req.Body should have contained the expectedBody string
		if value := buf.String(); value != c.expectedBody {
			t.Errorf("expected Request.Body %s, got %s", c.expectedBody, value)
		}
		// Header Content-Type should be expectedContentType ("" means no contentType expected)
		if actualHeader := req.Header.Get(contentType); actualHeader != c.expectedContentType && c.expectedContentType != "" {
			t.Errorf("Incorrect or missing header, expected %s, got %s", c.expectedContentType, actualHeader)
		}
	}
}

func TestRequest_bodyNoData(t *testing.T) {
	// test that Body is left nil when no bodyJSON or bodyStruct set
	slings := []*Sling{
		New(),
		New().BodyJSON(nil),
		New().BodyForm(nil),
	}
	for _, sling := range slings {
		req, _ := sling.Request()
		if req.Body != nil {
			t.Errorf("expected nil Request.Body, got %v", req.Body)
		}
		// Header Content-Type should not be set when bodyJSON argument was nil or never called
		if actualHeader := req.Header.Get(contentType); actualHeader != "" {
			t.Errorf("did not expect a Content-Type header, got %s", actualHeader)
		}
	}
}

func TestRequest_bodyEncodeErrors(t *testing.T) {
	cases := []struct {
		sling       *Sling
		expectedErr error
	}{
		// check that Encode errors are propagated, illegal JSON field
		{New().BodyJSON(FakeModel{Temperature: math.Inf(1)}), errors.New("json: unsupported value: +Inf")},
	}
	for _, c := range cases {
		req, err := c.sling.Request()
		if err == nil || err.Error() != c.expectedErr.Error() {
			t.Errorf("expected error %v, got %v", c.expectedErr, err)
		}
		if req != nil {
			t.Errorf("expected nil Request, got %+v", req)
		}
	}
}

func TestRequest_headers(t *testing.T) {
	cases := []struct {
		sling          *Sling
		expectedHeader map[string][]string
	}{
		{New().Add("authorization", "OAuth key=\"value\""), map[string][]string{"Authorization": []string{"OAuth key=\"value\""}}},
		// header keys should be canonicalized
		{New().Add("content-tYPE", "application/json").Add("User-AGENT", "sling"), map[string][]string{"Content-Type": []string{"application/json"}, "User-Agent": []string{"sling"}}},
		// values for existing keys should be appended
		{New().Add("A", "B").Add("a", "c"), map[string][]string{"A": []string{"B", "c"}}},
		// Add should add to values for keys added by parent Slings
		{New().Add("A", "B").Add("a", "c").New(), map[string][]string{"A": []string{"B", "c"}}},
		{New().Add("A", "B").New().Add("a", "c"), map[string][]string{"A": []string{"B", "c"}}},
		// Add and Set
		{New().Add("A", "B").Set("a", "c"), map[string][]string{"A": []string{"c"}}},
		{New().Set("content-type", "A").Set("Content-Type", "B"), map[string][]string{"Content-Type": []string{"B"}}},
		// Set should replace values received by copying parent Slings
		{New().Set("A", "B").Add("a", "c").New(), map[string][]string{"A": []string{"B", "c"}}},
		{New().Add("A", "B").New().Set("a", "c"), map[string][]string{"A": []string{"c"}}},
	}
	for _, c := range cases {
		req, _ := c.sling.Request()
		// type conversion from Header to alias'd map for deep equality comparison
		headerMap := map[string][]string(req.Header)
		if !reflect.DeepEqual(c.expectedHeader, headerMap) {
			t.Errorf("not DeepEqual: expected %v, got %v", c.expectedHeader, headerMap)
		}
	}
}

func TestAddQueryStructs(t *testing.T) {
	cases := []struct {
		rawurl       string
		queryStructs []interface{}
		expected     string
	}{
		{"http://a.io", []interface{}{}, "http://a.io"},
		{"http://a.io", []interface{}{paramsA}, "http://a.io?limit=30"},
		{"http://a.io", []interface{}{paramsA, paramsA}, "http://a.io?limit=30&limit=30"},
		{"http://a.io", []interface{}{paramsA, paramsB}, "http://a.io?count=25&kind_name=recent&limit=30"},
		// don't blow away query values on the rawURL (parsed into RawQuery)
		{"http://a.io?initial=7", []interface{}{paramsA}, "http://a.io?initial=7&limit=30"},
	}
	for _, c := range cases {
		reqURL, _ := url.Parse(c.rawurl)
		addQueryStructs(reqURL, c.queryStructs)
		if reqURL.String() != c.expected {
			t.Errorf("expected %s, got %s", c.expected, reqURL.String())
		}
	}
}

// Sending

type APIError struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func TestDo_onSuccess(t *testing.T) {
	const expectedText = "Some text"
	const expectedFavoriteCount int64 = 24

	client, mux, server := testServer()
	defer server.Close()
	mux.HandleFunc("/success", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"text": "Some text", "favorite_count": 24}`)
	})

	sling := New().Client(client)
	req, _ := http.NewRequest("GET", "http://example.com/success", nil)

	model := new(FakeModel)
	apiError := new(APIError)
	resp, err := sling.Do(req, model, apiError)

	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected %d, got %d", 200, resp.StatusCode)
	}
	if model.Text != expectedText {
		t.Errorf("expected %s, got %s", expectedText, model.Text)
	}
	if model.FavoriteCount != expectedFavoriteCount {
		t.Errorf("expected %d, got %d", expectedFavoriteCount, model.FavoriteCount)
	}
}

func TestDo_onSuccessWithNilValue(t *testing.T) {
	client, mux, server := testServer()
	defer server.Close()
	mux.HandleFunc("/success", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"text": "Some text", "favorite_count": 24}`)
	})

	sling := New().Client(client)
	req, _ := http.NewRequest("GET", "http://example.com/success", nil)

	apiError := new(APIError)
	resp, err := sling.Do(req, nil, apiError)

	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected %d, got %d", 200, resp.StatusCode)
	}
	expected := &APIError{}
	if !reflect.DeepEqual(expected, apiError) {
		t.Errorf("failureV should not be populated, exepcted %v, got %v", expected, apiError)
	}
}

func TestDo_onFailure(t *testing.T) {
	const expectedMessage = "Invalid argument"
	const expectedCode int = 215

	client, mux, server := testServer()
	defer server.Close()
	mux.HandleFunc("/failure", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(400)
		fmt.Fprintf(w, `{"message": "Invalid argument", "code": 215}`)
	})

	sling := New().Client(client)
	req, _ := http.NewRequest("GET", "http://example.com/failure", nil)

	model := new(FakeModel)
	apiError := new(APIError)
	resp, err := sling.Do(req, model, apiError)

	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
	if resp.StatusCode != 400 {
		t.Errorf("expected %d, got %d", 400, resp.StatusCode)
	}
	if apiError.Message != expectedMessage {
		t.Errorf("expected %s, got %s", expectedMessage, apiError.Message)
	}
	if apiError.Code != expectedCode {
		t.Errorf("expected %d, got %d", expectedCode, apiError.Code)
	}
}

func TestDo_onFailureWithNilValue(t *testing.T) {
	client, mux, server := testServer()
	defer server.Close()
	mux.HandleFunc("/failure", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(420)
		fmt.Fprintf(w, `{"message": "Enhance your calm", "code": 88}`)
	})

	sling := New().Client(client)
	req, _ := http.NewRequest("GET", "http://example.com/failure", nil)

	model := new(FakeModel)
	resp, err := sling.Do(req, model, nil)

	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
	if resp.StatusCode != 420 {
		t.Errorf("expected %d, got %d", 420, resp.StatusCode)
	}
	expected := &FakeModel{}
	if !reflect.DeepEqual(expected, model) {
		t.Errorf("successV should not be populated, exepcted %v, got %v", expected, model)
	}
}

func TestReceive_success(t *testing.T) {
	client, mux, server := testServer()
	defer server.Close()
	mux.HandleFunc("/foo/submit", func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, "POST", r)
		assertQuery(t, map[string]string{"kind_name": "vanilla", "count": "11"}, r)
		assertPostForm(t, map[string]string{"kind_name": "vanilla", "count": "11"}, r)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"text": "Some text", "favorite_count": 24}`)
	})

	endpoint := New().Client(client).Base("http://example.com/").Path("foo/").Post("submit")
	// encode url-tagged struct in query params and as post body for testing purposes
	params := FakeParams{KindName: "vanilla", Count: 11}
	model := new(FakeModel)
	apiError := new(APIError)
	resp, err := endpoint.New().QueryStruct(params).BodyForm(params).Receive(model, apiError)

	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected %d, got %d", 200, resp.StatusCode)
	}
	expectedModel := &FakeModel{Text: "Some text", FavoriteCount: 24}
	if !reflect.DeepEqual(expectedModel, model) {
		t.Errorf("expected %v, got %v", expectedModel, model)
	}
	expectedAPIError := &APIError{}
	if !reflect.DeepEqual(expectedAPIError, apiError) {
		t.Errorf("failureV should be zero valued, exepcted %v, got %v", expectedAPIError, apiError)
	}
}

func TestReceive_failure(t *testing.T) {
	client, mux, server := testServer()
	defer server.Close()
	mux.HandleFunc("/foo/submit", func(w http.ResponseWriter, r *http.Request) {
		assertMethod(t, "POST", r)
		assertQuery(t, map[string]string{"kind_name": "vanilla", "count": "11"}, r)
		assertPostForm(t, map[string]string{"kind_name": "vanilla", "count": "11"}, r)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(429)
		fmt.Fprintf(w, `{"message": "Rate limit exceeded", "code": 88}`)
	})

	endpoint := New().Client(client).Base("http://example.com/").Path("foo/").Post("submit")
	// encode url-tagged struct in query params and as post body for testing purposes
	params := FakeParams{KindName: "vanilla", Count: 11}
	model := new(FakeModel)
	apiError := new(APIError)
	resp, err := endpoint.New().QueryStruct(params).BodyForm(params).Receive(model, apiError)

	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
	if resp.StatusCode != 429 {
		t.Errorf("expected %d, got %d", 429, resp.StatusCode)
	}
	expectedAPIError := &APIError{Message: "Rate limit exceeded", Code: 88}
	if !reflect.DeepEqual(expectedAPIError, apiError) {
		t.Errorf("expected %v, got %v", expectedAPIError, apiError)
	}
	expectedModel := &FakeModel{}
	if !reflect.DeepEqual(expectedModel, model) {
		t.Errorf("successV should not be zero valued, expected %v, got %v", expectedModel, model)
	}
}

func TestReceive_errorCreatingRequest(t *testing.T) {
	expectedErr := errors.New("json: unsupported value: +Inf")
	resp, err := New().BodyJSON(FakeModel{Temperature: math.Inf(1)}).Receive(nil, nil)
	if err == nil || err.Error() != expectedErr.Error() {
		t.Errorf("expected %v, got %v", expectedErr, err)
	}
	if resp != nil {
		t.Errorf("expected nil resp, got %v", resp)
	}
}

// Testing Utils

// testServer returns an http Client, ServeMux, and Server. The client proxies
// requests to the server and handlers can be registered on the mux to handle
// requests. The caller must close the test server.
func testServer() (*http.Client, *http.ServeMux, *httptest.Server) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	transport := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(server.URL)
		},
	}
	client := &http.Client{Transport: transport}
	return client, mux, server
}

func assertMethod(t *testing.T, expectedMethod string, req *http.Request) {
	if actualMethod := req.Method; actualMethod != expectedMethod {
		t.Errorf("expected method %s, got %s", expectedMethod, actualMethod)
	}
}

// assertQuery tests that the Request has the expected url query key/val pairs
func assertQuery(t *testing.T, expected map[string]string, req *http.Request) {
	queryValues := req.URL.Query() // net/url Values is a map[string][]string
	expectedValues := url.Values{}
	for key, value := range expected {
		expectedValues.Add(key, value)
	}
	if !reflect.DeepEqual(expectedValues, queryValues) {
		t.Errorf("expected parameters %v, got %v", expected, req.URL.RawQuery)
	}
}

// assertPostForm tests that the Request has the expected key values pairs url
// encoded in its Body
func assertPostForm(t *testing.T, expected map[string]string, req *http.Request) {
	req.ParseForm() // parses request Body to put url.Values in r.Form/r.PostForm
	expectedValues := url.Values{}
	for key, value := range expected {
		expectedValues.Add(key, value)
	}
	if !reflect.DeepEqual(expectedValues, req.PostForm) {
		t.Errorf("expected parameters %v, got %v", expected, req.PostForm)
	}
}
