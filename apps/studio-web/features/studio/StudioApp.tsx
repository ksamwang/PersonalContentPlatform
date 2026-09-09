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
import { CollectionsPanel } from "../collections/CollectionsPanel";
import { StudioView } from "./Sidebar";
import { ContentFilters, ContentFilter, emptyContentFilter } from "./ContentFilters";
import { DiscoverPanel } from "./DiscoverPanel";
export function StudioApp() {
  const [user, setUser] = useState<Principal | null | undefined>(undefined),
    [items, setItems] = useState<Content[]>([]),
    [selected, setSelected] = useState<Content>(),
    [query, setQuery] = useState(""),
    [loadError, setLoadError] = useState(""),
    [view, setView] = useState<StudioView>("content"),
    [filters,setFilters]=useState<ContentFilter>(emptyContentFilter),
    [publicSite,setPublicSite] = useState("http://localhost:3000/zh");
  const searchRef = useRef<HTMLInputElement>(null);
  const load = useCallback(
    async (p = user, q = query) => {
      if (!p) return;
      try {
        const result = q ? await api.search(p.WorkspaceID, q, filters.locale) : await api.list(p.WorkspaceID,filters);
        const sorted=filters.sort==="title"?[...result.items].sort((a,b)=>(a.localizations[0]?.current_revision?.title??"").localeCompare(b.localizations[0]?.current_revision?.title??"","zh-CN")):result.items;
        setItems(sorted);
        setLoadError("");
        if (selected) {
          setSelected(result.items.find((i) => i.id === selected.id) ?? selected);
        }
      } catch (err) {
        setLoadError(err instanceof Error ? err.message : "内容加载失败");
      }
    },
    [user, query, selected, filters],
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
  }, [user, query, filters]);
  async function lifecycle(action:"archive"|"restore"|"delete",content:Content){
    if(action==="delete"&&!window.confirm("确认将此内容移至回收站？公开页面会立即下线。"))return;
    try{if(action==="archive")await api.archiveContent(user!.WorkspaceID,content.id);else if(action==="restore")await api.restoreContent(user!.WorkspaceID,content.id);else await api.deleteContent(user!.WorkspaceID,content.id);if(selected?.id===content.id)setSelected(undefined);await load(user," ".trim())}catch(error){setLoadError(error instanceof Error?error.message:"操作失败")}
  }
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
      <Sidebar view={view} publicSite={publicSite} onNavigate={setView} />
      <main id="main" className="workspace">
        {view === "settings" ? (
          <SettingsPanel workspace={user.WorkspaceID} role={user.Role} />
        ) : view === "inbox" ? (
          <InboxPanel workspace={user.WorkspaceID} role={user.Role} onOpenContent={async id=>{try{const content=await api.content(user.WorkspaceID,id);setSelected(content);setView("content");void load(user,"")}catch(err){setLoadError(err instanceof Error?err.message:"内容加载失败");setView("content")}}}/>
        ) : view === "collections" ? (
          <CollectionsPanel workspace={user.WorkspaceID} role={user.Role} onOpenContent={async id=>{try{const content=await api.content(user.WorkspaceID,id);setSelected(content);setView("content")}catch(err){setLoadError(err instanceof Error?err.message:"内容加载失败");setView("content")}}}/>
        ) : view === "discover" ? (
          <DiscoverPanel workspace={user.WorkspaceID} onOpen={content=>{setSelected(content);setView("content")}}/>
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
        <ContentFilters value={filters} onChange={setFilters}/>
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
              canEdit={user.Role==="owner"||user.Role==="editor"}
              onArchive={content=>void lifecycle("archive",content)}
              onRestore={content=>void lifecycle("restore",content)}
              onDelete={content=>void lifecycle("delete",content)}
            />
          </section>
          {selected ? (
            <EditorPanel
              key={selected.id}
              workspace={user.WorkspaceID}
              content={selected}
              role={user.Role}
              publicSite={publicSite}
              onChanged={() => void load(user, query)}
              onContentChanged={setSelected}
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
