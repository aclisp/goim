package main

type BilinAuther struct {
	BypassAuther
}

func NewBilinAuther() Auther {
	return &BilinAuther{}
}

func (a *BilinAuther) Auth(body []byte) (userId int64, roomId int64, err error) {
	return a.BypassAuther.Auth(body)
}
