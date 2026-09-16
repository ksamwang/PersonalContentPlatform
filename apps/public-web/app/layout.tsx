import type { Metadata } from "next";
import "@fontsource-variable/newsreader/wght.css";
import "@fontsource-variable/noto-sans-sc/wght.css";
import "@fontsource-variable/noto-serif-sc/wght.css";
import "@fontsource-variable/public-sans/wght.css";
import "./globals.css";
import "./index.css";
import "./styles/discovery-polish.css";
import "./styles/reading-polish.css";
import "./styles/hero-cosmos.css";
import { getPublicSettings, publicBase } from "../lib/content";
import { socialImageURL } from "../lib/social-image";
export async function generateMetadata():Promise<Metadata>{const settings=await getPublicSettings(),base=publicBase(settings),title=settings.site.name||"Field Notes",description=settings.site.description||"Notes, essays and working knowledge.",image=socialImageURL(base);return {metadataBase:new URL(base),title:{default:title,template:`%s — ${title}`},description,openGraph:{siteName:title,type:"website",title,description,images:[{url:image,width:1200,height:630,type:"image/png",alt:`${title} social card`}]},twitter:{card:"summary_large_image",title,description,images:[image]}}}
export default function Layout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="zh-CN">
      <body>
        <a className="skip-link" href="#content">
          Skip to content
        </a>
        {children}
      </body>
    </html>
  );
}
