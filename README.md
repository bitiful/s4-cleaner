# S3 临时文件清理工具

<div align="center">

![Version](https://img.shields.io/badge/version-0.1.0-blue.svg)
![Go Version](https://img.shields.io/badge/go-%3E%3D1.16-blue.svg)
![License](https://img.shields.io/badge/license-Apache%202.0-green.svg)

</div>

> 一个高效的命令行工具，用于清理 S3 存储桶中过期的临时文件。

## ✨ 功能特点

- **灵活的存储桶选择** - 支持清理单个或所有 S3 存储桶中的临时文件
- **自定义过期时间** - 可设置不同的时间范围（支持天和小时单位）
- **安全的操作模式** - 提供列出和删除两种操作模式，默认仅列出
- **多样化输出格式** - 支持表格、JSON、CSV 等多种输出格式
- **友好的用户体验** - 彩色终端输出，提升可读性
- **健壮的错误处理** - 优雅处理长文件名和各种错误情况

## 📦 安装

### 从源代码安装

```bash
#### 克隆仓库
git clone https://github.com/bitiful/s4-cleaner.git
cd s4-cleaner

#### 创建构建目录
mkdir -p build

#### Linux (AMD64 和 ARM64)
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o build/s4-cleaner-linux-amd64 .
GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o build/s4-cleaner-linux-arm64 .

#### Windows (AMD64 和 ARM64)
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o build/s4-cleaner-windows-amd64.exe .
GOOS=windows GOARCH=arm64 go build -ldflags="-s -w" -o build/s4-cleaner-windows-arm64.exe .

#### macOS (AMD64/Intel 和 ARM64/M芯片)
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o build/s4-cleaner-macos-intel .
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o build/s4-cleaner-macos-apple-silicon .
```

### 使用预编译二进制文件

从 [Releases](https://github.com/bitiful/s4-cleaner/releases) 页面下载适合您系统的预编译二进制文件。

## 🚀 使用方法

### 基本用法

```bash
# 列出所有桶中7天前的临时文件（默认）
AWS_ACCESS_KEY_ID=your_ak AWS_SECRET_ACCESS_KEY=your_sk s4-cleaner

# 列出指定桶中3天前的临时文件
AWS_ACCESS_KEY_ID=your_ak AWS_SECRET_ACCESS_KEY=your_sk s4-cleaner --bucket=my-bucket --olderThan=3d

# 删除指定桶中72小时前的临时文件
AWS_ACCESS_KEY_ID=your_ak AWS_SECRET_ACCESS_KEY=your_sk s4-cleaner --bucket=my-bucket --olderThan=72h --doDelete

# 以JSON格式输出
AWS_ACCESS_KEY_ID=your_ak AWS_SECRET_ACCESS_KEY=your_sk s4-cleaner --fmt=json
```

### 参数说明

| 参数 | 说明 | 默认值 |
|:------:|:------|:--------:|
| `--bucket` | 存储桶名称，为空表示所有桶 | `""` (所有桶) |
| `--olderThan` | 查找早于此时间的文件，如 '7d'（7天前）或 '72h'（72小时前） | `"7d"` |
| `--doDelete` | 是否执行删除操作，默认为false（仅列出） | `false` |
| `--fmt` | 输出格式：table, json, csv | `"table"` |
| `--version`, `-v` | 显示版本信息 | - |

### 环境变量

工具使用以下环境变量进行 AWS 认证：

- `AWS_ACCESS_KEY_ID`：AWS 访问密钥 ID（必需）
- `AWS_SECRET_ACCESS_KEY`：AWS 秘密访问密钥（必需）

## 📊 输出示例

### 表格输出（默认）

```
+-----------------+----------------------------------+-------------+---------------------+-----------------+
| 存储桶 | BUCKET |             键 | KEY             | 大小 | SIZE | 修改时间 | MOD TIME |  状态 | STATUS  |
+-----------------+----------------------------------+-------------+---------------------+-----------------+
| my-bucket       | temp/file1.txt                   | 1.2 MB      | 2023-01-01 12:00:00 | 🎯 Will delete  |
| my-bucket       | temp/file2.txt                   | 3.4 MB      | 2023-01-05 12:00:00 | 🔍 Won't delete |
+-----------------+----------------------------------+-------------+---------------------+-----------------+

+--------------------------------+------------+
|     统计信息 | STATISTICS      | 值 | VALUE |
+--------------------------------+------------+
| 总文件数 | Total files         | 2          |
| 总容量 | Total size            | 4.6 MB     |
| 应删除文件数 | Files to delete | 1          |
| 应删除容量 | Size to delete    | 1.2 MB     |
| 已删除文件数 | Files deleted   | 0          |
| 已删除容量 | Size deleted      | 0 B        |
+--------------------------------+------------+
```

### 状态说明

- 🎯 **Will delete**：文件将会被删除（如果使用 `--doDelete` 参数）
- 🔍 **Won't delete**：文件不会被删除（不符合删除条件）
- ✅ **Deleted**：文件已成功删除
- ❌ **Delete failed**：文件删除失败

### JSON 输出

```json
{
  "files": [
    {
      "bucket": "my-bucket",
      "key": "temp/file1.txt",
      "size": 1258291,
      "size_formatted": "1.2 MB",
      "mod_time": "2023-01-01T12:00:00Z",
      "should_delete": true,
      "delete_success": null
    },
    {
      "bucket": "my-bucket",
      "key": "temp/file2.txt",
      "size": 3564812,
      "size_formatted": "3.4 MB",
      "mod_time": "2023-01-05T12:00:00Z",
      "should_delete": false,
      "delete_success": null
    }
  ],
  "statistics": {
    "total_files": 2,
    "total_size": 4823103,
    "total_size_formatted": "4.6 MB",
    "files_to_delete": 1,
    "size_to_delete": 1258291,
    "size_to_delete_formatted": "1.2 MB",
    "files_deleted": 0,
    "size_deleted": 0,
    "size_deleted_formatted": "0 B"
  }
}
```

### CSV 输出

```
Bucket,Key,Size,SizeFormatted,ModTime,ShouldDelete,DeleteSuccess
my-bucket,temp/file1.txt,1258291,"1.2 MB",2023-01-01T12:00:00Z,true,null
my-bucket,temp/file2.txt,3564812,"3.4 MB",2023-01-05T12:00:00Z,false,null
```

## 📄 许可证

Apache License 2.0

## 🌐 其他语言

- [English](README.en.md)
