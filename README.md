# passwordbox

[English README](https://github.com/vearne/passwordbox/blob/master/README_en.md)

`passwordbox`是一个类似1password的密码管理工具。完全基于命令行交互执行。1password支持多端同步，而`passwordbox`是完全离线的。

提示：如何想要多个设备同步，可以将数据文件传到`dropbox`, 利用`dropbox`的自动同步来完成。

### 内部实现细节
首先将每个记录项加密存储在`SQLite`的数据文件中，然后再对整个数据文件进行二次加密。


### 快速开始

#### 编译
```
make build
```
#### 安装
```
make install
```
#### 启动
```
pwbox --data=/Users/vearne
```

* --data 设置加密数据文件的存储路径

建议你为`passwordbox`设置一个别名
```
alias pwbox='pwbox --data=/Users/vearne'
```

#### 同步到对象存储
如果你希望数据文件在多个设备中共享，你还可以通过配置对象存储来实现。    
##### 目前已支持

* [青云](https://www.qingcloud.com/products/qingstor/)  `qingstor.yaml`
* [阿里云](https://cn.aliyun.com/product/oss)  `oss.yaml`

```
pwbox --data=/Users/vearne --oss=/directory/qingstor.yaml
```
```
pwbox --data=/Users/vearne --oss=/directory/oss.yaml
```
* --oss 对象存储的配置文件

##### 注意: 
pwbox是通过配置文件的名称来识别对象存储所属的云厂商，所以配置文件的名称是固定的


程序启动以后，按照导引的要求创建数据库，所有的记录项都存储在数据库中
```
─$ ./pwbox --data /tmp/
---- login database ----
? Please type database's name: test
fullpath /tmp/6879630a7d56210d2cd2491cb99d781194689fed71d7890a8dabbcb3a678cb73
? Database is not exist.
Do you like to create database now? Yes
---- create database ----
? Please type database's name: test
? Please type password: *****
? Please type hint[optional]: test
---- login database ----
? Please type database's name: test
fullpath /tmp/6879630a7d56210d2cd2491cb99d781194689fed71d7890a8dabbcb3a678cb73
? Please type your password: *****
Hint for database test is test
```
登录数据库成功之后，可以执行如下的命令
##### help
获取所有的可用命令，以及它们的用法
##### add
添加一个记录项
```
test > add
--AddItem--
? Please type Item's title: google
? Please type Item's account: myaccount
? Please type Item's password: **********
? Please type Item's comment(optional):
+----+--------+-----------+------------+---------+---------------------------+
| ID | TITLE  |  ACCOUNT  |  PASSWORD  | COMMENT |        MODIFIEDAT         |
+----+--------+-----------+------------+---------+---------------------------+
|  0 | google | myaccount | mypassword |         | 2020-04-15T13:43:45+08:00 |
+----+--------+-----------+------------+---------+---------------------------+
AddItem-save to file
--SearchItem--
total: 2
pageSize: 20 currentPage: 1
+----+--------+---------+----------+---------+------------+
| ID | TITLE  | ACCOUNT | PASSWORD | COMMENT | MODIFIEDAT |
+----+--------+---------+----------+---------+------------+
|  1 | baidu  | ***     | ***      | ***     | ***        |
|  2 | google | ***     | ***      | ***     | ***        |
+----+--------+---------+----------+---------+------------+
```
##### delete
```
test1 > delete --itemId 2
--DeleteItem--
+----+--------+---------------+---------------+---------+---------------------------+
| ID | TITLE  |    ACCOUNT    |   PASSWORD    | COMMENT |        MODIFIEDAT         |
+----+--------+---------------+---------------+---------+---------------------------+
|  2 | google | googleAccount | googleAccount |         | 2020-04-15T13:55:25+08:00 |
+----+--------+---------------+---------------+---------+---------------------------+
? confirm delete? Yes
delete item 2 success
--SearchItem--
total: 1
pageSize: 20 currentPage: 1
+----+----------------+---------+----------+---------+------------+
| ID |     TITLE      | ACCOUNT | PASSWORD | COMMENT | MODIFIEDAT |
+----+----------------+---------+----------+---------+------------+
|  1 |  baidu account | ***     | ***      | ***     | ***        |
+----+----------------+---------+----------+---------+------------+
```
##### modify
```
test > modify --itemId 1
--ModifyItem--
If you don't want to make changes, you can just press Enter!
? Please type Item's title:["baidu"] baidu account
? Please type Item's account:["baiduAccount"]
? Please type Item's password:["*************"]
? Please type Item's comment(optional):[""]
+----+---------------+--------------+---------------+---------+---------------------------+
| ID |     TITLE     |   ACCOUNT    |   PASSWORD    | COMMENT |        MODIFIEDAT         |
+----+---------------+--------------+---------------+---------+---------------------------+
|  1 | baidu account | baiduAccount | cbaiduAccount |         | 2020-04-15T13:17:58+08:00 |
+----+---------------+--------------+---------------+---------+---------------------------+
```
##### search
```
test > search --pageId 1 --keyword "baidu"
--SearchItem--
total: 1
pageSize: 20 currentPage: 1
+----+-------+---------+----------+---------+------------+
| ID | TITLE | ACCOUNT | PASSWORD | COMMENT | MODIFIEDAT |
+----+-------+---------+----------+---------+------------+
|  1 | baidu | ***     | ***      | ***     | ***        |
+----+-------+---------+----------+---------+------------+
```   
* `pageId` 记录项是分页显示的，每页20条数据，`pageId`是页号，从1开始
* `keyword` 可以使用`keyword`来对记录项进行过滤，效果近似如下SQL语句
```
select * from item where title like "%keyword%"
```
##### view
以明文方式查看某个记录项的账号密码等信息。
除非执行`view`命令，否则一个记录项在内存中也是加密的。
```
test1 > view --itemId 3
--ViewItem--
+----+-------+---------+----------+---------+---------------------------+
| ID | TITLE | ACCOUNT | PASSWORD | COMMENT |        MODIFIEDAT         |
+----+-------+---------+----------+---------+---------------------------+
|  3 | baidu    | a3      | p3       |         | 2020-04-16T10:04:47+08:00 |
+----+-------+---------+----------+---------+---------------------------+
```
##### quit
**注意** 记住所有修改（CRUD）只有在执行`quit`命令时，才会被持久化到磁盘上。


### 对象存储配置文件模板

#### 1. 青云 QingCloud

`qingstor.yaml`

```
access_key: xxxx
secret_key: xxxxx
bucket_name: xxxxx
zone: sh1a
dir_path: pwbox
```

#### 2. 阿里云 aliyun

`oss.yaml`

```
access_key_id: xxxx
access_key_secret: xxxxx
bucket_name: xxxxx
endpoint: sh1a
dir_path: pwbox
```
