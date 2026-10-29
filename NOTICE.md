# 第三方声明

## 参考项目

本项目的功能定位是飞牛 fnOS 第三方应用「扩展图标管理」（`customIcon`）的功能增强实现。
UI 布局、交互流程与文案参考其设计。

原扩展版权声明：

```
Copyright © 2021-Present 晓勃网络(XIAOBONET). All Rights Reserved.
Developed by MR_XIAOBO (fnOS Community Member)
```

## 实现独立性

本项目**不包含、不修改、不反编译**原扩展的任何二进制或源代码：

- 通过飞牛官方公开的 [应用开发规范](https://developer.fnnas.com/docs/guide) 实现 fpk 应用结构
- 通过公开的数据库表 `appcenter.app_service` 读写图标数据
- 前端为独立实现的单文件 HTML（内嵌于 Go 二进制）

与原扩展的唯一共享是**数据**（同一张表、同一批图标记录），这是为了实现"接管管理"的目的。

## 商标与隶属关系

本项目与飞牛（fnOS）官方、以及原扩展作者均无隶属、合作或背书关系。
"飞牛"、"fnOS" 等名称归其各自所有者所有。
