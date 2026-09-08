import type { Metadata } from "next";
import "./globals.css";
import "./index.css";
export const metadata: Metadata = {
  title: { default: "Field Notes", template: "%s — Field Notes" },
  description: "Notes, essays and working knowledge.",
};
export default function Layout({ children }: { children: React.ReactNode }) {
  return (
    <html>
      <body>
        <a className="skip-link" href="#content">
          Skip to content
        </a>
        {children}
      </body>
    </html>
  );
}
