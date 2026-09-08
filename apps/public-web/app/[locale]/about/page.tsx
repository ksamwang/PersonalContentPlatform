import { SiteHeader } from "../../../components/SiteHeader";
import { getPublicSettings } from "../../../lib/content";

export default async function About({
  params,
}: {
  params: Promise<{ locale: string }>;
}) {
  const { locale } = await params;
  const zh = locale === "zh";
  const settings = await getPublicSettings();
  return (
    <>
      <SiteHeader locale={locale} name={settings.site.name} rss={settings.site.rss_enabled} />
      <main id="content" className="article-shell">
        <article>
          <header>
            <span className="kicker">FIELD NOTES · ABOUT</span>
            <h1>{zh ? "关于这个空间" : "About this space"}</h1>
            <p className="dek">
              {settings.site.description || (zh
                ? "这里保存值得长期维护的经验、思考与创作记录。"
                : "A durable home for experiences, ideas, and creative notes worth maintaining.")}
            </p>
          </header>
          <div className="article-body">
            <p>
              {settings.site.about || (zh
                ? "内容以可修订、可连接、可迁移的形式保存，并通过独立网站与 RSS 发布。"
                : "Content is stored in a revisable, connected, and portable form, then published through an independent website and RSS.")}
            </p>
          </div>
        </article>
      </main>
      <footer>
        <span>{settings.site.name}</span>
        <span>{settings.site.footer}</span>
      </footer>
    </>
  );
}
