package main

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/fiorix/go-diameter/diam"
	"github.com/fiorix/go-diameter/diam/avp"
	"github.com/fiorix/go-diameter/diam/datatype"
	"github.com/fiorix/go-diameter/diam/dict"
	"github.com/fiorix/go-diameter/diam/sm"
	"github.com/fiorix/go-diameter/diam/sm/smpeer"
)

var helloDictionary = xml.Header + `
<diameter>
	<application id="999">
		<command code="111" short="HM" name="Hello-Message">
			<request>
				<rule avp="Session-Id" required="true" max="1"/>
				<rule avp="Origin-Host" required="true" max="1"/>
				<rule avp="Origin-Realm" required="true" max="1"/>
				<rule avp="Destination-Realm" required="true" max="1"/>
				<rule avp="Destination-Host" required="true" max="1"/>
				<rule avp="User-Name" required="false" max="1"/>
			</request>
			<answer>
				<rule avp="Session-Id" required="true" max="1"/>
				<rule avp="Result-Code" required="true" max="1"/>
				<rule avp="Origin-Host" required="true" max="1"/>
				<rule avp="Origin-Realm" required="true" max="1"/>
				<rule avp="Error-Message" required="false" max="1"/>
			</answer>
		</command>
	</application>
</diameter>
`

type Query struct {
	in chan session
}

type session struct {
	id  string
	out chan data
}

func (q *Query) Handler(w rest.ResponseWriter, r *rest.Request) {
	id := r.PathParam("id")
	out := make(chan data)
	q.in <- session{id: id, out: out}
	select {
	case d := <-out:
		if d.err != nil {
			w.WriteJson(d.err)
			return
		}
		w.WriteJson(d.response)
	}
}

func main() {
	inCh := backgroundClient()

	q := &Query{inCh}

	router, _ := rest.MakeRouter(rest.Get("/q/:id", q.Handler))

	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	api.SetApp(router)

	http.ListenAndServe(":8080", api.MakeHandler())
}

type data struct {
	response response
	err      error
}

func backgroundClient() chan session {
	in := make(chan session, 1000)

	dict.Default.Load(bytes.NewBufferString(helloDictionary))
	cfg := &sm.Settings{
		OriginHost:       datatype.DiameterIdentity("client"),
		OriginRealm:      datatype.DiameterIdentity("go-diameter"),
		VendorID:         13,
		ProductName:      "go-diameter",
		FirmwareRevision: 1,
	}

	done := make(chan data, 1000)
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
					sess.out <- data{err: err}
				}

				select {
				case d := <-done:
					sess.out <- data{response: d.response, err: err}
				}

			}
		}
	}()
	return in
}

func sendHMR(conn diam.Conn, cfg *sm.Settings, sess session) error {
	meta, ok := smpeer.FromContext(conn.Context())
	if !ok {
		return errors.New("peer metadata unavailable")
	}
	sid := fmt.Sprintf("session;%s", sess.id)

	m := diam.NewRequest(111, 999, nil)
	m.NewAVP(avp.SessionID, avp.Mbit, 0, datatype.UTF8String(sid))
	m.NewAVP(avp.OriginHost, avp.Mbit, 0, cfg.OriginHost)
	m.NewAVP(avp.OriginRealm, avp.Mbit, 0, cfg.OriginRealm)
	m.NewAVP(avp.DestinationRealm, avp.Mbit, 0, meta.OriginRealm)
	m.NewAVP(avp.DestinationHost, avp.Mbit, 0, meta.OriginHost)
	m.NewAVP(avp.UserName, avp.Mbit, 0, datatype.UTF8String("foobar"))
	_, err := m.WriteTo(conn)

	return err
}

type response struct {
	SessionID datatype.UTF8String `avp:"Session-Id"`
}

func handleHMA(done chan data) diam.HandlerFunc {
	return func(c diam.Conn, m *diam.Message) {
		// log.Printf("Received HMA from %s\n%s", c.RemoteAddr(), m)
		var resp response
		m.Unmarshal(&resp)
		done <- data{response: resp}
	}
}
