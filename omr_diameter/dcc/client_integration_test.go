// +build integration

package dcc

import (
	"testing"
	"time"
)

func TestIntegration_Client_QueryBalance(t *testing.T) {
	c := DiameterClient{
		Endpoint: "10.89.111.40:6573",
	}

	donec, err := c.Run()
	if err != nil {
		t.Error(err)
		return
	}
	<-donec

	out := make(chan Response)
	request := Request{
		out: out,
		data: &QueryBalanceData{
			Number: SubscriberNumber("66944800119"),
			Time:   time.Now(),
		},
	}

	in <- request
	response := <-out
	var diamResponse DiamResponse
	err = response.Unmarshal(&diamResponse)
	if err != nil {
		t.Error(err)
	}

	expected := "20150427092203"
	if activeDate := diamResponse.ServiceInformation.BalanceInformation.FirstActiveDate; activeDate != expected {
		t.Errorf("expect %s but got %s", expected, activeDate)
	}
}
