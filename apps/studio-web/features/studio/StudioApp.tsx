"use client";
import { useCallback, useEffect, useRef, useState } from "react";
import { Search } from "lucide-react";
import { api, Content, Principal } from "../../lib/api";
import { AuthPanel } from "../auth/AuthPanel";
import { AddPasskey } from "../auth/PasskeyButton";
import { AssetUpload } from "../assets/AssetUpload";
import { EditorPanel } from "../editor/EditorPanel";
import { ContentList } from "./ContentList";
import { CreateContent } from "./CreateContent";
import { Sidebar } from "./Sidebar";
import { SettingsPanel } from "../settings/SettingsPanel";
import { InboxPanel } from "../inbox/InboxPanel";
import { StudioView } from "./Sidebar";
export function StudioApp() {
  const [user, setUser] = useState<Principal | null | undefined>(undefined),
    [items, setItems] = useState<Content[]>([]),
    [selected, setSelected] = useState<Content>(),
    [query, setQuery] = useState(""),
    [loadError, setLoadError] = useState(""),
    [view, setView] = useState<StudioView>("content"),
    [publicSite,setPublicSite] = useState("http://localhost:3000/zh");
  const searchRef = useRef<HTMLInputElement>(null);
  const load = useCallback(
    async (p = user, q = query) => {
      if (!p) return;
      try {
        const result = q
          ? await api.search(p.WorkspaceID, q)
          : await api.list(p.WorkspaceID);
        setItems(result.items);
        setLoadError("");
        if (selected) {
          setSelected(result.items.find((i) => i.id === selected.id) ?? selected);
        }
      } catch (err) {
        setLoadError(err instanceof Error ? err.message : "内容加载失败");
      }
    },
    [user, query, selected],
  );
  useEffect(() => {
    api
      .me()
      .then((p) => setUser(p))
      .catch(() => setUser(null));
  }, []);
  useEffect(() => {
    if (!user) return;
    void api.settings(user.WorkspaceID).then(v=>{if(v.general.site.public_url)setPublicSite(v.general.site.public_url)}).catch(()=>{});
    const timer = setTimeout(() => void load(user, query), query ? 300 : 0);
    return () => clearTimeout(timer);
  }, [user, query]);
  if (user === undefined)
    return (
      <main id="main" className="boot-screen" aria-busy="true">
        <span className="brand-mark">P</span>
        <p>正在打开 Content Studio…</p>
      </main>
    );
  if (!user) return <AuthPanel onAuthenticated={setUser} />;
  return (
    <div className="studio">
      <Sidebar view={view} publicSite={publicSite} onNavigate={setView} onSearch={() => { setView("content"); setTimeout(() => searchRef.current?.focus(), 0); }} />
      <main id="main" className="workspace">
        {view === "settings" ? (
          <SettingsPanel workspace={user.WorkspaceID} role={user.Role} />
        ) : view === "inbox" ? (
          <InboxPanel workspace={user.WorkspaceID} role={user.Role} onOpenContent={async id=>{try{const content=await api.content(user.WorkspaceID,id);setSelected(content);setView("content");void load(user,"")}catch(err){setLoadError(err instanceof Error?err.message:"内容加载失败");setView("content")}}}/>
        ) : (<>
        <header className="topbar">
          <div>
            <span className="eyebrow">YOUR LIBRARY</span>
            <h1>内容工作台</h1>
          </div>
          <div className="top-actions">
            <label className="search-box">
              <Search aria-hidden size={18} />
              <span className="sr-only">搜索内容</span>
              <input
                ref={searchRef}
                value={query}
                onChange={(e) => setQuery(e.target.value)}
                placeholder="搜索内容…"
              />
            </label>
            <AddPasskey />
            <AssetUpload workspace={user.WorkspaceID} />
            <CreateContent
              workspace={user.WorkspaceID}
              onCreated={(c) => {
                setItems((v) => [c, ...v]);
                setSelected(c);
              }}
            />
          </div>
        </header>
        <div className="studio-grid">
          <section className="library-pane" aria-label="内容列表">
            <div className="pane-meta">
              <span>{items.length} 项内容</span>
              <button onClick={() => api.logout().then(() => setUser(null))}>
                退出登录
              </button>
            </div>
            {loadError && (
              <div className="inline-error" role="alert">
                <span>{loadError}</span>
                <button onClick={() => void load(user, query)}>重试</button>
              </div>
            )}
            <ContentList
              items={items}
              searching={Boolean(query)}
              selected={selected?.id}
              onSelect={setSelected}
            />
          </section>
          {selected ? (
            <EditorPanel
              key={selected.id}
              workspace={user.WorkspaceID}
              content={selected}
              onChanged={() => void load(user, query)}
            />
          ) : (
            <section className="welcome-panel">
              <span className="oversized">Aa</span>
              <h2>选择一项内容开始编辑</h2>
              <p>你的修改会自动保存，也可以在准备好后封存为版本并发布。</p>
            </section>
          )}
        </div>
        </>)}
      </main>
    </div>
  );
}
