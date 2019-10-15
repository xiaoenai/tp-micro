 # 用户模块

 ## 1.1 通过姓名获取用户接口

**简要描述：** 

- 获取用户资料信息

**请求URL：** 
- ` http://xxx.xxx.com/user/v1/profile/get_by_name
  

**请求方式：**
- GET 
- Content-Type: application/json;charset=utf-8

**参数：** 

|参数名|必选|类型|说明|
|:----    |:---|:----- |-----   |
| name | 是 |string | 用户姓名    |

 **参数示例**

``` json
  {
      "name": "micro"
  }
```

 **返回参数说明** 

|  参数名 |是否必填   |描述   |
| ------------ | ------------ | ------------ |
| id |是|用户id|
| age |是|用户年龄|
| phone|是|用户手机号|

**返回示例**：

 ``` json
 {
     "id": 123,
     "age": 2,
     "phone": "xxxx"
 } 

 ```