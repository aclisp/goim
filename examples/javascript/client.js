(function(win) {
    var Client = function(options) {
        var MAX_CONNECT_TIME = 10;
        var DELAY = 15000;
        this.options = options || {};
        this.wsocket = null;
        this.callid = 0;
        this.pending = {};
        this.createConnect(MAX_CONNECT_TIME, DELAY);
    };

    Client.prototype.nextCallid = function() {
        if (this.callid === Number.MAX_SAFE_INTEGER) {
            this.callid = 1;
        } else {
            this.callid = 1 + this.callid;
        }
        return this.callid;
    };

    Client.prototype.createConnect = function(max, delay) {
        var self = this;
        if (max === 0) {
            return;
        }
        connect();

        var heartbeatInterval;

        function connect() {
            self.wsocket = new WebSocket('wss://test-goim.yy.com/sub');
            var ws = self.wsocket;
            var auth = false;

            ws.onopen = function() {
                getAuth();
            };

            ws.onmessage = function(evt) {
                var receives = JSON.parse(evt.data);
                for(var i=0; i<receives.length; i++) {
                    var data = receives[i];
                    if (data.op === 8) {
                        auth = true;
                        heartbeat();
                        heartbeatInterval = setInterval(heartbeat, 30 * 1000);
                    }
                    if (!auth) {
                        setTimeout(getAuth, delay);
                    }
                    if (auth && data.op === 5) {
                        if (data.seq > 0) {
                            var p = self.pending[data.seq];
                            delete self.pending[data.seq];
                            var rspT = p.rspT;
                            var rspM = rspT.decode(stringToBuffer(data.body.rsp));
                            var rsp = rspT.toObject(rspM);
                            p.callback(rsp);
                        } else {
                            var notify = self.options.notify;
                            if (notify && data.seq === 0) {
                                notify(data.body);
                            }
                        }
                    }
                    if (auth && data.op === 16) {
                        if (data.seq > 0) {
                            var p = self.pending[data.seq];
                            delete self.pending[data.seq];
                            p.callback(data.body.ret, data.body.msg)
                        }
                    }
                }
            };

            ws.onclose = function() {
                if (heartbeatInterval) clearInterval(heartbeatInterval);
                setTimeout(reConnect, delay);
            };

            function heartbeat() {
                ws.send(JSON.stringify({
                    'ver': 1,
                    'op': 2,
                    'seq': self.nextCallid(),
                    'body': {}
                }));
            }

            function getAuth() {
                var body = [self.options.appid, self.options.uid, self.options.roomid].join('|');
                if (self.options.ticket) {
                    body += '|';
                    body += self.options.ticket;
                }
                ws.send(JSON.stringify({
                    'ver': 1,
                    'op': 7,
                    'seq': self.nextCallid(),
                    'body': body
                }));
            }

        }

        function reConnect() {
            self.createConnect(--max, delay * 2);
        }
    };

    Client.prototype.send = function(rpc, seq) {
        this.wsocket.send(JSON.stringify({
            'ver': 1,
            'op': 4,
            'seq': seq,
            'body': rpc
        }));
    };

    Client.prototype.call = function(options) {
        var proto = options.proto;
        var reqtype = options.reqtype;
        var rsptype = options.rsptype;
        var req = options.req;
        var service = options.service;
        var method = options.method;
        var context = options.context;
        var callback = options.callback;

        var reqT = proto.lookupType(reqtype);
        var rspT = proto.lookupType(rsptype);
        var reqM = reqT.fromObject(req);
        var reqB = reqT.encode(reqM).finish();
        var seq = this.nextCallid();
        this.send({
            obj: service,
            func: method,
            req: bufferToString(reqB),
            opt: context
        }, seq);
        this.pending[seq] = {
            rspT: rspT,
            callback: callback
        };
    };

    Client.prototype.join = function(roomid, callback) {
        var seq = this.nextCallid();
        this.wsocket.send(JSON.stringify({
            'ver': 1,
            'op': 15,
            'seq': seq,
            'body': this.options.appid + '|' + roomid
        }));
        this.pending[seq] = {
            callback: callback
        }
    };

    win['TarsClient'] = Client;
})(window);

function bufferToString(buffer) {
    return btoa(String.fromCharCode.apply(null, Array.from(buffer)));
}

function stringToBuffer(string) {
    return new Uint8Array(atob(string).split('').map(function(c) { return c.charCodeAt(0); }));
}
