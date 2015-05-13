package diameter

import (
	"bytes"
	"fmt"
	"net"
	"time"

	"github.com/fiorix/go-diameter/diam"
	"github.com/fiorix/go-diameter/diam/avp"
	"github.com/fiorix/go-diameter/diam/datatype"
	"github.com/fiorix/go-diameter/diam/dict"
	"github.com/fiorix/go-diameter/diam/sm"
	"github.com/wingyplus/diameter_spike/diameter/dictionary"
)

type Session struct {
	ID      string
	OutChan chan Data
	Request Request
}

type Data struct {
	Conn     diam.Conn
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
	// dtnAddr = "10.89.90.88:3868"
)

func BackgroundClient() chan Session {
	in := make(chan Session, 1000)

	dict.Default.Load(bytes.NewBufferString(dictionary.AppDictionary))
	dict.Default.Load(bytes.NewBufferString(dictionary.CreditControlDictionary))

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
		fmt.Println("DWA Answer")
		fmt.Println(m)
	})

	ccadone := make(chan Data)
	diam.HandleFunc("CCA", func(c diam.Conn, m *diam.Message) {
		fmt.Println("CCA Answer")
		fmt.Println(m)
		ccadone <- Data{Response: &Response{Message: m}}
	})

	// TODO: answer watchdog from server.
	diam.HandleFunc("ALL", func(c diam.Conn, m *diam.Message) {
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

	go func() {
		for {
			select {
			case sess := <-in:
				sendCCR(conn, cfg, sess.Request)
				select {
				case d := <-ccadone:
					sess.OutChan <- d
				}
			case err := <-diam.ErrorReports():
				fmt.Println(err)
			}
		}
	}()

	return in
}

type Request interface {
	AVP() []*diam.AVP
}

const (
	BalanceInformation  = 21100
	AccessMethod        = 20340
	AccountQueryMethod  = 20346
	SSPTime             = 20386
	CallingPartyAddress = 20336
)

func sendCCR(conn diam.Conn, cfg *sm.Settings, req Request) error {
	var (
		// balanceExchange uint32 = 272
		appID uint32 = 4
	)

	m := diam.NewRequest(diam.CreditControl, appID, nil)

	sessionID := fmt.Sprintf("dtac.co.th;OMR%s", time.Now().Format("200601021504050000"))
	m.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String(sessionID))

	m.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(4))
	m.NewAVP(avp.DestinationRealm, avp.Mbit, 0, datatype.OctetString("www.huawei.com"))
	m.NewAVP(avp.OriginHost, avp.Mbit, 0, cfg.OriginHost)
	m.NewAVP(avp.OriginRealm, avp.Mbit, 0, cfg.OriginRealm)

	m.NewAVP(avp.CCRequestType, avp.Mbit, 0, datatype.Integer32(4))
	m.NewAVP(avp.ServiceContextID, avp.Mbit, 0, datatype.UTF8String("QueryBalance@huawei.com"))
	m.NewAVP(avp.RequestedAction, avp.Mbit, 0, datatype.Integer32(2))
	m.NewAVP(avp.EventTimestamp, avp.Mbit, 0, datatype.Time(time.Now()))
	m.NewAVP(avp.ServiceIdentifier, avp.Mbit, 0, datatype.Unsigned32(0))
	m.NewAVP(avp.CCRequestNumber, avp.Mbit, 0, datatype.Unsigned32(0))
	m.NewAVP(avp.RouteRecord, avp.Mbit, 0, datatype.OctetString("10.89.111.40"))
	m.NewAVP(avp.DestinationHost, avp.Mbit, 0, datatype.OctetString("cbp211"))

	for _, avp := range req.AVP() {
		m.AddAVP(avp)
	}

	fmt.Println(m)
	n, err := m.WriteTo(conn)
	if err != nil {
		fmt.Println("error", err.Error())
	}
	fmt.Printf("no error write %d\n", n)
	return err
}

func sendDWR(conn diam.Conn, cfg *sm.Settings) error {
	var (
		watchdogExchange uint32 = 280
		appID            uint32
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
	m.NewAVP(avp.OriginStateID, avp.Mbit, 0, datatype.Unsigned32(0))
	m.NewAVP(avp.SupportedVendorID, avp.Mbit, 0, datatype.Unsigned32(0))
	m.NewAVP(avp.AuthApplicationID, avp.Mbit, 0, datatype.Unsigned32(4))
	m.NewAVP(avp.AcctApplicationID, avp.Mbit, 0, datatype.Unsigned32(4))
	m.NewAVP(avp.FirmwareRevision, avp.Mbit, 0, cfg.FirmwareRevision)

	_, err := m.WriteTo(conn)

	return err
}

func handleCEA(d chan Data) diam.HandlerFunc {
	return func(c diam.Conn, m *diam.Message) {
		d <- Data{Conn: c, Response: &Response{Message: m}}
	}
}
