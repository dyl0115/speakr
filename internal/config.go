package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

/// Config는 speakr의 모든 설정정보를 저장하는 구조체
/// Json 파일로 저장될 때 각 필드가 json 태그 이름으로 기록됨.

/// Go는 문법상 구조체의 필드는 대문자로 시작해야 함.
/// 하지만 json 파일은 google_tts_api_key처럼 스네이크 방식으로 쓰고 싶은데, 오른쪽의 태그가 json으로 변환시 자동으로 변환해준다.
/// 결론적으로 아래와 같은 형식의 json 파일이 생성된다.
/// {
///  "google_tts_api_key": "...",
///  "default_voice": "Leda",
///  "output_dir": "/srv/audio"
/// }

type Config struct {
	GoogleTTSAPIKey string `json:"google_tts_api_key"`
	DefaultVoice    string `json:"default_voice"`
	DefaultModel    string `json:"default_model"`
	OutputDir       string `json:"output_dir"`
}

// / defaults는 기본값이 설정된 Config 구조체를 리턴한다.
// / 사용자가 아직 아무것도 설정하지 않았을 때, 이 값을 환경설정값으로 사용하도록 한다.
// / 기본값 환경설정 주머니라고 생각하자.
// / google tts api key는비워두었다. 기본키라는 것이 있을 수 없기 때문이다.
func defaults() Config {
	return Config{
		DefaultVoice: "Leda",
		DefaultModel: "gemini-2.5-pro-preview-tts",
		OutputDir:    "/srv/audio",
	}
}

// / configPath()는 이 설정 파일으 저장될 위치를 리턴한다.
// / os.UserConfigDir()이 OS별 적절한 설정 폴더 경로를 알아서 리턴해준다.
// / eg1. 리눅스: ~/.config/speakr/config.json
// / eg2. 윈도우: C:\Users\<유저명>\AppData\Roaming\speakr\config.json
func configPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "speakr", "config.json"), nil
}

// Load는 설정 파일을 읽어서 Config 객체에 바인딩 한다.
// 파일이 없다면, 에러 대신 기본값을 돌려주도록 한다. (첫 실행 대응)
func Load() (Config, error) {
	cfg := defaults() // 일단 기본값으로 시작

	path, err := configPath()
	if err != nil {
		return cfg, err
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return cfg, nil // 파일 없음 = 첫 실행. 기본값 그대로 반환.
	}
	if err != nil {
		return cfg, err // 그 외 진짜 에러 (권한 등)
	}

	// 파일 내용(JSON)을 cfg 구조체에 채워넣기
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("config 파싱 실패: %w", err)
	}

	// 파일에 빈 값이 있으면 기본값으로 보정
	d := defaults()
	if cfg.DefaultVoice == "" {
		cfg.DefaultVoice = d.DefaultVoice
	}
	if cfg.DefaultModel == "" {
		cfg.DefaultModel = d.DefaultModel
	}
	if cfg.OutputDir == "" {
		cfg.OutputDir = d.OutputDir
	}

	return cfg, nil
}

// Save는 현재 Config를 설정 파일에 저장
// 설정 폴더가 없으면 자동으로 만들고, 파일은 0600 권한(본인만 읽기)으로 작성한다.
// 앞의 (c Config)는 "Save()함수는 Config에 속한 함수다"라는 의미로, 이후에 cfg.Save()로 쓴다. 즉 cfg라는 Config의 Save()를 호출한다.
func (c Config) Save() error {
	path, err := configPath()
	if err != nil {
		return err
	}

	// ~/.config/speakr/ 폴더가 없으면 생성 (0700: 본인만 접근)
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	// 구조체 보기 좋은 Json 텍스트로 변환
	// 세번째 인자는 들여쓰기를 공백 2칸으로 지정한다.
	// 결과는 아래와 같을 것이다.
	// {
	//    "google_tts_api_key": "...",
	//    "default_voice": "Leda",
	//    "output_dir": "/srv/audio"
	// }
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	// 파일 쓰기 (0600: 본인만 읽고 쓰기 가능 — API 키 보호)
	return os.WriteFile(path, data, 0600)
}

// ResolveAPIKey는 실제로 사용할 API 키를 결정한다.
// 우선순위: 환경변수(GOOGLE_TTS_API_KEY) > 설정 파일.
// 환경변수가 있으면 그걸 우선 쓰고, 없으면 config에 저장된 키를 씀.
func (c Config) ResolveAPIKey() string {
	if v := os.Getenv("GOOGLE_TTS_API_KEY"); v != "" {
		return v
	}
	return c.GoogleTTSAPIKey
}

// MaskKey는 키를 화면에 보여줄 때 가운데를 가려줌.
// 예: "AIzaSyD...8xQ2"  (앞 4자 + 뒤 4자만 노출)
func MaskKey(key string) string {
	if key == "" {
		return "(설정 안 됨)"
	}
	if len(key) <= 8 {
		return "****" // 너무 짧으면 통째로 가림
	}
	return key[:4] + "..." + key[len(key)-4:]
}
