<!-- TOC -->

- [1. 说明](#1-说明)
- [2. 邮箱注册示例](#2-邮箱注册示例)
- [3. 手机注册示例](#3-手机注册示例)

<!-- /TOC -->

<a id="markdown-1-说明" name="1-说明"></a>
# 1. 说明

`/register`

注册接口,3种注册方式:
* 邮箱注册
* 手机号注册
* 邮箱+手机号注册


请求参数:
参数|类型|最大长度|描述|示例值
-|-|-|-|-
phone|string|11|手机号|13023252617
email|string|255|邮箱号|yqsy021@126.com
invite_code|string|255|邀请码|
passwd|string|不限,后台存储是hash|密码|abc123456


应答参数:
参数|类型|最大长度|描述|示例值
-|-|-|-|-
status|in32||应答状态码|100
msg|string||应答描述|成功
user_id|string|32|生成的用户ID|623f9143358f43d8bb670671d991c04f

<a id="markdown-2-邮箱注册示例" name="2-邮箱注册示例"></a>
# 2. 邮箱注册示例

请求:
```
{
	"email": "yqsy021@126.com",
	"passwd": "abc123456"
}
```
应答:

```
{
    "status": 100,
    "msg": "注册成功",
    "user_id": "b456abb573624f68a1f8078e3d006a7a"
}
```

<a id="markdown-3-手机注册示例" name="3-手机注册示例"></a>
# 3. 手机注册示例

请求:
```
{
	"phone": "13023252617",
	"passwd": "abc123456"
}
```
应答:

```
{
    "status": 100,
    "msg": "注册成功",
    "user_id": "384a32d2c5c148e5851984b34fbc6e04"
}
```



