# 队列调度器说明

## 概述

队列调度器是一个统一的任务管理系统，用于管理AlphaFold和I-TASSER的蛋白质结构预测任务。由于这些任务需要数小时的计算时间，调度器确保：

1. **避免重复计算**：同一时间只处理一个任务
2. **自动任务调度**：定时检查队列中的待处理任务
3. **状态跟踪**：实时跟踪任务状态（pending、processing、completed、failed）
4. **自动清理**：定期清理已完成和失败的任务

## 功能特性

### 任务状态管理
- `pending`: 等待处理的任务
- `processing`: 正在处理中的任务
- `completed`: 已完成的任务
- `failed`: 处理失败的任务

### 并发控制
- 每个队列（AlphaFold/I-TASSER）同时只能处理一个任务
- 避免资源冲突和重复计算

### 自动调度
- 每30秒检查一次队列状态
- 自动启动待处理任务
- 自动清理24小时前的已完成/失败任务

### 模型后处理
- 自动生成Ramachandran图
- 图片文件名与模型ID一致
- 保存到`static/ramachandran_plots/`目录

## 使用方法

### 1. 启动调度器
调度器在服务器启动时自动启动，无需手动操作。

### 2. 查看队列状态
```bash
GET /profasa/system/queuestatus
```

返回示例：
```json
{
  "code": 200,
  "message": "获取队列状态成功",
  "data": {
    "alphafold": {
      "pending": 2,
      "processing": 1,
      "completed": 5,
      "failed": 0
    },
    "itasser": {
      "pending": 1,
      "processing": 0,
      "completed": 3,
      "failed": 1
    },
    "is_running": true
  }
}
```

### 3. 添加任务到队列
任务通过现有的API接口添加到队列中，状态默认为`pending`。

## 技术实现

### 核心组件

1. **QueueScheduler**: 主调度器
   - 管理任务生命周期
   - 控制并发处理
   - 提供状态查询接口

2. **AlphaProcessor**: AlphaFold处理器
   - 处理AlphaFold模型生成
   - 验证FASTA格式
   - 移动生成的文件
   - 生成Ramachandran图

3. **ItasserProcessor**: I-TASSER处理器
   - 处理I-TASSER模型生成
   - 验证FASTA格式
   - 移动生成的文件
   - 生成Ramachandran图

### 数据库模型

#### AlphaFoldQueue
```go
type AlphaFoldQueue struct {
    gorm.Model
    Sequence string `gorm:"not null;uniqueIndex:uni_alpha_fold_queues_sequence;type:varchar(191)"`
    IsSubseq int64  `gorm:"not null;default:0"`
    ParentId *int64 `gorm:"default:null"`
    Status   string `gorm:"not null;default:'pending'"` // pending, processing, completed, failed
}
```

#### ITasserQueue
```go
type ITasserQueue struct {
    gorm.Model
    Sequence string `gorm:"not null;uniqueIndex:uni_i_tasser_queues_sequence;type:varchar(191)"`
    IsSubseq int64  `gorm:"not null;default:0"`
    ParentId *int64 `gorm:"default:null"`
    Status   string `gorm:"not null;default:'pending'"` // pending, processing, completed, failed
}
```

## 配置参数

### 调度器配置
- **检查间隔**: 30秒
- **清理保留时间**: 24小时
- **最大并发数**: 每个队列1个任务

### 处理器配置
- **AlphaFold最大工作线程**: 1
- **I-TASSER最大工作线程**: 1

## 监控和日志

### 日志记录
调度器会记录以下事件：
- 调度器启动/停止
- 任务开始处理
- 任务完成/失败
- 队列清理操作

### 状态监控
通过API接口可以实时监控：
- 各队列的任务数量
- 调度器运行状态
- 任务处理进度

## 故障处理

### 常见问题

1. **任务卡在processing状态**
   - 检查计算进程是否正常运行
   - 查看日志文件确认错误信息
   - 手动重置任务状态

2. **队列任务堆积**
   - 检查计算资源是否充足
   - 确认外部工具（AlphaFold/I-TASSER）是否正常
   - 考虑增加并发处理能力

3. **文件移动失败**
   - 检查磁盘空间
   - 确认文件权限
   - 验证目标路径是否存在

### 手动操作

如果需要手动干预，可以通过数据库直接操作：

```sql
-- 重置卡住的任务
UPDATE alpha_fold_queues SET status = 'pending' WHERE status = 'processing';

-- 查看任务状态
SELECT * FROM alpha_fold_queues WHERE status = 'processing';
```

## 扩展性

### 增加并发处理
修改`main.go`中的调度器配置：
```go
queueScheduler := services.NewQueueScheduler(2, 2) // 增加并发数
```

### 添加新的处理器
1. 创建新的处理器结构
2. 在调度器中添加相应的处理逻辑
3. 更新状态查询接口

### 自定义清理策略
修改`cleanupCompletedTasks`方法中的清理逻辑，可以：
- 调整保留时间
- 添加更复杂的清理条件
- 实现分级清理策略 