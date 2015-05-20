package dcc

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

var in = make(chan Request)

type DiameterClient struct {
	Endpoint string
}

type Request struct {
	out  chan Response
	data Data
}

type Data interface {
	AVP() []*diam.AVP
}

type Response interface {
	Unmarshal(v interface{}) error
}

const (
	identity    = datatype.DiameterIdentity("jenkin13_OMR_TEST01")
	realm       = datatype.DiameterIdentity("dtac.co.th")
	vendorID    = datatype.Unsigned32(0)
	productName = datatype.UTF8String("omr")
)

type Answer struct {
	*diam.Message
	Error error
}

func (client *DiameterClient) Run() (chan struct{}, error) {
	done := make(chan struct{})

	dict.Default.Load(bytes.NewBufferString(dictionary.AppDictionary))
	dict.Default.Load(bytes.NewBufferString(dictionary.CreditControlDictionary))

	cfg := &sm.Settings{
		OriginHost:       identity,
		OriginRealm:      realm,
		VendorID:         vendorID,
		ProductName:      productName,
		FirmwareRevision: 1,
	}

	ccadone := make(chan Answer)

	smux := diam.NewServeMux()
	smux.Handle("CEA", handleCEA(done))
	smux.Handle("DWA", handleDWA(done))
	smux.Handle("CCA", handleCCA(ccadone))
	smux.HandleFunc("DWR", handleDWR(done))

	conn, err := diam.Dial(client.Endpoint, smux, nil)
	if err != nil {
		return nil, err
	}
	err = sendCER(conn, cfg)
	if err != nil {
		return nil, err
	}

	err = sendDWR(conn, cfg)
	if err != nil {
		return nil, err
	}

	go func() {
		request := <-in
		sendCCR(conn, cfg, request)
		if err != nil {
			fmt.Println(err)
		}
		select {
		case answer := <-ccadone:
			request.out <- answer
		case err := <-diam.ErrorReports():
			fmt.Println(err)
		}
	}()

	return done, nil
}

func sendCER(conn diam.Conn, cfg *sm.Settings) error {
	m := diam.NewRequest(CapabilitiesExchange, AppID, nil)

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

func sendDWR(conn diam.Conn, cfg *sm.Settings) error {
	m := diam.NewRequest(WatchdogExchange, AppID, nil)
	m.NewAVP(avp.OriginHost, avp.Mbit, 0, cfg.OriginHost)
	m.NewAVP(avp.OriginRealm, avp.Mbit, 0, cfg.OriginRealm)

	_, err := m.WriteTo(conn)
	return err
}

func handleCEA(done chan struct{}) diam.HandlerFunc {
	return func(conn diam.Conn, m *diam.Message) {
		done <- struct{}{}
	}
}

func handleDWA(done chan struct{}) diam.HandlerFunc {
	return func(conn diam.Conn, m *diam.Message) {
		done <- struct{}{}
	}
}

func handleDWR(done chan struct{}) diam.HandlerFunc {
	return func(conn diam.Conn, m *diam.Message) {
		fmt.Println("handle dwr")
		fmt.Println(m)

		m.Answer(diam.Success)
		m.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.OctetString("client"))
		m.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.OctetString("localhost"))

		fmt.Println("send connection")
		m.WriteTo(conn)
		fmt.Println("after send connection")
		done <- struct{}{}
	}
}

func sendCCR(conn diam.Conn, cfg *sm.Settings, req Request) error {
	var appID uint32 = 4

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

	for _, avp := range req.data.AVP() {
		m.AddAVP(avp)
	}
	_, err := m.WriteTo(conn)
	return err
}

func handleCCA(ccadone chan Answer) diam.HandlerFunc {
	return func(conn diam.Conn, m *diam.Message) {
		ccadone <- Answer{Message: m}
	}
}
