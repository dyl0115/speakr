package cmd

import (
	"fmt"

	config "github.com/dyl0115/speakr/internal"
	"github.com/spf13/cobra"
)

// configCmd는 "speakr config" 부모 명령이다.
// 자기 혼자서는 하는 일이 없고, 아래 자식 명령들을 묶는 그룹 역할
// Run이 없는 이유는 부모 그룹 명령은 직접 하는 일이 없어서.
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "speakr 설정 관리 (키, 음성, 출력 경로)",
}

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "현재 설정 보기",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		fmt.Println("현재 speakr 설정:")
		fmt.Printf("  API 키      : %s\n", config.MaskKey(cfg.ResolveAPIKey()))
		fmt.Printf("  기본 음성   : %s\n", cfg.DefaultVoice)
		fmt.Printf("  기본 모델   : %s\n", cfg.DefaultModel)
		fmt.Printf("  출력 경로   : %s\n", cfg.OutputDir)
		return nil
	},
}

// setKeyCmd는 "speakr config set-key <KEY>" — API 키를 저장한다.
var setKeyCmd = &cobra.Command{
	Use:   "set-key <KEY>",
	Short: "Google TTS API 키 설정",
	Args:  cobra.ExactArgs(1), // 인자 딱 1개 필수
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		cfg.GoogleTTSAPIKey = args[0] // 사용자가 입력한 키
		if err := cfg.Save(); err != nil {
			return err
		}

		fmt.Printf("API 키 저장 완료: %s\n", config.MaskKey(cfg.GoogleTTSAPIKey))
		return nil
	},
}

// setVoiceCmd는 "speakr config set-voice <NAME>" — 기본 음성을 바꿔.
var setVoiceCmd = &cobra.Command{
	Use:   "set-voice <NAME>",
	Short: "기본 음성 설정 (예: Leda)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		cfg.DefaultVoice = args[0]
		if err := cfg.Save(); err != nil {
			return err
		}

		fmt.Printf("기본 음성 설정 완료: %s\n", cfg.DefaultVoice)
		return nil
	},
}

// setModelCmd는 "speakr config set-model <NAME>" — TTS 모델을 변경
var setModelCmd = &cobra.Command{
	Use:   "set-model <NAME>",
	Short: "TTS 모델 설정 (예: gemini-2.5-pro-preview-tts, gemini-3.1-flash-tts-preview, gemini-2.5-flash-tts-preview)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		cfg.DefaultModel = args[0]
		if err := cfg.Save(); err != nil {
			return err
		}

		fmt.Printf("TTS 모델 설정 완료: %s\n", cfg.DefaultModel)
		return nil
	},
}

// setOutputCmd는 "speakr config set-output <DIR>" — 출력 경로를 바꾼다.
var setOutputCmd = &cobra.Command{
	Use:   "set-output <DIR>",
	Short: "출력 디렉토리 설정 (예: /srv/audio)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		cfg.OutputDir = args[0]
		if err := cfg.Save(); err != nil {
			return err
		}

		fmt.Printf("출력 경로 설정 완료: %s\n", cfg.OutputDir)
		return nil
	},
}

// init()은 이 패키지가 로드될 때 자동으로 한 번 실행된다.
// 여기서 명령들을 트리 구조로 연결한다.
func init() {
	// config 아래에 자식 명령들 붙이기
	configCmd.AddCommand(showCmd)
	configCmd.AddCommand(setKeyCmd)
	configCmd.AddCommand(setVoiceCmd)
	configCmd.AddCommand(setModelCmd)
	configCmd.AddCommand(setOutputCmd)

	// root 아래에 config 붙이기
	rootCmd.AddCommand(configCmd)
}
