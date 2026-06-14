package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// rootCmd는 speakr의 기본 명령.
// 아무 서브커맨드 없이 그냥 "speakr"만 쳤을 때 실행됨.
var rootCmd = &cobra.Command{
	Use:   "speakr",
	Short: "Gemini TTS로 텍스트를 음성으로 변환하는 CLI",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("hello speakr 👋")
	},
}

// Execute는 main.go에서 호출하는 진입 함수이다.
func Execute() error {
	return rootCmd.Execute()
}
