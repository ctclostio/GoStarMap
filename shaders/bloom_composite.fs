#version 330

// Final bloom composite - combines scene with blurred bloom
in vec2 fragTexCoord;
out vec4 finalColor;

uniform sampler2D sceneTexture;   // Original scene
uniform sampler2D bloomTexture;   // Blurred bloom
uniform float bloomIntensity;     // Bloom strength (0.0 - 1.0)
uniform float exposure;           // HDR exposure/tone mapping

// ACES Filmic Tone Mapping
// Provides natural, film-like response to HDR values
vec3 ACESFilm(vec3 x) {
    float a = 2.51;
    float b = 0.03;
    float c = 2.43;
    float d = 0.59;
    float e = 0.14;
    return clamp((x * (a * x + b)) / (x * (c * x + d) + e), 0.0, 1.0);
}

void main()
{
    // Sample scene and bloom
    vec3 sceneColor = texture(sceneTexture, fragTexCoord).rgb;
    vec3 bloomColor = texture(bloomTexture, fragTexCoord).rgb;

    // Combine scene with bloom (additive blending)
    vec3 hdrColor = sceneColor + bloomColor * bloomIntensity;

    // Apply exposure
    hdrColor *= exposure;

    // Tone mapping to LDR (Low Dynamic Range for display)
    vec3 ldrColor = ACESFilm(hdrColor);

    // Gamma correction (convert from linear to sRGB)
    ldrColor = pow(ldrColor, vec3(1.0 / 2.2));

    finalColor = vec4(ldrColor, 1.0);
}
