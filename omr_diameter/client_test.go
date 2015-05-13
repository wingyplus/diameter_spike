package main

import (
	"io"
	"net"
	"testing"

	"github.com/fiorix/go-diameter/diam"
	"github.com/fiorix/go-diameter/diam/avp"
	"github.com/fiorix/go-diameter/diam/datatype"
	"github.com/fiorix/go-diameter/diam/diamtest"
)

func TestClientCallCER(t *testing.T) {
	errc := make(chan error, 1)

	smux := diam.NewServeMux()
	smux.Handle("CER", handleCER(errc))

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
