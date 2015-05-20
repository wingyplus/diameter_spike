package servtest

import (
	"github.com/fiorix/go-diameter/diam"
	"github.com/fiorix/go-diameter/diam/avp"
	"github.com/fiorix/go-diameter/diam/datatype"
)

type CCR struct {
	SessionID         *diam.AVP `avp:"Session-Id"`
	AuthApplicationID *diam.AVP `avp:"Auth-Application-Id"`
	DestinationRealm  string    `avp:"Destination-Realm"`
	OriginHost        string    `avp:"Origin-Host"`
	OriginRealm       string    `avp:"Origin-Realm"`
	CCRequestType     *diam.AVP `avp:"CC-Request-Type"`
	ServiceContextID  string    `avp:"Service-Context-Id"`
	RequestedAction   *diam.AVP `avp:"Requested-Action"`
	EventTimestamp    *diam.AVP `avp:"Event-Timestamp"`
	ServiceIdentifier *diam.AVP `avp:"Service-Identifier"`
	CCRequestNumber   *diam.AVP `avp:"CC-Request-Number"`
	RouteRecord       *diam.AVP `avp:"Route-Record"`
	DestinationHost   *diam.AVP `avp:"Destination-Host"`
}

func serverHandleCCR(errc chan error) diam.HandlerFunc {
	return func(c diam.Conn, m *diam.Message) {
		var req CCR
		err := m.Unmarshal(&req)
		if err != nil {
			errc <- err
			return
		}
		a := m.Answer(diam.Success)
		_, err = sendCCA(c, a, req)
		if err != nil {
			errc <- err
		}
	}
}

func sendCCA(c diam.Conn, m *diam.Message, req CCR) (n int64, err error) {
	m.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String("dtac.co.th;OMR200601021504050000"))
	return m.WriteTo(c)
}
