package tts

import (
	"bytes"
	"encoding/binary"
)

// PCMToWAV는 raw PCM 데이터 앞에 44바이트 WAV 헤더를 붙여서
// 재생 가능한 WAV 파일 바이트를 만들어 돌려줘.
//
// Gemini TTS 출력 스펙에 맞춰 고정값 사용:
//   - 샘플레이트: 24000 Hz
//   - 비트:       16-bit
//   - 채널:       1 (mono)
func PCMToWAV(pcm []byte) []byte {
	const (
		sampleRate    = 24000
		bitsPerSample = 16
		numChannels   = 1
	)

	// 파생 계산값
	byteRate := sampleRate * numChannels * bitsPerSample / 8
	blockAlign := numChannels * bitsPerSample / 8
	dataSize := len(pcm)      // PCM 데이터 크기
	fileSize := 36 + dataSize // 전체 파일 크기 - 8

	buf := new(bytes.Buffer)

	// ── RIFF 청크 ──
	buf.WriteString("RIFF")                                  // 파일 시작 표식
	binary.Write(buf, binary.LittleEndian, uint32(fileSize)) // 파일 크기
	buf.WriteString("WAVE")                                  // 포맷 종류

	// ── fmt 청크 (오디오 형식 정보) ──
	buf.WriteString("fmt ")                                       // (끝 공백 주의)
	binary.Write(buf, binary.LittleEndian, uint32(16))            // fmt 청크 크기
	binary.Write(buf, binary.LittleEndian, uint16(1))             // 1 = PCM
	binary.Write(buf, binary.LittleEndian, uint16(numChannels))   // 채널 수
	binary.Write(buf, binary.LittleEndian, uint32(sampleRate))    // 샘플레이트
	binary.Write(buf, binary.LittleEndian, uint32(byteRate))      // 초당 바이트
	binary.Write(buf, binary.LittleEndian, uint16(blockAlign))    // 블록 정렬
	binary.Write(buf, binary.LittleEndian, uint16(bitsPerSample)) // 비트 수

	// ── data 청크 (실제 오디오) ──
	buf.WriteString("data")                                  // 데이터 시작 표식
	binary.Write(buf, binary.LittleEndian, uint32(dataSize)) // 데이터 크기
	buf.Write(pcm)                                           // PCM 본체

	return buf.Bytes()
}
