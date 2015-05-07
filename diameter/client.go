package diameter

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/fiorix/go-diameter/diam"
	"github.com/fiorix/go-diameter/diam/avp"
	"github.com/fiorix/go-diameter/diam/datatype"
	"github.com/fiorix/go-diameter/diam/dict"
	"github.com/fiorix/go-diameter/diam/sm"
	"github.com/fiorix/go-diameter/diam/sm/smpeer"
	"github.com/wingyplus/diameter_spike/diameter/dictionary"
)

type Session struct {
	ID      string
	OutChan chan Data
}

type Data struct {
	Response Response
	Err      error
}

type Response struct {
	SessionID datatype.UTF8String `avp:"Session-Id"`
}

func BackgroundClient() chan Session {
	in := make(chan Session, 1000)

	dict.Default.Load(bytes.NewBufferString(dictionary.HelloDictionary))
	cfg := &sm.Settings{
		OriginHost:       datatype.DiameterIdentity("client"),
		OriginRealm:      datatype.DiameterIdentity("go-diameter"),
		VendorID:         13,
		ProductName:      "go-diameter",
		FirmwareRevision: 1,
	}

	done := make(chan Data, 1000)
	mux := sm.New(cfg)
	mux.Handle("HMA", handleHMA(done))

	cli := &sm.Client{
		Dict:               dict.Default,
		Handler:            mux,
		MaxRetransmits:     3,
		RetransmitInterval: time.Second,
		EnableWatchdog:     true,
		WatchdogInterval:   5 * time.Second,
		AcctApplicationID: []*diam.AVP{
			diam.NewAVP(avp.AcctApplicationID, avp.Mbit, 0, datatype.Unsigned32(4)),
			diam.NewAVP(avp.AcctApplicationID, avp.Mbit, 0, datatype.Unsigned32(999)),
		},
	}

	conn, err := cli.Dial("localhost:3868")
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			select {
			case sess := <-in:
				err := sendHMR(conn, cfg, sess)
				if err != nil {
					sess.OutChan <- Data{Err: err}
				}

				select {
				case d := <-done:
					sess.OutChan <- Data{Response: d.Response, Err: err}
				}

			}
		}
	}()
	return in
}

func sendHMR(conn diam.Conn, cfg *sm.Settings, sess Session) error {
	var (
		commandCode uint32 = 111
		appID       uint32 = 999
	)

	meta, ok := smpeer.FromContext(conn.Context())
	if !ok {
		return errors.New("peer metadata unavailable")
	}
	sid := fmt.Sprintf("session;%s", sess.ID)

	m := diam.NewRequest(commandCode, appID, nil)
	m.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String(sid))
	m.NewAVP(avp.OriginHost, avp.Mbit, 0, cfg.OriginHost)
	m.NewAVP(avp.OriginRealm, avp.Mbit, 0, cfg.OriginRealm)
	m.NewAVP(avp.DestinationRealm, avp.Mbit, 0, meta.OriginRealm)
	m.NewAVP(avp.DestinationHost, avp.Mbit, 0, meta.OriginHost)
	m.NewAVP(avp.UserName, avp.Mbit, 0, datatype.UTF8String("foobar"))
	_, err := m.WriteTo(conn)

	return err
}

func handleHMA(done chan Data) diam.HandlerFunc {
	return func(c diam.Conn, m *diam.Message) {
		// log.Printf("Received HMA from %s\n%s", c.RemoteAddr(), m)
		var resp Response
		m.Unmarshal(&resp)
		done <- Data{Response: resp}
	}
}
