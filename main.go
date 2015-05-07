package main

import (
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/fiorix/go-diameter/diam"
	"github.com/fiorix/go-diameter/diam/avp"
	"github.com/fiorix/go-diameter/diam/datatype"
	"github.com/wingyplus/diameter_spike/diameter"
)

type HMR struct {
	AccountCode int
}

func (hmr *HMR) AVP() []*diam.AVP {
	return []*diam.AVP{
		diam.NewAVP(31809, avp.Mbit, 0, datatype.Integer64(hmr.AccountCode)),
	}
}

type HMA struct {
	SessionID string `avp:"Session-Id"`
}

type Query struct {
	in chan diameter.Session
}

func (q *Query) Handler(w rest.ResponseWriter, r *rest.Request) {
	id := r.PathParam("id")
	out := make(chan diameter.Data)
	q.in <- diameter.Session{ID: id, OutChan: out, Request: &HMR{AccountCode: 555}}
	var hma HMA
	select {
	case d := <-out:
		if d.Err != nil {
			w.WriteJson(d.Err)
			return
		}
		d.Response.Unmarshal(&hma)
		w.WriteJson(hma)
	}
}

func main() {
	inCh := diameter.BackgroundClient()

	q := &Query{inCh}

	router, _ := rest.MakeRouter(rest.Get("/q/:id", q.Handler))

	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	api.SetApp(router)

	http.ListenAndServe(":8080", api.MakeHandler())
}
