
<!-- TOC -->

- [1. 说明](#1-说明)
- [2. 入金示例](#2-入金示例)
- [3. 流水号](#3-流水号)
- [4. 幂等,重复流水号示例](#4-幂等重复流水号示例)

<!-- /TOC -->


<a id="markdown-1-说明" name="1-说明"></a>
# 1. 说明

`/recharge`


入金接口,流水号由客户端生成


请求参数:
参数|类型|最大长度|描述|示例值
-|-|-|-|-
f_goldin_flow_id|string|64||`64位流水号`
user_id|string|32|用户号|623f9143358f43d8bb670671d991c04f
amount|string||转入金额|88.88

应答参数:
参数|类型|最大长度|描述|示例值
-|-|-|-|-
status|in32||应答状态码|100
msg|string||应答描述|成功



<a id="markdown-2-入金示例" name="2-入金示例"></a>
# 2. 入金示例

请求:
```
{
	"f_goldin_flow_id" : "2018070715274948d19bf0fdd340dcbaef86e8282e5839416f8ac2ec7646afb4",
	"user_id" : "48d19bf0fdd340dcbaef86e8282e5839" ,
	"amount" : "1.0"
}
```

应答:
```
{
    "status": 100,
    "msg": "入金成功"
}
```

<a id="markdown-3-流水号" name="3-流水号"></a>
# 3. 流水号

```
时间(14)|user_id(32)|random(18)
```

<a id="markdown-4-幂等重复流水号示例" name="4-幂等重复流水号示例"></a>
# 4. 幂等,重复流水号示例

请求:
```
{
	"f_goldin_flow_id" : "2018070715274948d19bf0fdd340dcbaef86e8282e5839416f8ac2ec7646afb4",
	"user_id" : "48d19bf0fdd340dcbaef86e8282e5839" ,
	"amount" : "1.0"
}
```

应答:
```
{
    "status": 1000,
    "msg": "产生流水失败"
}
```
