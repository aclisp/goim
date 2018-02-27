# 客户端长连接协议文档

客户端建立长连接与后端服务通信。目前只支持一种协议 websocket。

## websocket 

### 请求 URL

ws://183.3.211.40:8090/sub

### 协议格式

Websocket（JSON Frame）

**请求json**

```json
{
    "ver": 1,
    "op": 1,
    "seq": 1,
    "body": {}
}
```

**返回json**

```json
[
    {
        "ver": 1,
        "op": 1,
        "seq": 1,
        "body": {}
    }
    // 可以有多个
]
```

**请求和返回参数说明**

| 参数名     | 必选  | 类型 | 说明       |
| :-----     | :---  | :--- | :---       |
| ver        | true  | int | 协议版本号（目前为1) |
| op         | true  | int | 指令编号 |
| seq        | true  | int | 序列号（服务端返回和客户端发送一一对应，广播总是0） |
| body       | true  | 可变 | 与指令有关 |

### 指令编号

| 指令编号     | 说明  | 
| :-----     | :---  |
| 2 | 心跳请求 |
| 3 | 心跳响应 |
| 4 | 业务上行消息 |
| 5 | 业务下行消息 |
| 7 | 认证    |
| 8 | 认证返回 |
|15 | 切换房间 |
|16 | 切换房间响应 |

### 流程

1. 客户端向请求URL建立websocket连接
1. 客户端在onopen事件里，发送认证op=7，body为string：授权令牌，用于检验获取appid|uid|roomid
1. 服务端认证失败，会关闭连接
1. 客户端在onmessage事件里，
    * 收到认证返回op=8，定期发送心跳op=2
    * 收到心跳响应op=3，不处理
    * 收到下行消息op=5，
        * 如果seq=0，为服务端主动推送的消息（广播，组播，单播）
        * 如果seq!=0，为客户端发送请求对应的响应消息
    * 收到下行消息op=16，表示切换房间是否成功
1. 客户端在认证成功后，可以发送心跳op=2，上行消息op=4，切换房间op=15
1. 客户端在onclose事件里，重新建立websocket连接

### 认证消息格式

上行 `body` 暂时为明文编码的string: `appid|uid|roomid`，其中`roomid=-1`表示不进入任何房间

```
上行 {"ver":1,"op":7,"seq":1,"body":"10001|88889999|-1"}
下行 [{"ver":1,"op":8,"seq":1,"body":{}}]
```

### 心跳消息格式

心跳消息发送间隔不能超过5分钟

```
上行 {"ver":1,"op":2,"seq":2,"body":{}}
下行 [{"ver":1,"op":3,"seq":2,"body":{}}]
```

### 业务消息格式

目前业务消息转发到YYTars框架后端

上行 `req` 为请求参数用protobuf序列化之后的base64编码，
下行 `body` 为应答参数用protobuf序列化之后的base64编码 

```
上行 {"ver":1,"op":4,"seq":4,"body":{"obj":"YYLiteApp.AttentionSrv.AttentionObj@tcp -h 58.215.138.213 -p 22018 -t 6000","func":"QueryUserAttentionList","req":"CNvxnu4EEhBtb2JBdHRlbnRpb25MaXRl"}}
下行 [{"ver":1,"op":5,"seq":4,"body":"ENvxnu4EGhBtb2JBdHRlbnRpb25MaXRlIhoI+K3WBxITMjAxOC0wMS0wMiAxNjoyNDowMSIbCJ32hIICEhMyMDE4LTAxLTAzIDE3OjU0OjEwIhsIq4ri2QMSEzIwMTgtMDEtMDMgMTk6NDA6MjUqDAoHdXNldGltZRIBMA=="}]
```

### 切换房间消息格式

上行 `body` 暂时为明文编码的string: `appid|roomid`，其中`roomid=-1`表示不进入任何房间

```
上行 {"ver":1,"op":15,"seq":7,"body":"10001|20"}
下行 [{"ver":1,"op":16,"seq":7,"body":{"ret":0,"msg":"change roomid -1-\u003e2815031242083270676 ok"}}]
```

### 推送消息格式

推送消息只有下行，其中 `body` 为业务自定义，一般是json

```
下行 [{"ver":0,"op":5,"seq":0,"body":{"test":12345}}]
```

## 触发推送消息

业务向推送 URL 发送 POST http

### 推送 URL

http://183.3.211.40:7172

### 房间推送

```
curl -d "{\"test\": 12345}" 'http://183.3.211.40:7172/1/push/room?rid=20&appid=10001'
``` 

### 广播

```
curl -d "{\"test\": 12345}" 'http://183.3.211.40:7172/1/push/all'
```

### 单人推送

```
curl -d "{\"test\":12345}" 'http://183.3.211.40:7172/1/push?uid=88889999&appid=10001'
```

### 单消息多人推送

```
curl -d "{\"u\":[88889999,88889000],\"m\":{\"test\":12},\"a\":10001}" 'http://183.3.211.40:7172/1/pushs'
```
