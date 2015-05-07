package main

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"

	"github.com/ant0ine/go-json-rest/rest"
)

func TestQ(t *testing.T) {
	in, out := backgroundClient()
	q := &Query{in, out}

	router, _ := rest.MakeRouter(rest.Get("/q/:id", q.Handler))

	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	api.SetApp(router)

	ts := httptest.NewServer(api.MakeHandler())
	defer ts.Close()

	loop := 1000

	var wg sync.WaitGroup
	wg.Add(loop)

	for i := 0; i < loop; i++ {
		go testSessionID(t, ts.URL, strconv.Itoa(int(rand.Uint32())), &wg)
	}

	wg.Wait()
}

func testSessionID(t *testing.T, url string, id string, wg *sync.WaitGroup) {
	defer wg.Done()

	var m map[string]string
	resp, err := http.Get(url + "/q/" + id)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	json.NewDecoder(resp.Body).Decode(&m)

	if m["SessionID"] != "session;"+id {
		t.Errorf("expect session;%s but got %s", id, m["SessionID"])
	}
}
