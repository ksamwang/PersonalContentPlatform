import type { Metadata } from "next";
import "./globals.css";
import "./index.css";
import "./styles/discovery-polish.css";
import { getPublicSettings, publicBase } from "../lib/content";
export async function generateMetadata():Promise<Metadata>{const settings=await getPublicSettings();return {metadataBase:new URL(publicBase(settings)),title:{default:settings.site.name||"Field Notes",template:`%s — ${settings.site.name||"Field Notes"}`},description:settings.site.description||"Notes, essays and working knowledge.",openGraph:{siteName:settings.site.name,type:"website"}}}
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
