import type { Metadata } from "next";
import "./globals.css";
import "./styles/editor-polish.css";
export const metadata: Metadata = {
  title: "Content Studio",
  description: "Personal Content Platform Studio",
};
export default function Layout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="zh-CN">
      <body>
        <a className="skip-link" href="#main">
          跳到主要内容
        </a>
        {children}
      </body>
    </html>
  );
}
