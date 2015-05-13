package main

type QueryBalancer struct{}

func (q *QueryBalancer) Retrieve(number SubscriberNumber) BalanceInformation {
	out := make(chan Response)
	in <- Request{out: out}
	resp := <-out
	var diamResponse DiamResponse
	resp.Unmarshal(&diamResponse)
	return diamResponse.ServiceInformation.BalanceInformation
}

var in chan Request

type Request struct {
	out chan Response
}

type Response interface {
	Unmarshal(v interface{}) error
}

type DiameterClient interface {
	Run() chan Request
}
