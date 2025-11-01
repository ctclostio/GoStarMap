#version 330

// Vertex shader for Sun rendering
// Input attributes
in vec3 vertexPosition;
in vec3 vertexNormal;
in vec2 vertexTexCoord;

// Output to fragment shader
out vec3 fragPosition;
out vec3 fragNormal;
out vec2 fragTexCoord;

// Uniforms
uniform mat4 mvp;          // Model-View-Projection matrix
uniform mat4 matModel;     // Model matrix for world space position
uniform mat4 matNormal;    // Normal matrix for correct normal transformation

void main()
{
    // Transform vertex position to clip space
    gl_Position = mvp * vec4(vertexPosition, 1.0);

    // Pass world space position to fragment shader
    fragPosition = vec3(matModel * vec4(vertexPosition, 1.0));

    // Transform normal to world space (with proper handling of non-uniform scaling)
    fragNormal = normalize(vec3(matNormal * vec4(vertexNormal, 0.0)));

    // Pass texture coordinates
    fragTexCoord = vertexTexCoord;
}
