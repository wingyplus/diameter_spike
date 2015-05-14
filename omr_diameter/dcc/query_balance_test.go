package dcc

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/fiorix/go-diameter/diam"
	"github.com/fiorix/go-diameter/diam/avp"
	"github.com/fiorix/go-diameter/diam/datatype"
)

type mockDiameterClient struct {
	fn func(out chan Response)
	t  *testing.T
}

func (m *mockDiameterClient) Run() {
	request := <-in
	if data, ok := request.data.(*QueryBalanceData); ok {
		if data.Number == "" {
			m.t.Error("expect number is not empty")
		}
	} else {
		m.t.Error("expect data should be QueryBalanceData")
	}
	m.fn(request.out)
}

func newClient(t *testing.T, fn func(out chan Response)) *mockDiameterClient {
	in = make(chan Request)
	return &mockDiameterClient{
		fn: fn,
		t:  t,
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
	diameterClient := newClient(t, func(out chan Response) {
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
	diameterClient := newClient(t, func(out chan Response) {
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

func TestQueryBalanceData(t *testing.T) {
	currentTime := time.Now()
	request := QueryBalanceData{
		Number: "66949014731",
		Time:   currentTime,
	}

	const (
		balanceInformation  = 21100
		accessMethod        = 20340
		accountQueryMethod  = 20346
		sspTime             = 20386
		callingPartyAddress = 20336
	)

	balanceInfoRequest := diam.NewAVP(balanceInformation, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(callingPartyAddress, avp.Mbit, 0, datatype.UTF8String("66949014731")),
			diam.NewAVP(accessMethod, avp.Mbit, 0, datatype.Unsigned32(9)),
			diam.NewAVP(accountQueryMethod, avp.Mbit, 0, datatype.Unsigned32(1)),
			diam.NewAVP(sspTime, avp.Mbit, 0, datatype.Time(currentTime)),
		},
	})
	expected := []*diam.AVP{
		diam.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
			AVP: []*diam.AVP{
				diam.NewAVP(avp.SubscriptionIDType, avp.Mbit, 0, datatype.Integer32(0)),
				diam.NewAVP(avp.SubscriptionIDData, avp.Mbit, 0, datatype.UTF8String("66949014731")),
			},
		}),
		diam.NewAVP(avp.ServiceInformation, avp.Mbit, 0, &diam.GroupedAVP{
			AVP: []*diam.AVP{
				balanceInfoRequest,
			},
		}),
	}
	if !reflect.DeepEqual(request.AVP(), expected) {
		t.Errorf("expect %v but got %v", expected, request.AVP())
	}
}
