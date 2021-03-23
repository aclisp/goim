`goim` 

* 曾经稳定的支撑了一个国内语音社交应用两年以上
* 最大同时在线用户约3万，设计容量能支撑10万+
* 本身是一个单播广播系统，适用于实时IM场景的应用
* 分布式架构设计，有横向扩容的能力
* 代码简单，容易修改；项目自包含，没有外部依赖

参见：[长连接和广播系统设计文档](doc/arch.md)



~~~
BELOW IS THE ORIGINAL PROJECT INFORMATION
~~~

goim
==============
goim is a im server writen by golang.

## Features
 * Light weight
 * High performance
 * Pure Golang
 * Supports single push, multiple push and broadcasting
 * Supports one key to multiple subscribers (Configurable maximum subscribers count)
 * Supports heartbeats (Application heartbeats, TCP, KeepAlive, HTTP long pulling)
 * Supports authentication (Unauthenticated user can't subscribe)
 * Supports multiple protocols (WebSocket，TCP，HTTP）
 * Scalable architecture (Unlimited dynamic job and logic modules)
 * Asynchronous push notification based on Kafka

## Architecture
![arch](https://github.com/Terry-Mao/goim/blob/master/doc/arch.png)

Protocol:

[proto](https://github.com/Terry-Mao/goim/blob/master/doc/protocol.png)

## Document
[English](./README_en.md)

[中文](./README_cn.md)

## Examples
Websocket: [Websocket Client Demo](https://github.com/Terry-Mao/goim/tree/master/examples/javascript)

Android: [Android](https://github.com/roamdy/goim-sdk)

iOS: [iOS](https://github.com/roamdy/goim-oc-sdk)

## Benchmark
![benchmark](./doc/benchmark.jpg)

### Benchmark Server
| CPU | Memory | OS | Instance |
| :---- | :---- | :---- | :---- |
| Intel(R) Xeon(R) CPU E5-2630 v2 @ 2.60GHz  | DDR3 32GB | Debian GNU/Linux 8 | 1 |

### Benchmark Case
* Online: 1,000,000
* Duration: 15min
* Push Speed: 40/s (broadcast room)
* Push Message: {"test":1}
* Received calc mode: 1s per times, total 30 times

### Benchmark Resource
* CPU: 2000%~2300%
* Memory: 14GB
* GC Pause: 504ms
* Network: Incoming(450MBit/s), Outgoing(4.39GBit/s)

### Benchmark Result
* Received: 35,900,000/s

[中文](./doc/benchmark_cn.md)

[English](./doc/benchmark_en.md)

## LICENSE
goim is is distributed under the terms of the MIT License.
