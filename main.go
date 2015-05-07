package main

import (
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/wingyplus/diameter_spike/diameter"
)

type Query struct {
	in chan diameter.Session
}

func (q *Query) Handler(w rest.ResponseWriter, r *rest.Request) {
	id := r.PathParam("id")
	out := make(chan diameter.Data)
	q.in <- diameter.Session{ID: id, OutChan: out}
	select {
	case d := <-out:
		if d.Err != nil {
			w.WriteJson(d.Err)
			return
		}
		w.WriteJson(d.Response)
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
