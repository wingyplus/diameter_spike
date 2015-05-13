package main

import "github.com/ant0ine/go-json-rest/rest"

type DiamResponse struct {
	ServiceInformation ServiceInformation
}

type ServiceInformation struct {
	BalanceInformation BalanceInformation
}

type SubscriberNumber string

type Balancer interface {
	Retrieve(number SubscriberNumber) BalanceInformation
}

type BalanceHandler struct {
	Balancer Balancer
}

type BalanceInformation struct {
	FirstActiveDate string `json:"firstActiveDate"`
}

func (h *BalanceHandler) Handler(w rest.ResponseWriter, r *rest.Request) {
	subscriberNumber := SubscriberNumber(r.PathParam("subscribernumber"))
	balanceInfo := h.Balancer.Retrieve(subscriberNumber)
	w.WriteJson(balanceInfo)
}
