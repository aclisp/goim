(function(win) {
    var Client = function(options) {
        var MAX_CONNECT_TIME = 10;
        var DELAY = 15000;
        this.options = options || {};
        this.wsocket = null;
        this.callid = 0;
        this.pending = {};
        this._createConnect(MAX_CONNECT_TIME, DELAY);
    };

    Client.prototype._nextCallid = function() {
        if (this.callid === Number.MAX_SAFE_INTEGER) {
            this.callid = 1;
        } else {
            this.callid = 1 + this.callid;
        }
        return this.callid;
    };

    Client.prototype._createConnect = function(max, delay) {
        var self = this;
        if (max === 0) {
            return;
        }
        connect();

        var heartbeatInterval;

        function connect() {
            self.wsocket = new WebSocket('ws://localhost:8191/sub');
            var ws = self.wsocket;
            var auth = false;

            ws.onopen = function() {
                getAuth();
            };

            ws.onmessage = function(evt) {
                var receives = JSON.parse(evt.data);
                for(var i=0; i<receives.length; i++) {
                    var data = receives[i];
                    if (data.op === 8) { // OP_AUTH_REPLY 收到认证回包
                        // TODO 需要检查认证结果
                        auth = true;
                        heartbeat();
                        heartbeatInterval = setInterval(heartbeat, 30 * 1000);
                    }
                    if (!auth) {
                        setTimeout(getAuth, delay);
                    }
                    if (auth && data.op === 5) { // OP_SEND_SMS_REPLY 收到业务推送
                        var notify = self.options.notify;
                        if (notify && data.seq === 0) { // 业务主动发起推送包 seq 一定是 0
                            notify(data.body);
                        }
                    }
                    if (auth && data.op === 16) { // OP_ROOM_CHANGE_REPLY 收到进房回包
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

            function heartbeat() { // 发送心跳包
                ws.send(JSON.stringify({
                    'ver': 1,
                    'op': 2, // OP_HEARTBEAT
                    'seq': self._nextCallid(),
                    'body': {}
                }));
            }

            function getAuth() { // 发送认证包
                var body = [self.options.appid, self.options.uid, self.options.roomid].join('|');
                if (self.options.ticket) {
                    body += '|';
                    body += self.options.ticket;
                }
                ws.send(JSON.stringify({
                    'ver': 1,
                    'op': 7, // OP_AUTH
                    'seq': self._nextCallid(),
                    'body': body
                }));
            }

        }

        function reConnect() {
            self._createConnect(--max, delay * 2);
        }
    };

    /**
     * 切换房间
     */
    Client.prototype.switchRoom = function(roomid, callback) {
        var seq = this._nextCallid();
        this.wsocket.send(JSON.stringify({
            'ver': 1,
            'op': 15, // OP_ROOM_CHANGE
            'seq': seq,
            'body': this.options.appid + '|' + roomid
        }));
        this.pending[seq] = {
            callback: callback
        }
    };

    win['IMClient'] = Client;
})(window);

function bufferToString(buffer) {
    return btoa(String.fromCharCode.apply(null, Array.from(buffer)));
}

function stringToBuffer(string) {
    return new Uint8Array(atob(string).split('').map(function(c) { return c.charCodeAt(0); }));
}
