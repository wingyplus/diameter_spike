package main

import (
	"encoding/json"
	"fmt"
	"testing"
)

type mockDiameterClient struct {
	fn func(out chan Response)
}

func (m *mockDiameterClient) Run() {
	request := <-in
	if m.fn != nil {
		m.fn(request.out)
	}
}

func newClient(fn func(out chan Response)) *mockDiameterClient {
	in = make(chan Request)
	return &mockDiameterClient{
		fn: fn,
	}
}

type mockResponse struct {
	Result string
}

func (resp *mockResponse) Unmarshal(v interface{}) error {
	body := `{
	    "ServiceInformation": {
	        "BalanceInformation": {
	            "FirstActiveDate": "%s"
	        }
	    }
	}`
	body = fmt.Sprintf(body, resp.Result)
	return json.Unmarshal([]byte(body), v)
}

func TestBalancerUnmarshalResponseFromDiameterClient(t *testing.T) {
	balancer := &QueryBalancer{}
	diameterClient := newClient(func(out chan Response) {
		out <- &mockResponse{
			Result: "20150501",
		}
	})
	go diameterClient.Run()

	subscriberNumber := SubscriberNumber("66814060967")
	balanceInfo := balancer.Retrieve(subscriberNumber)
	expected := "20150501"
	if balanceInfo.FirstActiveDate != expected {
		t.Errorf("expect %s but got %s", expected, balanceInfo.FirstActiveDate)
	}
}

func TestBalancerCallDiameterClient(t *testing.T) {
	balancer := &QueryBalancer{}
	diameterClient := newClient(func(out chan Response) {
		out <- &mockResponse{
			Result: "20150101",
		}
	})
	go diameterClient.Run()

	subscriberNumber := SubscriberNumber("66949014731")
	balanceInfo := balancer.Retrieve(subscriberNumber)
	expected := "20150101"
	if balanceInfo.FirstActiveDate != expected {
		t.Errorf("expect %s but got %s", expected, balanceInfo.FirstActiveDate)
	}
}
