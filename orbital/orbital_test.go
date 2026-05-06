package orbital

import (
	"math"
	"testing"
)

func TestSolveKeplerEquation(t *testing.T) {
	tests := []struct {
		name          string
		meanAnomaly   float64
		eccentricity  float64
		wantE         float64 // Expected eccentric anomaly (approximate)
		tolerance     float64
	}{
		{
			name:         "Circular orbit (e=0), M=0",
			meanAnomaly:  0.0,
			eccentricity: 0.0,
			wantE:        0.0,
			tolerance:    1e-10,
		},
		{
			name:         "Circular orbit (e=0), M=π/2",
			meanAnomaly:  math.Pi / 2,
			eccentricity: 0.0,
			wantE:        math.Pi / 2,
			tolerance:    1e-10,
		},
		{
			name:         "Low eccentricity, M=1.0",
			meanAnomaly:  1.0,
			eccentricity: 0.1,
			wantE:        1.10184, // Approximate solution
			tolerance:    1e-4,
		},
		{
			name:         "Earth-like eccentricity, M=0",
			meanAnomaly:  0.0,
			eccentricity: 0.0167,
			wantE:        0.0,
			tolerance:    1e-10,
		},
		{
			name:         "Mars-like eccentricity, M=π",
			meanAnomaly:  math.Pi,
			eccentricity: 0.0934,
			wantE:        math.Pi, // At periapsis/apoapsis, M = E
			tolerance:    1e-10,
		},
		{
			name:         "High eccentricity (e=0.7), M=1.0",
			meanAnomaly:  1.0,
			eccentricity: 0.7,
			wantE:        1.88793, // Approximate solution
			tolerance:    1e-4,
		},
		{
			name:         "Normalized mean anomaly (M > 2π)",
			meanAnomaly:  10.0, // ~1.5 * 2π + remainder
			eccentricity: 0.2,
			wantE:        SolveKeplerEquation(10.0-2*math.Pi, 0.2), // Should normalize
			tolerance:    1e-6,
		},
		{
			name:         "Negative mean anomaly",
			meanAnomaly:  -1.0,
			eccentricity: 0.1,
			wantE:        SolveKeplerEquation(-1.0+2*math.Pi, 0.1), // Should normalize to [0, 2π]
			tolerance:    1e-6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SolveKeplerEquation(tt.meanAnomaly, tt.eccentricity)

			// Verify the solution satisfies Kepler's equation: M = E - e*sin(E)
			checkM := got - tt.eccentricity*math.Sin(got)

			// Normalize checkM to [0, 2π)
			checkM = math.Mod(checkM, 2*math.Pi)
			if checkM < 0 {
				checkM += 2 * math.Pi
			}

			// Normalize meanAnomaly to [0, 2π)
			wantM := math.Mod(tt.meanAnomaly, 2*math.Pi)
			if wantM < 0 {
				wantM += 2 * math.Pi
			}

			if math.Abs(checkM-wantM) > tt.tolerance {
				t.Errorf("SolveKeplerEquation() = %v, satisfies M = %v, want M = %v (tolerance %v)",
					got, checkM, wantM, tt.tolerance)
			}
		})
	}
}

func TestCalculateTrueAnomaly(t *testing.T) {
	tests := []struct {
		name             string
		eccentricAnomaly float64
		eccentricity     float64
		wantNu           float64 // Expected true anomaly
		tolerance        float64
	}{
		{
			name:             "Circular orbit, E=0",
			eccentricAnomaly: 0.0,
			eccentricity:     0.0,
			wantNu:           0.0,
			tolerance:        1e-10,
		},
		{
			name:             "Circular orbit, E=π/2",
			eccentricAnomaly: math.Pi / 2,
			eccentricity:     0.0,
			wantNu:           math.Pi / 2,
			tolerance:        1e-10,
		},
		{
			name:             "At periapsis (E=0), e=0.5",
			eccentricAnomaly: 0.0,
			eccentricity:     0.5,
			wantNu:           0.0,
			tolerance:        1e-10,
		},
		{
			name:             "At apoapsis (E=π), e=0.5",
			eccentricAnomaly: math.Pi,
			eccentricity:     0.5,
			wantNu:           math.Pi,
			tolerance:        1e-10,
		},
		{
			name:             "Quarter orbit, e=0.2",
			eccentricAnomaly: math.Pi / 2,
			eccentricity:     0.2,
			wantNu:           1.772, // Corrected approximate value
			tolerance:        1e-3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateTrueAnomaly(tt.eccentricAnomaly, tt.eccentricity)

			// Verify true anomaly is in [0, 2π)
			if got < 0 || got >= 2*math.Pi {
				t.Errorf("True anomaly = %v, should be in [0, 2π)", got)
			}

			if math.Abs(got-tt.wantNu) > tt.tolerance {
				t.Errorf("CalculateTrueAnomaly() = %v, want %v (±%v)", got, tt.wantNu, tt.tolerance)
			}
		})
	}
}

func TestOrbitalElementsToPosition(t *testing.T) {
	// Test circular orbit at 1 AU (should be roughly at position for given mean anomaly)
	circularOrbit := OrbitalElements{
		SemiMajorAxis:        1.0,
		Eccentricity:         0.0,
		Inclination:          0.0,
		LongitudeAscendingNode: 0.0,
		ArgumentOfPeriapsis:  0.0,
		MeanAnomalyAtEpoch:   0.0,
		OrbitalPeriod:        365.256,
	}

	t.Run("Circular orbit at epoch (t=0)", func(t *testing.T) {
		pos := OrbitalElementsToPosition(circularOrbit, 0.0)

		// At t=0, M=0, so should be at periapsis (1, 0, 0) in orbital plane
		// With no inclination or rotation, this should be (1, 0, 0) in ecliptic
		if math.Abs(pos[0]-1.0) > 1e-6 {
			t.Errorf("X position = %v, want 1.0", pos[0])
		}
		if math.Abs(pos[1]) > 1e-6 {
			t.Errorf("Y position = %v, want 0.0", pos[1])
		}
		if math.Abs(pos[2]) > 1e-6 {
			t.Errorf("Z position = %v, want 0.0", pos[2])
		}
	})

	t.Run("Distance for circular orbit", func(t *testing.T) {
		pos := OrbitalElementsToPosition(circularOrbit, 0.0)
		dist := math.Sqrt(pos[0]*pos[0] + pos[1]*pos[1] + pos[2]*pos[2])

		if math.Abs(dist-1.0) > 1e-6 {
			t.Errorf("Distance from Sun = %v AU, want 1.0 AU", dist)
		}
	})

	t.Run("Earth-like orbit at J2000.0", func(t *testing.T) {
		earthElements := GetPlanetaryElements("Earth")
		pos := OrbitalElementsToPosition(earthElements, 0.0) // J2000.0

		// Earth at J2000.0 should be roughly at (1, 0, 0) with small variations
		// due to eccentricity and other elements
		dist := math.Sqrt(pos[0]*pos[0] + pos[1]*pos[1] + pos[2]*pos[2])

		// Distance should be close to 1 AU (within 0.02 AU considering eccentricity)
		if math.Abs(dist-1.0) > 0.02 {
			t.Errorf("Earth distance from Sun = %v AU, expected ~1.0 AU", dist)
		}
	})
}

func TestPositionToRenderUnits(t *testing.T) {
	tests := []struct {
		name   string
		posAU  [3]float64
		auScale float64
		want   [3]float32
	}{
		{
			name:    "Origin",
			posAU:   [3]float64{0, 0, 0},
			auScale: 150.0,
			want:    [3]float32{0, 0, 0},
		},
		{
			name:    "1 AU on X axis (astronomy Z-up: x->x, y->z, z->y)",
			posAU:   [3]float64{1, 0, 0},
			auScale: 150.0,
			want:    [3]float32{150.0, 0, 0}, // Z-up to Y-up: y and z swap
		},
		{
			name:    "1 AU on Y axis (astronomy)",
			posAU:   [3]float64{0, 1, 0},
			auScale: 150.0,
			want:    [3]float32{0, 0, 150.0}, // Y (astronomy) -> Z (raylib)
		},
		{
			name:    "1 AU on Z axis (astronomy)",
			posAU:   [3]float64{0, 0, 1},
			auScale: 150.0,
			want:    [3]float32{0, 150.0, 0}, // Z (astronomy) -> Y (raylib)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PositionToRenderUnits(tt.posAU, tt.auScale)

			for i := 0; i < 3; i++ {
				if math.Abs(float64(got[i]-tt.want[i])) > 0.01 {
					t.Errorf("PositionToRenderUnits()[%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestCalculateOrbitalVelocity(t *testing.T) {
	t.Run("Circular orbit velocity", func(t *testing.T) {
		// For circular orbit at 1 AU, v = 2π/T (AU/day)
		// T = 365.256 days, so v ≈ 0.0172 AU/day
		circularOrbit := OrbitalElements{
			SemiMajorAxis:      1.0,
			Eccentricity:       0.0,
			OrbitalPeriod:      365.256,
			Inclination:        0.0,
			LongitudeAscendingNode: 0.0,
			ArgumentOfPeriapsis: 0.0,
			MeanAnomalyAtEpoch: 0.0,
		}

		vel := CalculateOrbitalVelocity(circularOrbit, 0.0)

		// At (1, 0, 0) with circular orbit, velocity should be in +Y direction (astronomy)
		// Magnitude should be ~0.0172 AU/day
		speed := math.Sqrt(vel[0]*vel[0] + vel[1]*vel[1] + vel[2]*vel[2])
		expectedSpeed := 2 * math.Pi / 365.256

		if math.Abs(speed-expectedSpeed) > 0.001 {
			t.Errorf("Orbital speed = %v AU/day, want ~%v AU/day", speed, expectedSpeed)
		}
	})
}

func TestGetPlanetaryElements(t *testing.T) {
	t.Run("Known planet returns valid elements", func(t *testing.T) {
		earth := GetPlanetaryElements("Earth")

		if earth.SemiMajorAxis != 1.0 {
			t.Errorf("Earth SemiMajorAxis = %v, want 1.0", earth.SemiMajorAxis)
		}
		if earth.Eccentricity < 0 || earth.Eccentricity > 1 {
			t.Errorf("Earth Eccentricity = %v, should be in [0, 1)", earth.Eccentricity)
		}
		if earth.OrbitalPeriod <= 0 {
			t.Errorf("Earth OrbitalPeriod = %v, should be positive", earth.OrbitalPeriod)
		}
	})

	t.Run("Unknown planet returns default", func(t *testing.T) {
		defaultElements := GetPlanetaryElements("Pluto") // Not in our map

		if defaultElements.SemiMajorAxis != 1.0 {
			t.Errorf("Default SemiMajorAxis = %v, want 1.0", defaultElements.SemiMajorAxis)
		}
		if defaultElements.Eccentricity != 0.0 {
			t.Errorf("Default Eccentricity = %v, want 0.0", defaultElements.Eccentricity)
		}
	})
}

func TestUpdateSimulationTime(t *testing.T) {
	t.Run("Time advances when not paused", func(t *testing.T) {
		info := &TimeScaleInfo{
			SimulationDays: 100.0,
			TimeScale:     1.0, // 1 day per second
			IsPaused:      false,
		}

		deltaReal := 0.5 // 0.5 real seconds
		newTime := UpdateSimulationTime(info, deltaReal)

		if math.Abs(newTime-100.5) > 1e-6 {
			t.Errorf("SimulationDays = %v, want 100.5", newTime)
		}
	})

	t.Run("Time does not advance when paused", func(t *testing.T) {
		info := &TimeScaleInfo{
			SimulationDays: 100.0,
			TimeScale:     1.0,
			IsPaused:      true,
		}

		deltaReal := 10.0 // 10 real seconds
		newTime := UpdateSimulationTime(info, deltaReal)

		if math.Abs(newTime-100.0) > 1e-6 {
			t.Errorf("SimulationDays = %v, want 100.0 (should not advance when paused)", newTime)
		}
	})
}
