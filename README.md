# passwordbox

[中文 README](https://github.com/vearne/passwordbox/blob/master/README_zh.md)

Like 1Password, passwordbox is a tool for managing passwords, but it only allows use offline.

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




