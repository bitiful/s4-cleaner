/*
 * Copyright (c) 2025 缤纷云S4 (Bitiful S4)
 *
 * 缤纷云S3临时文件清理工具
 * Bitiful S4 S3 Temporary File Cleaner
 */

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/bitiful/s4-cleaner/pkg/cleaner"
	"github.com/bitiful/s4-cleaner/pkg/config"
	"github.com/spf13/cobra"
)

var (
	// 版本信息 | Version information
	Version = "0.1.0"

	// 配置选项 | Configuration options
	cfg = &config.Config{}
)

// rootCmd 表示没有调用子命令时的基础命令
// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "s4-cleaner",
	Short: "S3临时文件清理工具 | S3 Temporary File Cleaner",
	Long: `S3临时文件清理工具 - 清理S3存储桶中的过期临时文件
S3 Temporary File Cleaner - Clean up expired temporary files in S3 buckets

版本 | Version: ` + Version + `

使用示例 | Usage examples:
  # 列出所有桶中7天前的临时文件（默认）
  # List temporary files older than 7 days in all buckets (default)
  AWS_ACCESS_KEY_ID=your_ak AWS_SECRET_ACCESS_KEY=your_sk s4-cleaner

  # 列出指定桶中3天前的临时文件
  # List temporary files older than 3 days in the specified bucket
  AWS_ACCESS_KEY_ID=your_ak AWS_SECRET_ACCESS_KEY=your_sk s4-cleaner --bucket=my-bucket --olderThan=3d

  # 删除指定桶中72小时前的临时文件
  # Delete temporary files older than 72 hours in the specified bucket
  AWS_ACCESS_KEY_ID=your_ak AWS_SECRET_ACCESS_KEY=your_sk s4-cleaner --bucket=my-bucket --olderThan=72h --doDelete

  # 以JSON格式输出
  # Output in JSON format
  AWS_ACCESS_KEY_ID=your_ak AWS_SECRET_ACCESS_KEY=your_sk s4-cleaner --fmt=json
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 检查必要的环境变量 | Check required environment variables
		accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
		secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")

		if accessKey == "" || secretKey == "" {
			return fmt.Errorf("必须设置环境变量 AWS_ACCESS_KEY_ID 和 AWS_SECRET_ACCESS_KEY\nEnvironment variables AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY must be set")
		}

		// 创建清理器 | Create cleaner
		s3Cleaner, err := cleaner.NewS3Cleaner(accessKey, secretKey, cfg)
		if err != nil {
			return err
		}

		// 执行清理操作 | Execute cleaning operation
		return s3Cleaner.Run()
	},
}

// Execute 添加所有子命令到根命令并设置标志
// Execute adds all child commands to the root command and sets flags appropriately
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// 初始化标志 | Initialize flags
	rootCmd.PersistentFlags().StringVar(&cfg.Bucket, "bucket", "", "存储桶名称，为空表示所有桶 | Bucket name, empty means all buckets")
	rootCmd.PersistentFlags().StringVar(&cfg.Time, "olderThan", "7d", "查找早于此时间的文件，如 '7d'（7天前）或 '72h'（72小时前） | Find files older than this time, e.g. '7d' (7 days ago) or '72h' (72 hours ago)")
	rootCmd.PersistentFlags().BoolVar(&cfg.DoDelete, "doDelete", false, "是否执行删除操作，默认为false（仅列出） | Whether to perform deletion, default is false (list only)")
	rootCmd.PersistentFlags().StringVar(&cfg.Format, "fmt", "table", "输出格式：table, json, csv | Output format: table, json, csv")

	// 添加版本标志 | Add version flag
	rootCmd.PersistentFlags().BoolP("version", "v", false, "显示版本信息 | Show version information")
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		versionFlag, _ := cmd.Flags().GetBool("version")
		if versionFlag {
			fmt.Printf("S3临时文件清理工具 | S3 Temporary File Cleaner\n版本 | Version: %s\n", Version)
			os.Exit(0)
		}

		// 验证格式标志 | Validate format flag
		validFormats := map[string]bool{"table": true, "json": true, "csv": true}
		if !validFormats[strings.ToLower(cfg.Format)] {
			fmt.Fprintf(os.Stderr, "错误：无效的格式 '%s'，有效选项为: table, json, csv\nError: Invalid format '%s', valid options are: table, json, csv\n", cfg.Format, cfg.Format)
			os.Exit(1)
		}
	}
}
