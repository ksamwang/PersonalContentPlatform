"use client";
import { useCallback, useEffect, useState } from "react";
import { Search } from "lucide-react";
import { api, Content, Principal } from "../../lib/api";
import { AuthPanel } from "../auth/AuthPanel";
import { AddPasskey } from "../auth/PasskeyButton";
import { AssetUpload } from "../assets/AssetUpload";
import { EditorPanel } from "../editor/EditorPanel";
import { ContentList } from "./ContentList";
import { CreateContent } from "./CreateContent";
import { Sidebar } from "./Sidebar";
export function StudioApp() {
  const [user, setUser] = useState<Principal | null | undefined>(undefined),
    [items, setItems] = useState<Content[]>([]),
    [selected, setSelected] = useState<Content>(),
    [query, setQuery] = useState("");
  const load = useCallback(
    async (p = user, q = query) => {
      if (!p) return;
      const result = q
        ? await api.search(p.WorkspaceID, q)
        : await api.list(p.WorkspaceID);
      setItems(result.items);
      if (selected) {
        setSelected(result.items.find((i) => i.id === selected.id) ?? selected);
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
      <Sidebar />
      <main id="main" className="workspace">
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
            <ContentList
              items={items}
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
      </main>
    </div>
  );
}
