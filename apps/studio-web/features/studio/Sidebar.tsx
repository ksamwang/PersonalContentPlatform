"use client";

import { useState } from "react";
import {
  BookOpen,
  Inbox,
  Library,
  Images,
  LogOut,
  Network,
  Search,
  Settings2,
  Sparkles,
  Shapes,
  Send,
  ArrowLeftRight,
  Menu,
  X,
} from "lucide-react";

export type StudioView = "content"|"inbox"|"collections"|"assets"|"knowledge"|"ai"|"discover"|"publication"|"migration"|"settings";
export function Sidebar({ onNavigate, view, publicSite, userLabel, onLogout }: { onNavigate: (view:StudioView) => void; view: StudioView; publicSite:string; userLabel:string; onLogout:()=>void }) {
  const [menuOpen,setMenuOpen]=useState(false);
  const navigate=(next:StudioView)=>{onNavigate(next);setMenuOpen(false)};
  return (
    <aside className={`sidebar${menuOpen?" menu-open":""}`}>
      <div className="brand">
        <span className="brand-mark">P</span>
        <span>Content Studio</span>
      </div>
      <button className="mobile-menu" type="button" aria-label={menuOpen?"关闭导航":"打开导航"} aria-expanded={menuOpen} onClick={()=>setMenuOpen(value=>!value)}>
        {menuOpen?<X aria-hidden size={20}/>:<Menu aria-hidden size={20}/>}<span>菜单</span>
      </button>
      <nav aria-label="主要导航">
        <button
          className={view === "content" ? "active" : ""}
          onClick={() => navigate("content")}
        >
          <Library aria-hidden size={19} />
          <span>内容</span>
        </button>
        <button className={view === "inbox" ? "active" : ""} onClick={()=>navigate("inbox")}>
          <Inbox aria-hidden size={19} />
          <span>收件箱</span>
        </button>
        <button className={view === "collections" ? "active" : ""} onClick={()=>navigate("collections")}>
          <Shapes aria-hidden size={19} />
          <span>合集</span>
        </button>
        <button className={view === "assets" ? "active" : ""} onClick={()=>navigate("assets")}><Images aria-hidden size={19}/><span>资产</span></button>
        <button className={view === "knowledge" ? "active" : ""} onClick={()=>navigate("knowledge")}><Network aria-hidden size={19}/><span>知识</span></button>
        <button className={view === "ai" ? "active" : ""} onClick={()=>navigate("ai")}><Sparkles aria-hidden size={19}/><span>AI 助手</span></button>
        <button className={view === "discover" ? "active" : ""} onClick={()=>navigate("discover")}>
          <Search aria-hidden size={19} />
          <span>发现</span>
        </button>
        <button className={view === "publication" ? "active" : ""} onClick={()=>navigate("publication")}><Send aria-hidden size={19}/><span>发布</span></button>
        <button className={view === "migration" ? "active" : ""} onClick={()=>navigate("migration")}><ArrowLeftRight aria-hidden size={19}/><span>迁移</span></button>
        <a href={publicSite} target="_blank" rel="noreferrer">
          <BookOpen aria-hidden size={19} />
          <span>发布站点</span>
        </a>
      </nav>
      <div className="sidebar-account">
        <button className={`settings ${view === "settings" ? "active" : ""}`} onClick={()=>navigate("settings")}>
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
