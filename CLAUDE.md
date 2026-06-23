# CLAUDE.md

Guidance for AI assistants working in this repository.

## What this is

`speakr` is a Go CLI tool that converts text to speech using Google's Gemini TTS API. It synthesizes Korean (or any) text into 24kHz/16-bit mono PCM audio and writes it out as a WAV file. Module: `github.com/dyl0115/speakr`, Go 1.26.1.

## Layout

```
main.go              entry point, delegates to cmd.Execute()
cmd/
  root.go            root Cobra command
  say.go             `speakr say <text>` — synthesizes speech and saves a WAV file
  config.go          `speakr config` subcommands: show / set-key / set-voice / set-model / set-output
internal/
  config.go          Config struct + Load/Save, persisted at ~/.config/speakr/config.json
  tts/
    client.go         Synthesize() — calls the Gemini generateContent TTS endpoint over net/http
    wav.go            PCMToWAV() — wraps raw PCM in a 44-byte WAV header
```

Standard `cmd` (CLI surface, Cobra) / `internal` (private implementation) split. Keep new commands in `cmd/`, business logic in `internal/`.

## Build & run

```bash
go build -o speakr .
./speakr say "안녕하세요"
./speakr config show
```

There are no Makefile targets and no `go test` files in the repo today — there is no automated test suite. If you add behavior, prefer adding a `_test.go` alongside the package rather than introducing a new test framework.

## CI / Deployment

`.github/workflows/deploy.yml` runs on push to `main`:
1. Cross-compiles `GOOS=linux GOARCH=arm64 go build -o speakr .`
2. SCPs the binary to the deploy target and installs it at `/usr/local/bin/speakr`

There is no separate CI lint/test gate — the build itself is the only check before deploy. Be careful with changes to `main`; a successful build is what ships.

## Configuration & secrets

- Config lives at `~/.config/speakr/config.json` (dir `0700`, file `0600`).
- Fields: `google_tts_api_key`, `default_voice` (default `Leda`), `default_model` (default `gemini-2.5-pro-preview-tts`), `output_dir` (default `/srv/audio`).
- The `GOOGLE_TTS_API_KEY` environment variable overrides the stored key.
- `MaskKey()` redacts the key in CLI output (shows first/last 4 chars only) — never print the raw key.
- `.gitignore` excludes the built binary, `.env`, and `config.json` — never commit secrets or local config.

## Conventions

- Comments and CLI copy are written in Korean; match that style for user-facing strings in this CLI.
- Errors are wrapped with `fmt.Errorf("...: %w", err)`, not swallowed or logged-and-continued.
- HTTP calls to Gemini use a 180-second timeout (TTS synthesis can be slow) — don't lower this without reason.
- No third-party HTTP client or JSON library; stick to `net/http` and `encoding/json` to match the existing minimal-dependency style (only `spf13/cobra` and `pflag` are external deps).
- Audio format is fixed at 24kHz/16-bit/mono PCM from Gemini — `PCMToWAV` assumes this; don't generalize it speculatively.
