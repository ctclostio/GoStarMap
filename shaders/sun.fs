#version 330

// Fragment shader for photorealistic Sun rendering
// Inputs from vertex shader
in vec3 fragPosition;
in vec3 fragNormal;
in vec2 fragTexCoord;

// Output
out vec4 finalColor;

// Uniforms
uniform vec3 cameraPosition;
uniform float time;              // Animation time
uniform float sunRadius;         // Sun radius in world units

// Sun appearance parameters
const vec3 sunColorCore = vec3(1.0, 0.98, 0.88);      // Core color (bright yellow-white)
const vec3 sunColorLimb = vec3(1.0, 0.75, 0.45);      // Limb color (orange)
const vec3 sunspotColor = vec3(0.35, 0.25, 0.15);     // Sunspot color (dark orange-brown)
const float sunTemperature = 5778.0;                   // Surface temperature in Kelvin
const float hdrIntensity = 8.0;                        // HDR brightness multiplier

// Solar rotation (25 days = 2160000 seconds, scaled for visible animation)
const float solarRotationSpeed = 0.05;

// ============================================================================
// NOISE FUNCTIONS - GPU-side procedural texture generation
// ============================================================================

// Hash function for noise generation
vec3 hash3(vec3 p) {
    p = vec3(dot(p, vec3(127.1, 311.7, 74.7)),
             dot(p, vec3(269.5, 183.3, 246.1)),
             dot(p, vec3(113.5, 271.9, 124.6)));
    return fract(sin(p) * 43758.5453123);
}

// 3D Perlin-like noise (smooth value noise)
float noise3D(vec3 p) {
    vec3 i = floor(p);
    vec3 f = fract(p);

    // Quintic interpolation for smoother results
    vec3 u = f * f * f * (f * (f * 6.0 - 15.0) + 10.0);

    // Sample 8 corners of cube
    float a = dot(hash3(i + vec3(0.0, 0.0, 0.0)), f - vec3(0.0, 0.0, 0.0));
    float b = dot(hash3(i + vec3(1.0, 0.0, 0.0)), f - vec3(1.0, 0.0, 0.0));
    float c = dot(hash3(i + vec3(0.0, 1.0, 0.0)), f - vec3(0.0, 1.0, 0.0));
    float d = dot(hash3(i + vec3(1.0, 1.0, 0.0)), f - vec3(1.0, 1.0, 0.0));
    float e = dot(hash3(i + vec3(0.0, 0.0, 1.0)), f - vec3(0.0, 0.0, 1.0));
    float f_val = dot(hash3(i + vec3(1.0, 0.0, 1.0)), f - vec3(1.0, 0.0, 1.0));
    float g = dot(hash3(i + vec3(0.0, 1.0, 1.0)), f - vec3(0.0, 1.0, 1.0));
    float h = dot(hash3(i + vec3(1.0, 1.0, 1.0)), f - vec3(1.0, 1.0, 1.0));

    // Trilinear interpolation
    return mix(mix(mix(a, b, u.x), mix(c, d, u.x), u.y),
               mix(mix(e, f_val, u.x), mix(g, h, u.x), u.y), u.z);
}

// Fractional Brownian Motion (fBm) - multiple octaves of noise
float fbm(vec3 p, int octaves) {
    float value = 0.0;
    float amplitude = 0.5;
    float frequency = 1.0;

    for (int i = 0; i < octaves; i++) {
        value += amplitude * noise3D(p * frequency);
        frequency *= 2.0;
        amplitude *= 0.5;
    }

    return value;
}

// Turbulence - absolute value of noise for sharper features
float turbulence(vec3 p, int octaves) {
    float value = 0.0;
    float amplitude = 0.5;
    float frequency = 1.0;

    for (int i = 0; i < octaves; i++) {
        value += amplitude * abs(noise3D(p * frequency));
        frequency *= 2.0;
        amplitude *= 0.5;
    }

    return value;
}

// ============================================================================
// SOLAR SURFACE FEATURES
// ============================================================================

// Solar granulation (convection cells) - creates mottled texture
float granulation(vec3 p) {
    // High-frequency noise for small convection cells
    float cells = fbm(p * 4.0, 4);

    // Add some larger scale variation
    cells += fbm(p * 2.0, 2) * 0.3;

    // Contrast enhancement
    cells = pow(cells * 0.5 + 0.5, 1.5);

    return cells;
}

// Sunspot generation with penumbra and umbra
float sunspots(vec3 p, float rotation) {
    // Rotate coordinates for solar rotation
    float s = sin(rotation);
    float c = cos(rotation);
    vec3 rotP = vec3(
        p.x * c - p.z * s,
        p.y,
        p.x * s + p.z * c
    );

    // Large-scale noise for sunspot locations (low frequency)
    float spotNoise = fbm(rotP * 1.2, 3);

    // Create distinct spots using threshold
    float spots = smoothstep(0.55, 0.65, spotNoise);

    // Add smaller spots
    spots += smoothstep(0.7, 0.75, fbm(rotP * 2.5, 2)) * 0.5;

    // Penumbra (outer region) using turbulence
    float penumbra = turbulence(rotP * 3.0, 2) * 0.3;

    return clamp(spots + penumbra, 0.0, 1.0);
}

// ============================================================================
// MAIN SHADER
// ============================================================================

void main()
{
    // Normalize the normal vector
    vec3 N = normalize(fragNormal);

    // View direction from surface to camera
    vec3 V = normalize(cameraPosition - fragPosition);

    // ========================================================================
    // LIMB DARKENING - Center of Sun is brighter than edges
    // ========================================================================
    // Physical basis: We see deeper (hotter) layers at center,
    // only cooler upper layers at limb
    // Use abs() so both hemispheres appear equally bright (Sun is self-luminous)
    float cosTheta = abs(dot(N, V));

    // Limb darkening coefficient (tuned for realistic appearance)
    // Uses polynomial approximation to solar limb darkening law
    float limbDarkening = 1.0 - 0.6 * (1.0 - pow(cosTheta, 0.7));
    limbDarkening = clamp(limbDarkening, 0.3, 1.0);

    // ========================================================================
    // PROCEDURAL SURFACE TEXTURE
    // ========================================================================
    // Convert world position to texture coordinates on sphere
    vec3 texCoord3D = normalize(fragPosition) * 2.0;

    // Solar rotation
    float rotation = time * solarRotationSpeed;

    // 1. Solar granulation (convection cells)
    float granule = granulation(texCoord3D);
    granule = granule * 0.15 + 0.85; // Subtle variation

    // 2. Sunspots (darker regions)
    float spots = sunspots(texCoord3D, rotation);

    // Combine granulation and sunspots
    float surfaceDetail = granule * (1.0 - spots * 0.7);

    // ========================================================================
    // COLOR COMPOSITION
    // ========================================================================
    // Interpolate between core and limb color based on viewing angle
    vec3 baseColor = mix(sunColorLimb, sunColorCore, cosTheta);

    // Apply surface details
    vec3 detailColor = mix(baseColor, sunspotColor, spots * 0.6);
    detailColor *= surfaceDetail;

    // Apply limb darkening
    vec3 finalSurfaceColor = detailColor * limbDarkening;

    // ========================================================================
    // HDR EMISSION
    // ========================================================================
    // Sun emits light - use high intensity values for HDR bloom
    // Center is much brighter than edges (exponential falloff)
    float emission = hdrIntensity * pow(cosTheta, 0.5);

    // Add chromosphere glow at limb (reddish glow at edges)
    float chromosphere = pow(1.0 - cosTheta, 3.0) * 0.8;
    vec3 chromosphereColor = vec3(1.0, 0.4, 0.2) * chromosphere;

    // Final HDR color
    vec3 hdrColor = finalSurfaceColor * emission + chromosphereColor;

    // ========================================================================
    // OUTPUT
    // ========================================================================
    // Output HDR color (values > 1.0 will be used for bloom)
    finalColor = vec4(hdrColor, 1.0);

    // Optional: Add subtle animation to granulation
    // This creates a slow "boiling" effect on the surface
    float animationOffset = sin(time * 0.1 + fragPosition.x * 0.1) * 0.02;
    finalColor.rgb += vec3(animationOffset);
}
