# iBoard_http_server

一个基于Go语言开发的建筑管理系统后端服务。

## 项目结构
```
├── controller/ # 控制器层，处理请求逻辑
├── database/ # 数据库相关配置和连接

├── grpc/ # gRPC 服务相关代码
├── middleware/ # 中间件
├── models/ # 数据模型
├── router/ # 路由配置

├── services/ # 业务服务层
    
├── utils/ # 工具函数
├── views/ # 视图层

├── docker-compose.yml # Docker 编排配置
├── Dockerfile # Docker 构建文件
├── init.sql # 数据库初始化脚本
└── main.go # 程序入口文件

```


