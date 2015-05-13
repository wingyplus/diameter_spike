package main

import (
	"net"

	"github.com/fiorix/go-diameter/diam"
	"github.com/fiorix/go-diameter/diam/avp"
	"github.com/fiorix/go-diameter/diam/datatype"
	"github.com/fiorix/go-diameter/diam/sm"
)

type DiameterClient struct {
	Endpoint string
}

const (
	identity    = datatype.DiameterIdentity("jenkin13_OMR_TEST01")
	realm       = datatype.DiameterIdentity("dtac.co.th")
	vendorID    = datatype.Unsigned32(0)
	productName = datatype.UTF8String("omr")
)

type Status struct {
	Error error
}

func (client *DiameterClient) Run() (chan struct{}, error) {
	done := make(chan struct{})

	cfg := &sm.Settings{
		OriginHost:       identity,
		OriginRealm:      realm,
		VendorID:         vendorID,
		ProductName:      productName,
		FirmwareRevision: 1,
	}

	diam.HandleFunc("CEA", handleCEA(done))
	conn, err := diam.Dial(client.Endpoint, nil, nil)
	if err != nil {
		return nil, err
	}
	err = sendCER(conn, cfg)
	if err != nil {
		return nil, err
	}
	return done, nil
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

func handleCEA(done chan struct{}) diam.HandlerFunc {
	return func(conn diam.Conn, m *diam.Message) {
		done <- struct{}{}
	}
}
