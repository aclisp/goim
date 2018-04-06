package main

import (
	"testing"
)

func TestReadHostInfo(t *testing.T) {
	hInfo, err := readHostInfo("/Users/huang/Temp/yyms.json")
	if err != nil {
		t.Error(err)
	}
	t.Logf("%+v\n", hInfo)
	t.Logf("%v", buildAdvertisedAddrs(hInfo))
}
