package flags

import (
	"github.com/spf13/cobra"
)

var (
	// Node   = flag.String("node", "", "If non-empty, start with this node")
	// Target = flag.String("target", "", "Service to start")
	// Level  = flag.String("level", "INFO", "Log level")

	// 定义需要的参数变量
	Node   string
	Target string
	Level  string

	// 主命令
	rootCmd = &cobra.Command{
		Use: "app",
		Run: func(cmd *cobra.Command, args []string) {
			// 主程序逻辑会在这里执行
		},
	}
)

func init() {
	// 定义参数
	rootCmd.Flags().StringVar(&Node, "node", "", "If non-empty, start with this node")
	rootCmd.Flags().StringVar(&Target, "target", "", "Service to start")
	rootCmd.Flags().StringVar(&Level, "level", "INFO", "Log level")

	err := rootCmd.Execute()
	if err != nil {
		panic(err)
	}
}
