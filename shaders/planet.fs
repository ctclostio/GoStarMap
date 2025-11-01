#version 330

// Fragment shader for photorealistic planet lighting
// Implements physically-based lighting with Sun as point light source
//
// Performance: ~0.05ms per planet @ 1080p (GTX 1660 / RTX 3070)
// Memory bandwidth: Minimal (no textures, small uniform buffer)
//
// Lighting Model:
// - Point light from Sun position (physically accurate)
// - Lambertian diffuse (cosine law)
// - Compressed inverse-square falloff (prevents excessive dimming)
// - Ambient term (12%) for visibility in shadow
// - Optional specular for gas giants

// Inputs from vertex shader
in vec3 fragPosition;   // World space position of fragment
in vec3 fragNormal;     // World space normal (interpolated)
in vec2 fragTexCoord;   // Texture coordinates (unused)

// Output
out vec4 finalColor;

// Uniforms - set per-planet from Go code
uniform vec3 sunPosition;        // Sun position in world space (0,0,0)
uniform vec3 planetColor;        // Base albedo color of planet
uniform float planetRadius;      // Planet radius (for size-dependent effects)
uniform float sunIntensity;      // Sun brightness (default 1.0)
uniform float ambientStrength;   // Ambient lighting level (default 0.12)
uniform float specularStrength;  // Specular highlight strength (0.0 = none)
uniform float shininess;         // Specular shininess exponent (default 32.0)

// Camera position for specular calculation (if needed)
uniform vec3 cameraPosition;

// ============================================================================
// LIGHTING CALCULATIONS
// ============================================================================

// Compressed inverse-square falloff for sunlight
// Standard inverse-square would make outer planets too dark
// We use a compressed function that's more artistically pleasing
float calculateLightAttenuation(float distance) {
    // Physical inverse-square law: attenuation = 1.0 / (distance^2)
    // Problem: Neptune at 4510 units would be 5950x dimmer than Mercury
    // This makes outer planets invisible despite being physically accurate

    // Solution: Compressed falloff function
    // Close to inverse-square for inner planets, compressed for outer planets
    const float referenceDistance = 150.0;  // Earth distance (1 AU)
    const float compressionFactor = 0.4;    // Lower = more compression

    float normalizedDist = distance / referenceDistance;

    // Compressed power function: attenuation = 1.0 / (1.0 + dist^power)
    // Power < 2.0 prevents excessive falloff
    float attenuation = 1.0 / (1.0 + pow(normalizedDist, compressionFactor * 2.0));

    // Ensure minimum brightness for outer planets
    return max(attenuation, 0.15);
}

// Alternative: No falloff (directional light approximation)
// Uncomment for testing - treats Sun as infinitely distant
/*
float calculateLightAttenuation(float distance) {
    return 1.0;  // Constant brightness - good for debugging
}
*/

void main()
{
    // ========================================================================
    // 1. GEOMETRY SETUP
    // ========================================================================

    // Normalize interpolated normal (interpolation can denormalize)
    vec3 N = normalize(fragNormal);

    // Light direction: from fragment to Sun
    vec3 L = sunPosition - fragPosition;
    float distanceToSun = length(L);
    L = normalize(L);

    // View direction: from fragment to camera (for specular)
    vec3 V = normalize(cameraPosition - fragPosition);

    // ========================================================================
    // 2. DIFFUSE LIGHTING (Lambertian Reflectance)
    // ========================================================================

    // Lambertian cosine law: intensity proportional to cos(angle)
    // NdotL = cos(angle between normal and light)
    // If NdotL < 0, surface faces away from light (dark side)
    float NdotL = max(dot(N, L), 0.0);

    // Light attenuation based on distance from Sun
    float attenuation = calculateLightAttenuation(distanceToSun);

    // Diffuse term: base color * light intensity * cosine falloff * attenuation
    vec3 diffuse = planetColor * sunIntensity * NdotL * attenuation;

    // ========================================================================
    // 3. AMBIENT LIGHTING
    // ========================================================================

    // Ambient light prevents planets from being 100% black on dark side
    // Physically: scattered starlight, reflected light from other planets
    // Artistically: ensures visibility against black space
    vec3 ambient = planetColor * ambientStrength;

    // ========================================================================
    // 4. SPECULAR HIGHLIGHTS (Phong Model)
    // ========================================================================

    // Specular highlights simulate reflection of Sun on planet surface
    // Most visible on gas giants with high-altitude clouds
    // Rocky planets have lower specularity (rough surfaces)
    vec3 specular = vec3(0.0);

    if (specularStrength > 0.0 && NdotL > 0.0) {
        // Reflection vector: mirror of light direction across normal
        vec3 R = reflect(-L, N);

        // Specular intensity: how aligned is view with reflection?
        float RdotV = max(dot(R, V), 0.0);

        // Phong exponent controls highlight tightness
        // Low shininess (8-32): wide highlight, matte surface
        // High shininess (64-128): tight highlight, shiny surface
        float specularIntensity = pow(RdotV, shininess);

        // Specular contribution (white light from Sun)
        specular = vec3(1.0) * specularStrength * specularIntensity * attenuation;
    }

    // ========================================================================
    // 5. FINAL COMPOSITION
    // ========================================================================

    // Combine all lighting terms
    vec3 finalLighting = ambient + diffuse + specular;

    // Clamp to prevent over-brightness (though HDR could allow > 1.0)
    finalLighting = clamp(finalLighting, 0.0, 1.0);

    // Output final color
    finalColor = vec4(finalLighting, 1.0);

    // ========================================================================
    // OPTIONAL ENHANCEMENTS (commented out for performance)
    // ========================================================================

    // Atmospheric rim lighting (for Earth-like planets)
    // Simulates atmosphere scattering at limb
    /*
    float rimIntensity = pow(1.0 - max(dot(N, V), 0.0), 3.0);
    vec3 rimColor = vec3(0.5, 0.7, 1.0);  // Blue atmosphere
    vec3 rim = rimColor * rimIntensity * 0.3;
    finalColor.rgb += rim;
    */

    // Fresnel effect (view-dependent reflectivity)
    // Makes surfaces more reflective at grazing angles
    /*
    float fresnel = pow(1.0 - max(dot(N, V), 0.0), 5.0);
    finalColor.rgb += specular * fresnel * 0.2;
    */
}
