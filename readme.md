# zoo 动物园框架

为简约而生, 核心文件只有3个  
它是继承于 zoo 框架，分离 kwdog(kwcat, kwrat), kwlog, front 等模块（独立出来）  
创建一个zoant(https://github.com/suisrc/zoant.git) 模块，用来孵化为模块  

## 框架介绍

这是一极精简的 web 服务框架， 默认没有 routing， 使用 map 进行 action 检索 ，为单接口而生。  
通过 query.action(query参数) 或者 path=action(path就是action) 两种方式确定 handle 函数。  

  
为什么需要它？  
在很多项目中，可能只需要几个接口， 而为这些接口无论使用 gin, echo, iris, fasthttp...我认为都是不值当的。因此它就诞生了。  

  
自动注入wire?  
wire 是一个依赖注入框架， 但是考虑到框架本身就比较小，本身不依赖任何第三方库，所以不会集成wire， 如果需要，可以考虑自行增加。  
但是，实现了一个简单的注入封装，`svckit:"auto"`, 可以自动注入依赖。例如：

```go

type HelloHandler struct {
	F1 any                           // 标记不注入，默认
	F2 any      `svckit:"-"`         // 标记不注入，默认
	F3 zoo.Module `svckit:"type"`      // 根据【类型】自动注入
	F4 zoo.SvcKit `svckit:"type"`      // 根据【类型】自动注入
	F5 any      `svckit:"hdl-hello"` // 根据【名称】自动注入
	F6 any      `svckit:"hdl-world"` // 根据【名称】自动注入
	F7 zoo.TplKit `svckit:"auto"`      // 根据【类型】自动注入
	f1 zoo.TplKit `svckit:"auto"`      // 私有【属性】不能注入
}

```

## 项目列表

[zodog](https://github.com/suisrc/zodog.git) 网关工具， 反向代理， 正向代理，透传代理  
[zolog](https://github.com/suisrc/zolog.git) 为 fluentbit 提供 http 协议的收集和展示方案  
[zofnt](https://github.com/suisrc/zofnt.git) 取代常规 nginx 部署前端静态资源，提供文件路由和展示  
[zobee](https://github.com/suisrc/zobee.git) 基于 ebpf 对 系统中的 http / https 流量进行免入侵形式监控  
