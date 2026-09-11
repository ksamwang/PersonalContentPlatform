export const vertexShader = `#version 300 es
in vec2 position;
void main() { gl_Position = vec4(position, 0.0, 1.0); }
`;

export const fragmentShader = `#version 300 es
precision highp float;

uniform vec2 resolution;
uniform vec2 pointer;
uniform float time;
uniform float pulse;
out vec4 fragColor;

float hash21(vec2 p) {
  p = fract(p * vec2(123.34, 456.21));
  p += dot(p, p + 45.32);
  return fract(p.x * p.y);
}

float noise(vec2 p) {
  vec2 i = floor(p), f = fract(p);
  f = f * f * (3.0 - 2.0 * f);
  return mix(mix(hash21(i), hash21(i + vec2(1, 0)), f.x), mix(hash21(i + vec2(0, 1)), hash21(i + 1.0), f.x), f.y);
}

float stars(vec2 uv, float scale, float threshold) {
  vec2 p = uv * scale;
  vec2 cell = floor(p);
  vec2 local = fract(p) - 0.5;
  float seed = hash21(cell);
  vec2 offset = vec2(hash21(cell + 7.1), hash21(cell + 19.7)) - 0.5;
  float star = smoothstep(0.065, 0.0, length(local - offset * 0.65));
  return star * smoothstep(threshold, 1.0, seed) * (0.55 + 0.8 * sin(seed * 31.0 + time * 1.8));
}

void main() {
  vec2 uv = (gl_FragCoord.xy * 2.0 - resolution.xy) / min(resolution.x, resolution.y);
  vec2 mouse = (pointer - 0.5) * vec2(0.22, -0.16);
  vec2 p = uv - mouse;
  float r = length(p);
  float angle = atan(p.y, p.x);

  float lens = 0.11 / max(r, 0.12);
  vec2 warped = p * (1.0 + lens * 0.34);
  warped += vec2(cos(angle * 2.0 + time * 0.055), sin(angle * 3.0 - time * 0.04)) * 0.018 / max(r, 0.2);

  vec3 color = vec3(0.0);
  float nebula = noise(warped * 2.4 + vec2(time * 0.012, 0.0));
  nebula *= noise(warped * 5.2 - vec2(0.0, time * 0.009));
  float fieldFade = 1.0 - smoothstep(0.68, 1.34, r);
  float nebulaMask = nebula * fieldFade * smoothstep(0.18, 0.62, r);
  color += vec3(0.22, 0.08, 0.16) * nebulaMask * 0.34;

  float starField = stars(warped + pointer * 0.035, 18.0, 0.93);
  starField += stars(warped * 1.17 - pointer * 0.05, 31.0, 0.975) * 0.72;
  float visibleStars = starField * smoothstep(0.20, 0.34, r) * fieldFade;
  color += vec3(0.34, 0.29, 0.36) * visibleStars * 1.25;

  float tilt = mix(3.9, 2.8, pointer.y);
  vec2 diskP = vec2(p.x, p.y * tilt + p.x * (pointer.x - 0.5) * 0.7);
  float diskR = length(diskP);
  float flow = noise(vec2(angle * 3.1 - time * 0.18, diskR * 22.0));
  float disk = exp(-abs(diskR - 0.50) * (28.0 + flow * 15.0));
  disk += exp(-abs(diskR - 0.36) * 44.0) * 0.46;
  disk *= smoothstep(0.19, 0.27, r) * smoothstep(0.91, 0.39, diskR);
  float doppler = clamp(0.62 + p.x * 0.95 + flow * 0.22, 0.18, 1.35);
  vec3 hot = mix(vec3(0.82, 0.035, 0.31), vec3(0.98, 0.49, 0.26), smoothstep(-0.45, 0.46, p.x));
  hot = mix(hot, vec3(0.89, 0.82, 0.72), smoothstep(0.82, 1.18, doppler));
  float diskLight = disk * (1.18 + doppler) * (0.62 + flow * 0.76);
  color += hot * diskLight;

  float photonRing = exp(-abs(r - 0.205) * 105.0);
  float lensArc = exp(-abs(r - 0.255) * 38.0) * (0.34 + 0.66 * pow(abs(sin(angle + time * 0.08)), 5.0));
  color += vec3(0.97, 0.18, 0.48) * photonRing * 1.45;
  color += vec3(0.40, 0.31, 0.52) * lensArc * 0.28;

  float horizon = 1.0 - smoothstep(0.165, 0.207, r);
  color = mix(color, vec3(0.0), horizon);
  float innerGlow = exp(-abs(r - 0.19) * 55.0) * smoothstep(0.15, 0.2, r);
  color += vec3(0.72, 0.04, 0.28) * innerGlow * 0.46;

  float wave = exp(-abs(r - (0.27 + pulse * 0.72)) * 55.0) * (1.0 - pulse);
  color += vec3(0.95, 0.12, 0.48) * wave * 0.72;
  color *= 1.0 - smoothstep(0.74, 1.36, length(uv));
  color = pow(max(color, 0.0), vec3(0.86));

  float light = max(color.r, max(color.g, color.b));
  float blackBody = smoothstep(0.218, 0.178, r);
  float atmosphere = (nebulaMask * 0.12 + visibleStars * 0.72 + lensArc * 0.18) * fieldFade;
  float alpha = max(blackBody, clamp(light * 1.12 + atmosphere, 0.0, 0.94));
  alpha *= 1.0 - smoothstep(0.76, 1.30, r);
  color = mix(color, vec3(0.012, 0.010, 0.014), blackBody);
  fragColor = vec4(color, clamp(alpha, 0.0, 1.0));
}
`;
