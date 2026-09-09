import {
  BookOpen,
  Inbox,
  Library,
  Search,
  Settings2,
  Shapes,
} from "lucide-react";

export type StudioView = "content"|"inbox"|"collections"|"settings";
export function Sidebar({ onNavigate, onSearch, view, publicSite }: { onNavigate: (view:StudioView) => void; onSearch: () => void; view: StudioView; publicSite:string }) {
  return (
    <aside className="sidebar">
      <div className="brand">
        <span className="brand-mark">P</span>
        <span>Content Studio</span>
      </div>
      <nav aria-label="主要导航">
        <button
          className={view === "content" ? "active" : ""}
          onClick={() => onNavigate("content")}
        >
          <Library aria-hidden size={19} />
          <span>内容</span>
        </button>
        <button className={view === "inbox" ? "active" : ""} onClick={()=>onNavigate("inbox")}>
          <Inbox aria-hidden size={19} />
          <span>收件箱</span>
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
      <button className={`settings ${view === "settings" ? "active" : ""}`} onClick={()=>onNavigate("settings")}>
        <Settings2 aria-hidden size={19} />
        <span>设置</span>
      </button>
    </aside>
  );
}
