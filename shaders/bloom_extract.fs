#version 330

// Bloom extraction shader - extracts bright regions for bloom effect
in vec2 fragTexCoord;
out vec4 finalColor;

uniform sampler2D texture0;    // Scene texture
uniform float threshold;       // Brightness threshold for bloom

void main()
{
    vec4 color = texture(texture0, fragTexCoord);

    // Calculate luminance (perceived brightness)
    float luminance = dot(color.rgb, vec3(0.2126, 0.7152, 0.0722));

    // Extract only bright pixels above threshold
    if (luminance > threshold) {
        // Preserve the bright color with HDR values
        finalColor = color;
    } else {
        // Dark pixels don't contribute to bloom
        finalColor = vec4(0.0, 0.0, 0.0, 1.0);
    }
}
