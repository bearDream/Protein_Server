# 全局日志系统使用说明

## 概述

项目已集成全局日志系统，位于 `logger/logger.go`，提供了统一的日志输出功能，支持不同级别的日志记录。

## 日志级别

系统支持以下日志级别（按优先级从低到高）：

- `DEBUG`: 调试信息
- `INFO`: 一般信息
- `WARN`: 警告信息
- `ERROR`: 错误信息
- `FATAL`: 致命错误（会终止程序）

## 使用方法

### 1. 导入日志包

```go
import "Protein_Server/logger"
```

### 2. 使用日志函数

```go
// 基本用法
logger.Debug("调试信息")
logger.Info("一般信息")
logger.Warn("警告信息")
logger.Error("错误信息")
logger.Fatal("致命错误") // 会终止程序

// 格式化输出
logger.Info("服务器启动成功，端口: %d", 3000)
logger.Error("数据库连接失败: %v", err)
```

### 3. 设置日志级别

```go
// 设置日志级别
logger.SetLogLevel(logger.DEBUG)  // 显示所有日志
logger.SetLogLevel(logger.INFO)   // 显示INFO及以上级别
logger.SetLogLevel(logger.WARN)   // 显示WARN及以上级别
logger.SetLogLevel(logger.ERROR)  // 只显示ERROR和FATAL
logger.SetLogLevel(logger.FATAL)  // 只显示FATAL
```

## 日志格式

日志输出格式为：
```
[时间戳] [日志级别] 日志内容
```

示例：
```
[2025-07-21 11:14:07] [INFO] 服务器启动成功，端口: 3000
[2025-07-21 11:14:07] [ERROR] 数据库连接失败: 连接超时
```

## 项目中的使用

项目中所有原有的 `fmt.Println`、`fmt.Printf`、`log.Println`、`log.Printf` 等输出都已替换为对应的日志函数：

- 错误信息使用 `logger.Error()`
- 一般信息使用 `logger.Info()`
- 调试信息使用 `logger.Debug()`
- 警告信息使用 `logger.Warn()`

## 优势

1. **统一格式**: 所有日志都有统一的时间戳和级别标识
2. **级别控制**: 可以通过设置日志级别来控制输出
3. **易于维护**: 集中管理所有日志输出
4. **性能优化**: 可以根据级别过滤不需要的日志

## 注意事项

1. 默认日志级别为 `INFO`
2. 只有当前设置的级别及以上的日志才会输出
3. `FATAL` 级别的日志会终止程序执行
4. 日志系统在程序启动时自动初始化 