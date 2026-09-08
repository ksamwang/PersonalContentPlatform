import Link from "next/link";
import { ArrowUpRight } from "lucide-react";
import { SiteHeader } from "../../components/SiteHeader";
import { getPublicSettings, listPages } from "../../lib/content";

export default async function Home({
  params,
}: {
  params: Promise<{ locale: string }>;
}) {
  const { locale } = await params;
  const zh = locale === "zh";
  const [items,settings] = await Promise.all([listPages(zh ? "zh-CN" : "en"),getPublicSettings()]);
  return (
    <>
      <SiteHeader locale={locale} name={settings.site.name} rss={settings.site.rss_enabled} />
      <main id="content">
        <section className="hero">
          <span className="kicker">
            {zh
              ? "持续记录 · 独立思考"
              : "ONGOING NOTES · INDEPENDENT THINKING"}
          </span>
          <h1>
            {zh ? (
              <>
                把零散经验，
                <br />
                写成可以返回的地方。
              </>
            ) : (
              <>
                Turn passing experience
                <br />
                into a place to return to.
              </>
            )}
          </h1>
          <p>
            {settings.site.description || (zh
              ? "这里收录关于创作、技术与生活实践的文章和笔记。内容会被更新、连接，也会在时间里重新出现。"
              : "Essays and notes on making, technology, and lived practice—revised, connected, and rediscovered over time.")}
          </p>
        </section>
        <section className="index">
          <div className="section-heading">
            <span>01</span>
            <h2>{zh ? "最近发布" : "Latest publications"}</h2>
          </div>
          {items.length ? (
            <div className="publication-index">
              {items.map((item) => (
                <Link
                  key={item.slug}
                  href={`/${locale}/${item.type}/${item.slug}`}
                >
                  <span>{item.type}</span>
                  <strong>{item.title}</strong>
                  <p>{item.summary}</p>
                  <ArrowUpRight aria-hidden />
                </Link>
              ))}
            </div>
          ) : (
            <div className="empty-public">
              <p>
                {zh
                  ? "第一篇内容正在路上。"
                  : "The first publication is on its way."}
              </p>
              <Link href={`/${locale}/about`}>
                {zh ? "了解这个空间" : "About this space"}
                <ArrowUpRight aria-hidden />
              </Link>
            </div>
          )}
        </section>
      </main>
      <footer>
        <span>{settings.site.name}</span>
        <span>{settings.site.footer}</span>
      </footer>
    </>
  );
}
