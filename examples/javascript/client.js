(function(win) {
    var Client = function(options) {
        var MAX_CONNECT_TIME = 10;
        var DELAY = 15000;
        this.options = options || {};
        this.wsocket = null;
        this.createConnect(MAX_CONNECT_TIME, DELAY);
    };

    Client.prototype.createConnect = function(max, delay) {
        var self = this;
        if (max === 0) {
            return;
        }
        connect();

        var heartbeatInterval;

        function connect() {
            self.wsocket = new WebSocket('ws://183.3.211.40:8090/sub');
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
                        heartbeatInterval = setInterval(heartbeat, 4 * 60 * 1000);
                    }
                    if (!auth) {
                        setTimeout(getAuth, delay);
                    }
                    if (auth && data.op === 5) {
                        var notify = self.options.notify;
                        if(notify) notify(data.body, data.seq);
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
                    'seq': 2,
                    'body': {}
                }));
            }

            function getAuth() {
                ws.send(JSON.stringify({
                    'ver': 1,
                    'op': 7,
                    'seq': 1,
                    'body': self.options.uid
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

    win['TarsClient'] = Client;
})(window);

function bufferToString(buffer) {
    return Array.from(buffer).map(function (x) {
        return ('00' + x.toString(16)).slice(-2)
    }).join('');
}
function stringToBuffer(string) {
    var buffer = new Uint8Array(string.length/2);
    for (var i=0; i < string.length/2; i++) {
        buffer[i] = parseInt(string.substr(i*2, 2), 16)
    }
    return buffer
}