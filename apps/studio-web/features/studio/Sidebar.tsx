import {
  BookOpen,
  Inbox,
  Library,
  Images,
  LogOut,
  Network,
  Search,
  Settings2,
  Shapes,
} from "lucide-react";

export type StudioView = "content"|"inbox"|"collections"|"assets"|"knowledge"|"discover"|"settings";
export function Sidebar({ onNavigate, view, publicSite, userLabel, onLogout }: { onNavigate: (view:StudioView) => void; view: StudioView; publicSite:string; userLabel:string; onLogout:()=>void }) {
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
        <button className={view === "collections" ? "active" : ""} onClick={()=>onNavigate("collections")}>
          <Shapes aria-hidden size={19} />
          <span>合集</span>
        </button>
        <button className={view === "assets" ? "active" : ""} onClick={()=>onNavigate("assets")}><Images aria-hidden size={19}/><span>资产</span></button>
        <button className={view === "knowledge" ? "active" : ""} onClick={()=>onNavigate("knowledge")}><Network aria-hidden size={19}/><span>知识</span></button>
        <button className={view === "discover" ? "active" : ""} onClick={()=>onNavigate("discover")}>
          <Search aria-hidden size={19} />
          <span>发现</span>
        </button>
        <a href={publicSite} target="_blank" rel="noreferrer">
          <BookOpen aria-hidden size={19} />
          <span>发布站点</span>
        </a>
      </nav>
      <div className="sidebar-account">
        <button className={`settings ${view === "settings" ? "active" : ""}`} onClick={()=>onNavigate("settings")}>
          <Settings2 aria-hidden size={19} />
          <span>设置</span>
        </button>
        <div className="account-meta" title={userLabel}>
          <span className="account-avatar" aria-hidden>{userLabel.slice(0, 1).toUpperCase()}</span>
          <span>{userLabel}</span>
        </div>
        <button className="logout" onClick={onLogout}>
          <LogOut aria-hidden size={19} />
          <span>退出登录</span>
        </button>
      </div>
    </aside>
  );
}
