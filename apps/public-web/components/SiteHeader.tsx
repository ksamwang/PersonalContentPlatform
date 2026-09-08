import Link from "next/link";
import { Languages, Rss } from "lucide-react";
export function SiteHeader({ locale, name="Field Notes", rss=true }: { locale: string; name?:string; rss?:boolean }) {
  const other = locale === "zh" ? "en" : "zh";
  return (
    <header className="site-header">
      <Link className="wordmark" href={`/${locale}`}>
        {name.toUpperCase()}<span>.</span>
      </Link>
      <nav aria-label={locale === "zh" ? "主要导航" : "Primary navigation"}>
        <Link href={`/${locale}`}>{locale === "zh" ? "最新" : "Latest"}</Link>
        <Link href={`/${locale}/about`}>
          {locale === "zh" ? "关于" : "About"}
        </Link>
        <Link
          className="icon-link"
          href={`/${other}`}
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
