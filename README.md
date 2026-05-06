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
go run .             # NOT `go run main.go` — sources span multiple files
go build -o gostarmap
./gostarmap          # run from the repo root; shaders are loaded via relative paths
```

## Controls

**Camera**

| Key | Action |
|-----|--------|
| `WASD` | Move horizontally |
| `Space` / `Left Ctrl` | Up / down |
| `Shift` | Speed boost (scales with distance from origin) |
| Mouse | Look |
| `Tab` | Toggle stats overlay |
| `Esc` | Exit |

**Time**

| Key | Action |
|-----|--------|
| `P` | Pause / resume |
| `[` / `]` | Halve / double speed |
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
- `orbital_mechanics.go` — Newton-Raphson Kepler solver and JPL J2000.0 orbital elements for the 8 planets
- `lighting_config.go` — `LightingConfig`, presets, keyboard hot-reload, shader uniform updates
- `shaders/` — GLSL: Sun (HDR emission), planets (`planet_enhanced.fs`, falling back to `planet.fs`), and the bloom pipeline scaffolding
- `CLAUDE.md` — guidance for Claude Code instances working in this repo

## Accuracy notes

Orbital elements are real (JPL Horizons, J2000.0 epoch). The Kepler solver converges to ~10⁻¹² rad. The Sun radius and AU scale are intentionally not mutually consistent — sizes are compressed for navigability.

## License

MIT — see [LICENSE](LICENSE).
