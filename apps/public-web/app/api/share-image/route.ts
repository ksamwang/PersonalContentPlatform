import { createElement } from "react";
import { ImageResponse } from "next/og";
import { SocialCard } from "../../../components/social-card/SocialCard";
import { getPage, getPublicSettings, publicBase } from "../../../lib/content";

export const runtime = "nodejs";
export const dynamic = "force-dynamic";

const imageSize = { width: 1200, height: 630 };

async function coverData(assetID?: string) {
  if (!assetID) return undefined;
  const api = process.env.API_ORIGIN ?? "http://localhost:8080";
  try {
    const response = await fetch(`${api}/v1/public/assets/${encodeURIComponent(assetID)}/content-1280`, {
      cache: "no-store",
      signal: AbortSignal.timeout(5000),
    });
    if (!response.ok) return undefined;
    const body = Buffer.from(await response.arrayBuffer());
    if (body.byteLength > 5 * 1024 * 1024) return undefined;
    return `data:${response.headers.get("content-type") || "image/jpeg"};base64,${body.toString("base64")}`;
  } catch {
    return undefined;
  }
}

export async function GET(request: Request) {
  const url = new URL(request.url);
  const locale = url.searchParams.get("locale");
  const type = url.searchParams.get("type");
  const slug = url.searchParams.get("slug");
  const settings = await getPublicSettings();
  const base = publicBase(settings);
  const domain = new URL(base).host;
  const page = locale && type && slug ? await getPage(locale, type, slug) : null;
  const accent = settings.site.accent_color || "#d82f76";
  const cover = await coverData(page?.metadata?.cover_asset_id);
  const zh = locale === "zh-CN" || locale === "zh" || (!locale && settings.workspace.default_locale === "zh-CN");

  return new ImageResponse(
    createElement(SocialCard, {
      accent,
      coverData: cover,
      date: page?.published_at,
      description: page?.metadata?.seo_description || page?.summary || settings.site.description || (zh ? "记录创作、技术与生活实践，并在时间里重新连接。" : "Notes, essays and working knowledge—connected over time."),
      domain,
      kind: page ? "article" : "site",
      locale: locale || settings.workspace.default_locale,
      siteName: settings.site.name || "Field Notes",
      title: page?.metadata?.seo_title || page?.title || settings.site.name || "Field Notes",
      type: page?.type,
    }),
    {
      ...imageSize,
      headers: {
        "Cache-Control": "public, max-age=0, s-maxage=300, stale-while-revalidate=86400",
      },
    },
  );
}
