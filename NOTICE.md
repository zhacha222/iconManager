# 声明

## 实现方式

本项目是飞牛 fnOS 的第三方应用，按官方公开规范独立实现：

- 应用结构遵循飞牛官方 [应用开发规范](https://developer.fnnas.com/docs/guide)
- 数据通过公开的数据库表 `appcenter.app_service` 读写
- 前端为独立实现的单文件 HTML（内嵌于 Go 二进制）
- 后端为 Go 编写的 CGI 程序，不监听端口、无常驻进程

## 商标与隶属关系

本项目与飞牛（fnOS）官方无隶属、合作或背书关系。
"飞牛"、"fnOS" 等名称归其各自所有者所有。

## 许可

本项目以 [MIT](LICENSE) 许可发布。
