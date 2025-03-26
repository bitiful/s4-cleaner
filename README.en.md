# S4 Cleaner - Clean Unfinished Bitiful S4 Multipart Uploads

<div align="center">

![Version](https://img.shields.io/badge/version-0.1.0-blue.svg)
![Go Version](https://img.shields.io/badge/go-%3E%3D1.16-blue.svg)
![License](https://img.shields.io/badge/license-Apache%202.0-green.svg)

</div>

> An efficient command-line tool for cleaning up long-standing unfinished multipart uploads in S3 buckets, freeing up storage space and reducing costs.

## 🌐 About Bitiful S4

[Bitiful S4](https://docs.bitiful.com/bitiful-s4/intro) is a high-performance object storage service with the following key features:

- **Cost-Effective**: Up to 80% cost reduction compared to Amazon S3 and Alibaba OSS
- **High Performance**: Built on [Directory Buckets](https://docs.aws.amazon.com/zh_cn/AmazonS3/latest/userguide/directory-buckets-overview.html) technology for superior metadata performance
- **Network Optimized**: Native support for Rename, HTTP/2, HTTP/3, TLS1.3
- **S3 Compatible**: Can be used like Amazon S3 in most scenarios
- **Simplified Storage Tiers**: No need to choose between standard, infrequent access, or archive storage types, reducing complexity and hidden costs

## ✨ S4 Cleaner Features

- **Flexible Bucket Selection** - Support for cleaning unfinished multipart uploads in a single or all Bitiful S4 buckets
- **Customizable Expiration Time** - Set different time ranges (supports days and hours units) to identify long-standing uploads
- **Safe Operation Modes** - Two operation modes: list and delete, default is list only to prevent accidental deletion
- **Multiple Output Formats** - Support for table, JSON, CSV and other output formats for easy integration with automation workflows
- **User-Friendly Experience** - Colorful terminal output for better readability and operation experience
- **Robust Error Handling** - Elegant handling of various error situations, ensuring safe and reliable operations

## 📦 Installation

### From Source Code

```bash
# Clone repository
git clone https://github.com/bitiful/s4-cleaner.git
cd s4-cleaner

#### Create build directory
mkdir -p build

#### Linux (AMD64 and ARM64)
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o build/s4-cleaner-linux-amd64 .
GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o build/s4-cleaner-linux-arm64 .

#### Windows (AMD64 and ARM64)
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o build/s4-cleaner-windows-amd64.exe .
GOOS=windows GOARCH=arm64 go build -ldflags="-s -w" -o build/s4-cleaner-windows-arm64.exe .

#### macOS (AMD64/Intel and ARM64/M芯片)
GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o build/s4-cleaner-macos-intel .
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o build/s4-cleaner-macos-apple-silicon .
```

### Using Pre-compiled Binaries

Download pre-compiled binaries for your system from the [Releases](https://github.com/bitiful/s4-cleaner/releases) page.

## 🚀 Usage

### Basic Usage

```bash
# List unfinished multipart uploads older than 7 days in all buckets (default)
AWS_ACCESS_KEY_ID=your_ak AWS_SECRET_ACCESS_KEY=your_sk s4-cleaner

# List unfinished multipart uploads older than 3 days in the specified bucket
AWS_ACCESS_KEY_ID=your_ak AWS_SECRET_ACCESS_KEY=your_sk s4-cleaner --bucket=my-bucket --olderThan=3d

# Delete unfinished multipart uploads older than 72 hours in the specified bucket
AWS_ACCESS_KEY_ID=your_ak AWS_SECRET_ACCESS_KEY=your_sk s4-cleaner --bucket=my-bucket --olderThan=72h --doDelete

# Output in JSON format
AWS_ACCESS_KEY_ID=your_ak AWS_SECRET_ACCESS_KEY=your_sk s4-cleaner --fmt=json
```

### Parameters

| Parameter | Description | Default Value |
|:-----------|:-------------|:---------------|
| `--bucket` | Bucket name, empty means all buckets | `""` (all buckets) |
| `--olderThan` | Find multipart uploads older than this time, e.g. '7d' (7 days ago) or '72h' (72 hours ago) | `"7d"` |
| `--doDelete` | Whether to perform deletion, default is false (list only) | `false` |
| `--fmt` | Output format: table, json, csv | `"table"` |
| `--version`, `-v` | Show version information | - |

### Environment Variables

The tool uses the following environment variables for AWS authentication:

- `AWS_ACCESS_KEY_ID`: AWS access key ID (required)
- `AWS_SECRET_ACCESS_KEY`: AWS secret access key (required)

## 📊 Output Examples

### Table Output (Default)

```
+-----------------+----------------------------------+-------------+---------------------+-----------------+
| BUCKET          |               KEY                |    SIZE     |      MOD TIME       |     STATUS      |
+-----------------+----------------------------------+-------------+---------------------+-----------------+
| my-bucket       | temp/file1.txt                   | 1.2 MB      | 2023-01-01 12:00:00 | Will delete  |
| my-bucket       | temp/file2.txt                   | 3.4 MB      | 2023-01-05 12:00:00 | Won't delete |
+-----------------+----------------------------------+-------------+---------------------+-----------------+

+--------------------------------+------------+
|          STATISTICS            |   VALUE    |
+--------------------------------+------------+
| Total files                    | 2          |
| Total size                     | 4.6 MB     |
| Files to delete                | 1          |
| Size to delete                 | 1.2 MB     |
| Files deleted                  | 0          |
| Size deleted                   | 0 B        |
+--------------------------------+------------+
```

### Status Explanation

- **Will delete**: File will be deleted (if using the `--doDelete` parameter)
- **Won't delete**: File will not be deleted (does not meet deletion criteria)
- **Deleted**: File has been successfully deleted
- **Delete failed**: File deletion failed

### JSON Output

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

### CSV Output

```
Bucket,Key,Size,SizeFormatted,ModTime,ShouldDelete,DeleteSuccess
my-bucket,temp/file1.txt,1258291,"1.2 MB",2023-01-01T12:00:00Z,true,null
my-bucket,temp/file2.txt,3564812,"3.4 MB",2023-01-05T12:00:00Z,false,null
```

## 📄 License

Apache License 2.0

## 🌐 Other Languages

- [简体中文](README.md)
