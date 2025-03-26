/*
 * Copyright (c) 2025 缤纷云S4 (Bitiful S4)
 * 
 * 缤纷云S3临时文件清理工具
 * Bitiful S4 S3 Temporary File Cleaner
 */

package config

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

// Config 存储命令行配置
// Config stores command line configuration
type Config struct {
	// Bucket 存储桶名称，为空表示所有桶
	// Bucket name, empty means all buckets
	Bucket string

	// Time 查找早于此时间的文件，如 '7d'（7天前）或 '72h'（72小时前）
	// Find files older than this time, e.g. '7d' (7 days ago) or '72h' (72 hours ago)
	Time string

	// DoDelete 是否执行删除操作，默认为false（仅列出）
	// Whether to perform deletion, default is false (list only)
	DoDelete bool

	// Format 输出格式：table, json, csv
	// Output format: table, json, csv
	Format string

	// ExpirationTime 解析后的过期时间
	// Parsed expiration time
	ExpirationTime time.Time
}

// ParseTime 解析时间字符串为时间对象
// ParseTime parses time string to time object
func (c *Config) ParseTime() error {
	// 正则表达式匹配时间格式，如 7d, 72h
	// Regular expression to match time format, e.g. 7d, 72h
	re := regexp.MustCompile(`^(\d+)([dh])$`)
	matches := re.FindStringSubmatch(c.Time)

	if len(matches) != 3 {
		return fmt.Errorf("无效的时间格式 '%s'，有效格式为: 数字+单位，如 '7d'（7天）或 '72h'（72小时）\nInvalid time format '%s', valid format is: number+unit, e.g. '7d' (7 days) or '72h' (72 hours)", c.Time, c.Time)
	}

	value, err := strconv.Atoi(matches[1])
	if err != nil {
		return fmt.Errorf("无效的时间值 '%s'\nInvalid time value '%s'", matches[1], matches[1])
	}

	unit := matches[2]
	now := time.Now()

	switch unit {
	case "d":
		c.ExpirationTime = now.AddDate(0, 0, -value)
	case "h":
		c.ExpirationTime = now.Add(-time.Duration(value) * time.Hour)
	default:
		return fmt.Errorf("无效的时间单位 '%s'，有效单位为: d（天）, h（小时）\nInvalid time unit '%s', valid units are: d (days), h (hours)", unit, unit)
	}

	return nil
}
