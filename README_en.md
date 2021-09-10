# passwordbox

[![golang-ci](https://github.com/vearne/passwordbox/actions/workflows/golang-ci.yml/badge.svg)](https://github.com/vearne/passwordbox/actions/workflows/golang-ci.yml)

[中文 README](https://github.com/vearne/passwordbox/blob/master/README.md)

Like 1Password, passwordbox is a tool for managing passwords.

## Warning
This program has not undergone rigorous security testing, there may be security risks, please use it with caution.



## Quickstart

### build
```
make build
```
### install
```
make install
```
You can also find the compiled file in [release](https://github.com/vearne/passwordbox/releases)

### start
```
pwbox --data=/Users/vearne
```
I advise you set alias for `passwordbox`
```
alias pwbox='pwbox --data=/Users/vearne'
```
After the program starts, create the database according to the manual requirements. In `passwordbox`, all items store in a database.

* `--data` set the data path of passwordbox

#### Synchronize to object storage

If you want data files to be Shared across multiple devices, 
you can also configure object storage.

##### Currently supported

* [QingCloud](https://www.qingcloud.com/products/qingstor/)  `qingstor.yaml`
* [aliyun](https://cn.aliyun.com/product/oss)  `oss.yaml`

```
pwbox --data=/Users/vearne --oss=/directory/qingstor.yaml
```
```
pwbox --data=/Users/vearne --oss=/directory/oss.yaml
```
* --oss Object store configuration file (optional)

##### Notice: 
1）Pwbox identifies the cloud vendor to which the object store belongs by the name of the configuration file.   
So the name of the configuration file is fixed.
2) For security, make sure that the Bucket where the object is stored is private (read and write using the key only)


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

In interactive mode, you can use the following commands.

####  help 
Get usage details of commands
#### add
Add a item

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
#### delete
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
  
#### modify
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

#### search

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

* `pageId` Records are displayed in pages, pageId is the number of page, start from 1.
* `keyword` You can use `keyword` to filter 
In `passwordbox`, the filter effect is like the following SQL statement
```
select * from item where title like "%keyword%"
```
#### view
view account and password as plaintext.
```
test1 > view --itemId 3
--ViewItem--
+----+-------+---------+----------+---------+---------------------------+
| ID | TITLE | ACCOUNT | PASSWORD | COMMENT |        MODIFIEDAT         |
+----+-------+---------+----------+---------+---------------------------+
|  3 | t3    | a3      | p3       |         | 2020-04-16T10:04:47+08:00 |
+----+-------+---------+----------+---------+---------------------------+
```

##### backup
Backup
```
test > backup
2021/09/10 22:23:09 [debug] commandLine:backup
Backup will be executed where it quit.
```
##### restore
Display a list of all backup files
```
test > restore
--RestoreItem--
+----+---------------------------+
| ID |            TAG            |
+----+---------------------------+
|  1 | 2021-09-10T22:24:34+08:00 |
|  2 | 2021-09-10T22:09:09+08:00 |
|  3 | 2021-09-10T21:57:03+08:00 |
|  4 | 2021-09-10T19:15:30+08:00 |
|  5 | 2021-09-10T18:31:27+08:00 |
|  6 | 2021-09-10T17:31:25+08:00 |
+----+---------------------------+
```
Restore from the specified backup file
```
test > restore -tagId 1
--RestoreItem--
? confirm restore? Yes
2021/09/10 22:26:46 [info] 1. RestoreItem-close DB
2021/09/10 22:26:46 [info] 2. RestoreItem-rename, oldName:/tmp/6879630a7d56210d2cd2491cb99d781194689fed71d7890a8dabbcb3a678cb73.2021-09-10T22:24:34+08:00, newName:/tmp/6879630a7d56210d2cd2491cb99d781194689fed71d7890a8dabbcb3a678cb73
2021/09/10 22:26:46 [info] 3. RestoreItem-upload, key:pwbox/6879630a7d56210d2cd2491cb99d781194689fed71d7890a8dabbcb3a678cb73
2021/09/10 22:26:46 [info] Restore success.Please login later...
```

#### quit
**Notice:** Remember, changes will only be saved when the quit command is executed.

## Detail
`passwordbox` use sqlite database as underlying storage, then encrypt sqlite data files.


### oss config template

#### 1. QingCloud
`qingstor.yaml`

```
access_key: xxxx
secret_key: xxxxx
bucket_name: xxxxx
zone: sh1a
dir_path: pwbox
```

#### 2. aliyun

`oss.yaml`

```
access_key_id: xxxx
access_key_secret: xxxxx
bucket_name: xxxxx
endpoint: sh1a
dir_path: pwbox
```



