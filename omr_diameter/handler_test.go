package main

import (
	"testing"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/ant0ine/go-json-rest/rest/test"
	"github.com/wingyplus/diameter_spike/omr_diameter/dcc"
)

type mockQueryBalancer struct{}

func (m *mockQueryBalancer) Retrieve(number dcc.SubscriberNumber) dcc.BalanceInformation {
	if number == "66949014731" {
		return dcc.BalanceInformation{
			FirstActiveDate: "20150201",
		}
	}
	return dcc.BalanceInformation{
		FirstActiveDate: "20150501",
	}
}

func TestQueryBalanceHandler(t *testing.T) {
	balanceHandler := &BalanceHandler{
		Balancer: &mockQueryBalancer{},
	}
	router, _ := rest.MakeRouter(
		rest.Get("/api/balance/:subscribernumber", balanceHandler.Handler),
	)
	api := rest.NewApi()
	api.SetApp(router)

	req := test.MakeSimpleRequest("GET", "http://localhost/api/balance/66949014731", nil)
	recorded := test.RunRequest(t, api.MakeHandler(), req)
	recorded.CodeIs(200)
	recorded.BodyIs(`{"firstActiveDate":"20150201"}`)
}

func TestCallQueryBalancer(t *testing.T) {
	balanceHandler := &BalanceHandler{
		Balancer: &mockQueryBalancer{},
	}

	router, _ := rest.MakeRouter(
		rest.Get("/api/balance/:subscribernumber", balanceHandler.Handler),
	)
	api := rest.NewApi()
	api.SetApp(router)
	req := test.MakeSimpleRequest("GET", "http://localhost/api/balance/66814060967", nil)
	recorded := test.RunRequest(t, api.MakeHandler(), req)
	recorded.CodeIs(200)
	recorded.BodyIs(`{"firstActiveDate":"20150501"}`)
}
