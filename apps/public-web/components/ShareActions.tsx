"use client";

import { useState } from "react";
import { Check, Download, Share2 } from "lucide-react";
import { renderSharePoster } from "./share-poster/canvas";

type Props = {
  locale: string;
  title: string;
  summary: string;
  siteName: string;
  cover?: string;
  accent: string;
  footer?: string;
  excerpt: string;
  type: string;
  publishedAt: string;
};

export function ShareActions({ locale, title, summary, siteName, cover, accent, footer, excerpt, type, publishedAt }: Props) {
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
      const canvas = await renderSharePoster({ locale, title, excerpt, siteName, cover, accent, footer, type, publishedAt, url: location.href, height: height as 1440 | 1920 });
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
        <button disabled={busy} onClick={() => void poster(1440)}><Download aria-hidden />{busy ? (zh ? "生成中…" : "Generating…") : (zh ? "朋友圈海报" : "Social poster")}</button>
        <button disabled={busy} onClick={() => void poster(1920)}><Download aria-hidden />{busy ? (zh ? "生成中…" : "Generating…") : (zh ? "竖屏海报" : "Vertical poster")}</button>
      </div>
    </aside>
  );
}
