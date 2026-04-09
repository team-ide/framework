# 说明

## common

### starter

* Start 方法，启动
    - 启动顺序
        * 系统 初始化
        * 触发 `EventSystemInitBefore` 事件
        * init config 初始化 配置 开始
            * 触发 `EventConfigInitBefore` 事件
            * 执行 config init func，按照 Order 正序执行
            * 触发 `EventConfigInitAfter` 事件
        * init component 初始化 组件 开始
            * 触发 `EventComponentInitBefore` 事件
            * 执行 component init func，按照 Order 正序执行
            * 触发 `EventComponentInitAfter` 事件
        * init factory 初始化 组件 开始
            * 触发 `EventFactoryInitBefore` 事件
            * 执行 factory init func，按照 Order 正序执行
            * 触发 `EventFactoryInitAfter` 事件
        * init table 初始化 表 开始
            * 触发 `EventTableInitBefore` 事件
            * 执行 table init func，按照 Order 正序执行
            * 触发 `EventTableInitAfter` 事件
        * init data 初始化 数据 开始
            * 触发 `EventDataInitBefore` 事件
            * 执行 data init func，按照 Order 正序执行
            * 触发 `EventDataInitAfter` 事件
        * 触发 `EventSystemInitAfter` 事件
        * 系统 启动服务
        * 触发 `EventServerStartBefore` 事件
            * 执行 server start func，按照 Order 正序执行
        * 触发 `EventServerStartAfter` 事件
        * 触发 `EventReady` 事件
* AddInitConfigFunc 添加 初始化 配置 函数
* AddInitComponentFunc 添加 初始化 组件 函数
* AddInitFactoryFunc 添加 初始化 工厂 函数
* AddInitTableFunc 添加 初始化 表 函数
* AddInitDataFunc 添加 初始化 数据 函数
* AddServerStartFunc 添加 服务 启动 函数
* OnEvent 添加 事件 监听
* CallEvent 触发 事件
* SetShouldWait 设置 是否需要 Wait 通常 系统包含服务时候需要设置 true
* Wait 阻塞线程 防止退出

### error

* 自定义异常，自定义属性
* code: 错误码
* msg: 错误信息

### logger

* 日志

## db

### dialect 数据库方言

* 数据库 DDL 语句
* 数据库 分页 语句
* 数据库 SQL 参数占位符
* 数据库 数据类型
* 数据库 常用函数

### service 数据库服务

* 表数据 增、删、改、查
* 结构体 > 增、改 SQL

### sql 动态 SQL

* 表数据 增、删、改、查
* 结构体 > 增、改 SQL
