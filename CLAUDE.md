# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build / run

```bash
go run .                              # NOT `go run main.go` — sources span multiple files
go build -o gostarmap
./gostarmap                           # run from the repo root; shaders load via relative paths
./gostarmap --width 2560 --height 1440 --stars 250000 --fullscreen
go vet ./...
go test ./orbital/... ./internal/...  # only the headless-friendly subpackages
```

Dependency: `github.com/gen2brain/raylib-go/raylib` v0.55.1 via the **purego** loader, not cgo — the binary dlopens `raylib.dll` (Windows) / `libraylib.so` / `libraylib.dylib` at runtime. First build needs a C compiler regardless (Go's runtime). Go 1.21.

## Architecture

Three Go modules in this repo:

- **`package main`** (root): `main.go` (~1300 lines) holds types (`Star` / `Planet` / `Galaxy`), `GenerateGalaxy(targetStars int)` (procedural 100k-star galactic disk + bulge + spiral arms, plus 8 planets and 9 hand-placed named stars), the Sun and planet shader renderers, star LOD, camera/input via yaw+pitch scalars, `CheckTargeting` reticle hit-testing, the `teleport` helper, and the main game loop. `lighting_config.go` holds `LightingConfig`, the `lightingPresets` map, keyboard hot-reload, and `ApplyToShader` (which uses uniform locations cached on `PlanetRenderData`).
- **`orbital`**: Newton-Raphson Kepler solver, `OrbitalElementsToPosition`, JPL Horizons J2000.0 element tables, `TimeScaleInfo` + `UpdateSimulationTime`, `PositionToRenderUnits`. Has its own tests.
- **`internal/celestial`**: `SpectralType` + Morgan-Keenan constants, `RandomType` (observed-distribution sampler), `FormatNumber`. Pure Go, no raylib import — split out so `go test` can run without raylib on the path. Has its own tests.

`shaders/` contains GLSL: `sun.vs/fs`, `planet.vs/fs`, `planet_enhanced.fs` (preferred — `InitPlanetRenderer` falls back to `planet.fs` if it fails to load), `default.vs`. The `bloom_*.fs` files exist but their loader was removed when the dead bloom pipeline was deleted; they're staged for a future re-implementation.

## Coordinate and scale conventions (easy to break)

- Astronomy uses **Z-up**, raylib uses **Y-up**. `orbital.PositionToRenderUnits` swaps Y↔Z on output. Anything that bridges orbital math and raylib coordinates must go through it.
- `AUScale = 150.0` render units per AU. Stars are scaled by `lyScale = 50.0` units per light-year (local const in `GenerateGalaxy`). The Sun sits at the origin with radius `SunRadiusUnits = 32.7` — *not* physically self-consistent with the AU scale, deliberately, for navigability.
- Simulation time is days since the **J2000.0 epoch** (Jan 1 2000, 12:00 TT). `TimeScale.TimeScale` is sim-days per real-second and may be **negative** (reverse time). The `[/]` clamp uses `|TimeScale|`. The `J` key resets `SimulationDays` to 0.

## Camera

Yaw and pitch are tracked as scalars (`yaw`, `pitch` locals in `main`). Each frame: mouse delta → update yaw/pitch → clamp pitch to ±~89.4° → derive `forward` (full 3D) and `right` (horizontal-only, derived from yaw alone). WASD moves `Position`; `Target` is rebuilt from `Position + forward` afterwards. This avoids the cross-product gimbal lock at zenith/nadir and keeps strafe motion level regardless of pitch.

## Performance shape

60 FPS budget at 1920×1080. The galaxy holds up to 100k stars but `MaxStarsPerFrame = 15000` and `MaxRenderDistance = 20000` cull aggressively in the main loop. Star LOD has four bands (full sphere → smaller sphere → tiny sphere → `DrawPoint3D`). `Star.Color` is precomputed at generation, so the per-frame star path doesn't do map lookups. Planet LOD reuses three pre-built sphere meshes (32/24/16 rings·slices) cached on `PlanetRenderData`. Per-frame planet shader uniforms split: globals (sunPosition, cameraPosition, lighting config) are written once before the per-planet loop; only color/radius/rotation are per-planet. Planet position/rotation updates are skipped while paused.

## CI

- `.github/workflows/lint.yml` runs `golangci-lint` (config in `.golangci.yml`) and `go test ./orbital/... ./internal/...` on push/PR.
- `.github/workflows/sonarcloud.yml` runs `go test -coverprofile=coverage.out` (with `continue-on-error`) and uploads to SonarCloud. Tests in `package main` won't run in CI because of the raylib loader; only the subpackage tests contribute to coverage.

## When extending

- New testable helpers that don't need raylib should live in `internal/celestial` (or a sibling `internal/...` package) so they stay testable in CI.
- New orbital-math features go in `orbital`. Keep that package raylib-free.
- New time-system features must respect that `TimeScale` can be negative.
- New camera features must update both `yaw`/`pitch` and `camera.Target`/`Position` (the `teleport` helper is the model).
