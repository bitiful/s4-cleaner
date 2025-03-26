/*
 * Copyright (c) 2025 缤纷云S4 (Bitiful S4)
 *
 * 缤纷云S3临时文件清理工具
 * Bitiful S4 S3 Temporary File Cleaner
 */

package main

import (
	"fmt"
	"os"

	"github.com/bitiful/s4-cleaner/cmd"
)

func main() {
	// 执行根命令 | Execute root command
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\nError: %v\n", err, err)
		os.Exit(1)
	}
}
