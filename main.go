package main

import (
	"fmt"
	"math"
	"math/rand"
	"strings"

	"github.com/ctclostio/GoStarMap/internal/celestial"
	"github.com/ctclostio/GoStarMap/orbital"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// spectralColors maps each Morgan-Keenan spectral class to a display color.
// Defined here (not in the celestial package) because rl.Color is a raylib
// type and we want celestial to stay raylib-free for testability.
var spectralColors = map[celestial.SpectralType]rl.Color{
	celestial.TypeO: rl.NewColor(155, 176, 255, 255), // Blue
	celestial.TypeB: rl.NewColor(170, 191, 255, 255), // Blue-white
	celestial.TypeA: rl.NewColor(202, 215, 255, 255), // White
	celestial.TypeF: rl.NewColor(248, 247, 255, 255), // Yellow-white
	celestial.TypeG: rl.NewColor(255, 244, 234, 255), // Yellow
	celestial.TypeK: rl.NewColor(255, 210, 161, 255), // Orange
	celestial.TypeM: rl.NewColor(255, 204, 111, 255), // Red
}

// GetSpectralColor returns the color for a spectral type.
func GetSpectralColor(stype celestial.SpectralType) rl.Color {
	return spectralColors[stype]
}

// Star represents a star with astronomical properties
type Star struct {
	X, Y, Z      float64                // Position in render units
	Magnitude    float32                // Visual magnitude (brightness)
	SpectralType celestial.SpectralType // Spectral classification
	Name         string                 // Name (if notable)
	IsNamed      bool                   // Whether this star has a name
	Color        rl.Color               // Precomputed color based on spectral type
}

// Planet represents a planet with scientific data and orbital mechanics
type Planet struct {
	Name            string
	X, Y, Z         float64
	Radius          float32
	Color           rl.Color
	MassKg          float64         // Mass in kilograms
	DiameterKm      float64         // Diameter in kilometers
	Description     string          // Brief description
	OrbitalElements orbital.OrbitalElements // Keplerian orbital parameters

	// Rotation parameters (NASA Planetary Fact Sheet validated)
	RotationPeriodDays float64 // Sidereal rotation period (negative if retrograde)
	AxialTiltDegrees   float64 // Obliquity (axial tilt from orbital plane)
	CurrentRotation    float64 // Current rotation angle in radians (updated each frame)
}

// Galaxy contains all celestial objects and simulation time
type Galaxy struct {
	Stars        []Star
	Planets      []Planet
	TimeScale    orbital.TimeScaleInfo // Simulation time control
}

// NewGalaxy creates a new galaxy with initial time set to J2000.0 epoch
func NewGalaxy() *Galaxy {
	return &Galaxy{
		Stars:   make([]Star, 0),
		Planets: make([]Planet, 0),
		TimeScale: orbital.TimeScaleInfo{
			SimulationDays: 0.0,    // Start at J2000.0 epoch (January 1, 2000, 12:00 TT)
			TimeScale:      365.256 / 60.0, // Default: Earth completes orbit in 60 seconds
			IsPaused:       false,
		},
	}
}

// AddNamedStar adds a notable star with a name
func (g *Galaxy) AddNamedStar(name string, x, y, z float64, mag float32, stype celestial.SpectralType) {
	g.Stars = append(g.Stars, Star{
		X:            x,
		Y:            y,
		Z:            z,
		Magnitude:    mag,
		SpectralType: stype,
		Name:         name,
		IsNamed:      true,
		Color:        GetSpectralColor(stype), // Precompute color
	})
}

// GenerateGalaxy creates a procedural galaxy with realistic structure
func GenerateGalaxy() *Galaxy {
	g := NewGalaxy()

	fmt.Println("Generating galaxy...")

	// 1. Add our solar system at origin
	g.AddNamedStar("Sun", 0, 0, 0, -26.74, celestial.TypeG)

	// Add planets with validated scientific data and orbital mechanics
	// Data validated against JPL Horizons System and NASA Planetary Fact Sheets (2025-10-31)
	// Radii now NASA-accurate! Jupiter is 11x Earth, Saturn is 9x Earth
	planets := []struct {
		name              string
		radius            float32 // Render radius (NASA-validated relative sizes)
		color             rl.Color
		massKg            float64
		diameterKm        float64
		description       string
		rotationDays      float64 // Sidereal rotation period (negative = retrograde)
		axialTiltDegrees  float64 // Obliquity (axial tilt)
	}{
		{"Mercury", 0.115, rl.Gray, 3.3011e23, 4879, "Smallest planet, closest to Sun", 58.646, 0.034},
		{"Venus", 0.285, rl.Beige, 4.8675e24, 12104, "Hottest planet, thick atmosphere", -243.018, 177.36},  // RETROGRADE
		{"Earth", 0.300, rl.Blue, 5.9724e24, 12742, "Our home, only known life", 0.99727, 23.4393},
		{"Mars", 0.160, rl.Red, 6.4171e23, 6779, "The Red Planet, has polar ice", 1.02596, 25.19},
		{"Jupiter", 3.291, rl.Orange, 1.8982e27, 139820, "Largest planet, gas giant", 0.41354, 3.13},      // Now 11x Earth!
		{"Saturn", 2.742, rl.Gold, 5.6834e26, 116460, "Famous rings, gas giant", 0.44401, 26.73},          // Now 9x Earth!
		{"Uranus", 1.194, rl.SkyBlue, 8.6810e25, 50724, "Ice giant, tilted axis", -0.71833, 97.77},       // RETROGRADE, sideways!
		{"Neptune", 1.158, rl.DarkBlue, 1.02413e26, 49244, "Windiest planet, ice giant", 0.67125, 28.32},
	}

	// Initialize planets with validated orbital elements
	for _, p := range planets {
		// Get validated orbital elements from JPL data
		elements := orbital.GetPlanetaryElements(p.name)

		// Calculate initial position at J2000.0 epoch (time = 0)
		positionAU := orbital.OrbitalElementsToPosition(elements, 0.0)

		// Convert to render units (AU to your scale)
		renderPos := orbital.PositionToRenderUnits(positionAU, AUScale)

		g.Planets = append(g.Planets, Planet{
			Name:               p.name,
			X:                  float64(renderPos[0]),
			Y:                  float64(renderPos[1]),
			Z:                  float64(renderPos[2]),
			Radius:             p.radius,
			Color:              p.color,
			MassKg:             p.massKg,
			DiameterKm:         p.diameterKm,
			Description:        p.description,
			OrbitalElements:    elements,
			RotationPeriodDays: p.rotationDays,
			AxialTiltDegrees:   p.axialTiltDegrees,
			CurrentRotation:    0.0, // Start at 0, will update each frame
		})
	}

	// 2. Add nearby named stars
	const lyScale = 50.0
	nearbyStars := []struct {
		name    string
		x, y, z float64 // in light-years
		mag     float32
		stype   celestial.SpectralType
	}{
		{"Proxima Centauri", 4.24, 0.5, -1.2, 11.13, celestial.TypeM},
		{"Alpha Centauri A", 4.37, 0.6, -1.1, -0.01, celestial.TypeG},
		{"Barnard's Star", 5.96, 1.2, 0.8, 9.53, celestial.TypeM},
		{"Sirius", 8.6, -2.5, 3.1, -1.46, celestial.TypeA},
		{"Procyon", 11.4, 3.0, 2.0, 0.38, celestial.TypeF},
		{"61 Cygni", 11.4, 5.0, -3.0, 5.2, celestial.TypeK},
		{"Epsilon Eridani", 10.5, -2.0, 4.0, 3.73, celestial.TypeK},
		{"Tau Ceti", 11.9, 1.0, -5.0, 3.49, celestial.TypeG},
	}

	for _, s := range nearbyStars {
		g.AddNamedStar(s.name, s.x*lyScale, s.y*lyScale, s.z*lyScale, s.mag, s.stype)
	}

	fmt.Printf("Added %d named stars\n", len(g.Stars))

	// 3. Generate procedural stars with galactic disk structure
	const targetStars = 100000 // With optimizations, we can handle 100k!
	fmt.Printf("Generating %d procedural stars...\n", targetStars)

	// Galactic parameters
	const (
		diskRadius      = 50000.0 // Disk radius in render units (~1000 light-years scaled)
		diskThickness   = 2000.0  // Disk thickness
		bulgeRadius     = 10000.0 // Central bulge radius
		bulgeThickness  = 5000.0  // Bulge thickness
		spiralArms      = 2       // Number of spiral arms
		spiralTightness = 0.3     // How tight the spiral is
	)

	for i := 0; i < targetStars; i++ {
		var x, y, z float64
		var mag float32

		// Decide which galactic component this star belongs to
		r := rand.Float64()

		if r < 0.15 { // 15% in central bulge
			// Spherical distribution for bulge
			radius := bulgeRadius * math.Pow(rand.Float64(), 0.333)
			theta := rand.Float64() * 2 * math.Pi
			phi := math.Acos(2*rand.Float64() - 1)

			x = radius * math.Sin(phi) * math.Cos(theta)
			y = radius * math.Sin(phi) * math.Sin(theta) * 0.5 // Flatten slightly
			z = radius * math.Cos(phi) * 0.5

			mag = rand.Float32()*8 - 2 // Brighter stars in bulge
		} else { // 85% in galactic disk with spiral structure
			// Distance from galactic center
			radius := math.Sqrt(rand.Float64()) * diskRadius

			// Add spiral arm structure
			armIndex := rand.Intn(spiralArms)
			armAngle := (float64(armIndex) / float64(spiralArms)) * 2 * math.Pi
			spiralAngle := armAngle + spiralTightness*radius/diskRadius*2*math.Pi

			// Add some randomness to spiral arms
			angleOffset := (rand.Float64() - 0.5) * 0.5
			angle := spiralAngle + angleOffset

			x = radius * math.Cos(angle)
			z = radius * math.Sin(angle)

			// Height follows gaussian distribution (thinner disk)
			y = (rand.Float64() + rand.Float64() - 1) * diskThickness

			// Magnitude based on distance (farther = dimmer)
			mag = rand.Float32()*10 + 4
		}

		// Get spectral type
		stype := celestial.RandomType()

		g.Stars = append(g.Stars, Star{
			X:            x,
			Y:            y,
			Z:            z,
			Magnitude:    mag,
			SpectralType: stype,
			Color:        GetSpectralColor(stype), // Precompute color
			IsNamed:      false,
		})
	}

	fmt.Printf("Galaxy generation complete: %d total stars\n", len(g.Stars))
	return g
}

// GetCameraSpeed returns camera speed - normal by default, boosted when shift is held
func GetCameraSpeed(camera rl.Camera3D, shiftHeld bool) float32 {
	baseSpeed := float32(CameraBaseSpeed)

	// If shift is not held, just return normal speed
	if !shiftHeld {
		return baseSpeed
	}

	// Shift is held - apply distance-based speed boost
	dist := float32(math.Sqrt(float64(camera.Position.X*camera.Position.X +
		camera.Position.Y*camera.Position.Y +
		camera.Position.Z*camera.Position.Z)))

	if dist < CameraDistThreshold {
		return baseSpeed * float32(CameraSpeedBoost)  // 3x boost for close range
	}
	return baseSpeed + dist*float32(CameraDistBoost)  // Distance-based boost for far range
}

// RenderStarLOD renders a star with level-of-detail based on camera distance
func RenderStarLOD(star Star, cameraPos rl.Vector3) {
	starPos := rl.Vector3{X: float32(star.X), Y: float32(star.Y), Z: float32(star.Z)}

	// Calculate distance to camera
	dx := starPos.X - cameraPos.X
	dy := starPos.Y - cameraPos.Y
	dz := starPos.Z - cameraPos.Z
	dist := float32(math.Sqrt(float64(dx*dx + dy*dy + dz*dz)))

	// Use precomputed color from Star struct (set during generation)
	color := star.Color

	// Brightness factor based on magnitude (lower magnitude = brighter)
	brightnessFactor := float32(math.Pow(2.512, float64(-star.Magnitude/5.0)))
	if brightnessFactor > 1.0 {
		brightnessFactor = 1.0
	}

	// Adjust color brightness
	color.R = uint8(float32(color.R) * (0.3 + 0.7*brightnessFactor))
	color.G = uint8(float32(color.G) * (0.3 + 0.7*brightnessFactor))
	color.B = uint8(float32(color.B) * (0.3 + 0.7*brightnessFactor))

	// LOD levels based on distance - much more aggressive
	switch {
	case dist < StarLODClose: // Very close - full sphere
		radius := 2.0 + brightnessFactor*1.0
		rl.DrawSphere(starPos, radius, color)

	case dist < StarLODMedium: // Medium distance - medium sphere
		radius := 1.0 + brightnessFactor*0.4
		rl.DrawSphere(starPos, radius, color)

	case dist < StarLODFar: // Far - small sphere
		rl.DrawSphere(starPos, 0.6, color)

	default: // Everything else - single point (most efficient)
		rl.DrawPoint3D(starPos, color)
	}
}

// targetCosThreshold is cos(2°) — vectors whose dot product exceeds this are
// within 2° of the camera forward, so we can avoid acos in the hot path.
var targetCosThreshold = float32(math.Cos(2.0 * math.Pi / 180.0))

// CheckTargeting determines if the camera is looking at a celestial body.
// Returns the targeted planet (or nil) and whether the Sun is targeted.
// Compares cos(angle) directly instead of taking acos — angle < threshold
// is equivalent to dot >= cos(threshold) for unit vectors, with no NaN risk.
func CheckTargeting(camera rl.Camera3D, planets []Planet, sunPos rl.Vector3) (*Planet, bool) {
	forward := rl.Vector3Subtract(camera.Target, camera.Position)
	forward = rl.Vector3Normalize(forward)

	// Sun first
	sunDir := rl.Vector3Subtract(sunPos, camera.Position)
	if rl.Vector3Length(sunDir) > 0.01 {
		sunDir = rl.Vector3Normalize(sunDir)
		dot := forward.X*sunDir.X + forward.Y*sunDir.Y + forward.Z*sunDir.Z
		if dot >= targetCosThreshold {
			return nil, true
		}
	}

	for i := range planets {
		planetPos := rl.Vector3{X: float32(planets[i].X), Y: float32(planets[i].Y), Z: float32(planets[i].Z)}
		planetDir := rl.Vector3Subtract(planetPos, camera.Position)
		if rl.Vector3Length(planetDir) > 0.01 {
			planetDir = rl.Vector3Normalize(planetDir)
			dot := forward.X*planetDir.X + forward.Y*planetDir.Y + forward.Z*planetDir.Z
			if dot >= targetCosThreshold {
				return &planets[i], false
			}
		}
	}

	return nil, false
}

// DrawCelestialInfo displays information about a targeted celestial body
func DrawCelestialInfo(screenWidth, screenHeight int32, planet *Planet, isSun bool) {
	// Position the info box in the lower right quadrant
	boxX := screenWidth - int32(UIInfoBoxWidth) - 20 // 20px margin
	boxY := screenHeight - int32(UIInfoBoxHeight) - 20

	// Draw semi-transparent background
	rl.DrawRectangle(boxX, boxY, UIInfoBoxWidth, UIInfoBoxHeight, rl.NewColor(0, 0, 0, 200))
	rl.DrawRectangleLines(boxX, boxY, UIInfoBoxWidth, UIInfoBoxHeight, rl.NewColor(100, 150, 255, 255))

	yOffset := boxY + 15

	if isSun {
		// Display Sun information
		rl.DrawText("THE SUN", boxX+15, yOffset, 24, rl.Gold)
		yOffset += 35

		rl.DrawText("Our Star", boxX+15, yOffset, 16, rl.Gray)
		yOffset += 25

		rl.DrawText(fmt.Sprintf("Mass: 1.989 x 10^30 kg"), boxX+15, yOffset, 14, rl.RayWhite)
		yOffset += 20

		rl.DrawText(fmt.Sprintf("Diameter: 1,392,700 km"), boxX+15, yOffset, 14, rl.RayWhite)
		yOffset += 20

		rl.DrawText(fmt.Sprintf("Type: G-type main-sequence star"), boxX+15, yOffset, 14, rl.RayWhite)
		yOffset += 20

		rl.DrawText(fmt.Sprintf("Age: ~4.6 billion years"), boxX+15, yOffset, 14, rl.RayWhite)
		yOffset += 20

		rl.DrawText(fmt.Sprintf("Temperature: 5,778 K (surface)"), boxX+15, yOffset, 14, rl.RayWhite)
		yOffset += 25

		rl.DrawText("Center of our solar system", boxX+15, yOffset, 12, rl.DarkGray)

	} else if planet != nil {
		// Display planet information
		rl.DrawText(strings.ToUpper(planet.Name), boxX+15, yOffset, 24, planet.Color)
		yOffset += 35

		rl.DrawText(planet.Description, boxX+15, yOffset, 14, rl.Gray)
		yOffset += 25

		// Format mass in scientific notation
		massExp := math.Log10(planet.MassKg)
		massMantissa := planet.MassKg / math.Pow(10, math.Floor(massExp))
		rl.DrawText(fmt.Sprintf("Mass: %.2f x 10^%d kg", massMantissa, int(massExp)), boxX+15, yOffset, 14, rl.RayWhite)
		yOffset += 20

		rl.DrawText(fmt.Sprintf("Diameter: %s km", celestial.FormatNumber(planet.DiameterKm)), boxX+15, yOffset, 14, rl.RayWhite)
		yOffset += 20

		rl.DrawText(fmt.Sprintf("Semi-major Axis: %.3f AU", planet.OrbitalElements.SemiMajorAxis), boxX+15, yOffset, 14, rl.RayWhite)
		yOffset += 20

		rl.DrawText(fmt.Sprintf("Orbital Period: %s days", celestial.FormatNumber(planet.OrbitalElements.OrbitalPeriod)), boxX+15, yOffset, 14, rl.RayWhite)
		yOffset += 20

		earthYears := planet.OrbitalElements.OrbitalPeriod / 365.256
		rl.DrawText(fmt.Sprintf("  (%.2f Earth years)", earthYears), boxX+15, yOffset, 12, rl.DarkGray)
		yOffset += 20

		rl.DrawText(fmt.Sprintf("Eccentricity: %.4f", planet.OrbitalElements.Eccentricity), boxX+15, yOffset, 14, rl.RayWhite)
		yOffset += 20

		inclinationDeg := planet.OrbitalElements.Inclination * (180.0 / math.Pi)
		rl.DrawText(fmt.Sprintf("Inclination: %.2f degrees", inclinationDeg), boxX+15, yOffset, 14, rl.RayWhite)
	}
}

// UpdatePlanetPositions recalculates planet positions based on current simulation time
// Expected cost: ~0.02ms for 8 planets on modern CPU
func UpdatePlanetPositions(galaxy *Galaxy) {
	currentTime := galaxy.TimeScale.SimulationDays

	for i := range galaxy.Planets {
		planet := &galaxy.Planets[i]

		// Calculate position in AU based on orbital elements
		positionAU := orbital.OrbitalElementsToPosition(planet.OrbitalElements, currentTime)

		// Convert to render coordinates
		renderPos := orbital.PositionToRenderUnits(positionAU, AUScale)

		// Update planet position
		planet.X = float64(renderPos[0])
		planet.Y = float64(renderPos[1])
		planet.Z = float64(renderPos[2])
	}
}

// UpdatePlanetRotations updates rotation angles for all planets based on simulation time
// Expected cost: ~0.005ms for 8 planets (trivial arithmetic)
// Memory: No allocations, pure computation
func UpdatePlanetRotations(galaxy *Galaxy) {
	currentTime := galaxy.TimeScale.SimulationDays

	for i := range galaxy.Planets {
		planet := &galaxy.Planets[i]

		// Skip if no rotation period defined (use threshold to handle near-zero values)
		if math.Abs(planet.RotationPeriodDays) < 1e-6 {
			planet.CurrentRotation = 0.0
			continue
		}

		// Calculate rotation angle in radians
		// rotationAngle = 2π * (time / period)
		// Negative period (retrograde) automatically gives opposite rotation direction

		// IMPORTANT: Normalize rotations FIRST, then multiply by 2π
		// This prevents floating-point precision loss on long simulations
		rotationsCompleted := currentTime / planet.RotationPeriodDays
		normalizedRotations := math.Mod(rotationsCompleted, 1.0)
		if normalizedRotations < 0 {
			normalizedRotations += 1.0
		}
		planet.CurrentRotation = normalizedRotations * 2.0 * math.Pi
	}
}

// ============================================================================
// SUN RENDERING SYSTEM
// ============================================================================

// SunRenderData contains the shader and parameters for the Sun's HDR-emissive
// rendering. The bloom post-process pipeline that previously lived here was
// never actually used at the call site; it has been removed. The GLSL files
// in shaders/ remain in case bloom is reintroduced later.
type SunRenderData struct {
	Shader        rl.Shader
	SunRadius     float32
	IsInitialized bool
}

// PlanetRenderData contains resources for planet lighting system
type PlanetRenderData struct {
	Shader        rl.Shader
	IsInitialized bool

	// Cached uniform locations (for performance)
	LocSunPosition      int32
	LocPlanetColor      int32
	LocPlanetRadius     int32
	LocSunIntensity     int32
	LocAmbientStrength  int32
	LocSpecularStrength int32
	LocShininess        int32
	LocCameraPosition   int32

	// Rotation uniform locations
	LocRotationAngle    int32
	LocAxialTilt        int32

	// Lighting config uniform locations (cached for performance)
	LocSunColorTemp       int32
	LocAmbientColorSat    int32
	LocHemisphericAmbient int32
	LocFresnelStrength    int32
	LocTerminatorSoftness int32
	LocRimLightStrength   int32
	LocRimLightColor      int32
	LocProximityWarmth    int32
	LocProximityStrength  int32

	// Cached sphere models (LOD - Level of Detail)
	// These are created once and reused every frame for performance
	SphereModelHigh   rl.Model // 32 rings/slices - close up
	SphereModelMedium rl.Model // 24 rings/slices - medium distance
	SphereModelLow    rl.Model // 16 rings/slices - far away
}

// Physical Sun parameters
const (
	SunRadiusUnits   = 32.7 // Physically accurate: 109x Earth radius (0.3 units)
	SunDistanceScale = 1.0  // Sun at origin

	// AU Scale: Distance scaling for orbital distances
	// True scale would be 7,030 units/AU (215 Sun radii per AU)
	// Using compressed scale for better navigation and visibility
	AUScale = 150.0 // Units per AU - keeps planets visible and navigable

	// Planet rendering parameters
	PlanetSizeScale      = 4.0   // Scale factor for visibility (makes planets 4x larger)
	PlanetAmbientLight   = 0.12  // Ambient lighting strength (12% to keep dark sides visible)
	PlanetSunIntensity   = 1.0   // Sun brightness multiplier
	PlanetSpecularRocky  = 0.1   // Specular strength for rocky planets (subtle)
	PlanetSpecularGas    = 0.3   // Specular strength for gas giants (more reflective)
	PlanetShininessRocky = 16.0  // Specular shininess for rocky planets (wider highlight)
	PlanetShininessGas   = 64.0  // Specular shininess for gas giants (tighter highlight)
)

// Rendering constants
const (
	// Star rendering
	MaxStarsPerFrame   = 15000  // Maximum stars rendered per frame
	MaxRenderDistance   = 20000.0 // Maximum distance for star rendering (render units)
	
	// Star LOD distances (render units)
	StarLODClose       = 100.0  // Full sphere
	StarLODMedium      = 300.0  // Medium sphere
	StarLODFar         = 800.0  // Small sphere
	// Beyond StarLODFar: point rendering
	
	// Planet LOD screen sizes (fraction of screen)
	PlanetLODClose     = 0.05  // High detail (32 rings/slices)
	PlanetLODMedium    = 0.02  // Medium detail (24 rings/slices)
	// Below PlanetLODMedium: low detail (16 rings/slices)
	
	// Camera speed
	CameraBaseSpeed    = 0.5   // Base camera movement speed
	CameraSpeedBoost   = 3.0   // Speed multiplier when shift is held (close range)
	CameraDistBoost    = 0.05  // Distance-based speed boost factor
	CameraDistThreshold = 50.0  // Distance threshold for speed boost
)

// UI Constants
const (
	UIInfoBoxWidth     = 400
	UIInfoBoxHeight    = 320
	UIReticleSize      = 10
	UIReticleThickness = 2
	UIReticleGap       = 4
	UITimeBoxWidth     = 480
	UITimeBoxHeight    = 110
)

// InitSunRenderer loads the Sun shader and prepares the render data.
// The screenWidth/screenHeight parameters are kept on the signature for
// compatibility but are no longer used now that the bloom render textures
// have been removed.
func InitSunRenderer(screenWidth, screenHeight int32) (*SunRenderData, error) {
	_ = screenWidth
	_ = screenHeight
	sun := &SunRenderData{SunRadius: SunRadiusUnits}

	sunShader := rl.LoadShader("shaders/sun.vs", "shaders/sun.fs")
	if sunShader.ID == 0 {
		return nil, fmt.Errorf("failed to load Sun shader")
	}
	sun.Shader = sunShader

	sun.IsInitialized = true
	fmt.Println("Sun renderer initialized successfully")
	return sun, nil
}

// Unload releases the Sun shader.
func (sun *SunRenderData) Unload() {
	if !sun.IsInitialized {
		return
	}
	rl.UnloadShader(sun.Shader)
	sun.IsInitialized = false
}

// RenderSun renders the photorealistic Sun with custom shaders
// Expected cost: 0.3-0.5ms @ 1080p on RTX 3070 / GTX 1660
// Memory bandwidth: ~15MB/frame (render texture reads/writes)
func RenderSun(sun *SunRenderData, camera rl.Camera3D, time float32) {
	if !sun.IsInitialized {
		return
	}

	sunPos := rl.Vector3{X: 0, Y: 0, Z: 0}

	// Calculate distance from camera to Sun for LOD
	camDist := rl.Vector3Distance(camera.Position, sunPos)

	// LOD: Reduce sphere quality at far distances
	var sphereRings, sphereSlices int32
	if camDist < 100 {
		sphereRings = 32
		sphereSlices = 32
	} else if camDist < 500 {
		sphereRings = 24
		sphereSlices = 24
	} else {
		sphereRings = 16
		sphereSlices = 16
	}

	// Set shader uniforms
	cameraLoc := rl.GetShaderLocation(sun.Shader, "cameraPosition")
	rl.SetShaderValue(sun.Shader, cameraLoc, []float32{camera.Position.X, camera.Position.Y, camera.Position.Z}, rl.ShaderUniformVec3)

	timeLoc := rl.GetShaderLocation(sun.Shader, "time")
	rl.SetShaderValue(sun.Shader, timeLoc, []float32{time}, rl.ShaderUniformFloat)

	radiusLoc := rl.GetShaderLocation(sun.Shader, "sunRadius")
	rl.SetShaderValue(sun.Shader, radiusLoc, []float32{sun.SunRadius}, rl.ShaderUniformFloat)

	// Render Sun with custom shader
	rl.BeginShaderMode(sun.Shader)
	rl.DrawSphereEx(sunPos, sun.SunRadius, sphereRings, sphereSlices, rl.White)
	rl.EndShaderMode()
}


// ============================================================================
// PLANET RENDERING SYSTEM
// ============================================================================

// InitPlanetRenderer initializes the planet lighting shader system
// Expected cost: One-time initialization, ~5ms
// Memory: ~100KB for shader program
func InitPlanetRenderer() (*PlanetRenderData, error) {
	planet := &PlanetRenderData{}

	// Load planet shader - try enhanced version first, fall back to standard
	planetShader := rl.LoadShader("shaders/planet.vs", "shaders/planet_enhanced.fs")
	if planetShader.ID == 0 {
		fmt.Println("Warning: Enhanced shader not found, trying standard shader...")
		planetShader = rl.LoadShader("shaders/planet.vs", "shaders/planet.fs")
		if planetShader.ID == 0 {
			return nil, fmt.Errorf("failed to load planet shader")
		}
	}
	planet.Shader = planetShader

	// Cache uniform locations for performance
	// Looking up locations every frame is expensive, so we cache them
	planet.LocSunPosition = rl.GetShaderLocation(planetShader, "sunPosition")
	planet.LocPlanetColor = rl.GetShaderLocation(planetShader, "planetColor")
	planet.LocPlanetRadius = rl.GetShaderLocation(planetShader, "planetRadius")
	planet.LocSunIntensity = rl.GetShaderLocation(planetShader, "sunIntensity")
	planet.LocAmbientStrength = rl.GetShaderLocation(planetShader, "ambientStrength")
	planet.LocSpecularStrength = rl.GetShaderLocation(planetShader, "specularStrength")
	planet.LocShininess = rl.GetShaderLocation(planetShader, "shininess")
	planet.LocCameraPosition = rl.GetShaderLocation(planetShader, "cameraPosition")

	// Cache rotation uniform locations
	planet.LocRotationAngle = rl.GetShaderLocation(planetShader, "rotationAngle")
	planet.LocAxialTilt = rl.GetShaderLocation(planetShader, "axialTilt")

	// Cache lighting config uniform locations (for performance)
	planet.LocSunColorTemp = rl.GetShaderLocation(planetShader, "sunColorTemperature")
	planet.LocAmbientColorSat = rl.GetShaderLocation(planetShader, "ambientColorSaturation")
	planet.LocHemisphericAmbient = rl.GetShaderLocation(planetShader, "hemisphericAmbient")
	planet.LocFresnelStrength = rl.GetShaderLocation(planetShader, "fresnelStrength")
	planet.LocTerminatorSoftness = rl.GetShaderLocation(planetShader, "terminatorSoftness")
	planet.LocRimLightStrength = rl.GetShaderLocation(planetShader, "rimLightStrength")
	planet.LocRimLightColor = rl.GetShaderLocation(planetShader, "rimLightColor")
	planet.LocProximityWarmth = rl.GetShaderLocation(planetShader, "proximityWarmth")
	planet.LocProximityStrength = rl.GetShaderLocation(planetShader, "proximityWarmthStrength")

	// Create and cache sphere models for LOD (Level of Detail).
	// BeginShaderMode in RenderPlanet overrides any per-mesh material shader,
	// so we don't bother assigning Materials.Shader here.
	meshHigh := rl.GenMeshSphere(1.0, 32, 32)
	planet.SphereModelHigh = rl.LoadModelFromMesh(meshHigh)

	meshMedium := rl.GenMeshSphere(1.0, 24, 24)
	planet.SphereModelMedium = rl.LoadModelFromMesh(meshMedium)

	meshLow := rl.GenMeshSphere(1.0, 16, 16)
	planet.SphereModelLow = rl.LoadModelFromMesh(meshLow)

	planet.IsInitialized = true
	fmt.Println("Planet renderer initialized successfully")
	return planet, nil
}

// UnloadPlanetRenderer cleans up planet rendering resources
func (planet *PlanetRenderData) Unload() {
	if !planet.IsInitialized {
		return
	}
	// Unload cached sphere models
	rl.UnloadModel(planet.SphereModelHigh)
	rl.UnloadModel(planet.SphereModelMedium)
	rl.UnloadModel(planet.SphereModelLow)

	rl.UnloadShader(planet.Shader)
	planet.IsInitialized = false
}

// PlanetRenderParams contains rendering parameters for a single planet
type PlanetRenderParams struct {
	Position         rl.Vector3
	Radius           float32
	Color            rl.Color
	SpecularStrength float32
	Shininess        float32
	IsGasGiant       bool
	RotationAngle    float32 // Current rotation angle in radians
	AxialTilt        float32 // Axial tilt in radians
}

// RenderPlanet renders a single planet with physically-based lighting.
// Per-frame globals (sunPosition, cameraPosition) are expected to have been
// set once by RenderAllPlanets — this function only writes per-planet uniforms.
// Expected cost: ~0.05ms per planet @ 1080p on RTX 3070 / GTX 1660.
func RenderPlanet(planet *PlanetRenderData, params PlanetRenderParams, camera rl.Camera3D) {
	if !planet.IsInitialized {
		// Fallback: render simple colored sphere
		rl.DrawSphere(params.Position, params.Radius, params.Color)
		return
	}

	// Normalize color to 0-1 range for shader
	colorVec := []float32{
		float32(params.Color.R) / 255.0,
		float32(params.Color.G) / 255.0,
		float32(params.Color.B) / 255.0,
	}

	// Per-planet uniforms only.
	rl.SetShaderValue(planet.Shader, planet.LocPlanetColor, colorVec, rl.ShaderUniformVec3)
	rl.SetShaderValue(planet.Shader, planet.LocPlanetRadius, []float32{params.Radius}, rl.ShaderUniformFloat)
	rl.SetShaderValue(planet.Shader, planet.LocSpecularStrength, []float32{params.SpecularStrength}, rl.ShaderUniformFloat)
	rl.SetShaderValue(planet.Shader, planet.LocShininess, []float32{params.Shininess}, rl.ShaderUniformFloat)
	rl.SetShaderValue(planet.Shader, planet.LocRotationAngle, []float32{params.RotationAngle}, rl.ShaderUniformFloat)
	rl.SetShaderValue(planet.Shader, planet.LocAxialTilt, []float32{params.AxialTilt}, rl.ShaderUniformFloat)

	// Calculate distance to camera for LOD (Level of Detail)
	camDist := rl.Vector3Distance(camera.Position, params.Position)
	screenSize := params.Radius / camDist // Approximate angular size on screen

	// Select appropriate LOD model based on screen size
	// Using cached models for maximum performance (created once at init)
	var selectedModel rl.Model
	if screenSize > PlanetLODClose { // Close up - high detail (32 rings/slices)
		selectedModel = planet.SphereModelHigh
	} else if screenSize > PlanetLODMedium { // Medium distance (24 rings/slices)
		selectedModel = planet.SphereModelMedium
	} else { // Far/small planet - low detail (16 rings/slices)
		selectedModel = planet.SphereModelLow
	}

	// Draw model at planet position with radius as scale
	// Models are unit sphere (radius 1.0), so scale by actual planet radius
	rl.BeginShaderMode(planet.Shader)
	rl.DrawModel(selectedModel, params.Position, params.Radius, rl.White)
	rl.EndShaderMode()
}

// RenderAllPlanets renders all planets in the solar system with proper lighting.
// Sun position and camera position are uniform across all planets, so we write
// them once here instead of repeating per-planet inside RenderPlanet.
// Expected total cost: ~0.64ms for 8 planets @ 1080p (with enhanced lighting).
func RenderAllPlanets(planetRenderer *PlanetRenderData, planets []Planet, sunPos rl.Vector3, camera rl.Camera3D) {
	if planetRenderer.IsInitialized {
		rl.SetShaderValue(planetRenderer.Shader, planetRenderer.LocSunPosition,
			[]float32{sunPos.X, sunPos.Y, sunPos.Z}, rl.ShaderUniformVec3)
		rl.SetShaderValue(planetRenderer.Shader, planetRenderer.LocCameraPosition,
			[]float32{camera.Position.X, camera.Position.Y, camera.Position.Z}, rl.ShaderUniformVec3)
	}

	for _, p := range planets {
		// Determine if planet is a gas giant (affects specularity)
		isGasGiant := p.Name == "Jupiter" || p.Name == "Saturn" || p.Name == "Uranus" || p.Name == "Neptune"

		// Set rendering parameters based on planet type
		// Note: These can be overridden by the lighting config system
		var specular, shininess float32
		if isGasGiant {
			specular = PlanetSpecularGas
			shininess = PlanetShininessGas
		} else {
			specular = PlanetSpecularRocky
			shininess = PlanetShininessRocky
		}

		// Apply size scaling for visibility
		scaledRadius := p.Radius * PlanetSizeScale

		// Convert axial tilt from degrees to radians
		axialTiltRad := float32(p.AxialTiltDegrees * (math.Pi / 180.0))

		params := PlanetRenderParams{
			Position:         rl.Vector3{X: float32(p.X), Y: float32(p.Y), Z: float32(p.Z)},
			Radius:           scaledRadius,
			Color:            p.Color,
			SpecularStrength: specular,
			Shininess:        shininess,
			IsGasGiant:       isGasGiant,
			RotationAngle:    float32(p.CurrentRotation),
			AxialTilt:        axialTiltRad,
		}

		RenderPlanet(planetRenderer, params, camera)
	}
}

func main() {
	// Initialize window
	screenWidth := int32(1920)
	screenHeight := int32(1080)

	rl.InitWindow(screenWidth, screenHeight, "GoStarMap - 3D Galactic Navigation [100k Stars - Optimized]")
	defer rl.CloseWindow()
	rl.SetTargetFPS(60)

	// Initialize camera - positioned to view Sun and inner planets
	// Distance ~378 units from origin allows viewing Sun (radius 32.7) and inner planets
	camera := rl.Camera3D{
		Position:   rl.Vector3{X: 250, Y: 150, Z: 250},
		Target:     rl.Vector3{X: 0, Y: 0, Z: 0},
		Up:         rl.Vector3{X: 0, Y: 1, Z: 0},
		Fovy:       60.0,
		Projection: rl.CameraPerspective,
	}

	// Generate galaxy (this will take a moment)
	galaxy := GenerateGalaxy()

	// Initialize Sun renderer
	sunRenderer, err := InitSunRenderer(screenWidth, screenHeight)
	if err != nil {
		fmt.Printf("Warning: Failed to initialize Sun renderer: %v\n", err)
		fmt.Println("Falling back to basic Sun rendering")
		sunRenderer = &SunRenderData{
			SunRadius:     SunRadiusUnits,
			IsInitialized: false,
		}
	}
	defer sunRenderer.Unload()

	// Initialize Planet renderer
	planetRenderer, err := InitPlanetRenderer()
	if err != nil {
		fmt.Printf("Warning: Failed to initialize Planet renderer: %v\n", err)
		fmt.Println("Falling back to basic planet rendering")
		planetRenderer = &PlanetRenderData{
			IsInitialized: false,
		}
	}
	defer planetRenderer.Unload()

	// Initialize lighting configuration system
	lightingConfig := NewLightingConfig()
	lightingConfig.ApplyPreset(PresetRealistic) // Start with realistic preset
	fmt.Println("\nLighting System Initialized - Preset: Realistic")
	fmt.Println("Presets: [6]Realistic [7]Cinematic [8]Educational [9]Stylized [0]Dark")

	fmt.Println("\nGoStarMap - 3D Galactic Navigation System with Orbital Mechanics")
	fmt.Println("=================================================================")
	fmt.Printf("Total Stars: %d\n", len(galaxy.Stars))
	fmt.Printf("Planets: %d (with validated JPL orbital elements)\n", len(galaxy.Planets))
	fmt.Println("\nCamera Controls:")
	fmt.Println("  WASD - Move camera horizontally")
	fmt.Println("  SPACE - Move up")
	fmt.Println("  LEFT CTRL - Move down")
	fmt.Println("  SHIFT - Speed boost")
	fmt.Println("  Mouse - Look around")
	fmt.Println("  TAB - Toggle stats")
	fmt.Println("\nTime Controls:")
	fmt.Println("  P - Pause/Resume time")
	fmt.Println("  [ ] - Decrease/Increase time speed")
	fmt.Println("  1 - 1 day per second")
	fmt.Println("  2 - 1 week per second")
	fmt.Println("  3 - 1 month per second")
	fmt.Println("  4 - Earth orbit in 60 seconds (default)")
	fmt.Println("  5 - 1 year per second")
	fmt.Println("\nFeatures:")
	fmt.Println("  - Realistic Keplerian orbital mechanics")
	fmt.Println("  - Validated orbital elements from JPL Horizons System")
	fmt.Println("  - Elliptical orbits with proper eccentricity")
	fmt.Println("  - 3D orbital inclinations")
	fmt.Println("  - Point camera at planets/Sun for detailed info")
	fmt.Println("\nNavigate through the solar system and watch planets orbit!")
	fmt.Println()

	rl.DisableCursor()

	starsRendered := 0
	showStats := true

	startTime := rl.GetTime()

	// Track camera orientation as yaw/pitch scalars derived from the initial
	// Position→Target direction. This avoids the forward×up gimbal lock when
	// looking straight up or down, and keeps strafe motion horizontal.
	initFwd := rl.Vector3Normalize(rl.Vector3Subtract(camera.Target, camera.Position))
	yaw := float32(math.Atan2(float64(initFwd.X), float64(initFwd.Z)))
	pitch := float32(math.Asin(float64(initFwd.Y)))
	const mouseSensitivity = float32(0.003)
	const maxPitch = float32(math.Pi/2 - 0.01) // clamp ~89.4° off the pole

	// Main game loop
	for !rl.WindowShouldClose() {
		// Check if shift is held for speed boost
		shiftHeld := rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift)
		cameraSpeed := GetCameraSpeed(camera, shiftHeld)

		// Mouse look — update yaw/pitch and clamp pitch to avoid the pole singularity.
		mouseDelta := rl.GetMouseDelta()
		yaw += mouseDelta.X * mouseSensitivity
		pitch -= mouseDelta.Y * mouseSensitivity // mouse up = look up
		if pitch > maxPitch {
			pitch = maxPitch
		} else if pitch < -maxPitch {
			pitch = -maxPitch
		}

		// Derive forward from yaw+pitch and a horizontal right from yaw alone.
		// A horizontal right keeps strafe motion level regardless of pitch.
		cp := float32(math.Cos(float64(pitch)))
		forward := rl.Vector3{
			X: cp * float32(math.Sin(float64(yaw))),
			Y: float32(math.Sin(float64(pitch))),
			Z: cp * float32(math.Cos(float64(yaw))),
		}
		right := rl.Vector3{
			X: float32(math.Cos(float64(yaw))),
			Y: 0,
			Z: -float32(math.Sin(float64(yaw))),
		}

		// Movement — only Position moves; Target is rebuilt from forward below.
		if rl.IsKeyDown(rl.KeyW) {
			camera.Position = rl.Vector3Add(camera.Position, rl.Vector3Scale(forward, cameraSpeed))
		}
		if rl.IsKeyDown(rl.KeyS) {
			camera.Position = rl.Vector3Subtract(camera.Position, rl.Vector3Scale(forward, cameraSpeed))
		}
		if rl.IsKeyDown(rl.KeyD) {
			camera.Position = rl.Vector3Add(camera.Position, rl.Vector3Scale(right, cameraSpeed))
		}
		if rl.IsKeyDown(rl.KeyA) {
			camera.Position = rl.Vector3Subtract(camera.Position, rl.Vector3Scale(right, cameraSpeed))
		}
		if rl.IsKeyDown(rl.KeySpace) {
			camera.Position.Y += cameraSpeed
		}
		if rl.IsKeyDown(rl.KeyLeftControl) {
			camera.Position.Y -= cameraSpeed
		}
		camera.Target = rl.Vector3Add(camera.Position, forward)

		// Toggle stats with TAB
		if rl.IsKeyPressed(rl.KeyTab) {
			showStats = !showStats
		}

		// Update lighting configuration with keyboard controls
		lightingChanged := lightingConfig.UpdateWithKeyboard()
		if lightingChanged {
			fmt.Printf("Lighting changed - Preset: %s\n", lightingConfig.CurrentPreset)
		}

		// Time control keys
		if rl.IsKeyPressed(rl.KeyP) {
			galaxy.TimeScale.IsPaused = !galaxy.TimeScale.IsPaused
		}
		if rl.IsKeyPressed(rl.KeyLeftBracket) {
			// Slow down time (divide by 2)
			galaxy.TimeScale.TimeScale /= 2.0
			if galaxy.TimeScale.TimeScale < 0.01 {
				galaxy.TimeScale.TimeScale = 0.01 // Minimum speed
			}
		}
		if rl.IsKeyPressed(rl.KeyRightBracket) {
			// Speed up time (multiply by 2)
			galaxy.TimeScale.TimeScale *= 2.0
			if galaxy.TimeScale.TimeScale > 36525.6 {
				galaxy.TimeScale.TimeScale = 36525.6 // Maximum: 100 years/second
			}
		}
		if rl.IsKeyPressed(rl.KeyOne) {
			galaxy.TimeScale.TimeScale = 1.0 // 1 day per second
		}
		if rl.IsKeyPressed(rl.KeyTwo) {
			galaxy.TimeScale.TimeScale = 7.0 // 1 week per second
		}
		if rl.IsKeyPressed(rl.KeyThree) {
			galaxy.TimeScale.TimeScale = 30.0 // 1 month per second
		}
		if rl.IsKeyPressed(rl.KeyFour) {
			galaxy.TimeScale.TimeScale = 365.256 / 60.0 // Earth orbit in 60 seconds
		}
		if rl.IsKeyPressed(rl.KeyFive) {
			galaxy.TimeScale.TimeScale = 365.256 // 1 year per second
		}

		// Update simulation time
		frameTime := rl.GetFrameTime() // Real seconds elapsed this frame
		orbital.UpdateSimulationTime(&galaxy.TimeScale, float64(frameTime))

		// Skip Kepler resolves while paused — positions are unchanged.
		if !galaxy.TimeScale.IsPaused {
			UpdatePlanetPositions(galaxy)
			UpdatePlanetRotations(galaxy)
		}

		// Draw
		rl.BeginDrawing()
		rl.ClearBackground(rl.Black)

		rl.BeginMode3D(camera)

		starsRendered = 0

		// Calculate elapsed time for animations
		elapsedTime := float32(rl.GetTime() - startTime)

		// Define Sun position for lighting calculations
		sunPos := rl.Vector3{X: 0, Y: 0, Z: 0}

		// Apply lighting configuration to planet shader
		lightingConfig.ApplyToShader(planetRenderer)

		// Draw planets first (before Sun, so Sun glows over them)
		RenderAllPlanets(planetRenderer, galaxy.Planets, sunPos, camera)

		// Draw stars with LOD and aggressive culling
		const maxRenderDistance = MaxRenderDistance // Much more aggressive culling
		const maxStarsPerFrame = MaxStarsPerFrame // Limit total stars rendered

		for _, star := range galaxy.Stars {
			if starsRendered >= maxStarsPerFrame {
				break // Stop rendering if we hit the limit
			}

			// Skip the Sun (it's rendered separately with special effects)
			if star.IsNamed && star.Name == "Sun" {
				continue
			}

			// Distance culling
			dx := float32(star.X) - camera.Position.X
			dy := float32(star.Y) - camera.Position.Y
			dz := float32(star.Z) - camera.Position.Z
			distSq := dx*dx + dy*dy + dz*dz

			if distSq < maxRenderDistance*maxRenderDistance {
				RenderStarLOD(star, camera.Position)
				starsRendered++
			}
		}

		// Draw the Sun last so it glows brightly on top.
		if sunRenderer.IsInitialized {
			RenderSun(sunRenderer, camera, elapsedTime)
		} else {
			rl.DrawSphere(rl.Vector3{}, sunRenderer.SunRadius, rl.Gold)
		}

		rl.EndMode3D()

		// Draw center reticle/crosshair
		centerX := screenWidth / 2
		centerY := screenHeight / 2
		reticleSize := int32(UIReticleSize)
		reticleThickness := int32(UIReticleThickness)
		reticleGap := int32(UIReticleGap)
		reticleColor := rl.NewColor(255, 255, 255, 180) // Semi-transparent white

		// Draw crosshair lines (horizontal and vertical)
		// Top line
		rl.DrawRectangle(centerX-reticleThickness/2, centerY-reticleSize-reticleGap, reticleThickness, reticleSize-reticleGap, reticleColor)
		// Bottom line
		rl.DrawRectangle(centerX-reticleThickness/2, centerY+reticleGap, reticleThickness, reticleSize-reticleGap, reticleColor)
		// Left line
		rl.DrawRectangle(centerX-reticleSize-reticleGap, centerY-reticleThickness/2, reticleSize-reticleGap, reticleThickness, reticleColor)
		// Right line
		rl.DrawRectangle(centerX+reticleGap, centerY-reticleThickness/2, reticleSize-reticleGap, reticleThickness, reticleColor)

		// Draw center dot
		rl.DrawCircle(centerX, centerY, 2, reticleColor)

		// Check if player is targeting a celestial body
		targetedPlanet, targetingSun := CheckTargeting(camera, galaxy.Planets, sunPos)

		// Draw celestial body info if something is targeted
		if targetingSun || targetedPlanet != nil {
			DrawCelestialInfo(screenWidth, screenHeight, targetedPlanet, targetingSun)
		}

		// Draw Time Control UI (always visible)
		timeBoxX := int32(10)
		timeBoxY := screenHeight - int32(UITimeBoxHeight) - 10

		rl.DrawRectangle(timeBoxX, timeBoxY, UITimeBoxWidth, UITimeBoxHeight, rl.NewColor(0, 0, 0, 200))
		rl.DrawRectangleLines(timeBoxX, timeBoxY, UITimeBoxWidth, UITimeBoxHeight, rl.NewColor(100, 200, 100, 255))

		timeYPos := timeBoxY + 10
		if galaxy.TimeScale.IsPaused {
			rl.DrawText("TIME: PAUSED", timeBoxX+15, timeYPos, 20, rl.Red)
		} else {
			rl.DrawText("TIME: RUNNING", timeBoxX+15, timeYPos, 20, rl.Green)
		}
		timeYPos += 25

		// Calculate simulation date (days since J2000.0)
		years := galaxy.TimeScale.SimulationDays / 365.256
		displayYear := 2000.0 + years
		rl.DrawText(fmt.Sprintf("Date: %.1f (J2000 + %.1f days)", displayYear, galaxy.TimeScale.SimulationDays), timeBoxX+15, timeYPos, 14, rl.RayWhite)
		timeYPos += 20

		// Show time scale
		var timeScaleStr string
		if galaxy.TimeScale.TimeScale < 1.0 {
			timeScaleStr = fmt.Sprintf("%.2f days/sec", galaxy.TimeScale.TimeScale)
		} else if galaxy.TimeScale.TimeScale < 365.256 {
			timeScaleStr = fmt.Sprintf("%.1f days/sec", galaxy.TimeScale.TimeScale)
		} else {
			yearsPerSec := galaxy.TimeScale.TimeScale / 365.256
			timeScaleStr = fmt.Sprintf("%.2f years/sec", yearsPerSec)
		}
		rl.DrawText(fmt.Sprintf("Speed: %s", timeScaleStr), timeBoxX+15, timeYPos, 14, rl.RayWhite)
		timeYPos += 20

		rl.DrawText("P=Pause  [/]=Slower/Faster  1-5=Presets", timeBoxX+15, timeYPos, 12, rl.Gray)

		// Draw UI
		if showStats {
			yPos := int32(10)
			rl.DrawText(fmt.Sprintf("GoStarMap - Milky Way Galaxy [%d total stars]", len(galaxy.Stars)), 10, yPos, 20, rl.RayWhite)
			yPos += 25

			rl.DrawText("TAB=Toggle Stats | WASD=Move | Shift=Boost | Space/Ctrl=Up/Down | Mouse=Look | ESC=Exit", 10, yPos, 14, rl.Gray)
			yPos += 20

			camInfo := fmt.Sprintf("Pos: (%.0f, %.0f, %.0f) | Speed: %.1f",
				camera.Position.X, camera.Position.Y, camera.Position.Z, cameraSpeed)
			rl.DrawText(camInfo, 10, yPos, 14, rl.DarkGray)
			yPos += 20

			renderInfo := fmt.Sprintf("Stars Rendered: %d / %d (%.1f%%)",
				starsRendered, len(galaxy.Stars), float64(starsRendered)/float64(len(galaxy.Stars))*100)
			rl.DrawText(renderInfo, 10, yPos, 14, rl.DarkGray)
			yPos += 20

			distFromSun := math.Sqrt(float64(camera.Position.X*camera.Position.X +
				camera.Position.Y*camera.Position.Y +
				camera.Position.Z*camera.Position.Z))
			rl.DrawText(fmt.Sprintf("Distance from Sol: %.1f light-years", distFromSun/50.0), 10, yPos, 14, rl.DarkGray)
			yPos += 25

			// Lighting information
			rl.DrawText(fmt.Sprintf("Lighting Preset: %s", lightingConfig.CurrentPreset), 10, yPos, 14, rl.Yellow)
			yPos += 15
			rl.DrawText(fmt.Sprintf("Sun: %.2f | Ambient: %.2f | Terminator: %.2f | Rim: %.2f",
				lightingConfig.SunIntensity, lightingConfig.AmbientStrength,
				lightingConfig.TerminatorSoftness, lightingConfig.RimLightStrength), 10, yPos, 12, rl.Gold)
			yPos += 15
			rl.DrawText("[6-0]=Presets | [L/K]=Sun | [O/I]=Ambient | [T/Y]=Terminator | [R]=Rim", 10, yPos, 11, rl.DarkGray)

			rl.DrawFPS(screenWidth-100, 10)
		} else {
			rl.DrawText("Press TAB to show stats", 10, 10, 16, rl.Gray)
			rl.DrawFPS(screenWidth-100, 10)
		}

		rl.EndDrawing()
	}
}
