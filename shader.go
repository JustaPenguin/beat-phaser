package main

import "github.com/go-gl/mathgl/mgl32"

var iTime float32
var iMouse mgl32.Vec4
var iLightPos mgl32.Vec2

// NOTE: I have added #version and all "uniforms". I also renamed the main() function
// and moved it's imports to "out" and "in" package variables

//@TODO I think utexture needs to be brought in to the output render (see second shader for example)
var fragmentShaderLighting = `
#version 330 core

out vec4 fragColor;

// Steps is pretty much shadow smoothness
#define STEPS 16
//#define LIGHT_SPEED 1.5
#define LIGHT_SPEED 0.5
#define HARDNESS 2.0

uniform sampler2D uTexture;
uniform vec4 uTexBounds;

// custom uniforms
uniform float iTime;
uniform vec4 iMouse;
uniform vec2 iLightPos;

struct ray {
	vec2 t;
	vec2 p;
	vec2 d;
};

float scene (in vec2 p) {
	// p is position of rotating square
	p -= vec2 (70,20);
	
	// probably something to do with rotation
    // used in conjunction sin/cos (iTime / x) changing x changes square rotation speed
	// changing one x makes the square change size on a sin/cos curve
	float sn = sin (iTime / 2.0);
	float cs = cos (iTime / 2.0);
	
	float f1 = dot (vec2 (sn,-cs), p);
	float f2 = dot (vec2 (cs, sn), p);
	float f = max (f1*f1, f2*f2);

    // Make an extra circle, needs to be placed variably @TODO many broken pls fix
	//p += vec2 (70, 20);
	//f = min (f, dot (p, p) * 2400.0);
	
	// this p sets circle position, equation makes it follow a circular arc
	p += vec2 (sn + 0.5, cs) * 200.0;
	// increase multiplier to decrease circle size
	f = min (f, dot (p, p) * 2.0);
	
	// something to do with square / circle sizes / (f + x)
	return 1000.0 / (f + 1000.0);
}

ray newRay (in vec2 origin, in vec2 target) {
	ray r;
	
	r.t = target;
	r.p = origin;
	r.d = (target - origin) / float (STEPS);
	
	return r;
}

void rayMarch (inout ray r) {
	r.p += r.d * clamp (HARDNESS - scene (r.p) * HARDNESS * 2.0, 0.0, LIGHT_SPEED);
}

vec3 light (in ray r, in vec3 color) {
	return color / (dot (r.p, r.p) + color);
}

void main() {
	// window position / size from origin
	vec2 uv = (gl_FragCoord.xy-225) / uTexBounds.zw * vec2 (1, 1);
	
	// light positions
	ray r0 = newRay (uv, vec2 (200, 250));
	ray r1 = newRay (uv, vec2 (0, -250));
	ray r2 = newRay (uv, iLightPos);
	
	for (int i = 0; i < STEPS; i++) {
		rayMarch (r0);
		rayMarch (r1);
		rayMarch (r2);
	}
	
	r0.p -= r0.t;
	r1.p -= r1.t;
	r2.p -= r2.t;
	
	// (r0, vec3(r, g, b) * intensity)
	vec3 light1 = light (r0, vec3 (0.34, 0.06, 0.55) * 2000.0);
	vec3 light2 = light (r1, vec3 (0.1, 0.2, 0.3) * 1500.0);
	vec3 light3 = light (r2, vec3 (0.1, 0.1, 0.1) * 0.0001);
	
    // affects light intensity on multiple objects
	// final number is most useful, affects square/circle roughness
	float f = clamp (scene (uv) * 200.0 - 100.0, 0.0, 4.0);
	
	vec3 col = texture(uTexture, uv).rgb;
	fragColor = vec4 ((col + light1 + light2 + light3) * (1.0 + f), 1.0);
	//fragColor = vec4 (scene (uv));
}
`

var fragmentShaderDrunkered = `
#version 330 core

out vec4 fragColor;

uniform sampler2D uTexture;
uniform vec4 uTexBounds;

// custom uniforms
uniform float uSpeed;
uniform float uTime;

void main() {
    vec2 t = gl_FragCoord.xy / uTexBounds.zw;
	vec3 influence = texture(uTexture, t).rgb;

    if (influence.r + influence.g + influence.b > 0.3) {
		t.y += cos(t.x * 40.0 + (uTime * uSpeed))*0.005;
		t.x += cos(t.y * 40.0 + (uTime * uSpeed))*0.01;
	}

    vec3 col = texture(uTexture, t).rgb;
	fragColor = vec4(col * vec3(0.6, 0.6, 1.2),1.0);
}
`

/* Drunkered requires :

var uTime, uSpeed float32

win.Canvas().SetUniform("uTime", &uTime)
win.Canvas().SetUniform("uSpeed", &uSpeed)
uSpeed = 5.0

 */

var fragmentShaderGreyScale = `
#version 330 core

in vec2  vTexCoords;

out vec4 fragColor;

uniform vec4 uTexBounds;
uniform sampler2D uTexture;

void main() {
	// Get our current screen coordinate
	vec2 t = (vTexCoords - uTexBounds.xy) / uTexBounds.zw;

	// Sum our 3 color channels
	float sum  = texture(uTexture, t).r;
	      sum += texture(uTexture, t).g;
	      sum += texture(uTexture, t).b;

	// Divide by 3, and set the output to the result
	vec4 color = vec4( sum/3, sum/3, sum/3, 1.0);
	fragColor = color;
}
`
