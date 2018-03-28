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
}
```

**服务器推送body**
```protobuf
message ServerPush {
    sint32 messageType    = 1;
    bytes  pushBuffer     = 2;
    map<string, string> headers = 3;
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

`body` 为空，跳过认证

### 心跳消息格式

心跳消息发送间隔不能超过5分钟

### 业务消息格式

目前业务消息转发到YYTars框架后端

`requestBuffer`  是业务消息请求的protobuf编码
`responseBuffer` 是业务消息应答的protobuf编码

### 切换房间消息格式

TODO

### 推送消息格式

推送消息只有下行，其中 `body` 为业务自定义，一般是json

## 触发推送消息

业务向推送 URL 发送 POST http

### 推送 URL

http://test-goim.yy.com:7172

### 房间推送

```
curl -d "{\"test\": 12345}" 'http://test-goim.yy.com:7172/1/push/room?rid=20&appid=10001'
``` 

### 广播

```
curl -d "{\"test\": 12345}" 'http://test-goim.yy.com:7172/1/push/all'
```

### 单人推送

```
curl -d "{\"test\":12345}" 'http://test-goim.yy.com:7172/1/push?uid=88889999&appid=10001'
```

### 单消息多人推送

```
curl -d "{\"u\":[88889999,88889000],\"m\":{\"test\":12},\"a\":10001}" 'http://test-goim.yy.com:7172/1/pushs'
```
