package main

import (
	"net/http"
	"time"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/fiorix/go-diameter/diam"
	"github.com/fiorix/go-diameter/diam/avp"
	"github.com/fiorix/go-diameter/diam/datatype"
	"github.com/wingyplus/diameter_spike/diameter"
)

type Query struct {
	in chan diameter.Session
}

var dtnSubscriberNumber SubscriberNumber = "66949014731"

func (q *Query) Handler(w rest.ResponseWriter, r *rest.Request) {
	balanceInfo := RetrieveBalance(q.in, dtnSubscriberNumber)
	w.WriteJson(balanceInfo)
}

type SubscriberNumber string

type BalanceInformation struct {
	ServiceInformation struct {
		BalanceInformation struct {
			FirstActiveDate string `avp:"First-Active-Date"`
		} `avp:"Balance-Information"`
	} `avp:"Service-Information"`
}

type BalanceInformationRequest struct {
	number SubscriberNumber
}

func (balanceInfo *BalanceInformationRequest) AVP() []*diam.AVP {
	const (
		balanceInformation  = 21100
		accessMethod        = 20340
		accountQueryMethod  = 20346
		sspTime             = 20386
		callingPartyAddress = 20336
	)

	number := string(balanceInfo.number)
	balanceInfoRequest := diam.NewAVP(balanceInformation, avp.Mbit, 0, &diam.GroupedAVP{
		AVP: []*diam.AVP{
			diam.NewAVP(callingPartyAddress, avp.Mbit, 0, datatype.UTF8String(number)),
			diam.NewAVP(accessMethod, avp.Mbit, 0, datatype.Unsigned32(9)),
			diam.NewAVP(accountQueryMethod, avp.Mbit, 0, datatype.Unsigned32(1)),
			diam.NewAVP(sspTime, avp.Mbit, 0, datatype.Time(time.Now())),
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

func RetrieveBalance(c chan diameter.Session, number SubscriberNumber) (balanceInfo BalanceInformation) {
	out := make(chan diameter.Data)
	c <- diameter.Session{OutChan: out, Request: &BalanceInformationRequest{number}}

	d := <-out
	d.Response.Unmarshal(&balanceInfo)
	return
}

func main() {
	inCh := diameter.BackgroundClient()

	q := &Query{inCh}

	router, _ := rest.MakeRouter(rest.Get("/q", q.Handler))

	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	api.SetApp(router)

	http.ListenAndServe(":8080", api.MakeHandler())
}
