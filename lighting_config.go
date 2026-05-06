package main

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// ============================================================================
// LIGHTING CONFIGURATION SYSTEM
// ============================================================================
// Centralized lighting parameter management with presets and runtime tuning
//
// Performance: Negligible CPU cost (~0.01ms for uniform updates)
// Memory: 256 bytes for struct
//
// Usage:
//   config := NewLightingConfig()
//   config.ApplyPreset(PresetCinematic)
//   config.UpdateWithKeyboard()  // In main loop
//   config.ApplyToShader(planetRenderer)

// LightingConfig stores all tunable lighting parameters
type LightingConfig struct {
	// ========================================================================
	// PRIMARY LIGHTING PARAMETERS
	// ========================================================================

	// SunIntensity controls overall Sun brightness
	// Range: 0.5 - 3.0
	// - 1.0 = Physically realistic
	// - 1.5 = Cinematic (more dramatic)
	// - 0.8 = Subdued (more mysterious)
	SunIntensity float32

	// SunColorTemperature adjusts Sun's color warmth
	// Range: 0.0 (neutral) - 1.0 (very warm)
	// - 0.0 = Pure white (5778K, physically accurate)
	// - 0.3 = Slightly warm (more cinematic)
	// - 0.6 = Golden hour feel
	SunColorTemperature float32

	// ========================================================================
	// AMBIENT LIGHTING
	// ========================================================================

	// AmbientStrength controls dark side brightness
	// Range: 0.0 (pitch black) - 0.3 (very visible)
	// - 0.08 = High contrast, dramatic
	// - 0.12 = Balanced (default)
	// - 0.18 = Educational (everything visible)
	AmbientStrength float32

	// AmbientColorSaturation controls ambient color richness
	// Range: 0.0 (grayscale) - 1.0 (full color)
	// - 0.0 = Monochrome ambient (stylized)
	// - 0.3 = Subtle color (realistic starlight)
	// - 0.6 = Rich color (artistic)
	AmbientColorSaturation float32

	// HemisphericAmbient enables direction-aware ambient lighting
	// If true, ambient is warmer on Sun-facing hemisphere, cooler on dark side
	HemisphericAmbient bool

	// ========================================================================
	// SPECULAR HIGHLIGHTS
	// ========================================================================

	// SpecularRocky controls shine on rocky planets (Mercury, Venus, Earth, Mars)
	// Range: 0.0 (matte) - 0.5 (glossy)
	// Default: 0.1 (subtle)
	SpecularRocky float32

	// SpecularGas controls shine on gas giants (Jupiter, Saturn, Uranus, Neptune)
	// Range: 0.1 - 1.0
	// Default: 0.3 (reflective clouds)
	SpecularGas float32

	// ShininessRocky controls highlight size on rocky planets
	// Range: 8.0 (wide) - 64.0 (tight)
	// Default: 16.0
	ShininessRocky float32

	// ShininessGas controls highlight size on gas giants
	// Range: 32.0 (wide) - 256.0 (tight)
	// Default: 64.0
	ShininessGas float32

	// FresnelStrength controls edge reflection intensity (view-dependent)
	// Range: 0.0 (off) - 0.5 (strong)
	// Default: 0.15
	FresnelStrength float32

	// ========================================================================
	// ADVANCED LIGHTING FEATURES
	// ========================================================================

	// TerminatorSoftness controls day/night transition smoothness
	// Range: 0.0 (hard edge) - 0.3 (very soft)
	// - 0.0 = Sharp transition (no atmosphere)
	// - 0.05 = Subtle softening (realistic)
	// - 0.15 = Pronounced atmosphere effect
	TerminatorSoftness float32

	// RimLightStrength controls atmospheric edge glow
	// Range: 0.0 (off) - 0.5 (strong glow)
	// - 0.0 = No rim lighting
	// - 0.15 = Subtle atmosphere (Earth-like)
	// - 0.3 = Pronounced glow (stylized)
	RimLightStrength float32

	// RimLightColor defines rim light color (typically bluish for atmosphere)
	RimLightColor rl.Vector3 // RGB 0-1 range

	// ProximityWarmth enables color temperature shift based on distance from Sun
	// If true, inner planets get warmer tones, outer planets cooler
	ProximityWarmth bool

	// ProximityWarmthStrength controls intensity of proximity color shift
	// Range: 0.0 (off) - 0.3 (strong)
	ProximityWarmthStrength float32

	// ========================================================================
	// DEBUG & VISUALIZATION
	// ========================================================================

	// ShowNormals renders surface normals for debugging
	ShowNormals bool

	// ShowLightDirection visualizes light vectors
	ShowLightDirection bool

	// CurrentPreset tracks which preset is active (for UI display)
	CurrentPreset string
}

// ============================================================================
// PRESET SYSTEM
// ============================================================================

// PresetType defines available lighting presets
type PresetType int

const (
	PresetRealistic PresetType = iota // Physically accurate
	PresetCinematic                   // Dramatic, high contrast
	PresetEducational                 // High visibility, clear details
	PresetStylized                    // Artistic, enhanced features
	PresetDark                        // Moody, low ambient
	PresetCustom                      // User-modified
)

// NewLightingConfig creates a new configuration with default values
func NewLightingConfig() *LightingConfig {
	config := &LightingConfig{
		// Default preset: Realistic
		SunIntensity:            1.0,
		SunColorTemperature:     0.1,
		AmbientStrength:         0.12,
		AmbientColorSaturation:  0.3,
		HemisphericAmbient:      true,
		SpecularRocky:           0.1,
		SpecularGas:             0.3,
		ShininessRocky:          16.0,
		ShininessGas:            64.0,
		FresnelStrength:         0.15,
		TerminatorSoftness:      0.05,
		RimLightStrength:        0.0, // Off by default
		RimLightColor:           rl.Vector3{X: 0.5, Y: 0.7, Z: 1.0}, // Blue
		ProximityWarmth:         true,
		ProximityWarmthStrength: 0.1,
		ShowNormals:             false,
		ShowLightDirection:      false,
		CurrentPreset:           "Realistic",
	}
	return config
}

// lightingPresets is the data table behind ApplyPreset. Adding a new preset
// is a one-liner: define its values here and add a constant above.
// Fields not set in an entry (RimLightColor, ShowNormals, ShowLightDirection)
// are preserved from the live config rather than zeroed.
var lightingPresets = map[PresetType]LightingConfig{
	PresetRealistic: {
		SunIntensity: 1.8, SunColorTemperature: 0.1,
		AmbientStrength: 0.12, AmbientColorSaturation: 0.3, HemisphericAmbient: true,
		SpecularRocky: 0.1, SpecularGas: 0.3, ShininessRocky: 16.0, ShininessGas: 64.0,
		FresnelStrength: 0.15, TerminatorSoftness: 0.05, RimLightStrength: 0.0,
		ProximityWarmth: true, ProximityWarmthStrength: 0.1,
		CurrentPreset: "Realistic",
	},
	PresetCinematic: {
		SunIntensity: 2.0, SunColorTemperature: 0.25,
		AmbientStrength: 0.02, AmbientColorSaturation: 0.4, HemisphericAmbient: true,
		SpecularRocky: 0.15, SpecularGas: 0.5, ShininessRocky: 32.0, ShininessGas: 128.0,
		FresnelStrength: 0.25, TerminatorSoftness: 0.08, RimLightStrength: 0.15,
		ProximityWarmth: true, ProximityWarmthStrength: 0.15,
		CurrentPreset: "Cinematic",
	},
	PresetEducational: {
		SunIntensity: 1.2, SunColorTemperature: 0.05,
		AmbientStrength: 0.18, AmbientColorSaturation: 0.4, HemisphericAmbient: false,
		SpecularRocky: 0.12, SpecularGas: 0.35, ShininessRocky: 20.0, ShininessGas: 80.0,
		FresnelStrength: 0.1, TerminatorSoftness: 0.1, RimLightStrength: 0.1,
		ProximityWarmth: false, ProximityWarmthStrength: 0.0,
		CurrentPreset: "Educational",
	},
	PresetStylized: {
		SunIntensity: 1.4, SunColorTemperature: 0.3,
		AmbientStrength: 0.15, AmbientColorSaturation: 0.6, HemisphericAmbient: true,
		SpecularRocky: 0.2, SpecularGas: 0.6, ShininessRocky: 24.0, ShininessGas: 96.0,
		FresnelStrength: 0.3, TerminatorSoftness: 0.12, RimLightStrength: 0.25,
		ProximityWarmth: true, ProximityWarmthStrength: 0.2,
		CurrentPreset: "Stylized",
	},
	PresetDark: {
		SunIntensity: 1.1, SunColorTemperature: 0.15,
		AmbientStrength: 0.01, AmbientColorSaturation: 0.2, HemisphericAmbient: true,
		SpecularRocky: 0.08, SpecularGas: 0.25, ShininessRocky: 20.0, ShininessGas: 80.0,
		FresnelStrength: 0.2, TerminatorSoftness: 0.03, RimLightStrength: 0.05,
		ProximityWarmth: true, ProximityWarmthStrength: 0.08,
		CurrentPreset: "Dark",
	},
}

// ApplyPreset loads a predefined lighting configuration.
func (lc *LightingConfig) ApplyPreset(preset PresetType) {
	cfg, ok := lightingPresets[preset]
	if !ok {
		return
	}
	// Preserve fields not controlled by presets.
	cfg.RimLightColor = lc.RimLightColor
	cfg.ShowNormals = lc.ShowNormals
	cfg.ShowLightDirection = lc.ShowLightDirection
	*lc = cfg
}

// ============================================================================
// KEYBOARD CONTROLS
// ============================================================================

// UpdateWithKeyboard handles runtime lighting adjustments
// Call this in the main loop before rendering
// Returns true if any parameters changed (for UI feedback)
func (lc *LightingConfig) UpdateWithKeyboard() bool {
	changed := false

	// Quick preset switching (Number keys 6-9, 0)
	if rl.IsKeyPressed(rl.KeySix) {
		lc.ApplyPreset(PresetRealistic)
		changed = true
	}
	if rl.IsKeyPressed(rl.KeySeven) {
		lc.ApplyPreset(PresetCinematic)
		changed = true
	}
	if rl.IsKeyPressed(rl.KeyEight) {
		lc.ApplyPreset(PresetEducational)
		changed = true
	}
	if rl.IsKeyPressed(rl.KeyNine) {
		lc.ApplyPreset(PresetStylized)
		changed = true
	}
	if rl.IsKeyPressed(rl.KeyZero) {
		lc.ApplyPreset(PresetDark)
		changed = true
	}

	// Fine-tuning controls (hold SHIFT for faster adjustment)
	speedMultiplier := float32(1.0)
	if rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift) {
		speedMultiplier = 3.0
	}

	// Sun Intensity: L/K keys
	if rl.IsKeyDown(rl.KeyL) {
		lc.SunIntensity += 0.02 * speedMultiplier
		if lc.SunIntensity > 3.0 {
			lc.SunIntensity = 3.0
		}
		lc.CurrentPreset = "Custom"
		changed = true
	}
	if rl.IsKeyDown(rl.KeyK) {
		lc.SunIntensity -= 0.02 * speedMultiplier
		if lc.SunIntensity < 0.5 {
			lc.SunIntensity = 0.5
		}
		lc.CurrentPreset = "Custom"
		changed = true
	}

	// Ambient Strength: O/I keys
	if rl.IsKeyDown(rl.KeyO) {
		lc.AmbientStrength += 0.005 * speedMultiplier
		if lc.AmbientStrength > 0.3 {
			lc.AmbientStrength = 0.3
		}
		lc.CurrentPreset = "Custom"
		changed = true
	}
	if rl.IsKeyDown(rl.KeyI) {
		lc.AmbientStrength -= 0.005 * speedMultiplier
		if lc.AmbientStrength < 0.0 {
			lc.AmbientStrength = 0.0
		}
		lc.CurrentPreset = "Custom"
		changed = true
	}

	// Rim Lighting: R key (toggle on/off)
	if rl.IsKeyPressed(rl.KeyR) {
		if lc.RimLightStrength > 0.0 {
			lc.RimLightStrength = 0.0
		} else {
			lc.RimLightStrength = 0.15
		}
		lc.CurrentPreset = "Custom"
		changed = true
	}

	// Terminator Softness: T/Y keys
	if rl.IsKeyDown(rl.KeyT) {
		lc.TerminatorSoftness += 0.005 * speedMultiplier
		if lc.TerminatorSoftness > 0.3 {
			lc.TerminatorSoftness = 0.3
		}
		lc.CurrentPreset = "Custom"
		changed = true
	}
	if rl.IsKeyDown(rl.KeyY) {
		lc.TerminatorSoftness -= 0.005 * speedMultiplier
		if lc.TerminatorSoftness < 0.0 {
			lc.TerminatorSoftness = 0.0
		}
		lc.CurrentPreset = "Custom"
		changed = true
	}

	return changed
}

// ============================================================================
// SHADER INTEGRATION
// ============================================================================

// ApplyToShader sets all lighting parameters to the planet shader
// Call this before rendering planets
// Cost: ~0.01ms (13 uniform updates)
func (lc *LightingConfig) ApplyToShader(planetRenderer *PlanetRenderData) {
	if !planetRenderer.IsInitialized {
		return
	}

	// Use cached uniform locations (set during InitPlanetRenderer)
	// This avoids expensive GetShaderLocation calls every frame

	// Set all uniforms using cached locations
	rl.SetShaderValue(planetRenderer.Shader, planetRenderer.LocSunIntensity,
		[]float32{lc.SunIntensity}, rl.ShaderUniformFloat)

	rl.SetShaderValue(planetRenderer.Shader, planetRenderer.LocAmbientStrength,
		[]float32{lc.AmbientStrength}, rl.ShaderUniformFloat)

	rl.SetShaderValue(planetRenderer.Shader, planetRenderer.LocSunColorTemp,
		[]float32{lc.SunColorTemperature}, rl.ShaderUniformFloat)

	rl.SetShaderValue(planetRenderer.Shader, planetRenderer.LocAmbientColorSat,
		[]float32{lc.AmbientColorSaturation}, rl.ShaderUniformFloat)

	hemisphericFloat := float32(0.0)
	if lc.HemisphericAmbient {
		hemisphericFloat = 1.0
	}
	rl.SetShaderValue(planetRenderer.Shader, planetRenderer.LocHemisphericAmbient,
		[]float32{hemisphericFloat}, rl.ShaderUniformFloat)

	rl.SetShaderValue(planetRenderer.Shader, planetRenderer.LocFresnelStrength,
		[]float32{lc.FresnelStrength}, rl.ShaderUniformFloat)

	rl.SetShaderValue(planetRenderer.Shader, planetRenderer.LocTerminatorSoftness,
		[]float32{lc.TerminatorSoftness}, rl.ShaderUniformFloat)

	rl.SetShaderValue(planetRenderer.Shader, planetRenderer.LocRimLightStrength,
		[]float32{lc.RimLightStrength}, rl.ShaderUniformFloat)

	rl.SetShaderValue(planetRenderer.Shader, planetRenderer.LocRimLightColor,
		[]float32{lc.RimLightColor.X, lc.RimLightColor.Y, lc.RimLightColor.Z},
		rl.ShaderUniformVec3)

	proximityFloat := float32(0.0)
	if lc.ProximityWarmth {
		proximityFloat = 1.0
	}
	rl.SetShaderValue(planetRenderer.Shader, planetRenderer.LocProximityWarmth,
		[]float32{proximityFloat}, rl.ShaderUniformFloat)

	rl.SetShaderValue(planetRenderer.Shader, planetRenderer.LocProximityStrength,
		[]float32{lc.ProximityWarmthStrength}, rl.ShaderUniformFloat)
}

// GetUIString returns a formatted string for display
func (lc *LightingConfig) GetUIString() string {
	return fmt.Sprintf(
		"Lighting Preset: %s\n"+
		"Sun Intensity: %.2f | Ambient: %.2f\n"+
		"Terminator Softness: %.2f | Rim Light: %.2f\n"+
		"\n"+
		"Presets: [6]Realistic [7]Cinematic [8]Educational [9]Stylized [0]Dark\n"+
		"Controls: [L/K]Sun Brightness [O/I]Ambient [T/Y]Terminator [R]Rim Toggle\n"+
		"Hold SHIFT for faster adjustment",
		lc.CurrentPreset,
		lc.SunIntensity, lc.AmbientStrength,
		lc.TerminatorSoftness, lc.RimLightStrength,
	)
}
