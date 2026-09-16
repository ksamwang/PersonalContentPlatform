import Link from "next/link";
import { Languages, Rss, Search } from "lucide-react";
import { BrandMark } from "./BrandMark";
export function SiteHeader({ locale, name="Field Notes", rss=true,alternate }: { locale: string; name?:string; rss?:boolean;alternate?:string }) {
  const other = locale === "zh" ? "en" : "zh";
  return (
    <header className="site-header">
      <Link className="wordmark" href={`/${locale}`}>
        <BrandMark className="brand-mark" size={34}/><b>{name.toUpperCase()}<span>.</span></b>
      </Link>
      <nav aria-label={locale === "zh" ? "主要导航" : "Primary navigation"}>
        <Link href={`/${locale}`}>{locale === "zh" ? "最新" : "Latest"}</Link>
        <Link className="collections-link" href={`/${locale}/collections`}>{locale === "zh" ? "合集" : "Collections"}</Link>
        <Link className="icon-link" href={`/${locale}/search`} aria-label={locale === "zh" ? "搜索" : "Search"}><Search aria-hidden /></Link>
        <Link href={`/${locale}/about`}>
          {locale === "zh" ? "关于" : "About"}
        </Link>
        <Link
          className="icon-link"
          href={alternate||`/${other}`}
          aria-label={locale === "zh" ? "Switch to English" : "切换到中文"}
        >
          <Languages aria-hidden />
        </Link>
        {rss && <a className="icon-link" href={`/api/rss/${locale}`} aria-label="RSS">
          <Rss aria-hidden />
        </a>}
      </nav>
    </header>
  );
}
