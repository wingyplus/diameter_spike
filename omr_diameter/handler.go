package main

import (
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/wingyplus/diameter_spike/omr_diameter/dcc"
)

type BalanceHandler struct {
	Balancer dcc.Balancer
}

func (h *BalanceHandler) Handler(w rest.ResponseWriter, r *rest.Request) {
	subscriberNumber := dcc.SubscriberNumber(r.PathParam("subscribernumber"))
	balanceInfo := h.Balancer.Retrieve(subscriberNumber)
	w.WriteJson(balanceInfo)
}
