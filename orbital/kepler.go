package orbital

import "math"

// SolveKeplerEquation solves Kepler's equation: M = E - e*sin(E)
// Uses Newton-Raphson iteration to find eccentric anomaly E from mean anomaly M
// Expected precision: < 10^-12 radians after ~5-8 iterations
func SolveKeplerEquation(meanAnomaly, eccentricity float64) float64 {
	const tolerance = 1e-12 // Precision threshold (radians)
	const maxIterations = 20

	// Normalize mean anomaly to [0, 2π]
	M := math.Mod(meanAnomaly, 2*math.Pi)
	if M < 0 {
		M += 2 * math.Pi
	}

	// Initial guess for eccentric anomaly
	// For low eccentricity: E ≈ M
	// For high eccentricity: better initial guess improves convergence
	var E float64
	if eccentricity < 0.8 {
		E = M // Good enough for most planets
	} else {
		// Better initial guess for high eccentricity
		E = math.Pi
	}

	// Newton-Raphson iteration: E_n+1 = E_n - f(E_n)/f'(E_n)
	// where f(E) = E - e*sin(E) - M
	// and f'(E) = 1 - e*cos(E)
	for i := 0; i < maxIterations; i++ {
		denominator := 1.0 - eccentricity*math.Cos(E)

		// Safety check for near-singular case (parabolic/near-parabolic orbits)
		if math.Abs(denominator) < 1e-10 {
			// Denominator too small, use current E as best estimate
			break
		}

		deltaE := (E - eccentricity*math.Sin(E) - M) / denominator
		E -= deltaE

		// Check convergence
		if math.Abs(deltaE) < tolerance {
			return E
		}
	}

	// If we reach here, convergence failed (extremely rare)
	// Return best estimate
	return E
}

// CalculateTrueAnomaly computes true anomaly from eccentric anomaly
// True anomaly ν is the actual angle from periapsis to the body
func CalculateTrueAnomaly(eccentricAnomaly, eccentricity float64) float64 {
	// Formula: tan(ν/2) = √[(1+e)/(1-e)] * tan(E/2)
	// More stable formula: ν = 2 * atan2(√(1+e)*sin(E/2), √(1-e)*cos(E/2))
	sqrtFactor1 := math.Sqrt(1.0 + eccentricity)
	sqrtFactor2 := math.Sqrt(1.0 - eccentricity)

	trueAnomaly := 2.0 * math.Atan2(
		sqrtFactor1*math.Sin(eccentricAnomaly/2.0),
		sqrtFactor2*math.Cos(eccentricAnomaly/2.0),
	)

	return trueAnomaly
}

// OrbitalElementsToPosition converts Keplerian elements to Cartesian position
// Returns position in AU (heliocentric ecliptic coordinates)
// Time is in days since J2000.0 epoch (January 1, 2000, 12:00 TT)
func OrbitalElementsToPosition(elements OrbitalElements, timeDays float64) [3]float64 {
	// 1. Calculate mean anomaly at current time
	// M(t) = M₀ + n*t, where n = 2π/T (mean motion)
	meanMotion := (2.0 * math.Pi) / elements.OrbitalPeriod
	meanAnomaly := elements.MeanAnomalyAtEpoch + meanMotion*timeDays

	// 2. Solve Kepler's equation for eccentric anomaly E
	eccentricAnomaly := SolveKeplerEquation(meanAnomaly, elements.Eccentricity)

	// 3. Calculate true anomaly ν
	trueAnomaly := CalculateTrueAnomaly(eccentricAnomaly, elements.Eccentricity)

	// 4. Calculate distance r from Sun
	// r = a(1 - e²) / (1 + e*cos(ν))
	r := elements.SemiMajorAxis * (1.0 - elements.Eccentricity*elements.Eccentricity) /
		(1.0 + elements.Eccentricity*math.Cos(trueAnomaly))

	// 5. Calculate position in orbital plane (x', y', z'=0)
	// x' = r * cos(ν + ω)
	// y' = r * sin(ν + ω)
	angleInOrbit := trueAnomaly + elements.ArgumentOfPeriapsis

	// 6. Rotate from orbital plane to ecliptic plane
	// This involves three rotations:
	//   - Rotate by ω (argument of periapsis) - already done above
	//   - Rotate by i (inclination) around x-axis
	//   - Rotate by Ω (longitude of ascending node) around z-axis

	cosI := math.Cos(elements.Inclination)
	sinI := math.Sin(elements.Inclination)
	cosO := math.Cos(elements.LongitudeAscendingNode)
	sinO := math.Sin(elements.LongitudeAscendingNode)

	// Compute ecliptic coordinates
	// This is the combined rotation matrix application
	x := (cosO*math.Cos(angleInOrbit) - sinO*math.Sin(angleInOrbit)*cosI) * r
	y := (sinO*math.Cos(angleInOrbit) + cosO*math.Sin(angleInOrbit)*cosI) * r
	z := math.Sin(angleInOrbit) * sinI * r

	return [3]float64{x, y, z}
}

// CalculateOrbitalVelocity computes velocity vector for a given position
// This is useful for physics simulations, though not required for simple visualization
func CalculateOrbitalVelocity(elements OrbitalElements, timeDays float64) [3]float64 {
	// Using vis-viva equation: v² = μ(2/r - 1/a)
	// where μ = GM_sun (standard gravitational parameter)
	// For simplicity, we'll compute this numerically using position derivative

	// Small time step for numerical derivative (0.01 days = 14.4 minutes)
	dt := 0.01

	// Get positions at t and t+dt
	pos1 := OrbitalElementsToPosition(elements, timeDays)
	pos2 := OrbitalElementsToPosition(elements, timeDays+dt)

	// Numerical derivative: v ≈ (pos2 - pos1) / dt
	vx := (pos2[0] - pos1[0]) / dt
	vy := (pos2[1] - pos1[1]) / dt
	vz := (pos2[2] - pos1[2]) / dt

	return [3]float64{vx, vy, vz}
}
