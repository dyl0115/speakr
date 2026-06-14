package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	// "github.com/dyl0115/speakr/internal/config"
	config "github.com/dyl0115/speakr/internal"
	"github.com/dyl0115/speakr/internal/tts"
	"github.com/spf13/cobra"
)

// sayCmd는 "speakr say <텍스트>" — 텍스트를 음성(WAV)으로 변환해서 저장해.
var sayCmd = &cobra.Command{
	Use:   "say <텍스트>",
	Short: "텍스트를 음성으로 변환해 WAV 파일로 저장",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		text := args[0]

		// 1. 설정 읽기
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		// 2. API 키 확인 (없으면 친절히 안내)
		apiKey := cfg.ResolveAPIKey()
		if apiKey == "" {
			return fmt.Errorf("API 키가 없습니다. 'speakr config set-key <KEY>'로 먼저 설정하세요")
		}

		// 3. 텍스트 → PCM (Gemini TTS 호출)
		fmt.Printf("음성 생성 중... (음성: %s, 모델: %s)\n", cfg.DefaultVoice, cfg.DefaultModel)

		pcm, err := tts.Synthesize(apiKey, cfg.DefaultModel, cfg.DefaultVoice, text)
		if err != nil {
			return fmt.Errorf("음성 생성 실패: %w", err)
		}

		// 4. PCM → WAV
		wav := tts.PCMToWAV(pcm)

		// 5. 출력 디렉토리 준비 (없으면 생성)
		if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
			return fmt.Errorf("출력 폴더 생성 실패: %w", err)
		}

		// 6. 타임스탬프 파일명 만들기
		filename := fmt.Sprintf("speakr_%s.wav", time.Now().Format("20060102_150405"))
		outPath := filepath.Join(cfg.OutputDir, filename)

		// 7. 파일 저장
		if err := os.WriteFile(outPath, wav, 0644); err != nil {
			return fmt.Errorf("파일 저장 실패: %w", err)
		}

		fmt.Printf("✅ 저장 완료: %s\n", outPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(sayCmd)
}
