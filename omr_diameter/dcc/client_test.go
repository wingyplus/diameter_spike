package dcc

import (
	"bytes"
	"testing"
	"time"

	"github.com/fiorix/go-diameter/diam"
	"github.com/fiorix/go-diameter/diam/diamtest"
	"github.com/fiorix/go-diameter/diam/dict"
	"github.com/wingyplus/diameter_spike/diameter/dictionary"
)

func newServer() (*diamtest.Server, chan error) {
	errc := make(chan error, 1)

	dict.Default.Load(bytes.NewBufferString(dictionary.AppDictionary))
	dict.Default.Load(bytes.NewBufferString(dictionary.CreditControlDictionary))

	smux := diam.NewServeMux()
	smux.Handle("CER", serverHandleCER(errc))
	smux.Handle("CCR", serverHandleCCR(errc))
	smux.Handle("DWR", serverHandleDWR(errc))

	return diamtest.NewServer(smux, dict.Default), errc
}

func TestClientCallCERAndDWR(t *testing.T) {
	srv, errc := newServer()
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
	srv, errc := newServer()
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
