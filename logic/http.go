package main

import (
	"encoding/json"
	inet "goim/libs/net"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"time"

	log "github.com/thinkboy/log4go"
	"goim/libs/proto"
)

func InitHTTP() (err error) {
	// http listen
	var network, addr string
	for i := 0; i < len(Conf.HTTPAddrs); i++ {
		httpServeMux := http.NewServeMux()
		httpServeMux.HandleFunc("/1/push", Push)
		httpServeMux.HandleFunc("/1/pushs", Pushs)
		httpServeMux.HandleFunc("/1/push/all", PushAll)
		httpServeMux.HandleFunc("/1/push/room", PushRoom)
		httpServeMux.HandleFunc("/1/server/del", DelServer)
		httpServeMux.HandleFunc("/1/count", Count)
		httpServeMux.HandleFunc("/1/session", Session)
		httpServeMux.HandleFunc("/1/list", List)
		log.Info("start http listen:\"%s\"", Conf.HTTPAddrs[i])
		if network, addr, err = inet.ParseNetwork(Conf.HTTPAddrs[i]); err != nil {
			log.Error("inet.ParseNetwork() error(%v)", err)
			return
		}
		go httpListen(httpServeMux, network, addr)
	}
	return
}

func httpListen(mux *http.ServeMux, network, addr string) {
	httpServer := &http.Server{Handler: mux, ReadTimeout: Conf.HTTPReadTimeout, WriteTimeout: Conf.HTTPWriteTimeout}
	httpServer.SetKeepAlivesEnabled(true)
	l, err := net.Listen(network, addr)
	if err != nil {
		log.Error("net.Listen(\"%s\", \"%s\") error(%v)", network, addr, err)
		panic(err)
	}
	if err := httpServer.Serve(l); err != nil {
		log.Error("server.Serve() error(%v)", err)
		panic(err)
	}
}

// retWrite marshal the result and write to client(get).
func retWrite(w http.ResponseWriter, r *http.Request, res map[string]interface{}, start time.Time) {
	data, err := json.Marshal(res)
	if err != nil {
		log.Error("json.Marshal(\"%v\") error(%v)", res, err)
		return
	}
	dataStr := string(data)
	if _, err := w.Write([]byte(dataStr)); err != nil {
		log.Error("w.Write(\"%s\") error(%v)", dataStr, err)
	}
	log.Debug("req: \"%s\", get: res:\"%s\", ip:\"%s\", time:\"%fs\"", r.URL.String(), dataStr, r.RemoteAddr, time.Now().Sub(start).Seconds())
}

// retPWrite marshal the result and write to client(post).
func retPWrite(w http.ResponseWriter, r *http.Request, res map[string]interface{}, body *string, start time.Time) {
	data, err := json.Marshal(res)
	if err != nil {
		log.Error("json.Marshal(\"%v\") error(%v)", res, err)
		return
	}
	dataStr := string(data)
	if _, err := w.Write([]byte(dataStr)); err != nil {
		log.Error("w.Write(\"%s\") error(%v)", dataStr, err)
	}
	log.Debug("req: \"%s\", post: \"%s\", res:\"%s\", ip:\"%s\", time:\"%fs\"", r.URL.String(), *body, dataStr, r.RemoteAddr, time.Now().Sub(start).Seconds())
}

func Push(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", 405)
		return
	}
	var (
		body      string
		serverId  int32
		keys      []string
		subKeys   map[int32][]string
		bodyBytes []byte
		userId    int64
		appId     int64
		err       error
		uidStr    = r.URL.Query().Get("uid")
		appidStr  = r.URL.Query().Get("appid")
		res       = map[string]interface{}{"ret": OK}
	)
	defer retPWrite(w, r, res, &body, time.Now())
	if bodyBytes, err = ioutil.ReadAll(r.Body); err != nil {
		log.Error("ioutil.ReadAll() failed (%s)", err)
		res["ret"] = InternalErr
		return
	}
	body = string(bodyBytes)
	if userId, err = strconv.ParseInt(uidStr, 10, 48); err != nil {
		log.Error("strconv.Atoi(\"%s\") error(%v)", uidStr, err)
		res["ret"] = InternalErr
		return
	}
	if appId, err = strconv.ParseInt(appidStr, 10, 16); err != nil {
		log.Error("strconv.Atoi(\"%s\") error(%v)", appidStr, err)
		res["ret"] = InternalErr
		return
	}
	userId = (appId << 48) | userId
	subKeys = genSubKey(userId)
	for serverId, keys = range subKeys {
		if err = mpushKafka(serverId, keys, bodyBytes); err != nil {
			res["ret"] = InternalErr
			return
		}
	}
	res["ret"] = OK
	return
}

type pushsBodyMsg struct {
	Msg     json.RawMessage `json:"m"`
	UserIds []int64         `json:"u"`
	AppId   int16           `json:"a"`
}

func parsePushsBody(body []byte) (msg []byte, userIds []int64, err error) {
	tmp := pushsBodyMsg{}
	if err = json.Unmarshal(body, &tmp); err != nil {
		return
	}
	msg = tmp.Msg
	for _, userId := range tmp.UserIds {
		userIds = append(userIds, (int64(tmp.AppId)<<48)|userId)
	}
	return
}

// {"m":{"test":1},"u":"1,2,3","a":1}
func Pushs(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", 405)
		return
	}
	var (
		body      string
		bodyBytes []byte
		serverId  int32
		userIds   []int64
		err       error
		res       = map[string]interface{}{"ret": OK}
		subKeys   map[int32][]string
		keys      []string
	)
	defer retPWrite(w, r, res, &body, time.Now())
	if bodyBytes, err = ioutil.ReadAll(r.Body); err != nil {
		log.Error("ioutil.ReadAll() failed (%s)", err)
		res["ret"] = InternalErr
		return
	}
	body = string(bodyBytes)
	if bodyBytes, userIds, err = parsePushsBody(bodyBytes); err != nil {
		log.Error("parsePushsBody(\"%s\") error(%s)", body, err)
		res["ret"] = InternalErr
		return
	}
	subKeys = genSubKeys(userIds)
	for serverId, keys = range subKeys {
		if err = mpushKafka(serverId, keys, bodyBytes); err != nil {
			res["ret"] = InternalErr
			return
		}
	}
	res["ret"] = OK
	return
}

func PushRoom(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", 405)
		return
	}
	var (
		bodyBytes []byte
		body      string
		rid       int64
		appid     int64
		err       error
		param     = r.URL.Query()
		res       = map[string]interface{}{"ret": OK}
	)
	defer retPWrite(w, r, res, &body, time.Now())
	if bodyBytes, err = ioutil.ReadAll(r.Body); err != nil {
		log.Error("ioutil.ReadAll() failed (%v)", err)
		res["ret"] = InternalErr
		return
	}
	body = string(bodyBytes)
	ridStr := param.Get("rid")
	enable, _ := strconv.ParseBool(param.Get("ensure"))
	appidStr := param.Get("appid")
	// push room
	if rid, err = strconv.ParseInt(ridStr, 10, 48); err != nil {
		log.Error("strconv.Atoi(\"%s\") error(%v)", ridStr, err)
		res["ret"] = InternalErr
		return
	}
	if appid, err = strconv.ParseInt(appidStr, 10, 16); err != nil {
		log.Error("strconv.Atoi(\"%s\") error(%v)", appidStr, err)
		res["ret"] = InternalErr
		return
	}
	rid = (appid << 48) | rid
	if err = broadcastRoomKafka(rid, bodyBytes, enable); err != nil {
		log.Error("broadcastRoomKafka(\"%s\",\"%s\",\"%d\") error(%s)", rid, body, enable, err)
		res["ret"] = InternalErr
		return
	}
	res["ret"] = OK
	return
}

func PushAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", 405)
		return
	}
	var (
		bodyBytes []byte
		body      string
		err       error
		res       = map[string]interface{}{"ret": OK}
	)
	defer retPWrite(w, r, res, &body, time.Now())
	if bodyBytes, err = ioutil.ReadAll(r.Body); err != nil {
		log.Error("ioutil.ReadAll() failed (%v)", err)
		res["ret"] = InternalErr
		return
	}
	body = string(bodyBytes)
	// push all
	if err := broadcastKafka(bodyBytes); err != nil {
		log.Error("broadcastKafka(\"%s\") error(%s)", body, err)
		res["ret"] = InternalErr
		return
	}
	res["ret"] = OK
	return
}

type RoomCounter struct {
	RoomId int64
	Count  int32
}

type ServerCounter struct {
	Server int32
	Count  int32
}

func Count(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", 405)
		return
	}
	var (
		typeStr = r.URL.Query().Get("type")
		res     = map[string]interface{}{"ret": OK}
	)
	defer retWrite(w, r, res, time.Now())
	if typeStr == "room" {
		d := make([]*RoomCounter, 0, len(RoomCountMap))
		for roomId, count := range RoomCountMap {
			d = append(d, &RoomCounter{RoomId: roomId, Count: count})
		}
		res["data"] = d
	} else if typeStr == "server" {
		d := make([]*ServerCounter, 0, len(ServerCountMap))
		for server, count := range ServerCountMap {
			d = append(d, &ServerCounter{Server: server, Count: count})
		}
		res["data"] = d
		m, _ := allServerInfo()
		res["meta"] = m
	}
	return
}

func Session(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", 405)
		return
	}
	var (
		session   *proto.UserSession
		err       error
		userId    int64
		uidStr    = r.URL.Query().Get("uid")
		res       = map[string]interface{}{"ret": OK}
	)
	defer retWrite(w, r, res, time.Now())
	if userId, err = strconv.ParseInt(uidStr, 10, 64); err != nil {
		log.Error("strconv.Atoi(\"%s\") error(%v)", uidStr, err)
		res["ret"] = InternalErr
		return
	}
	if session, err = userSession(userId); err != nil {
		res["ret"] = InternalErr
		return
	}
	res["session"] = session
	return
}

func List(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", 405)
		return
	}
	type Session struct {
		UserId int64
		Seq    int32
		Comet  int32
	}
	var (
		res       = map[string]interface{}{"ret": OK}
		nodes     []Sessions
		data      = map[string][]Session{}
	)
	defer retWrite(w, r, res, time.Now())
	nodes, _ = listUserSession()
	for _, node := range nodes {
		total := 0
		if len(node.userIds) > 0 {
			total = len(node.userIds) * len(node.seqs[0])
		}
		sessions := make([]Session, 0, total)
		for i := range node.userIds {
			for j := range node.seqs[i] {
				sessions = append(sessions, Session{
					UserId: node.userIds[i],
					Seq:    node.seqs[i][j],
					Comet:  node.servers[i][j],
				})
			}
		}
		data[node.node] = sessions
	}
	res["nodes"] = data
	return
}

func DelServer(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", 405)
		return
	}
	var (
		err       error
		serverStr = r.URL.Query().Get("server")
		server    int64
		res       = map[string]interface{}{"ret": OK}
	)
	if server, err = strconv.ParseInt(serverStr, 10, 32); err != nil {
		log.Error("strconv.Atoi(\"%s\") error(%v)", serverStr, err)
		res["ret"] = InternalErr
		return
	}
	defer retWrite(w, r, res, time.Now())
	if err = delServer(int32(server)); err != nil {
		res["ret"] = InternalErr
		return
	}
	return
}
