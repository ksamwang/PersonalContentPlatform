import {
  BookOpen,
  Inbox,
  Library,
  Search,
  Settings2,
  Shapes,
} from "lucide-react";

export function Sidebar({ onSearch }: { onSearch: () => void }) {
  const publicSite =
    process.env.NEXT_PUBLIC_PUBLIC_SITE_URL ?? "http://localhost:3000/zh";
  return (
    <aside className="sidebar">
      <div className="brand">
        <span className="brand-mark">P</span>
        <span>Content Studio</span>
      </div>
      <nav aria-label="主要导航">
        <button
          className="active"
          onClick={() => document.querySelector("#main")?.scrollIntoView()}
        >
          <Library aria-hidden size={19} />
          <span>内容</span>
        </button>
        <button disabled title="收件箱将在后续版本开放">
          <Inbox aria-hidden size={19} />
          <span>收件箱</span>
          <small>规划中</small>
        </button>
        <button disabled title="合集将在后续版本开放">
          <Shapes aria-hidden size={19} />
          <span>合集</span>
          <small>规划中</small>
        </button>
        <button onClick={onSearch}>
          <Search aria-hidden size={19} />
          <span>发现</span>
        </button>
        <a href={publicSite} target="_blank" rel="noreferrer">
          <BookOpen aria-hidden size={19} />
          <span>发布站点</span>
        </a>
      </nav>
      <button className="settings" disabled title="设置界面将在后续版本开放">
        <Settings2 aria-hidden size={19} />
        <span>设置</span>
        <small>规划中</small>
      </button>
    </aside>
  );
}
