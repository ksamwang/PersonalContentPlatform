"use client";

import { PointerEvent, useEffect, useRef } from "react";

type Props = {
  locale: string;
  publicationCount: number;
  collectionCount: number;
  year: number;
};

export function HeroOrbit({ locale, publicationCount, collectionCount, year }: Props) {
  const stage = useRef<HTMLDivElement>(null);
  const frame = useRef<number | null>(null);
  const zh = locale === "zh";

  useEffect(() => () => {
    if (frame.current !== null) cancelAnimationFrame(frame.current);
  }, []);

  function updateTilt(event: PointerEvent<HTMLDivElement>) {
    const bounds = event.currentTarget.getBoundingClientRect();
    const x = (event.clientX - bounds.left) / bounds.width - 0.5;
    const y = (event.clientY - bounds.top) / bounds.height - 0.5;
    if (frame.current !== null) cancelAnimationFrame(frame.current);
    frame.current = requestAnimationFrame(() => {
      stage.current?.style.setProperty("--orbit-tilt-x", `${(-y * 18).toFixed(2)}deg`);
      stage.current?.style.setProperty("--orbit-tilt-y", `${(x * 22).toFixed(2)}deg`);
      stage.current?.style.setProperty("--orbit-shift-x", `${(x * 10).toFixed(2)}px`);
      stage.current?.style.setProperty("--orbit-shift-y", `${(y * 10).toFixed(2)}px`);
    });
  }

  function resetTilt() {
    if (frame.current !== null) cancelAnimationFrame(frame.current);
    stage.current?.style.removeProperty("--orbit-tilt-x");
    stage.current?.style.removeProperty("--orbit-tilt-y");
    stage.current?.style.removeProperty("--orbit-shift-x");
    stage.current?.style.removeProperty("--orbit-shift-y");
  }

  return (
    <div className="hero-orbit" aria-hidden="true" onPointerMove={updateTilt} onPointerLeave={resetTilt}>
      <div className="hero-orbit-stage" ref={stage}>
        <div className="hero-orbit-spinner">
          <span className="orbit-ring orbit-ring-primary"><i /></span>
          <span className="orbit-ring orbit-ring-latitude"><i /></span>
          <span className="orbit-ring orbit-ring-meridian"><i /></span>
          <span className="orbit-axis" />
        </div>
        <span className="orbit-core">
          <small>{zh ? "已发布" : "PUBLISHED"}</small>
          <strong>{String(publicationCount).padStart(2, "0")}</strong>
          <em>{year}</em>
        </span>
      </div>
      <div className="hero-orbit-caption">
        <span>FIELD NOTES / {year}</span>
        <span>{String(collectionCount).padStart(2, "0")} {zh ? "个合集" : "COLLECTIONS"}</span>
      </div>
    </div>
  );
}
