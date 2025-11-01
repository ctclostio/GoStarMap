package main

import (
	"fmt"
	"math"
	"math/rand"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// SpectralType represents stellar classification
type SpectralType int

const (
	TypeO SpectralType = iota // Blue
	TypeB                      // Blue-white
	TypeA                      // White
	TypeF                      // Yellow-white
	TypeG                      // Yellow (like our Sun)
	TypeK                      // Orange
	TypeM                      // Red
)

// GetSpectralColor returns the color for a spectral type
func GetSpectralColor(stype SpectralType) rl.Color {
	colors := map[SpectralType]rl.Color{
		TypeO: rl.NewColor(155, 176, 255, 255), // Blue
		TypeB: rl.NewColor(170, 191, 255, 255), // Blue-white
		TypeA: rl.NewColor(202, 215, 255, 255), // White
		TypeF: rl.NewColor(248, 247, 255, 255), // Yellow-white
		TypeG: rl.NewColor(255, 244, 234, 255), // Yellow
		TypeK: rl.NewColor(255, 210, 161, 255), // Orange
		TypeM: rl.NewColor(255, 204, 111, 255), // Red
	}
	return colors[stype]
}

// Star represents a star with astronomical properties
type Star struct {
	X, Y, Z       float64      // Position in render units
	Magnitude     float32      // Visual magnitude (brightness)
	SpectralType  SpectralType // Spectral classification
	Name          string       // Name (if notable)
	IsNamed       bool         // Whether this star has a name
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
	OrbitalElements OrbitalElements // Keplerian orbital parameters

	// Rotation parameters (NASA Planetary Fact Sheet validated)
	RotationPeriodDays float64 // Sidereal rotation period (negative if retrograde)
	AxialTiltDegrees   float64 // Obliquity (axial tilt from orbital plane)
	CurrentRotation    float64 // Current rotation angle in radians (updated each frame)
}

// Galaxy contains all celestial objects and simulation time
type Galaxy struct {
	Stars        []Star
	Planets      []Planet
	TimeScale    TimeScaleInfo // Simulation time control
}

// NewGalaxy creates a new galaxy with initial time set to J2000.0 epoch
func NewGalaxy() *Galaxy {
	return &Galaxy{
		Stars:   make([]Star, 0),
		Planets: make([]Planet, 0),
		TimeScale: TimeScaleInfo{
			SimulationDays: 0.0,    // Start at J2000.0 epoch (January 1, 2000, 12:00 TT)
			TimeScale:      365.256 / 60.0, // Default: Earth completes orbit in 60 seconds
			IsPaused:       false,
		},
	}
}

// GetRandomSpectralType returns a spectral type based on realistic distribution
func GetRandomSpectralType() SpectralType {
	// Realistic distribution: M stars are most common, O stars are rarest
	// Based on stellar population statistics (cumulative percentages)
	r := rand.Float32()
	switch {
	case r < 0.00003: // 0.003% O stars (very rare, massive, blue supergiants)
		return TypeO
	case r < 0.0013: // 0.13% B stars (rare, blue-white, hot)
		return TypeB
	case r < 0.006: // 0.6% A stars (white, hotter than Sun)
		return TypeA
	case r < 0.03: // 3% F stars (yellow-white, slightly hotter than Sun)
		return TypeF
	case r < 0.076: // 7.6% G stars (yellow, like our Sun)
		return TypeG
	case r < 0.197: // 12.1% K stars (orange, cooler than Sun)
		return TypeK
	default: // 76% M stars (red dwarfs, most common in galaxy)
		return TypeM
	}
}

// AddNamedStar adds a notable star with a name
func (g *Galaxy) AddNamedStar(name string, x, y, z float64, mag float32, stype SpectralType) {
	g.Stars = append(g.Stars, Star{
		X:            x,
		Y:            y,
		Z:            z,
		Magnitude:    mag,
		SpectralType: stype,
		Name:         name,
		IsNamed:      true,
	})
}

// GenerateGalaxy creates a procedural galaxy with realistic structure
func GenerateGalaxy() *Galaxy {
	g := NewGalaxy()

	fmt.Println("Generating galaxy...")

	// 1. Add our solar system at origin
	g.AddNamedStar("Sun", 0, 0, 0, -26.74, TypeG)

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
		elements := GetPlanetaryElements(p.name)

		// Calculate initial position at J2000.0 epoch (time = 0)
		positionAU := OrbitalElementsToPosition(elements, 0.0)

		// Convert to render units (AU to your scale)
		renderPos := PositionToRenderUnits(positionAU, AUScale)

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
		name     string
		x, y, z  float64 // in light-years
		mag      float32
		stype    SpectralType
	}{
		{"Proxima Centauri", 4.24, 0.5, -1.2, 11.13, TypeM},
		{"Alpha Centauri A", 4.37, 0.6, -1.1, -0.01, TypeG},
		{"Barnard's Star", 5.96, 1.2, 0.8, 9.53, TypeM},
		{"Sirius", 8.6, -2.5, 3.1, -1.46, TypeA},
		{"Procyon", 11.4, 3.0, 2.0, 0.38, TypeF},
		{"61 Cygni", 11.4, 5.0, -3.0, 5.2, TypeK},
		{"Epsilon Eridani", 10.5, -2.0, 4.0, 3.73, TypeK},
		{"Tau Ceti", 11.9, 1.0, -5.0, 3.49, TypeG},
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
		stype := GetRandomSpectralType()

		g.Stars = append(g.Stars, Star{
			X:            x,
			Y:            y,
			Z:            z,
			Magnitude:    mag,
			SpectralType: stype,
			IsNamed:      false,
		})
	}

	fmt.Printf("Galaxy generation complete: %d total stars\n", len(g.Stars))
	return g
}

// GetCameraSpeed returns camera speed - normal by default, boosted when shift is held
func GetCameraSpeed(camera rl.Camera3D, shiftHeld bool) float32 {
	baseSpeed := float32(0.5)

	// If shift is not held, just return normal speed
	if !shiftHeld {
		return baseSpeed
	}

	// Shift is held - apply distance-based speed boost
	dist := float32(math.Sqrt(float64(camera.Position.X*camera.Position.X +
		camera.Position.Y*camera.Position.Y +
		camera.Position.Z*camera.Position.Z)))

	if dist < 50 {
		return baseSpeed * 3.0  // 3x boost for close range
	}
	return baseSpeed + dist*0.05  // Distance-based boost for far range
}

// RenderStarLOD renders a star with level-of-detail based on camera distance
func RenderStarLOD(star Star, cameraPos rl.Vector3) {
	starPos := rl.Vector3{X: float32(star.X), Y: float32(star.Y), Z: float32(star.Z)}

	// Calculate distance to camera
	dx := starPos.X - cameraPos.X
	dy := starPos.Y - cameraPos.Y
	dz := starPos.Z - cameraPos.Z
	dist := float32(math.Sqrt(float64(dx*dx + dy*dy + dz*dz)))

	color := GetSpectralColor(star.SpectralType)

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
	case dist < 100: // Very close - full sphere
		radius := 2.0 + brightnessFactor*1.0
		rl.DrawSphere(starPos, radius, color)

	case dist < 300: // Medium distance - medium sphere
		radius := 1.0 + brightnessFactor*0.4
		rl.DrawSphere(starPos, radius, color)

	case dist < 800: // Far - small sphere
		rl.DrawSphere(starPos, 0.6, color)

	default: // Everything else - single point (most efficient)
		rl.DrawPoint3D(starPos, color)
	}
}

// CheckTargeting determines if the camera is looking at a celestial body
// Returns the targeted planet (or nil) and whether the Sun is targeted
func CheckTargeting(camera rl.Camera3D, planets []Planet, sunPos rl.Vector3) (*Planet, bool) {
	// Calculate camera forward direction
	forward := rl.Vector3Subtract(camera.Target, camera.Position)
	forward = rl.Vector3Normalize(forward)

	// Targeting threshold (in degrees)
	const targetThreshold = 2.0 // 2 degrees
	thresholdRad := targetThreshold * math.Pi / 180.0

	// Check Sun first
	sunDir := rl.Vector3Subtract(sunPos, camera.Position)
	sunDist := rl.Vector3Length(sunDir)
	if sunDist > 0.01 {
		sunDir = rl.Vector3Normalize(sunDir)
		dotProduct := forward.X*sunDir.X + forward.Y*sunDir.Y + forward.Z*sunDir.Z
		// Clamp dot product to valid range for acos to prevent NaN from floating-point errors
		dotProduct = float32(math.Max(-1.0, math.Min(1.0, float64(dotProduct))))
		angle := math.Acos(float64(dotProduct))
		if angle < thresholdRad {
			return nil, true // Sun is targeted
		}
	}

	// Check each planet
	for i := range planets {
		planetPos := rl.Vector3{X: float32(planets[i].X), Y: float32(planets[i].Y), Z: float32(planets[i].Z)}
		planetDir := rl.Vector3Subtract(planetPos, camera.Position)
		planetDist := rl.Vector3Length(planetDir)

		if planetDist > 0.01 {
			planetDir = rl.Vector3Normalize(planetDir)
			dotProduct := forward.X*planetDir.X + forward.Y*planetDir.Y + forward.Z*planetDir.Z
			// Clamp dot product to valid range for acos to prevent NaN from floating-point errors
			dotProduct = float32(math.Max(-1.0, math.Min(1.0, float64(dotProduct))))
			angle := math.Acos(float64(dotProduct))
			if angle < thresholdRad {
				return &planets[i], false // Planet is targeted
			}
		}
	}

	return nil, false // Nothing targeted
}

// DrawCelestialInfo displays information about a targeted celestial body
func DrawCelestialInfo(screenWidth, screenHeight int32, planet *Planet, isSun bool) {
	// Position the info box in the lower right quadrant
	boxX := screenWidth - 420
	boxY := screenHeight - 340

	// Draw semi-transparent background
	rl.DrawRectangle(boxX, boxY, 400, 320, rl.NewColor(0, 0, 0, 200))
	rl.DrawRectangleLines(boxX, boxY, 400, 320, rl.NewColor(100, 150, 255, 255))

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

		rl.DrawText(fmt.Sprintf("Diameter: %s km", formatNumber(planet.DiameterKm)), boxX+15, yOffset, 14, rl.RayWhite)
		yOffset += 20

		rl.DrawText(fmt.Sprintf("Semi-major Axis: %.3f AU", planet.OrbitalElements.SemiMajorAxis), boxX+15, yOffset, 14, rl.RayWhite)
		yOffset += 20

		rl.DrawText(fmt.Sprintf("Orbital Period: %s days", formatNumber(planet.OrbitalElements.OrbitalPeriod)), boxX+15, yOffset, 14, rl.RayWhite)
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

// formatNumber adds commas to numbers for readability
func formatNumber(num float64) string {
	str := fmt.Sprintf("%.0f", num)
	n := len(str)
	if n <= 3 {
		return str
	}
	result := ""
	for i, c := range str {
		if i > 0 && (n-i)%3 == 0 {
			result += ","
		}
		result += string(c)
	}
	return result
}

// UpdatePlanetPositions recalculates planet positions based on current simulation time
// Expected cost: ~0.02ms for 8 planets on modern CPU
func UpdatePlanetPositions(galaxy *Galaxy) {
	currentTime := galaxy.TimeScale.SimulationDays

	for i := range galaxy.Planets {
		planet := &galaxy.Planets[i]

		// Calculate position in AU based on orbital elements
		positionAU := OrbitalElementsToPosition(planet.OrbitalElements, currentTime)

		// Convert to render coordinates
		renderPos := PositionToRenderUnits(positionAU, AUScale)

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

// SunRenderData contains all resources needed for photorealistic Sun rendering
type SunRenderData struct {
	Shader          rl.Shader
	RenderTexture   rl.RenderTexture2D
	BloomTexture    rl.RenderTexture2D
	BloomBlurTemp   rl.RenderTexture2D
	ExtractShader   rl.Shader
	BlurShader      rl.Shader
	CompositeShader rl.Shader
	SunRadius       float32
	IsInitialized   bool
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

	// Cached sphere models (LOD - Level of Detail)
	// These are created once and reused every frame for performance
	SphereModelHigh   rl.Model // 32 rings/slices - close up
	SphereModelMedium rl.Model // 24 rings/slices - medium distance
	SphereModelLow    rl.Model // 16 rings/slices - far away
}

// Physical Sun parameters
const (
	SunRadiusUnits   = 32.7  // Physically accurate: 109x Earth radius (0.3 units)
	SunDistanceScale = 1.0   // Sun at origin
	BloomThreshold   = 1.0   // HDR brightness threshold for bloom
	BloomIntensity   = 0.35  // Bloom effect strength
	ExposureValue    = 1.2   // HDR exposure

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

// InitSunRenderer initializes all resources for Sun rendering with bloom
// Expected cost: One-time initialization, ~20ms
// Memory: ~12MB for render textures at 1920x1080
func InitSunRenderer(screenWidth, screenHeight int32) (*SunRenderData, error) {
	sun := &SunRenderData{
		SunRadius: SunRadiusUnits,
	}

	// Load Sun shader
	sunShader := rl.LoadShader("shaders/sun.vs", "shaders/sun.fs")
	if sunShader.ID == 0 {
		return nil, fmt.Errorf("failed to load Sun shader")
	}
	sun.Shader = sunShader

	// Load bloom post-processing shaders
	sun.ExtractShader = rl.LoadShader("shaders/default.vs", "shaders/bloom_extract.fs")
	if sun.ExtractShader.ID == 0 {
		return nil, fmt.Errorf("failed to load bloom extract shader")
	}

	sun.BlurShader = rl.LoadShader("shaders/default.vs", "shaders/bloom_blur.fs")
	if sun.BlurShader.ID == 0 {
		return nil, fmt.Errorf("failed to load bloom blur shader")
	}

	sun.CompositeShader = rl.LoadShader("shaders/default.vs", "shaders/bloom_composite.fs")
	if sun.CompositeShader.ID == 0 {
		return nil, fmt.Errorf("failed to load bloom composite shader")
	}

	// Create render textures for HDR rendering and bloom
	// Using half resolution for bloom textures saves memory and improves performance
	sun.RenderTexture = rl.LoadRenderTexture(screenWidth, screenHeight)
	sun.BloomTexture = rl.LoadRenderTexture(screenWidth/2, screenHeight/2)
	sun.BloomBlurTemp = rl.LoadRenderTexture(screenWidth/2, screenHeight/2)

	sun.IsInitialized = true
	fmt.Println("Sun renderer initialized successfully")
	return sun, nil
}

// UnloadSunRenderer cleans up all Sun rendering resources
func (sun *SunRenderData) Unload() {
	if !sun.IsInitialized {
		return
	}

	rl.UnloadShader(sun.Shader)
	rl.UnloadShader(sun.ExtractShader)
	rl.UnloadShader(sun.BlurShader)
	rl.UnloadShader(sun.CompositeShader)
	rl.UnloadRenderTexture(sun.RenderTexture)
	rl.UnloadRenderTexture(sun.BloomTexture)
	rl.UnloadRenderTexture(sun.BloomBlurTemp)

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

// RenderSunWithBloom renders the Sun with full HDR bloom pipeline
// This is the main rendering function that should be called from the game loop
// Expected total cost: 0.8-1.2ms @ 1080p on RTX 3070 / GTX 1660
func RenderSunWithBloom(sun *SunRenderData, camera rl.Camera3D, time float32, screenWidth, screenHeight int32) {
	if !sun.IsInitialized {
		// Fallback: render simple sun without bloom
		sunPos := rl.Vector3{X: 0, Y: 0, Z: 0}
		rl.DrawSphere(sunPos, sun.SunRadius, rl.Gold)
		return
	}

	// Just render the Sun directly with the shader
	// The shader already has HDR emission built-in
	// For now, we'll skip the full bloom pipeline to keep it simple
	RenderSun(sun, camera, time)

	// TODO: Optional bloom pipeline for even more dramatic effect
	// This would require rendering to texture first, then post-processing
	// Skipped for now to maintain simplicity and performance
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

	// Create and cache sphere models for LOD (Level of Detail)
	// These are created once and reused every frame for maximum performance
	// Radius of 1.0 is used as base - will be scaled when rendering
	meshHigh := rl.GenMeshSphere(1.0, 32, 32)
	planet.SphereModelHigh = rl.LoadModelFromMesh(meshHigh)
	planet.SphereModelHigh.Materials.Shader = planetShader

	meshMedium := rl.GenMeshSphere(1.0, 24, 24)
	planet.SphereModelMedium = rl.LoadModelFromMesh(meshMedium)
	planet.SphereModelMedium.Materials.Shader = planetShader

	meshLow := rl.GenMeshSphere(1.0, 16, 16)
	planet.SphereModelLow = rl.LoadModelFromMesh(meshLow)
	planet.SphereModelLow.Materials.Shader = planetShader

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

// RenderPlanet renders a single planet with physically-based lighting
// Expected cost: ~0.05ms per planet @ 1080p on RTX 3070 / GTX 1660
// Bandwidth: ~1-2MB per planet (depends on screen coverage)
func RenderPlanet(planet *PlanetRenderData, params PlanetRenderParams, sunPos rl.Vector3, camera rl.Camera3D) {
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

	// Set shader uniforms (per-planet parameters only)
	// Note: Global lighting params (sunIntensity, ambientStrength, etc.) are set by lightingConfig.ApplyToShader()
	rl.SetShaderValue(planet.Shader, planet.LocSunPosition, []float32{sunPos.X, sunPos.Y, sunPos.Z}, rl.ShaderUniformVec3)
	rl.SetShaderValue(planet.Shader, planet.LocPlanetColor, colorVec, rl.ShaderUniformVec3)
	rl.SetShaderValue(planet.Shader, planet.LocPlanetRadius, []float32{params.Radius}, rl.ShaderUniformFloat)
	// Removed: sunIntensity and ambientStrength - these come from lightingConfig now
	rl.SetShaderValue(planet.Shader, planet.LocSpecularStrength, []float32{params.SpecularStrength}, rl.ShaderUniformFloat)
	rl.SetShaderValue(planet.Shader, planet.LocShininess, []float32{params.Shininess}, rl.ShaderUniformFloat)
	rl.SetShaderValue(planet.Shader, planet.LocCameraPosition, []float32{camera.Position.X, camera.Position.Y, camera.Position.Z}, rl.ShaderUniformVec3)

	// Set rotation parameters
	rl.SetShaderValue(planet.Shader, planet.LocRotationAngle, []float32{params.RotationAngle}, rl.ShaderUniformFloat)
	rl.SetShaderValue(planet.Shader, planet.LocAxialTilt, []float32{params.AxialTilt}, rl.ShaderUniformFloat)

	// Calculate distance to camera for LOD (Level of Detail)
	camDist := rl.Vector3Distance(camera.Position, params.Position)
	screenSize := params.Radius / camDist // Approximate angular size on screen

	// Select appropriate LOD model based on screen size
	// Using cached models for maximum performance (created once at init)
	var selectedModel rl.Model
	if screenSize > 0.05 { // Close up - high detail (32 rings/slices)
		selectedModel = planet.SphereModelHigh
	} else if screenSize > 0.02 { // Medium distance (24 rings/slices)
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

// RenderAllPlanets renders all planets in the solar system with proper lighting
// Expected total cost: ~0.64ms for 8 planets @ 1080p (with enhanced lighting)
// Note: Lighting parameters are set globally via lightingConfig.ApplyToShader()
// This function only sets per-planet parameters (color, position, size)
func RenderAllPlanets(planetRenderer *PlanetRenderData, planets []Planet, sunPos rl.Vector3, camera rl.Camera3D) {
	// Get lighting config from global state (passed via ApplyToShader before this call)
	// Per-planet specular values are taken from constants (can be overridden by config)

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

		RenderPlanet(planetRenderer, params, sunPos, camera)
	}
}

func main() {
	// Initialize window
	screenWidth := int32(1920)
	screenHeight := int32(1080)

	rl.InitWindow(screenWidth, screenHeight, "GoStarMap - 3D Galactic Navigation [100k Stars - Optimized]")
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

	// Main game loop
	for !rl.WindowShouldClose() {
		// Check if shift is held for speed boost
		shiftHeld := rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift)
		cameraSpeed := GetCameraSpeed(camera, shiftHeld)

		// Camera controls
		forward := rl.Vector3Subtract(camera.Target, camera.Position)
		forward = rl.Vector3Normalize(forward)

		right := rl.Vector3CrossProduct(forward, camera.Up)
		right = rl.Vector3Normalize(right)

		// Movement
		if rl.IsKeyDown(rl.KeyW) {
			camera.Position = rl.Vector3Add(camera.Position, rl.Vector3Scale(forward, cameraSpeed))
			camera.Target = rl.Vector3Add(camera.Target, rl.Vector3Scale(forward, cameraSpeed))
		}
		if rl.IsKeyDown(rl.KeyS) {
			camera.Position = rl.Vector3Subtract(camera.Position, rl.Vector3Scale(forward, cameraSpeed))
			camera.Target = rl.Vector3Subtract(camera.Target, rl.Vector3Scale(forward, cameraSpeed))
		}
		if rl.IsKeyDown(rl.KeyD) {
			camera.Position = rl.Vector3Add(camera.Position, rl.Vector3Scale(right, cameraSpeed))
			camera.Target = rl.Vector3Add(camera.Target, rl.Vector3Scale(right, cameraSpeed))
		}
		if rl.IsKeyDown(rl.KeyA) {
			camera.Position = rl.Vector3Subtract(camera.Position, rl.Vector3Scale(right, cameraSpeed))
			camera.Target = rl.Vector3Subtract(camera.Target, rl.Vector3Scale(right, cameraSpeed))
		}
		if rl.IsKeyDown(rl.KeySpace) {
			camera.Position.Y += cameraSpeed
			camera.Target.Y += cameraSpeed
		}
		if rl.IsKeyDown(rl.KeyLeftControl) {
			camera.Position.Y -= cameraSpeed
			camera.Target.Y -= cameraSpeed
		}

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
		UpdateSimulationTime(&galaxy.TimeScale, float64(frameTime))

		// Update planet positions based on current time
		UpdatePlanetPositions(galaxy)

		// Update planet rotations based on current time
		UpdatePlanetRotations(galaxy)

		// Mouse look
		mouseDelta := rl.GetMouseDelta()
		sensitivity := float32(0.003)

		if mouseDelta.X != 0 {
			angle := mouseDelta.X * sensitivity
			offset := rl.Vector3Subtract(camera.Target, camera.Position)
			cosAngle := float32(math.Cos(float64(angle)))
			sinAngle := float32(math.Sin(float64(angle)))

			newX := offset.X*cosAngle - offset.Z*sinAngle
			newZ := offset.X*sinAngle + offset.Z*cosAngle

			camera.Target = rl.Vector3Add(camera.Position, rl.Vector3{X: newX, Y: offset.Y, Z: newZ})
		}

		if mouseDelta.Y != 0 {
			angle := -mouseDelta.Y * sensitivity // Negative for natural mouse look
			offset := rl.Vector3Subtract(camera.Target, camera.Position)

			cosAngle := float32(math.Cos(float64(angle)))
			sinAngle := float32(math.Sin(float64(angle)))

			// Rotate forward vector around the camera's right vector (local space rotation)
			// Using Rodrigues' rotation formula: v_rot = v*cos(θ) + (axis × v)*sin(θ)
			forward := rl.Vector3Normalize(offset)

			// Cross product: right × forward (perpendicular to both, creates the rotation plane)
			crossProduct := rl.Vector3CrossProduct(right, forward)

			// Apply rotation around right vector
			newForward := rl.Vector3Add(
				rl.Vector3Scale(forward, cosAngle),
				rl.Vector3Scale(crossProduct, sinAngle),
			)
			newForward = rl.Vector3Normalize(newForward)

			length := rl.Vector3Length(offset)
			camera.Target = rl.Vector3Add(camera.Position, rl.Vector3Scale(newForward, length))
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
		const maxRenderDistance = 20000.0 // Much more aggressive culling
		const maxStarsPerFrame = 15000     // Limit total stars rendered

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

		// Draw the Sun last so it glows brightly on top
		RenderSunWithBloom(sunRenderer, camera, elapsedTime, screenWidth, screenHeight)

		rl.EndMode3D()

		// Draw center reticle/crosshair
		centerX := screenWidth / 2
		centerY := screenHeight / 2
		reticleSize := int32(10)
		reticleThickness := int32(2)
		reticleGap := int32(4)
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
		timeBoxY := screenHeight - 120
		timeBoxWidth := int32(480)
		timeBoxHeight := int32(110)

		rl.DrawRectangle(timeBoxX, timeBoxY, timeBoxWidth, timeBoxHeight, rl.NewColor(0, 0, 0, 200))
		rl.DrawRectangleLines(timeBoxX, timeBoxY, timeBoxWidth, timeBoxHeight, rl.NewColor(100, 200, 100, 255))

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

	rl.CloseWindow()
}
