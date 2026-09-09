import Link from "next/link";
import { ArrowUpRight } from "lucide-react";
import { SiteHeader } from "../../components/SiteHeader";
import { getPublicSettings, listCollections, listPages, listTags } from "../../lib/content";
import { ContentCards } from "../../components/ContentCards";

export default async function Home({
  params,
}: {
  params: Promise<{ locale: string }>;
}) {
  const { locale } = await params;
  const zh = locale === "zh";
  const apiLocale=zh?"zh-CN":"en";const [items,settings,tags,collections] = await Promise.all([listPages(apiLocale),getPublicSettings(),listTags(apiLocale),listCollections(apiLocale)]);
  return (
    <div data-theme={settings.site.theme||"paper"} style={{"--accent":settings.site.accent_color||"#d82f76"} as React.CSSProperties}>
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
            <ContentCards items={items} locale={locale}/>
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
        {(collections.length>0||tags.length>0)&&<section className="discover-sections">{collections.length>0&&<div><div className="section-heading"><span>02</span><h2>{zh?"内容合集":"Collections"}</h2></div><div className="collection-links">{collections.map(item=><Link key={item.slug} href={`/${locale}/collections/${item.slug}`}>{item.title}<ArrowUpRight/></Link>)}</div></div>}{tags.length>0&&<div><div className="section-heading"><span>03</span><h2>{zh?"按标签发现":"Browse tags"}</h2></div><div className="tag-cloud">{tags.map(item=><Link key={item.name} href={`/${locale}/search?tag=${encodeURIComponent(item.name)}`}>{item.name}<small>{item.count}</small></Link>)}</div></div>}</section>}
      </main>
      <footer>
        <span>{settings.site.name}</span>
        <span>{settings.site.footer}</span>
      </footer>
    </div>
  );
}
