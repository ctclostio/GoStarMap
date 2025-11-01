#version 330

// ============================================================================
// PLANET VERTEX SHADER WITH ROTATION
// ============================================================================
// Handles planet rotation and axial tilt in GPU for optimal performance
//
// Performance: GPU-native transformation, no additional CPU cost
// Memory: 8 bytes per planet per frame (rotationAngle + axialTilt uniforms)
//
// Rotation Strategy:
// 1. Apply axial tilt rotation (orient rotation axis)
// 2. Apply rotation around tilted Y-axis
// 3. Transform to world space for lighting
//
// This handles:
// - Normal rotation (positive period)
// - Retrograde rotation (negative period)
// - Axial tilt (Uranus rotating on its side at 97.77°)
// ============================================================================

// Input attributes from vertex buffer
in vec3 vertexPosition;  // Local space position
in vec3 vertexNormal;    // Local space normal
in vec2 vertexTexCoord;  // Texture coordinates (unused for now)

// Output to fragment shader
out vec3 fragPosition;   // World space position
out vec3 fragNormal;     // World space normal
out vec2 fragTexCoord;   // Pass through for future texturing

// Uniforms provided by raylib
uniform mat4 mvp;        // Model-View-Projection matrix (for clip space)
uniform mat4 matModel;   // Model matrix (local -> world space)
uniform mat4 matNormal;  // Normal matrix (handles non-uniform scaling)

// Rotation uniforms
uniform float rotationAngle;  // Current rotation angle in radians
uniform float axialTilt;      // Axial tilt in radians (obliquity)

// ============================================================================
// ROTATION MATRIX CONSTRUCTION
// ============================================================================

// Build rotation matrix around Y-axis (standard planet rotation)
// This is more efficient than using cos/sin in multiple places
mat3 rotateY(float angle) {
    float c = cos(angle);
    float s = sin(angle);
    return mat3(
        c,  0.0,  s,
        0.0, 1.0, 0.0,
        -s, 0.0,  c
    );
}

// Build rotation matrix around X-axis (for axial tilt)
// Tilts the planet's rotation axis
mat3 rotateX(float angle) {
    float c = cos(angle);
    float s = sin(angle);
    return mat3(
        1.0, 0.0, 0.0,
        0.0,  c,  -s,
        0.0,  s,   c
    );
}

// ============================================================================
// MAIN VERTEX TRANSFORMATION
// ============================================================================

void main()
{
    // ========================================================================
    // 1. APPLY ROTATION IN LOCAL SPACE
    // ========================================================================
    // We rotate vertices BEFORE world transformation to handle planet spin
    // Order: tilt axis -> rotate around tilted axis

    // Build axial tilt rotation matrix
    // This tilts the planet's rotation axis (e.g., Uranus tilted 97.77°)
    mat3 tiltMatrix = rotateX(axialTilt);

    // Build rotation matrix around tilted Y-axis
    // This is the actual planet spin
    mat3 spinMatrix = rotateY(rotationAngle);

    // Combine rotations: first tilt, then spin
    // Matrix multiplication order: spin * tilt
    // (Right-to-left: tilt is applied first, then spin around tilted axis)
    mat3 rotationMatrix = spinMatrix * tiltMatrix;

    // Apply rotation to vertex position
    vec3 rotatedPosition = rotationMatrix * vertexPosition;

    // Apply rotation to normal for correct lighting
    // Normals must rotate with the surface to maintain lighting accuracy
    vec3 rotatedNormal = rotationMatrix * vertexNormal;

    // ========================================================================
    // 2. TRANSFORM TO WORLD SPACE
    // ========================================================================

    // Transform rotated vertex to clip space for rasterization
    gl_Position = mvp * vec4(rotatedPosition, 1.0);

    // Transform rotated position to world space for lighting calculations
    // Fragment shader needs this to calculate light direction
    fragPosition = vec3(matModel * vec4(rotatedPosition, 1.0));

    // Transform rotated normal to world space
    // matNormal handles non-uniform scaling correctly
    // Must normalize in case of scaling
    fragNormal = normalize(vec3(matNormal * vec4(rotatedNormal, 0.0)));

    // ========================================================================
    // 3. PASS THROUGH TEXTURE COORDINATES
    // ========================================================================
    // Texture coordinates also need rotation for surface features to spin
    // For now, we pass them through (no textures yet)
    // Future: rotate UV coordinates for texture mapping
    fragTexCoord = vertexTexCoord;
}

// ============================================================================
// PERFORMANCE NOTES
// ============================================================================
// Cost per vertex: 2 matrix constructions + 2 matrix-vector multiplications
// - rotateX: 1 sin, 1 cos, 9 multiplications
// - rotateY: 1 sin, 1 cos, 9 multiplications
// - Matrix multiply: 18 multiplications, 12 additions
// - Total: ~40 operations per vertex
//
// For 32x32 sphere (2048 vertices):
// - ~82k operations total
// - On modern GPU: <0.001ms (trivial cost)
// - Memory bandwidth: 8 bytes (2 floats) per planet
//
// Alternative (CPU matrix): Would require 64 bytes per planet + CPU cycles
// GPU approach is 8x more bandwidth efficient and uses parallel hardware
// ============================================================================
