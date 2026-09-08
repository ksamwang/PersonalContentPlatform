const origin = process.env.API_ORIGIN ?? "http://localhost:8080";
const workspace = process.env.WORKSPACE_SLUG ?? "personal";

export async function GET(
  _request: Request,
  { params }: { params: Promise<{ locale: string }> },
) {
  const { locale } = await params;
  const apiLocale = locale === "zh" ? "zh-CN" : "en";
  try {
    const upstream = await fetch(
      `${origin}/v1/public/${workspace}/${apiLocale}/rss.xml`,
      { cache: "no-store" },
    );
    if (!upstream.ok) {
      return new Response("RSS feed is temporarily unavailable", {
        status: upstream.status,
      });
    }
    return new Response(await upstream.arrayBuffer(), {
      headers: {
        "Content-Type": "application/rss+xml; charset=utf-8",
        "Cache-Control": "public, max-age=60",
      },
    });
  } catch {
    return new Response("RSS feed is temporarily unavailable", { status: 503 });
  }
}
