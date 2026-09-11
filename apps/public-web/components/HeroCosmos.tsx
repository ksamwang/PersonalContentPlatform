"use client";

import { PointerEvent, useEffect, useRef, useState } from "react";
import { createBlackHoleRenderer } from "./hero-cosmos/webgl";

type Props = { locale: string; publicationCount: number; collectionCount: number; year: number };

export function HeroCosmos({ locale, publicationCount, collectionCount, year }: Props) {
  const canvas = useRef<HTMLCanvasElement>(null);
  const target = useRef({ x: 0.5, y: 0.5 });
  const current = useRef({ x: 0.5, y: 0.5 });
  const pulseStarted = useRef(-10);
  const [supported, setSupported] = useState(true);
  const zh = locale === "zh";

  useEffect(() => {
    if (!canvas.current) return;
    let renderer: ReturnType<typeof createBlackHoleRenderer>;
    try { renderer = createBlackHoleRenderer(canvas.current); } catch { renderer = null; }
    if (!renderer) { setSupported(false); return; }
    const reduced = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
    let frame = 0;
    let visible = true;
    const observer = new IntersectionObserver(([entry]) => { visible = entry.isIntersecting; }, { threshold: 0.05 });
    observer.observe(canvas.current);
    const started = performance.now();
    const draw = (now: number) => {
      if (visible) {
        current.current.x += (target.current.x - current.current.x) * 0.055;
        current.current.y += (target.current.y - current.current.y) * 0.055;
        const seconds = reduced ? 4.2 : (now - started) / 1000;
        const pulse = Math.min(1, Math.max(0, (now - pulseStarted.current) / 1250));
        renderer?.render(seconds, current.current.x, current.current.y, pulse);
      }
      frame = requestAnimationFrame(draw);
    };
    frame = requestAnimationFrame(draw);
    return () => { cancelAnimationFrame(frame); observer.disconnect(); renderer?.destroy(); };
  }, []);

  function point(event: PointerEvent<HTMLDivElement>) {
    const bounds = event.currentTarget.getBoundingClientRect();
    target.current = {
      x: Math.min(1, Math.max(0, (event.clientX - bounds.left) / bounds.width)),
      y: Math.min(1, Math.max(0, (event.clientY - bounds.top) / bounds.height)),
    };
  }

  return (
    <div
      className={`hero-cosmos${supported ? "" : " is-fallback"}`}
      onPointerMove={point}
      onPointerLeave={() => { target.current = { x: 0.5, y: 0.5 }; }}
      onPointerDown={(event) => { point(event); pulseStarted.current = performance.now(); }}
      aria-label={zh ? `真相观测站：${publicationCount} 篇已发布内容，移动鼠标探索黑洞` : `Truth observatory: ${publicationCount} published notes. Move the pointer to explore the black hole.`}
      role="img"
    >
      <canvas ref={canvas} />
      <span className="cosmos-orbit orbit-one" aria-hidden />
      <span className="cosmos-orbit orbit-two" aria-hidden />
      <div className="cosmos-hud" aria-hidden>
        <div className="cosmos-heading"><span>TRUTH OBSERVATORY</span><b>01</b></div>
        <p className="cosmos-instruction">{zh ? "移动光标改变观测角度 / 点击释放引力波" : "MOVE TO SHIFT THE VIEW / PRESS TO RELEASE A GRAVITY WAVE"}</p>
        <div className="cosmos-data"><span>{String(publicationCount).padStart(2, "0")} NOTES</span><span>{String(collectionCount).padStart(2, "0")} ORBITS</span><span>{year}</span></div>
      </div>
      <span className="cosmos-crosshair" aria-hidden />
      <span className="cosmos-index" aria-hidden>BH–01 / FIELD NOTE</span>
      <span className="cosmos-truth" aria-hidden>{zh ? "真相存在于事件视界之外" : "TRUTH LIVES BEYOND THE EVENT HORIZON"}</span>
    </div>
  );
}
