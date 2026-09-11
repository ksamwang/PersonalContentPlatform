import QRCode from "qrcode";

export type PosterOptions = {
  locale: string;
  title: string;
  excerpt: string;
  siteName: string;
  type: string;
  publishedAt: string;
  url: string;
  height: 1440 | 1920;
  accent: string;
  cover?: string;
  footer?: string;
};

const WIDTH = 1080;
const PAPER = "#f4f0e8";
const INK = "#171416";
const MUTED = "#686168";

function loadImage(url: string) {
  return new Promise<HTMLImageElement>((resolve, reject) => {
    const image = new Image();
    image.crossOrigin = "anonymous";
    image.onload = () => resolve(image);
    image.onerror = reject;
    image.src = url;
  });
}

function wrap(ctx: CanvasRenderingContext2D, text: string, maxWidth: number, maxLines: number) {
  const lines: string[] = [];
  let line = "";
  let truncated = false;
  const characters = [...text.replace(/\s+/g, " ").trim()];
  for (const character of characters) {
    if (ctx.measureText(line + character).width > maxWidth && line) {
      lines.push(line.trimEnd());
      line = character.trimStart();
      if (lines.length === maxLines) {
        truncated = true;
        break;
      }
    } else {
      line += character;
    }
  }
  if (!truncated && line && lines.length < maxLines) lines.push(line.trimEnd());
  if (truncated && lines.length) {
    let last = lines.length - 1;
    while (ctx.measureText(`${lines[last]}…`).width > maxWidth && lines[last]) {
      lines[last] = lines[last].slice(0, -1);
    }
    lines[last] += "…";
  }
  return lines;
}

function drawLines(ctx: CanvasRenderingContext2D, lines: string[], x: number, y: number, lineHeight: number) {
  for (const line of lines) {
    ctx.fillText(line, x, y);
    y += lineHeight;
  }
  return y;
}

function drawCover(ctx: CanvasRenderingContext2D, image: HTMLImageElement, y: number, height: number) {
  const x = 72;
  const width = WIDTH - 144;
  const scale = Math.max(width / image.width, height / image.height);
  const imageWidth = image.width * scale;
  const imageHeight = image.height * scale;
  ctx.save();
  ctx.beginPath();
  ctx.rect(x, y, width, height);
  ctx.clip();
  ctx.drawImage(image, x + (width - imageWidth) / 2, y + (height - imageHeight) / 2, imageWidth, imageHeight);
  ctx.restore();
}

function formatPublished(locale: string, value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  return new Intl.DateTimeFormat(locale === "zh" ? "zh-CN" : "en", {
    year: "numeric",
    month: "long",
    day: "numeric",
  }).format(date);
}

export async function renderSharePoster(options: PosterOptions) {
  const canvas = document.createElement("canvas");
  canvas.width = WIDTH;
  canvas.height = options.height;
  const ctx = canvas.getContext("2d");
  if (!ctx) throw new Error("Canvas is unavailable");

  const zh = options.locale === "zh";
  ctx.fillStyle = PAPER;
  ctx.fillRect(0, 0, WIDTH, options.height);
  const footerTop = options.height - 300;
  ctx.fillStyle = options.accent;
  ctx.fillRect(0, 0, 20, options.height);
  ctx.fillRect(WIDTH - 168, 0, 168, 18);

  ctx.fillStyle = options.accent;
  ctx.font = "700 25px Arial, sans-serif";
  ctx.fillText(options.siteName.toUpperCase(), 72, 82);
  ctx.textAlign = "right";
  ctx.fillStyle = MUTED;
  ctx.font = "600 20px Arial, sans-serif";
  ctx.fillText("FIELD NOTES / 01", WIDTH - 72, 82);
  ctx.textAlign = "left";

  let y = 142;
  if (options.cover) {
    try {
      const image = await loadImage(options.cover);
      const coverHeight = options.height === 1920 ? 430 : 260;
      drawCover(ctx, image, y, coverHeight);
      ctx.fillStyle = options.accent;
      ctx.fillRect(72, y + coverHeight - 12, 260, 12);
      y += coverHeight + 62;
    } catch {
      y += 18;
    }
  } else {
    ctx.fillStyle = options.accent;
    ctx.fillRect(WIDTH - 254, y, 182, 164);
    ctx.fillStyle = PAPER;
    ctx.font = "700 21px Arial, sans-serif";
    ctx.fillText(zh ? "本期阅读" : "READING", WIDTH - 228, y + 43);
    ctx.font = "500 82px Georgia, serif";
    ctx.fillText("01", WIDTH - 228, y + 128);
  }

  const metadataWidth = options.cover ? WIDTH - 144 : WIDTH - 360;
  ctx.fillStyle = MUTED;
  ctx.font = "600 21px Arial, sans-serif";
  const date = formatPublished(options.locale, options.publishedAt);
  ctx.fillText(`${options.type.toUpperCase()}${date ? `  /  ${date}` : ""}`, 72, y + 18);
  y += 70;

  ctx.fillStyle = INK;
  ctx.font = `600 ${options.height === 1920 ? 70 : 64}px Georgia, "Noto Serif SC", serif`;
  const titleLines = wrap(ctx, options.title, metadataWidth, options.cover ? 3 : 4);
  y = drawLines(ctx, titleLines, 72, y, options.height === 1920 ? 86 : 78);
  y += 24;

  const panelTop = y;
  const panelBottom = footerTop - 62;
  ctx.fillStyle = INK;
  ctx.fillRect(72, panelTop, WIDTH - 144, panelBottom - panelTop);
  ctx.fillStyle = options.accent;
  ctx.fillRect(72, panelTop, WIDTH - 144, 12);
  ctx.font = "700 20px Arial, sans-serif";
  ctx.fillText(zh ? "正文摘录 / ARTICLE EXCERPT" : "ARTICLE EXCERPT / 正文摘录", 112, panelTop + 58);
  ctx.globalAlpha = 0.28;
  ctx.font = "500 176px Georgia, serif";
  ctx.fillText("“", 94, panelTop + 188);
  ctx.globalAlpha = 1;
  ctx.fillStyle = PAPER;
  ctx.font = `${options.height === 1920 ? 34 : 31}px Georgia, "Noto Serif SC", serif`;
  const excerpt = options.excerpt || (zh ? "正文尚未填写，扫码查看最新发布版本。" : "No excerpt is available. Scan to view the latest version.");
  const excerptLineHeight = options.height === 1920 ? 55 : 50;
  const excerptY = panelTop + 128;
  const excerptLineLimit = Math.max(2, Math.min(options.height === 1920 ? 9 : 6, Math.floor((panelBottom - 92 - excerptY) / excerptLineHeight)));
  const excerptLines = wrap(ctx, excerpt, WIDTH - 190, excerptLineLimit);
  drawLines(ctx, excerptLines, 112, excerptY, excerptLineHeight);
  ctx.save();
  ctx.globalAlpha = 0.07;
  ctx.fillStyle = PAPER;
  ctx.textAlign = "right";
  ctx.font = `500 ${options.height === 1920 ? 360 : 280}px Georgia, serif`;
  ctx.fillText("01", WIDTH - 108, panelBottom - 74);
  ctx.restore();
  ctx.strokeStyle = "#50494d";
  ctx.lineWidth = 2;
  ctx.beginPath();
  ctx.moveTo(112, panelBottom - 58);
  ctx.lineTo(WIDTH - 112, panelBottom - 58);
  ctx.stroke();
  ctx.fillStyle = "#bcb4ae";
  ctx.font = "600 17px Arial, sans-serif";
  ctx.fillText(zh ? "继续阅读 · 扫描下方二维码" : "CONTINUE READING · SCAN BELOW", 112, panelBottom - 27);

  ctx.strokeStyle = "#c9c1b7";
  ctx.lineWidth = 2;
  ctx.beginPath();
  ctx.moveTo(72, footerTop - 34);
  ctx.lineTo(WIDTH - 72, footerTop - 34);
  ctx.stroke();
  ctx.fillStyle = options.accent;
  ctx.beginPath();
  ctx.arc(WIDTH - 94, footerTop - 34, 10, 0, Math.PI * 2);
  ctx.fill();

  const qr = await loadImage(await QRCode.toDataURL(options.url, {
    margin: 1,
    width: 210,
    color: { dark: INK, light: PAPER },
  }));
  ctx.drawImage(qr, 72, footerTop, 210, 210);
  ctx.fillStyle = INK;
  ctx.font = "600 29px Arial, sans-serif";
  ctx.fillText(options.footer || (zh ? "扫码继续阅读完整文章" : "Scan to continue reading"), 326, footerTop + 70);
  ctx.fillStyle = MUTED;
  ctx.font = "23px Arial, sans-serif";
  ctx.fillText(new URL(options.url).host, 326, footerTop + 112);
  ctx.font = "18px Arial, sans-serif";
  ctx.fillText(zh ? "保存海报 · 分享给想读的人" : "Save · Share · Read", 326, footerTop + 158);

  return canvas;
}
