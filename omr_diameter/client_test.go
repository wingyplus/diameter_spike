package main

import (
	"bytes"
	"io"
	"net"
	"testing"
	"time"

	"github.com/fiorix/go-diameter/diam"
	"github.com/fiorix/go-diameter/diam/avp"
	"github.com/fiorix/go-diameter/diam/datatype"
	"github.com/fiorix/go-diameter/diam/diamtest"
	"github.com/fiorix/go-diameter/diam/dict"
	"github.com/wingyplus/diameter_spike/diameter/dictionary"
)

func TestClientCallCER(t *testing.T) {
	errc := make(chan error, 1)

	smux := diam.NewServeMux()
	smux.Handle("CER", handleCER(errc))
	smux.Handle("CCR", handleCCR(errc))

	srv := diamtest.NewServer(smux, nil)
	defer srv.Close()

	client := &DiameterClient{
		Endpoint: srv.Address,
	}
	done, err := client.Run()
	if err != nil {
		t.Error("Cannot connect to server")
		return
	}
	select {
	case err := <-errc:
		t.Error(err)
	case <-done:
	}
}

func handleCER(errc chan error) diam.HandlerFunc {
	type CER struct {
		OriginHost        string    `avp:"Origin-Host"`
		OriginRealm       string    `avp:"Origin-Realm"`
		VendorID          int       `avp:"Vendor-Id"`
		ProductName       string    `avp:"Product-Name"`
		OriginStateID     *diam.AVP `avp:"Origin-State-Id"`
		AcctApplicationID *diam.AVP `avp:"Acct-Application-Id"`
	}
	return func(c diam.Conn, m *diam.Message) {
		var req CER
		err := m.Unmarshal(&req)
		if err != nil {
			errc <- err
			return
		}

		a := m.Answer(diam.Success)
		_, err = sendCEA(c, a, req.OriginStateID, req.AcctApplicationID)
		if err != nil {
			errc <- err
		}
	}
}

func sendCEA(w io.Writer, m *diam.Message, OriginStateID, AcctApplicationID *diam.AVP) (n int64, err error) {
	m.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.OctetString("srv"))
	m.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.OctetString("localhost"))
	m.NewAVP(avp.HostIPAddress, avp.Mbit, 0, datatype.Address(net.ParseIP("127.0.0.1")))
	m.NewAVP(avp.VendorID, avp.Mbit, 0, datatype.Unsigned32(99))
	m.NewAVP(avp.ProductName, avp.Mbit, 0, datatype.UTF8String("go-diameter"))
	m.AddAVP(OriginStateID)
	m.AddAVP(AcctApplicationID)
	return m.WriteTo(w)
}

func TestClientCallCCR(t *testing.T) {
	errc := make(chan error, 1)

	dict.Default.Load(bytes.NewBufferString(dictionary.AppDictionary))
	dict.Default.Load(bytes.NewBufferString(dictionary.CreditControlDictionary))

	smux := diam.NewServeMux()
	smux.Handle("CER", handleCER(errc))
	smux.Handle("CCR", handleCCR(errc))

	srv := diamtest.NewServer(smux, dict.Default)
	defer srv.Close()

	client := &DiameterClient{
		Endpoint: srv.Address,
	}
	done, err := client.Run()
	if err != nil {
		t.Error("Cannot connect to server")
		return
	}
	select {
	case err := <-errc:
		t.Error(err)
		return
	case <-done:

	}

	out := make(chan Response)
	number := SubscriberNumber("66814060967")
	currentTime := time.Now()
	request := Request{out: out, data: &QueryBalanceData{Number: number, Time: currentTime}}

	in <- request
	select {
	case response := <-out:
		var diamResponse DiamResponse
		err = response.Unmarshal(&diamResponse)
		if err != nil {
			t.Error(err)
		}
		expected := "dtac.co.th;OMR200601021504050000"

		sessionId := diamResponse.SessionID
		if sessionId != expected {
			t.Errorf("expect %s but got %s", expected, sessionId)
		}
	case err := <-errc:
		t.Error(err)
	case err := <-smux.ErrorReports():
		t.Error(err)
	}
}

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

func handleCCR(errc chan error) diam.HandlerFunc {
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
