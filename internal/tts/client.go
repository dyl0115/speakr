package tts

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

/* gemini에게 보낼 json request 형식은 아래와 같음
	{
  "contents": [{ "parts": [{ "text": "읽을 텍스트" }] }],
  "generationConfig": {
    "responseModalities": ["AUDIO"],
    "speechConfig": {
      "voiceConfig": {
        "prebuiltVoiceConfig": { "voiceName": "Leda" }
      }
    }
  }
}

gemini가 주는 json response 형식은 아래와 같음.
{
  "candidates": [{
    "content": {
      "parts": [{
        "inlineData": {
          "data": "여기에_base64로_인코딩된_오디오..."
        }
      }]
    }
  }]
}
*/

// -------------요청 메시지 (우리서버 -> Google 서버)-----------------------------------
type generateRequest struct {
	Contents         []content        `json:"contents"`
	GenerationConfig generationConfig `json:"generationConfig"`
}

type content struct {
	Parts []part `json:"parts"`
}

type part struct {
	Text string `json:"text"`
}

type generationConfig struct {
	ResponseModalities []string     `json:"responseModalities"`
	SpeechConfig       speechConfig `json:"speechConfig"`
}

type speechConfig struct {
	VoiceConfig voiceConfig `json:"voiceConfig"`
}

type voiceConfig struct {
	PrebuiltVoiceConfig prebuiltVoiceConfig `json:"prebuiltVoiceConfig"`
}

type prebuiltVoiceConfig struct {
	VoiceName string `json:"voiceName"`
}

// -----------------------응답 메시지 (우리서버 <- Google 서버) ---------------------------

// Gemini가 돌려주는 JSON에서 우리가 필요한 부분만 추려서 구조체로 만든다.
// (응답엔 다른 필드도 많지만, 안 쓰는 건 그냥 생략해도 됨)

type generateResponse struct {
	Candidates []candidate `json:"candidates"`
}

type candidate struct {
	Content responseContent `json:"content"`
}

type responseContent struct {
	Parts []responsePart `json:"parts"`
}

type responsePart struct {
	InlineData inlineData `json:"inlineData"`
}

type inlineData struct {
	Data string `json:"data"` // base64로 인코딩된 오디오 데이터
}

/*
 Synthesize는 텍스트를 음성(raw PCM 바이트)으로 변환해서 돌려줌.

 매개변수는 아래와 같다.
	apiKey: 구글 TTS API 키
	model: 사용할 모델 (ex. gemini-2.5 pro perview -tts)
	voice: 음성 이름 (ex. Leda)
	text: 읽을 텍스트

 반환값은 아래와 같다.
	[] byte: raw PCM 오디오 데이터 (24kHz, 16-bit, mono)
	error: 실패 시 에러
*/
// go 문법에서는 매개변수 타입이 모두 다 string이면 마지막에 한번 string을 써주면 된다.\
// PCM -> WAV 변환은 /interal/wav.go가 맡도록 한다.
func Synthesize(apiKey, model, voice, text string) ([]byte, error) {

	// 1단계: 요청 구조체에 값 채우기
	reqBody := generateRequest{
		Contents: []content{
			{
				Parts: []part{
					{Text: text}, // tts가 읽을 텍스트
				},
			},
		},
		GenerationConfig: generationConfig{
			ResponseModalities: []string{"AUDIO"}, // 오디오로 받겠다.
			SpeechConfig: speechConfig{
				VoiceConfig: voiceConfig{
					PrebuiltVoiceConfig: prebuiltVoiceConfig{
						VoiceName: voice, // 음성 이름
					},
				},
			},
		},
	}

	// 2단계: 구조체 → JSON 텍스트(바이트)로 변환
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("요청 JSON 생성 실패: %w", err)
	}

	// 3단계: HTTP POST 요청 만들기
	url := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent",
		model,
	)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("요청 생성 실패: %w", err)
	}

	// 헤더 설정
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", apiKey)

	// 요청 전송 (타임아웃 60초)
	client := &http.Client{Timeout: 180 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("요청 전송 실패: %w", err)
	}
	defer resp.Body.Close()

	// 4단계: 응답 본문 전체 읽기
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("응답 읽기 실패: %w", err)
	}

	// HTTP 상태코드 확인 (200이 아니면 에러)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API 에러 (상태 %d): %s", resp.StatusCode, string(respData))
	}

	// 응답 JSON → 구조체로 변환
	var result generateResponse
	if err := json.Unmarshal(respData, &result); err != nil {
		return nil, fmt.Errorf("응답 파싱 실패: %w", err)
	}

	// 5단계: 응답 속 base64 오디오를 꺼내서 디코딩
	if len(result.Candidates) == 0 ||
		len(result.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("응답에 오디오 데이터가 없습니다")
	}

	base64Audio := result.Candidates[0].Content.Parts[0].InlineData.Data

	pcm, err := base64.StdEncoding.DecodeString(base64Audio)
	if err != nil {
		return nil, fmt.Errorf("오디오 디코딩 실패: %w", err)
	}

	return pcm, nil
}
