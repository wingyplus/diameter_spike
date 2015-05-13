package main

import "github.com/ant0ine/go-json-rest/rest"

type SubscriberNumber string

type BalanceHandler struct {
	Balancer Balancer
}

func (h *BalanceHandler) Handler(w rest.ResponseWriter, r *rest.Request) {
	subscriberNumber := SubscriberNumber(r.PathParam("subscribernumber"))
	balanceInfo := h.Balancer.Retrieve(subscriberNumber)
	w.WriteJson(balanceInfo)
}
