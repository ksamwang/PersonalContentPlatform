import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { SiteHeader } from "../../../../components/SiteHeader";
import { ShareActions } from "../../../../components/ShareActions";
import { articleExcerpt } from "../../../../lib/article-text";
import { getPage, getPublicSettings, publicBase } from "../../../../lib/content";
import { socialImageURL } from "../../../../lib/social-image";
type Props = {
  params: Promise<{ locale: string; type: string; slug: string }>;
};
export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const p = await params,
    page = await getPage(p.locale === "zh" ? "zh-CN" : "en", p.type, p.slug);
  if (!page) return {};
  const settings=await getPublicSettings(),base=publicBase(settings),title=page.metadata?.seo_title||page.title,description=page.metadata?.seo_description||page.summary,image=socialImageURL(base,{locale:page.locale,type:p.type,slug:p.slug,publishedAt:page.published_at});
  const languages=Object.fromEntries((page.alternates??[]).map(item=>[item.locale==="zh-CN"?"zh":item.locale,`/${item.locale==="zh-CN"?"zh":item.locale}/${item.type}/${item.slug}`]));
  return {
    title,
    description,
    alternates: { canonical: new URL(`/${p.locale}/${p.type}/${p.slug}`,base),languages },
    openGraph:{type:"article",title,description,url:`/${p.locale}/${p.type}/${p.slug}`,images:[{url:image,width:1200,height:630,type:"image/png",alt:p.locale==="zh"?`${page.title} 分享卡片`:`${page.title} social card`}]},
    twitter:{card:"summary_large_image",title,description,images:[image]},
  };
}
export default async function Published({ params }: Props) {
  const p = await params,
    apiLocale = p.locale === "zh" ? "zh-CN" : "en",
    [page,settings] = await Promise.all([getPage(apiLocale, p.type, p.slug),getPublicSettings()]);
  if (!page) notFound();
  const otherLocale=p.locale==="zh"?"en":"zh",alternate=page.alternates?.find(item=>(item.locale==="zh-CN"?"zh":item.locale)===otherLocale),alternatePath=alternate?`/${otherLocale}/${alternate.type}/${alternate.slug}`:undefined;
  const cover=page.metadata?.cover_asset_id?`/media/${page.metadata.cover_asset_id}/content-1280`:undefined;
  return (
    <div data-theme={settings.site.theme||"paper"} style={{"--accent":settings.site.accent_color||"#d82f76"} as React.CSSProperties}>
      <SiteHeader locale={p.locale} name={settings.site.name} rss={settings.site.rss_enabled} alternate={alternatePath}/>
      <main id="content" className="article-shell">
        <article>
          <header className="article-header">
            <span className="kicker">
              {p.type} ·{" "}
              {new Intl.DateTimeFormat(p.locale, { dateStyle: "long" }).format(
                new Date(page.published_at),
              )}
            </span>
            <h1>{page.title}</h1>
            {page.summary && <p className="dek">{page.summary}</p>}
            {page.metadata?.tags?.length?<div className="article-tags">{page.metadata.tags.map(tag=><a key={tag} href={`/${p.locale}/search?tag=${encodeURIComponent(tag)}`}>{tag}</a>)}</div>:null}
          </header>
          {cover&&<picture className="article-cover"><source media="(max-width: 640px)" srcSet={`/media/${page.metadata.cover_asset_id}/thumbnail`}/><img src={cover} alt=""/></picture>}
          <div
            className="article-body"
            dangerouslySetInnerHTML={{ __html: page.html }}
          />
          {!alternate&&<p className="translation-note">{p.locale==="zh"?"此内容暂时没有英文版本。":"A Chinese version is not available yet."}</p>}
          <ShareActions locale={p.locale} title={page.title} summary={page.summary} excerpt={articleExcerpt(page.html,page.summary)} type={p.type} publishedAt={page.published_at} siteName={settings.site.name} cover={cover} accent={settings.site.accent_color||"#d82f76"} footer={settings.site.share_footer}/>
        </article>
      </main>
      <footer>
        <span>{settings.site.name}</span>
        <span>{settings.site.footer}</span>
      </footer>
    </div>
  );
}
