#version 330

// Gaussian blur shader - separable blur for efficient bloom
in vec2 fragTexCoord;
out vec4 finalColor;

uniform sampler2D texture0;
uniform vec2 resolution;       // Screen resolution
uniform bool horizontal;       // Horizontal or vertical pass

// 9-tap Gaussian blur weights (optimized for performance/quality balance)
// Weights sum to 1.0
const float weights[5] = float[](0.227027, 0.1945946, 0.1216216, 0.054054, 0.016216);

void main()
{
    // Calculate texel size
    vec2 texelSize = 1.0 / resolution;

    // Current pixel contribution
    vec3 result = texture(texture0, fragTexCoord).rgb * weights[0];

    // Determine blur direction
    vec2 direction = horizontal ? vec2(texelSize.x, 0.0) : vec2(0.0, texelSize.y);

    // Sample surrounding pixels with Gaussian weights
    for(int i = 1; i < 5; i++) {
        // Sample both directions from center
        vec2 offset = direction * float(i);

        result += texture(texture0, fragTexCoord + offset).rgb * weights[i];
        result += texture(texture0, fragTexCoord - offset).rgb * weights[i];
    }

    finalColor = vec4(result, 1.0);
}
