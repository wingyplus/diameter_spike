package main

import (
	"time"

	"github.com/fiorix/go-diameter/diam"
	"github.com/fiorix/go-diameter/diam/avp"
	"github.com/fiorix/go-diameter/diam/datatype"
)

type Balancer interface {
	Retrieve(number SubscriberNumber) BalanceInformation
}

type DiamResponse struct {
	ServiceInformation ServiceInformation
}

type ServiceInformation struct {
	BalanceInformation BalanceInformation
}

type BalanceInformation struct {
	FirstActiveDate string `json:"firstActiveDate"`
}

type QueryBalancer struct{}

func (q *QueryBalancer) Retrieve(number SubscriberNumber) BalanceInformation {
	out := make(chan Response)
	in <- Request{out: out, data: &QueryBalanceData{Number: number, Time: time.Now()}}
	resp := <-out
	var diamResponse DiamResponse
	resp.Unmarshal(&diamResponse)
	return diamResponse.ServiceInformation.BalanceInformation
}

var in chan Request

type Request struct {
	out  chan Response
	data Data
}

type QueryBalanceData struct {
	Number SubscriberNumber
	Time   time.Time
}

func (d *QueryBalanceData) AVP() []*diam.AVP {
	const (
		balanceInformation  = 21100
		accessMethod        = 20340
		accountQueryMethod  = 20346
		sspTime             = 20386
		callingPartyAddress = 20336
	)

	number := string(d.Number)
	balanceInfoRequest := diam.NewAVP(balanceInformation, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(callingPartyAddress, avp.Mbit, 0, datatype.UTF8String(number)),
			diam.NewAVP(accessMethod, avp.Mbit, 0, datatype.Unsigned32(9)),
			diam.NewAVP(accountQueryMethod, avp.Mbit, 0, datatype.Unsigned32(1)),
			diam.NewAVP(sspTime, avp.Mbit, 0, datatype.Time(d.Time)),
		},
	})
	return []*diam.AVP{
		diam.NewAVP(avp.SubscriptionID, avp.Mbit, 0, &diam.GroupedAVP{
			AVP: []*diam.AVP{
				diam.NewAVP(avp.SubscriptionIDType, avp.Mbit, 0, datatype.Integer32(0)),
				diam.NewAVP(avp.SubscriptionIDData, avp.Mbit, 0, datatype.UTF8String(number)),
			},
		}),
		diam.NewAVP(avp.ServiceInformation, avp.Mbit, 0, &diam.GroupedAVP{
			AVP: []*diam.AVP{
				balanceInfoRequest,
			},
		}),
	}
}

type Data interface {
	AVP() []*diam.AVP
}

type Response interface {
	Unmarshal(v interface{}) error
}
