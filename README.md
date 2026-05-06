# GoStarMap

Real-time 3D solar system and procedural-galaxy simulator in Go, using raylib for rendering and Keplerian orbital mechanics for planet motion.

## What it does

- Renders the Sun and 8 planets with physically-based lighting and custom GLSL shaders
- Procedurally generates a 100,000-star Milky Way (galactic disk, central bulge, two spiral arms) plus a handful of hand-placed nearby stars
- Solves Kepler's equation each frame to advance planet positions; orbital elements come from JPL Horizons J2000.0 data
- Adjustable simulation time, from paused to 100 years per real second
- Five tunable lighting presets (Realistic, Cinematic, Educational, Stylized, Dark) with runtime keyboard adjustment
- Targeting reticle: point the camera at the Sun or a planet to see mass, diameter, orbital period, eccentricity, and inclination

## Requirements

- Go 1.21+
- A C compiler (raylib-go uses cgo)
  - Windows: MinGW-w64 or MSVC
  - Linux: `build-essential`, plus `libgl1-mesa-dev libxi-dev libxcursor-dev libxrandr-dev libxinerama-dev`
  - macOS: Xcode command-line tools

## Build and run

```bash
go mod download
go run .                              # NOT `go run main.go` — sources span multiple files
go build -o gostarmap
./gostarmap                           # run from the repo root; shaders are loaded via relative paths
./gostarmap --width 2560 --height 1440 --stars 250000
./gostarmap --fullscreen
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--width N` | `1920` | Initial window width |
| `--height N` | `1080` | Initial window height |
| `--stars N` | `100000` | Procedural background star count |
| `--fullscreen` | off | Open fullscreen at the requested resolution |

The window is resizable; the HUD reflows when you drag the edges.

## Controls

**Camera and navigation**

| Key | Action |
|-----|--------|
| `WASD` | Move horizontally |
| `Space` / `Left Ctrl` | Up / down |
| `Shift` | Speed boost (scales with distance from origin) |
| Mouse | Look (yaw/pitch, pitch clamped to ±89°) |
| `F` | Teleport to whatever the reticle is targeting (planet or Sun) |
| `C` | Toggle constellation lines from Sun to nearby named stars |
| `Tab` | Toggle stats overlay |
| `Esc` | Exit |

**Time**

| Key | Action |
|-----|--------|
| `P` | Pause / resume |
| `[` / `]` | Halve / double speed |
| `\` | Reverse time direction |
| `J` | Jump to J2000.0 epoch (reset clock) |
| `1`–`5` | Speed presets (1 day/sec → 1 year/sec) |

**Lighting**

| Key | Action |
|-----|--------|
| `6` `7` `8` `9` `0` | Realistic / Cinematic / Educational / Stylized / Dark |
| `L` / `K` | Sun intensity up / down |
| `O` / `I` | Ambient up / down |
| `T` / `Y` | Terminator softness up / down |
| `R` | Toggle rim lighting |
| Hold `Shift` | 3× adjustment speed |

## Project layout

- `main.go` — types (`Star`, `Planet`, `Galaxy`), procedural galaxy generation, Sun and planet renderers, camera/input, game loop
- `lighting_config.go` — `LightingConfig`, preset table, keyboard hot-reload, shader uniform updates
- `orbital/` — package: Newton-Raphson Kepler solver, JPL J2000.0 orbital elements, time-scale helpers (with tests)
- `internal/celestial/` — package: pure-Go helpers (`SpectralType`, `RandomType`, `FormatNumber`) split out so they're testable without raylib on the path
- `shaders/` — GLSL: Sun (HDR emission), planets (`planet_enhanced.fs`, falling back to `planet.fs`)
- `CLAUDE.md` — guidance for Claude Code instances working in this repo

## Tests and lint

```bash
go test ./orbital/... ./internal/...   # pure-Go subpackages — no raylib needed
go vet ./...
golangci-lint run                      # config in .golangci.yml
```

Tests in `package main` can't run in a headless environment because raylib-go's
loader pulls in the native library at package init. CI runs the subpackage
tests only (`.github/workflows/lint.yml`).

## Accuracy notes

Orbital elements are real (JPL Horizons, J2000.0 epoch). The Kepler solver converges to ~10⁻¹² rad. The Sun radius and AU scale are intentionally not mutually consistent — sizes are compressed for navigability.

## License

MIT — see [LICENSE](LICENSE).
