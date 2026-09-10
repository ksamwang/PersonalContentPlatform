"use client";

import { useState } from "react";
import { Check, Download, Share2 } from "lucide-react";
import QRCode from "qrcode";

type Props = {
  locale: string;
  title: string;
  summary: string;
  siteName: string;
  cover?: string;
  accent: string;
  footer?: string;
};

async function loadImage(url: string) {
  return new Promise<HTMLImageElement>((resolve, reject) => {
    const value = new Image();
    value.crossOrigin = "anonymous";
    value.onload = () => resolve(value);
    value.onerror = reject;
    value.src = url;
  });
}

function wrap(ctx: CanvasRenderingContext2D, text: string, maxWidth: number, maxLines: number) {
  const lines: string[] = [];
  let line = "";
  for (const character of [...text]) {
    if (ctx.measureText(line + character).width > maxWidth && line) {
      lines.push(line);
      line = character;
      if (lines.length === maxLines - 1) break;
    } else {
      line += character;
    }
  }
  if (line) lines.push(line);
  return lines.slice(0, maxLines);
}

export function ShareActions({ locale, title, summary, siteName, cover, accent, footer }: Props) {
  const [copied, setCopied] = useState(false);
  const [busy, setBusy] = useState(false);
  const zh = locale === "zh";

  async function share() {
    const data = { title, text: summary, url: location.href };
    if (navigator.share) {
      await navigator.share(data).catch(() => {});
      return;
    }
    await navigator.clipboard.writeText(location.href);
    setCopied(true);
    setTimeout(() => setCopied(false), 2200);
  }

  async function poster(height: number) {
    setBusy(true);
    try {
      const canvas = document.createElement("canvas");
      canvas.width = 1080;
      canvas.height = height;
      const ctx = canvas.getContext("2d")!;
      ctx.fillStyle = "#f4f0e8";
      ctx.fillRect(0, 0, 1080, height);
      ctx.fillStyle = accent;
      ctx.fillRect(0, 0, 18, height);
      let y = 110;
      if (cover) {
        try {
          const coverImage = await loadImage(cover);
          const targetHeight = height === 1920 ? 690 : 460;
          const scale = Math.max(980 / coverImage.width, targetHeight / coverImage.height);
          const width = coverImage.width * scale;
          const imageHeight = coverImage.height * scale;
          ctx.save();
          ctx.beginPath();
          ctx.roundRect(50, y, 980, targetHeight, 24);
          ctx.clip();
          ctx.drawImage(coverImage, 50 + (980 - width) / 2, y + (targetHeight - imageHeight) / 2, width, imageHeight);
          ctx.restore();
          y += targetHeight + 82;
        } catch {}
      }
      ctx.fillStyle = accent;
      ctx.font = "700 28px Arial";
      ctx.fillText(siteName.toUpperCase(), 60, y);
      y += 78;
      ctx.fillStyle = "#171416";
      ctx.font = "600 64px Georgia";
      for (const line of wrap(ctx, title, 950, 4)) {
        ctx.fillText(line, 60, y);
        y += 82;
      }
      ctx.fillStyle = "#686168";
      ctx.font = "32px Georgia";
      y += 18;
      for (const line of wrap(ctx, summary, 930, 4)) {
        ctx.fillText(line, 60, y);
        y += 50;
      }
      const qr = await loadImage(await QRCode.toDataURL(location.href, { margin: 1, width: 220, color: { dark: "#171416", light: "#f4f0e8" } }));
      ctx.drawImage(qr, 60, height - 290, 220, 220);
      ctx.fillStyle = "#171416";
      ctx.font = "600 30px Arial";
      ctx.fillText(footer || (zh ? "扫码阅读完整内容" : "Scan to read the full story"), 320, height - 190);
      ctx.fillStyle = "#686168";
      ctx.font = "24px Arial";
      ctx.fillText(new URL(location.href).host, 320, height - 145);
      const link = document.createElement("a");
      link.download = `${title.slice(0, 24)}-${height === 1920 ? "vertical" : "social"}.png`;
      link.href = canvas.toDataURL("image/png");
      link.click();
    } finally {
      setBusy(false);
    }
  }

  return (
    <aside className="share-card" aria-label={zh ? "分享文章" : "Share this article"}>
      <div>
        <span className="kicker">SHARE</span>
        <strong>{zh ? "把这篇内容带到别处" : "Take this story with you"}</strong>
      </div>
      <div className="share-actions">
        <button onClick={() => void share()}>
          {copied ? <Check aria-hidden /> : <Share2 aria-hidden />}
          {copied ? (zh ? "链接已复制" : "Link copied") : (zh ? "分享链接" : "Share link")}
        </button>
        <button disabled={busy} onClick={() => void poster(1440)}><Download aria-hidden />{zh ? "朋友圈海报" : "Social poster"}</button>
        <button disabled={busy} onClick={() => void poster(1920)}><Download aria-hidden />{zh ? "竖屏海报" : "Vertical poster"}</button>
      </div>
    </aside>
  );
}
