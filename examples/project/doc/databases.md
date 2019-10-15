## MySql

## 1.1 用户信息表(user)
| 参数 | 说明 | 类型 | 默认值 | 版本 |
|------|------|------|------|------|
| id | 自增Id| `bigint(20)` | - | V1 |
| name | 用户姓名 | `varchar(20)` | 0 | V1 |
| age | 用户年龄 | `int(10)` | 0 | V1 |
| phone | 用户手机号 | `varchar(15)` | '' | V1 |

```sql
CREATE TABLE `user` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `name` varchar(20) NOT NULL DEFAULT '0' COMMENT '姓名',
  `age` int(10) NOT NULL DEFAULT '' COMMENT '年龄',
  `phone` varchar(15) NOT NULL DEFAULT '' COMMENT '手机号',
  `updated_at` bigint(11) NOT NULL DEFAULT '0' COMMENT '更新时间',
  `created_at` bigint(11) NOT NULL COMMENT '创建时间',
  `deleted_ts` bigint(11) NOT NULL DEFAULT '0' COMMENT '删除时间(0表示未删除)',
  PRIMARY KEY (`id`),
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户信息表';

```


## Mongodb

## 1.1 用户元信息表(user_meta)

| 参数   | 说明     | 类型           | 默认值 | 版本 |
| ------ | -------- | -------------- | ------ | ---- |
| avatar | 用户头像 | `varchar(200)` | 0      | V1   |

