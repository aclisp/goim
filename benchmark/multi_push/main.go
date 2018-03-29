package main

// Start Command eg : ./multi_push 0 20000 localhost:7172 60

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"
	pb "github.com/golang/protobuf/proto"
)

var (
	lg         *log.Logger
	httpClient *http.Client
	t          int
)

const TestContent = "{\"test\":1}"

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

type MultiPush struct {
	Msg     *ServerPush `protobuf:"bytes,1,opt,name=msg" json:"msg,omitempty"`
	UserIDs []int64     `protobuf:"varint,2,rep,packed,name=userIDs" json:"userIDs,omitempty"`
	AppID   int32       `protobuf:"varint,3,opt,name=appID" json:"appID,omitempty"`
}

func (m *MultiPush) Reset()                    { *m = MultiPush{} }
func (m *MultiPush) String() string            { return pb.CompactTextString(m) }
func (*MultiPush) ProtoMessage()               {}

func init() {
	httpTransport := &http.Transport{
		Dial: func(netw, addr string) (net.Conn, error) {
			deadline := time.Now().Add(30 * time.Second)
			c, err := net.DialTimeout(netw, addr, 20*time.Second)
			if err != nil {
				return nil, err
			}

			c.SetDeadline(deadline)
			return c, nil
		},
		DisableKeepAlives: false,
	}
	httpClient = &http.Client{
		Transport: httpTransport,
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	infoLogfi, err := os.OpenFile("./multi_push.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	lg = log.New(infoLogfi, "", log.LstdFlags|log.Lshortfile)

	begin, err := strconv.Atoi(os.Args[1])
	if err != nil {
		panic(err)
	}
	length, err := strconv.Atoi(os.Args[2])
	if err != nil {
		panic(err)
	}

	t, err = strconv.Atoi(os.Args[4])
	if err != nil {
		panic(err)
	}

	num := runtime.NumCPU() * 8

	l := length / num
	b, e := begin, begin+l
	time.AfterFunc(time.Duration(t)*time.Second, stop)
	for i := 0; i < num; i++ {
		go startPush(b, e)
		b += l
		e += l
	}
	if b < begin+length {
		go startPush(b, begin+length)
	}

	time.Sleep(9999 * time.Hour)
}

func stop() {
	os.Exit(-1)
}

func startPush(b, e int) {
	l := make([]int64, 0, e-b)
	for i := b; i < e; i++ {
		l = append(l, int64(i))
	}
	msg := &MultiPush{
		Msg: &ServerPush{
			PushBuffer: []byte("abc"),
		},
		UserIDs: l,
	}
	body, err := pb.Marshal(msg)
	if err != nil {
		panic(err)
	}
	for {
		resp, err := httpPost(fmt.Sprintf("http://%s/1/pushs", os.Args[3]), "application/x-www-form-urlencoded", bytes.NewBuffer(body))
		if err != nil {
			lg.Printf("post error (%v)", err)
			continue
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			lg.Printf("post error (%v)", err)
			return
		}
		resp.Body.Close()

		lg.Printf("response %s", string(body))
	}
}

func httpPost(url string, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", contentType)
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
