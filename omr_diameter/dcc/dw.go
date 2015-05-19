package dcc

import (
	"io"

	"github.com/fiorix/go-diameter/diam"
	"github.com/fiorix/go-diameter/diam/avp"
	"github.com/fiorix/go-diameter/diam/datatype"
)

func serverHandleDWR(errc chan error) diam.HandlerFunc {
	type DWR struct {
		OriginHost  string `avp:"Origin-Host"`
		OriginRealm string `avp:"Origin-Realm"`
	}
	return func(c diam.Conn, m *diam.Message) {
		var dwr DWR
		err := m.Unmarshal(&dwr)
		if err != nil {
			errc <- err
			return
		}

		a := m.Answer(diam.Success)
		_, err = sendDWA(c, a)
		if err != nil {
			errc <- err
		}
	}
}

func sendDWA(w io.Writer, m *diam.Message) (int64, error) {
	m.NewAVP(avp.OriginHost, avp.Mbit, 0, datatype.OctetString("srv"))
	m.NewAVP(avp.OriginRealm, avp.Mbit, 0, datatype.OctetString("localhost"))

	return m.WriteTo(w)
}
