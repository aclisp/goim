# 客户端长连接协议文档

客户端建立长连接与后端服务通信。协议使用 TCP。

## TCP 

### 请求 IP:Port

test-goim.yy.com:8080

### 协议格式

TCP

**消息头**

```cpp
#pragma pack(push, 1)
struct __STNetMsgXpHeader {
    uint32_t    head_length;    // always == 20
    uint32_t    client_version;
    uint32_t    cmdid;
    uint32_t    seq;
    uint32_t	body_length;
};
#pragma pack(pop)
```

**RPC请求body**

```protobuf
syntax = "proto3";
message RPCInput {
    string serviceName    = 1;
    string methodName     = 2;
    bytes  requestBuffer  = 3;
    map<string, string> headers = 4;
}
```

**RPC返回body**

```protobuf
message RPCOutput {
    sint32 retCode        = 1;
    bytes  responseBuffer = 2;
    map<string, string> headers = 3;
    string retDesc        = 4;
    string serviceName    = 5;
    string methodName     = 6;
}
```

**服务器推送body**
```protobuf
message ServerPush {
    sint32 messageType    = 1;
    bytes  pushBuffer     = 2;
    map<string, string> headers = 3;
    string messageDesc    = 4;
    string serviceName    = 5;
    string methodName     = 6;
}
```

**请求和返回参数说明**

| 参数名     | 必选  | 类型 | 说明       |
| :-----     | :---  | :--- | :---       |
| head_length   | true  | int32 | 总是20 |
| client_version| true  | int32 | 协议版本号（目前为1) |
| cmdid         | true  | int32 | 指令编号 |
| seq           | true  | int32 | 序列号（服务端返回和客户端发送一一对应，广播总是0） |
| body_length   | true  | int32 | body长度 |
| body          | false | bytes | 与指令有关 |

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

1. 客户端向TCP的IP:Port建立长连接
1. 连接建立成功之后，立刻发送认证cmdid=7，body为token授权令牌
1. 服务端认证失败，会关闭连接
1. 客户端，
    * 收到认证返回cmdid=8，定期发送心跳cmdid=2
    * 收到心跳响应cmdid=3，不处理
    * 收到下行消息cmdid=5，
        * 如果seq=0，为服务端主动推送的消息（广播，组播，单播）
        * 如果seq!=0，为客户端发送请求对应的响应消息
    * 收到下行消息cmdid=16，表示切换房间是否成功
1. 客户端在认证成功后，可以发送心跳cmdid=2，上行消息cmdid=4，切换房间cmdid=15
1. 客户端断线，可以重新建立长连接

### 认证消息格式

长连接经过认证之后，才与uid建立对应关系，才能发送消息和接收推送。

`body` 为空，使用匿名认证；
`body` 不为空，需要填写 `RPCInput.headers`

|header key|必选     |类型   |说明        | 
| :-----   | :---   | :--- | :---       |
|uid       | true   | uint | uid==0表示匿名认证，一般用于用户登录之前|
|token     | true   | string|匿名是空，登录之后是账号系统的票据|
|subscribe-room-push|false|int|是否订阅房间推送。-1表示取消订阅，其他值为房间号|

认证返回的 `RPCOutput.retCode` 为0，则通过。否则失败，失败原因在 `RPCOutput.retDesc`。
如果认证失败，服务器会关闭连接。

### 心跳消息格式

心跳消息发送间隔不能超过5分钟

### 业务消息格式

目前业务消息转发到YYTars框架后端

`serviceName` 和 `methodName` 用于后端服务的路由；
`requestBuffer`  是业务消息请求的protobuf编码；
`responseBuffer` 是业务消息应答的protobuf编码。

业务消息带有这些 `RPCInput.headers` 

|header key|必选     |类型   |说明        | 
| :-----   | :---   | :--- | :---       |
|is-anonymous-user|true|bool|是否是匿名用户，即uid==0|
|uid       | false   | uint | 当前连接对应的uid|
|connid    | false   | uint | 区分同一个uid的多条连接（多设备登录） |
|subscribe-room-push|false|int|当前连接订阅的房间推送。-1表示没有订阅，其他值为房间号|
|heartbeat-threshold|false|string|最长心跳间隔，客户端超过这个时间不发心跳，服务端会关闭连接|
|client-ip|false          |string|客户端IP|
|client-port|false        |uint  |客户端PORT|
|access-point-ip|false    |string|接入点IP|
|access-point-port|false  |uint  |接入点PORT|

注意：业务相关headers，推荐用大写，不要覆盖这些key。

### 切换房间消息格式

长连接建立之后，可以订阅、取消订阅房间推送。需要填写 `RPCInput.headers`

|header key|必选     |类型   |说明        | 
| :-----   | :---   | :--- | :---       |
|subscribe-room-push|true|int|-1表示取消订阅，其他值为房间号|

注意：如果 `RPCInput.serviceName` 不为空，则会调用下游服务，用于业务相关的直播间状态管理。

### 推送消息格式

推送消息只有下行，其中 `body` 为 protobuf 编码的 `ServerPush`。业务使用
`ServerPush.messageType` 识别内层的 `pushBuffer`

## 触发推送消息

业务向推送 URL 发送 POST http。
POST body 为 protobuf 编码的 `ServerPush`。
POST Content-Type 为 `application/x-protobuf`

### 推送 URL

http://test-goim.yy.com:7172

### 房间推送

```
curl -d "<protobuf bytes>" 'http://test-goim.yy.com:7172/1/push/room?rid=20'
``` 

### 广播

```
curl -d "<protobuf bytes>" 'http://test-goim.yy.com:7172/1/push/all'
```

### 单人推送

```
curl -d "<protobuf bytes>" 'http://test-goim.yy.com:7172/1/push?uid=88889999'
```

### 单消息多人推送

可以批量向多个用户发推送消息。POST body 为 protobuf 编码的 `MultiPush`

```protobuf
message MultiPush {
    ServerPush msg         = 1;
    repeated int64 userIDs = 2;
    int32 appID            = 3;
}
```

```
curl -d "<protobuf bytes>" 'http://test-goim.yy.com:7172/1/pushs'
```
