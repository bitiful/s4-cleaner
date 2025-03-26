/*
 * Copyright (c) 2025 缤纷云S4 (Bitiful S4)
 *
 * 缤纷云S3临时文件清理工具
 * Bitiful S4 S3 Temporary File Cleaner
 */

package cleaner

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bitiful/s4-cleaner/pkg/config"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
)

// S3Cleaner S3清理器
// S3Cleaner is a cleaner for S3 buckets
type S3Cleaner struct {
	client *s3.Client
	cfg    *config.Config
}

// FileInfo 文件信息
// FileInfo contains information about a file
type FileInfo struct {
	Bucket        string    `json:"bucket"`
	Key           string    `json:"key"`
	Size          int64     `json:"size"`
	ModTime       time.Time `json:"mod_time"`
	ShouldDelete  bool      `json:"should_delete"`
	DeleteSuccess *bool     `json:"delete_success,omitempty"`
}

// NewS3Cleaner 创建新的S3清理器
// NewS3Cleaner creates a new S3 cleaner
func NewS3Cleaner(accessKey, secretKey string, cfg *config.Config) (*S3Cleaner, error) {
	// 解析过期时间
	// Parse expiration time
	if err := cfg.ParseTime(); err != nil {
		return nil, err
	}

	// 创建AWS配置
	// Create AWS configuration
	awsCfg, err := awsconfig.LoadDefaultConfig(context.TODO(),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		awsconfig.WithRegion("us-east-1"), // 默认区域，会根据桶自动调整 | Default region, will be adjusted automatically based on bucket
	)
	if err != nil {
		return nil, fmt.Errorf("无法加载AWS配置: %v\nFailed to load AWS configuration: %v", err, err)
	}

	// 创建S3客户端
	// Create S3 client
	client := s3.NewFromConfig(awsCfg)

	return &S3Cleaner{
		client: client,
		cfg:    cfg,
	}, nil
}

// Run 执行清理操作
// Run executes the cleaning operation
func (c *S3Cleaner) Run() error {
	var buckets []string
	var err error

	// 如果指定了桶，则只处理该桶
	// If bucket is specified, only process that bucket
	if c.cfg.Bucket != "" {
		buckets = []string{c.cfg.Bucket}
	} else {
		// 否则获取所有桶
		// Otherwise get all buckets
		buckets, err = c.listBuckets()
		if err != nil {
			return err
		}
	}

	// 收集所有文件信息
	// Collect all file information
	allFiles := []FileInfo{}

	// 处理每个桶
	// Process each bucket
	for _, bucket := range buckets {
		files, err := c.processOneBucket(bucket)
		if err != nil {
			color.Red("处理桶 %s 时出错: %v\nError processing bucket %s: %v", bucket, err, bucket, err)
			continue
		}
		allFiles = append(allFiles, files...)
	}

	// 根据格式输出结果
	// Output results based on format
	switch strings.ToLower(c.cfg.Format) {
	case "json":
		return c.outputJSON(allFiles)
	case "csv":
		return c.outputCSV(allFiles)
	default: // table
		return c.outputTable(allFiles)
	}
}

// listBuckets 列出所有桶
// listBuckets lists all buckets
func (c *S3Cleaner) listBuckets() ([]string, error) {
	resp, err := c.client.ListBuckets(context.TODO(), &s3.ListBucketsInput{})
	if err != nil {
		return nil, fmt.Errorf("无法列出桶: %v\nFailed to list buckets: %v", err, err)
	}

	buckets := make([]string, 0, len(resp.Buckets))
	for _, bucket := range resp.Buckets {
		buckets = append(buckets, *bucket.Name)
	}

	return buckets, nil
}

// processOneBucket 处理一个桶
// processOneBucket processes one bucket
func (c *S3Cleaner) processOneBucket(bucket string) ([]FileInfo, error) {
	// color.Cyan("正在处理桶: %s\nProcessing bucket: %s", bucket, bucket)

	var keyMarker *string
	var uploadIdMarker *string
	files := []FileInfo{}

	// 分页列出所有未完成的分段上传
	// List all multipart uploads with pagination
	for {
		resp, err := c.client.ListMultipartUploads(context.TODO(), &s3.ListMultipartUploadsInput{
			Bucket:         aws.String(bucket),
			KeyMarker:      keyMarker,
			UploadIdMarker: uploadIdMarker,
		})
		if err != nil {
			return nil, fmt.Errorf("无法列出桶 %s 中的未完成上传: %v\nFailed to list multipart uploads in bucket %s: %v", bucket, err, bucket, err)
		}

		// 处理当前页的未完成上传
		// Process uploads in current page
		for _, upload := range resp.Uploads {
			// 检查是否过期
			// Check if it's expired
			shouldDelete := upload.Initiated.Before(c.cfg.ExpirationTime)

			// 获取对象大小
			// Get object size
			var size int64
			parts, err := c.client.ListParts(context.TODO(), &s3.ListPartsInput{
				Bucket:   aws.String(bucket),
				Key:      upload.Key,
				UploadId: upload.UploadId,
			})
			if err == nil {
				for _, part := range parts.Parts {
					if part.Size != nil {
						size += *part.Size
					}
				}
			}

			fileInfo := FileInfo{
				Bucket:       bucket,
				Key:          *upload.Key,
				Size:         size,
				ModTime:      *upload.Initiated,
				ShouldDelete: shouldDelete,
			}

			// 如果需要删除且do标志为true，则执行删除
			// If should delete and do flag is true, perform deletion
			if shouldDelete && c.cfg.DoDelete {
				success := c.abortMultipartUpload(bucket, *upload.Key, *upload.UploadId)
				fileInfo.DeleteSuccess = &success
			}

			files = append(files, fileInfo)
		}

		// 如果没有更多页，则退出循环
		// If no more pages, exit loop
		if resp.IsTruncated == nil || !*resp.IsTruncated {
			break
		}
		keyMarker = resp.NextKeyMarker
		uploadIdMarker = resp.NextUploadIdMarker
	}

	return files, nil
}

// abortMultipartUpload 中止分段上传
// abortMultipartUpload aborts a multipart upload
func (c *S3Cleaner) abortMultipartUpload(bucket, key, uploadId string) bool {
	_, err := c.client.AbortMultipartUpload(context.TODO(), &s3.AbortMultipartUploadInput{
		Bucket:   aws.String(bucket),
		Key:      aws.String(key),
		UploadId: aws.String(uploadId),
	})
	return err == nil
}

// truncateString 截断字符串，确保中日韩文字符占两个字节
// truncateString truncates string, ensuring CJK characters count as two bytes
func truncateString(s string, maxBytes int) string {
	// 计算字符串的字节宽度
	// Calculate string byte width
	var totalWidth int
	for _, r := range s {
		var runeWidth int
		if r >= 0x4E00 && r <= 0x9FFF || // CJK统一表意文字 | CJK Unified Ideographs
			r >= 0x3040 && r <= 0x309F || // 平假名 | Hiragana
			r >= 0x30A0 && r <= 0x30FF || // 片假名 | Katakana
			r >= 0x3400 && r <= 0x4DBF || // CJK统一表意文字扩展A | CJK Unified Ideographs Extension A
			r >= 0x20000 && r <= 0x2A6DF || // CJK统一表意文字扩展B | CJK Unified Ideographs Extension B
			r >= 0x2A700 && r <= 0x2B73F || // CJK统一表意文字扩展C | CJK Unified Ideographs Extension C
			r >= 0x2B740 && r <= 0x2B81F || // CJK统一表意文字扩展D | CJK Unified Ideographs Extension D
			r >= 0x2B820 && r <= 0x2CEAF || // CJK统一表意文字扩展E | CJK Unified Ideographs Extension E
			r >= 0xAC00 && r <= 0xD7AF || // 朝鲜文音节 | Hangul Syllables
			r >= 0xF900 && r <= 0xFAFF || // CJK兼容表意文字 | CJK Compatibility Ideographs
			r >= 0xFF00 && r <= 0xFFEF {
			runeWidth = 2
		} else {
			runeWidth = 1
		}
		totalWidth += runeWidth
	}

	// 如果总宽度不超过最大宽度，则不需要截断
	// If total width does not exceed max width, no need to truncate
	if totalWidth <= maxBytes {
		return s
	}

	// 需要截断
	// Need to truncate
	var currentBytes int
	var truncated []rune
	for _, r := range s {
		var runeWidth int
		if r >= 0x4E00 && r <= 0x9FFF || // CJK统一表意文字 | CJK Unified Ideographs
			r >= 0x3040 && r <= 0x309F || // 平假名 | Hiragana
			r >= 0x30A0 && r <= 0x30FF || // 片假名 | Katakana
			r >= 0x3400 && r <= 0x4DBF || // CJK统一表意文字扩展A | CJK Unified Ideographs Extension A
			r >= 0x20000 && r <= 0x2A6DF || // CJK统一表意文字扩展B | CJK Unified Ideographs Extension B
			r >= 0x2A700 && r <= 0x2B73F || // CJK统一表意文字扩展C | CJK Unified Ideographs Extension C
			r >= 0x2B740 && r <= 0x2B81F || // CJK统一表意文字扩展D | CJK Unified Ideographs Extension D
			r >= 0x2B820 && r <= 0x2CEAF || // CJK统一表意文字扩展E | CJK Unified Ideographs Extension E
			r >= 0xAC00 && r <= 0xD7AF || // 朝鲜文音节 | Hangul Syllables
			r >= 0xF900 && r <= 0xFAFF || // CJK兼容表意文字 | CJK Compatibility Ideographs
			r >= 0xFF00 && r <= 0xFFEF {
			runeWidth = 2
		} else {
			runeWidth = 1
		}

		if currentBytes+runeWidth > maxBytes-3 { // 为省略号预留空间 | Reserve space for ellipsis
			break
		}

		truncated = append(truncated, r)
		currentBytes += runeWidth
	}

	return string(truncated) + "..."
}

// outputTable 以表格形式输出结果
// outputTable outputs results in table format
func (c *S3Cleaner) outputTable(files []FileInfo) error {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"存储桶 | Bucket", "键 | Key", "大小 | Size", "修改时间 | Mod Time", "状态 | Status"})
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderColor(
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgCyanColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgCyanColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgCyanColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgCyanColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgCyanColor},
	)

	// 设置列宽，增加 Key 列的宽度
	// Set column width, increase Key column width
	table.SetColWidth(120)

	for _, file := range files {
		// 截断过长的键
		// Truncate long keys
		key := file.Key
		if utf8.RuneCountInString(key)*3 > 120 {
			key = truncateString(key, 120)
		}

		// 格式化时间
		// Format time
		timeStr := file.ModTime.Format("2006-01-02 15:04:05")

		// 格式化大小
		// Format size
		sizeStr := formatSize(file.Size)

		// 格式化状态，使用表情符号和文字
		// Format status with emoji and text
		var statusStr string
		var statusColor tablewriter.Colors

		if file.DeleteSuccess == nil {
			// 未执行删除操作时，显示是否会被命中删除
			// When deletion is not executed, show if it would be targeted for deletion
			if file.ShouldDelete {
				statusStr = "🎯 Will delete" // 会被命中删除 | Would be targeted for deletion
				statusColor = tablewriter.Colors{tablewriter.FgHiRedColor}
			} else {
				statusStr = "🔍 Won't delete" // 不会被命中删除 | Would not be targeted for deletion
				statusColor = tablewriter.Colors{tablewriter.FgYellowColor}
			}
		} else if *file.DeleteSuccess {
			statusStr = "✅ Deleted" // 删除成功 | Deletion Success
			statusColor = tablewriter.Colors{tablewriter.FgGreenColor}
		} else {
			statusStr = "❌ Delete failed" // 删除失败 | Deletion Failed
			statusColor = tablewriter.Colors{tablewriter.FgRedColor}
		}

		// 根据是否应删除设置时间列的颜色
		// Set time column color based on should delete
		var timeColor tablewriter.Colors
		if file.ShouldDelete {
			timeColor = tablewriter.Colors{tablewriter.FgHiRedColor}
		} else {
			timeColor = tablewriter.Colors{tablewriter.FgYellowColor}
		}

		row := []string{file.Bucket, key, sizeStr, timeStr, statusStr}
		colors := []tablewriter.Colors{
			tablewriter.Colors{tablewriter.FgHiBlueColor},
			tablewriter.Colors{tablewriter.FgWhiteColor},
			tablewriter.Colors{tablewriter.FgHiCyanColor},
			timeColor,
			statusColor,
		}

		table.Rich(row, colors)
	}

	if len(files) == 0 {
		color.Yellow("未找到临时文件\nNo temporary files found")
	} else {
		table.Render()

		// 计算总容量和应删除的容量
		// Calculate total size and size to delete
		var totalSize, sizeToDelete, sizeDeleted int64
		var countToDelete, countDeleted int
		for _, file := range files {
			totalSize += file.Size
			if file.ShouldDelete {
				sizeToDelete += file.Size
				countToDelete++
			}
			// 计算已删除的文件数量和大小
			// Calculate deleted files count and size
			if file.DeleteSuccess != nil && *file.DeleteSuccess {
				sizeDeleted += file.Size
				countDeleted++
			}
		}

		// 输出统计信息
		// Output statistics
		fmt.Println()

		// 创建统计信息表格
		// Create statistics table
		statTable := tablewriter.NewWriter(os.Stdout)
		statTable.SetHeader([]string{"统计信息 | Statistics", "值 | Value"})
		statTable.SetAutoWrapText(false)
		statTable.SetAutoFormatHeaders(true)
		statTable.SetHeaderColor(
			tablewriter.Colors{tablewriter.Bold, tablewriter.FgCyanColor},
			tablewriter.Colors{tablewriter.Bold, tablewriter.FgCyanColor},
		)
		statTable.SetColumnColor(
			tablewriter.Colors{tablewriter.FgHiWhiteColor},
			tablewriter.Colors{tablewriter.FgHiWhiteColor},
		)

		// 添加统计数据行
		// Add statistics data rows
		statTable.Append([]string{"总文件数 | Total files", fmt.Sprintf("%d", len(files))})
		statTable.Append([]string{"总容量 | Total size", formatSize(totalSize)})
		statTable.Append([]string{"应删除文件数 | Files to delete", fmt.Sprintf("%d", countToDelete)})
		statTable.Append([]string{"应删除容量 | Size to delete", formatSize(sizeToDelete)})
		statTable.Append([]string{"已删除文件数 | Files deleted", fmt.Sprintf("%d", countDeleted)})
		statTable.Append([]string{"已删除容量 | Size deleted", formatSize(sizeDeleted)})

		// 渲染统计表格
		// Render statistics table
		statTable.Render()
	}
	return nil
}

// formatSize 格式化文件大小
// formatSize formats file size
func formatSize(size int64) string {
	const (
		B  = 1
		KB = 1024 * B
		MB = 1024 * KB
		GB = 1024 * MB
		TB = 1024 * GB
	)

	switch {
	case size >= TB:
		return fmt.Sprintf("%.2f TB", float64(size)/float64(TB))
	case size >= GB:
		return fmt.Sprintf("%.2f GB", float64(size)/float64(GB))
	case size >= MB:
		return fmt.Sprintf("%.2f MB", float64(size)/float64(MB))
	case size >= KB:
		return fmt.Sprintf("%.2f KB", float64(size)/float64(KB))
	default:
		return fmt.Sprintf("%d B", size)
	}
}

// outputJSON 以JSON格式输出结果
// outputJSON outputs results in JSON format
func (c *S3Cleaner) outputJSON(files []FileInfo) error {
	result := struct {
		Files []FileInfo `json:"files"`
		Total int        `json:"total"`
	}{
		Files: files,
		Total: len(files),
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("无法序列化为JSON: %v\nFailed to serialize to JSON: %v", err, err)
	}

	fmt.Println(string(jsonData))
	return nil
}

// outputCSV 以CSV格式输出结果
// outputCSV outputs results in CSV format
func (c *S3Cleaner) outputCSV(files []FileInfo) error {
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	// 写入表头
	// Write header
	if err := writer.Write([]string{"Bucket", "Key", "Size", "ModTime", "ShouldDelete", "DeleteSuccess"}); err != nil {
		return fmt.Errorf("无法写入CSV表头: %v\nFailed to write CSV header: %v", err, err)
	}

	// 写入数据
	// Write data
	for _, file := range files {
		shouldDelete := "false"
		if file.ShouldDelete {
			shouldDelete = "true"
		}

		deleteSuccess := "not_executed"
		if file.DeleteSuccess != nil {
			if *file.DeleteSuccess {
				deleteSuccess = "true"
			} else {
				deleteSuccess = "false"
			}
		}

		if err := writer.Write([]string{
			file.Bucket,
			file.Key,
			fmt.Sprintf("%d", file.Size),
			file.ModTime.Format(time.RFC3339),
			shouldDelete,
			deleteSuccess,
		}); err != nil {
			return fmt.Errorf("无法写入CSV数据: %v\nFailed to write CSV data: %v", err, err)
		}
	}

	return nil
}
