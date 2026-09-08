import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { SiteHeader } from "../../../../components/SiteHeader";
import { getPage, getPublicSettings } from "../../../../lib/content";
type Props = {
  params: Promise<{ locale: string; type: string; slug: string }>;
};
export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const p = await params,
    page = await getPage(p.locale === "zh" ? "zh-CN" : "en", p.type, p.slug);
  if (!page) return {};
  return {
    title: page.title,
    description: page.summary,
    alternates: { canonical: `/${p.locale}/${p.type}/${p.slug}` },
  };
}
export default async function Published({ params }: Props) {
  const p = await params,
    apiLocale = p.locale === "zh" ? "zh-CN" : "en",
    [page,settings] = await Promise.all([getPage(apiLocale, p.type, p.slug),getPublicSettings()]);
  if (!page) notFound();
  return (
    <>
      <SiteHeader locale={p.locale} name={settings.site.name} rss={settings.site.rss_enabled} />
      <main id="content" className="article-shell">
        <article>
          <header>
            <span className="kicker">
              {p.type} ·{" "}
              {new Intl.DateTimeFormat(p.locale, { dateStyle: "long" }).format(
                new Date(page.published_at),
              )}
            </span>
            <h1>{page.title}</h1>
            {page.summary && <p className="dek">{page.summary}</p>}
          </header>
          <div
            className="article-body"
            dangerouslySetInnerHTML={{ __html: page.html }}
          />
        </article>
      </main>
      <footer>
        <span>{settings.site.name}</span>
        <span>{settings.site.footer}</span>
      </footer>
    </>
  );
}
