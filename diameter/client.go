package diameter

import (
	"fmt"
	"net"

	"github.com/fiorix/go-diameter/diam"
	"github.com/fiorix/go-diameter/diam/avp"
	"github.com/fiorix/go-diameter/diam/datatype"
	"github.com/fiorix/go-diameter/diam/sm"
)

type Session struct {
	ID      string
	OutChan chan Data
	Request Request
}

type Data struct {
	Response *Response
	Err      error
}

type Response struct {
	*diam.Message
}

const (
	identity    = datatype.DiameterIdentity("jenkin13_OMR_TEST01")
	realm       = datatype.DiameterIdentity("dtac.co.th")
	vendorID    = datatype.Unsigned32(0)
	productName = datatype.UTF8String("omr")

	host = "localhost"
	port = "3868"

	dtn     = "66949014731"
	dtnAddr = "10.89.111.40:6573"
)

func BackgroundClient() chan Session {
	in := make(chan Session, 1000)

	// dict.Default.Load(bytes.NewBufferString(dictionary.HelloDictionary))
	cfg := &sm.Settings{
		OriginHost:       identity,
		OriginRealm:      realm,
		VendorID:         vendorID,
		ProductName:      productName,
		FirmwareRevision: 1,
	}

	done := make(chan Data)

	diam.HandleFunc("CEA", handleCEA(done))
	diam.HandleFunc("DWA", func(c diam.Conn, m *diam.Message) {
		fmt.Println(m)
	})

	conn, err := diam.Dial(dtnAddr, nil, nil)
	if err != nil {
		fmt.Println(err.Error())
		return in
	}

	sendCER(conn, cfg)
	data := <-done
	fmt.Printf("%v\n", data.Response.Message)

	sendDWR(conn, cfg)

	return in
}

type Request interface {
	AVP() []*diam.AVP
}

func sendDWR(conn diam.Conn, cfg *sm.Settings) error {
	var (
		watchdogExchange uint32 = 280
		appID                uint32 = 0
	)

	m := diam.NewRequest(watchdogExchange, appID, nil)

	m.NewAVP(avp.OriginHost, avp.Mbit, 0, cfg.OriginHost)
	m.NewAVP(avp.OriginRealm, avp.Mbit, 0, cfg.OriginRealm)

	_, err := m.WriteTo(conn)

	return err
}

func sendCER(conn diam.Conn, cfg *sm.Settings) error {
	var (
		capabilitiesExchange uint32 = 257
		appID                uint32 = 0
	)

	m := diam.NewRequest(capabilitiesExchange, appID, nil)

	m.NewAVP(avp.OriginHost, avp.Mbit, 0, cfg.OriginHost)
	m.NewAVP(avp.OriginRealm, avp.Mbit, 0, cfg.OriginRealm)

	ip, _, _ := net.SplitHostPort(conn.LocalAddr().String())
	m.NewAVP(avp.HostIPAddress, avp.Mbit, 0, datatype.Address(net.ParseIP(ip)))
	m.NewAVP(avp.VendorID, avp.Mbit, 0, vendorID)
	m.NewAVP(avp.ProductName, 0, 0, productName)
	m.NewAVP(avp.SupportedVendorID, avp.Mbit, 0, datatype.Unsigned32(0))
	m.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(4))
	m.NewAVP(avp.AcctApplicationID, avp.Mbit, 0, datatype.Unsigned32(4))
	m.NewAVP(avp.FirmwareRevision, avp.Mbit, 0, cfg.FirmwareRevision)

	_, err := m.WriteTo(conn)

	return err
}

func handleCEA(d chan Data) diam.HandlerFunc {
	return func(c diam.Conn, m *diam.Message) {
		d <- Data{Response: &Response{Message: m}}
	}
}
