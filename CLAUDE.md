# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Note on the README

`README.md` is stale. It describes a much smaller "star distance toy." The actual code is a real-time 3D solar-system + procedural-galaxy simulator with custom shaders, Keplerian orbital mechanics, and runtime-tunable lighting. Trust the source over the README.

## Build / run

This is a single `package main` spread across three files in the repo root, plus GLSL shaders in `shaders/`. The files reference each other's symbols, so:

```bash
go run .          # NOT `go run main.go` — that fails with undefined refs
go build -o gostarmap
./gostarmap       # must be invoked from the repo root; shaders are loaded via relative paths like "shaders/sun.vs"
go vet ./...
go test ./...     # no *_test.go files exist yet — passes vacuously
```

Dependency: `github.com/gen2brain/raylib-go/raylib` v0.55.1, which is **cgo** — first build is slow and needs a C compiler + OpenGL headers (MinGW/MSVC on Windows, build-essential + libgl/libxi/libxrandr/libxcursor/libxinerama dev packages on Linux). Go 1.21.

## Architecture

Three source files, one package:

- **`main.go`** — entry point and most of the engine. Contains: `Star` / `Planet` / `Galaxy` types, `GenerateGalaxy` (procedural 100k-star galactic disk + bulge + spiral arms, plus 8 planets and ~9 hand-named nearby stars), the Sun and Planet shader renderers (`InitSunRenderer` / `InitPlanetRenderer` / `RenderAllPlanets` / `RenderSunWithBloom`), star LOD (`RenderStarLOD`), camera input, `CheckTargeting` reticle hit-testing, and the main game loop.
- **`orbital_mechanics.go`** — pure math, no rendering deps. Newton-Raphson Kepler solver (`SolveKeplerEquation`), `OrbitalElementsToPosition` (Keplerian → Cartesian), `GetPlanetaryElements` (hard-coded JPL Horizons J2000.0 elements for the 8 planets), `TimeScaleInfo` + `UpdateSimulationTime`, and `PositionToRenderUnits`.
- **`lighting_config.go`** — `LightingConfig` with 5 presets (Realistic / Cinematic / Educational / Stylized / Dark), keyboard hot-reload via `UpdateWithKeyboard`, and `ApplyToShader` which pushes ~13 uniforms to the planet shader each frame.

`shaders/` contains GLSL: `sun.vs/fs`, `planet.vs/fs`, `planet_enhanced.fs` (preferred — `InitPlanetRenderer` falls back to `planet.fs` if it fails to load), `default.vs`, and the bloom pipeline (`bloom_extract.fs` / `bloom_blur.fs` / `bloom_composite.fs`).

## Coordinate and scale conventions (easy to break)

- Astronomy uses **Z-up**, raylib uses **Y-up**. `PositionToRenderUnits` swaps Y↔Z on output. Anything that bridges orbital math and raylib coordinates must go through it.
- `AUScale = 150.0` render units per AU. Stars are scaled by `lyScale = 50.0` units per light-year (defined locally in `GenerateGalaxy`). The Sun sits at the origin with radius `SunRadiusUnits = 32.7` (109× Earth, but compressed against the 150 units/AU scale — *not* physically self-consistent with the AU scale, deliberately, for navigability).
- Simulation time is days since the **J2000.0 epoch** (Jan 1 2000, 12:00 TT). `TimeScale.TimeScale` is sim-days per real-second; planet positions are re-solved from Keplerian elements every frame in `UpdatePlanetPositions`.

## Performance shape

Designed around a 60 FPS budget at 1920×1080. The galaxy holds 100k stars but `maxStarsPerFrame = 15000` and `maxRenderDistance = 20000` cull aggressively in the main loop. Star LOD has four bands (full sphere → smaller sphere → tiny sphere → `DrawPoint3D`). Planet LOD reuses three pre-built sphere meshes (32/24/16 rings·slices) cached on `PlanetRenderData`. Bloom render textures are half-resolution.

Note: `RenderSunWithBloom` currently bypasses the bloom post-process and just calls `RenderSun` directly — the bloom textures are loaded but the multi-pass pipeline is a TODO in code.

## CI

`.github/workflows/sonarcloud.yml` runs `go test -coverprofile=coverage.out` (with `continue-on-error`) and uploads to SonarCloud on pushes to `main` / `master` / `develop` / `claude/**` and on PRs. Project key is `ctclostio_GoStarMap`. There are no tests yet, so coverage will read as 0%.
