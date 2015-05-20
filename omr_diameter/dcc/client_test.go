package dcc

import (
	"testing"
	"time"

	"github.com/fiorix/go-diameter/diam"
	"github.com/wingyplus/diameter_spike/omr_diameter/dcc/servtest"
)

var dwach = make(chan *diam.Message)

func TestClientCallCERAndDWR(t *testing.T) {
	srv, errc := servtest.NewServer()
	defer srv.Close()

	client := &DiameterClient{
		Endpoint: srv.Address,
	}
	done, err := client.Run()
	if err != nil {
		t.Error("Cannot connect to server")
		return
	}
	for i := 0; i < 2; {
		select {
		case err := <-errc:
			t.Error(err)
			return
		case <-done:
			i++
		}
	}
}

func TestClientCallCCR(t *testing.T) {
	srv, errc := servtest.NewServer()
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

		id := diamResponse.SessionID
		if id != expected {
			t.Errorf("expect %s but got %s", expected, id)
		}
	case err := <-errc:
		t.Error(err)
	}
}

// TODO: DWR has a problem, It's hangup when server send DWR and client send DWA back.
//
// func TestClient_DWR(t *testing.T) {
// 	srv, errc := servtest.NewServer()
// 	defer srv.Close()
//
// 	client := &DiameterClient{
// 		Endpoint: srv.Address,
// 	}
// 	done, err := client.Run()
// 	if err != nil {
// 		t.Error("Cannot connect to server")
// 		return
// 	}
// 	select {
// 	case err := <-errc:
// 		t.Error(err)
// 		return
// 	case <-done:
// 	}
//
// 	err = srv.sendDWR()
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	<-done
// 	<-srv.ReceiveDWA()
// }
