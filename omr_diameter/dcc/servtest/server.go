package servtest

import (
	"bytes"
	"fmt"

	"github.com/fiorix/go-diameter/diam"
	"github.com/fiorix/go-diameter/diam/avp"
	"github.com/fiorix/go-diameter/diam/datatype"
	"github.com/fiorix/go-diameter/diam/diamtest"
	"github.com/fiorix/go-diameter/diam/dict"
	"github.com/wingyplus/diameter_spike/diameter/dictionary"
)

type Server struct {
	*diamtest.Server

	conn  diam.Conn
	dwach chan struct{}
}

func (s *Server) sendDWR() error {
	var appID uint32
	watchdogExchange := uint32(280)

	m := diam.NewRequest(watchdogExchange, appID, nil)
	m.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.DiameterIdentity("local"))
	m.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.DiameterIdentity("srv"))
	_, err := m.WriteTo(s.conn)
	return err
}

func (s *Server) ReceiveDWA() chan struct{} {
	return s.dwach
}

func NewServer() (*Server, chan error) {
	errc := make(chan error, 1)

	dict.Default.Load(bytes.NewBufferString(dictionary.AppDictionary))
	dict.Default.Load(bytes.NewBufferString(dictionary.CreditControlDictionary))

	serv := &Server{
		dwach: make(chan struct{}),
	}

	smux := diam.NewServeMux()
	smux.Handle("CER", serverHandleCER(errc, serv))
	smux.Handle("CCR", serverHandleCCR(errc))
	smux.Handle("DWR", serverHandleDWR(errc))
	smux.Handle("DWA", serverHandleDWA(errc, serv.dwach))

	serv.Server = diamtest.NewServer(smux, dict.Default)
	return serv, errc
}

func serverHandleDWA(errc chan error, done chan struct{}) diam.HandlerFunc {
	return func(c diam.Conn, m *diam.Message) {
		fmt.Println(m)
		done <- struct{}{}
	}
}
