package main

// Start Commond eg: ./push_room 1 20 localhost:7172
// first parameter: room id
// second parameter: num per seconds
// third parameter: logic server ip

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
	pb "github.com/golang/protobuf/proto"
)

type ServerPush struct {
	MessageType int32             `protobuf:"zigzag32,1,opt,name=messageType" json:"messageType,omitempty"`
	PushBuffer  []byte            `protobuf:"bytes,2,opt,name=pushBuffer,proto3" json:"pushBuffer,omitempty"`
	Headers     map[string]string `protobuf:"bytes,3,rep,name=headers" json:"headers,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	MessageDesc string            `protobuf:"bytes,4,opt,name=messageDesc" json:"messageDesc,omitempty"`
	ServiceName string            `protobuf:"bytes,5,opt,name=serviceName" json:"serviceName,omitempty"`
	MethodName  string            `protobuf:"bytes,6,opt,name=methodName" json:"methodName,omitempty"`
}

func (m *ServerPush) Reset()                    { *m = ServerPush{} }
func (m *ServerPush) String() string            { return pb.CompactTextString(m) }
func (*ServerPush) ProtoMessage()               {}

func main() {
	rountineNum, err := strconv.Atoi(os.Args[2])
	if err != nil {
		panic(err)
	}
	addr := os.Args[3]

	gap := time.Second / time.Duration(rountineNum)
	delay := time.Duration(0)

	go run(addr, time.Duration(0)*time.Second)
	for i := 0; i < rountineNum-1; i++ {
		go run(addr, delay)
		delay += gap
		fmt.Println("delay:", delay)
	}
	time.Sleep(9999 * time.Hour)
}

func run(addr string, delay time.Duration) {
	time.Sleep(delay)
	i := int64(0)
	for {
		go post(addr, i)
		time.Sleep(10 * time.Second)
		i++
	}
}

func post(addr string, i int64) {
	msg := &ServerPush{
		PushBuffer: []byte("abc"),
		ServiceName: "push.test",
		Headers: map[string]string{
			"room": os.Args[1],
		},
	}
	body, err := pb.Marshal(msg)
	if err != nil {
		panic(err)
	}
	resp, err := http.Post("http://"+addr+"/1/push/room?rid="+os.Args[1], "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Println("Error: http.post() error(%s)", err)
		return
	}
	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error: http.post() error(%s)", err)
		return
	}

	fmt.Printf("%s postId:%d, response:%s\n", time.Now().Format("2006-01-02 15:04:05"), i, string(body))
}
