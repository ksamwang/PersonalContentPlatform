"use client";
import { useEffect, useRef, useState } from "react";
import { useEditor, EditorContent } from "@tiptap/react";
import StarterKit from "@tiptap/starter-kit";
import { Bold, Eye, Heading2, History, Italic, List, Quote, Save, Send } from "lucide-react";
import { api, Content, Draft } from "../../lib/api";
import { RevisionPanel } from "./RevisionPanel";
export function EditorPanel({
  workspace,
  content,
  onChanged,
  role,
  publicSite,
}: {
  workspace: string;
  content: Content;
  onChanged: () => void;
  role: string;
  publicSite: string;
}) {
  const localization = content.localizations[0];
  const canEdit = role === "owner" || role === "editor";
  const [draft, setDraft] = useState<Draft | null>(null),
    [title, setTitle] = useState(""),
    [summary, setSummary] = useState(""),
    [status, setStatus] = useState("加载中"),
    [busy, setBusy] = useState(false),
    [loadError, setLoadError] = useState(""),
    [reloadToken, setReloadToken] = useState(0),
    [historyOpen,setHistoryOpen]=useState(false),
    [revisionRefresh,setRevisionRefresh]=useState(0);
  const timer = useRef<ReturnType<typeof setTimeout> | null>(null);
  const draftRef = useRef<Draft | null>(null);
  const titleRef = useRef("");
  const summaryRef = useRef("");
  const localizationRef = useRef(localization.id);
  const saveQueue = useRef<Promise<Draft | null>>(Promise.resolve(null));
  localizationRef.current = localization.id;
  const editor = useEditor({
    extensions: [StarterKit],
    content: "",
    immediatelyRender: false,
    editable: canEdit,
    onUpdate: () => scheduleSave(),
  });
  useEffect(()=>{editor?.setEditable(canEdit)},[editor,canEdit]);
  useEffect(() => {
    let active = true;
    setDraft(null);
    setLoadError("");
    api
      .draft(workspace, localization.id)
      .then((d) => {
        if (!active) return;
        setDraft(d);
        draftRef.current = d;
        setTitle(d.title);
        titleRef.current = d.title;
        setSummary(d.summary);
        summaryRef.current = d.summary;
        editor?.commands.setContent(d.body, { emitUpdate: false });
        setStatus("已保存");
      })
      .catch((err) => {
        if (!active) return;
        const message = err instanceof Error ? err.message : "请检查服务后重试";
        setLoadError(message);
        setStatus("草稿加载失败");
      });
    return () => {
      active = false;
      if (timer.current) clearTimeout(timer.current);
    };
  }, [workspace, localization.id, editor, reloadToken]);
  function scheduleSave() {
    if (!canEdit) return;
    setStatus("等待保存");
    if (timer.current) clearTimeout(timer.current);
    timer.current = setTimeout(() => void save(), 900);
  }
  async function performSave(
    targetLocalizationID: string,
    values: Pick<Draft, "title" | "summary" | "body" | "metadata">,
  ) {
    if (localizationRef.current !== targetLocalizationID || !draftRef.current)
      return null;
    setStatus("正在保存");
    try {
      const next = await api.saveDraft(workspace, targetLocalizationID, {
        version: draftRef.current.version,
        ...values,
      });
      if (localizationRef.current !== targetLocalizationID) return null;
      setDraft(next);
      draftRef.current = next;
      setStatus("已保存");
      return next;
    } catch (err) {
      setStatus(
        err instanceof Error ? `保存失败：${err.message}` : "保存失败，请重试",
      );
      return null;
    }
  }
  function save() {
    if (!canEdit || !draftRef.current || !editor) return Promise.resolve(null);
    const targetLocalizationID = localization.id;
    const values = {
      title: titleRef.current,
      summary: summaryRef.current,
      body: editor.getJSON(),
      metadata: draftRef.current.metadata,
    };
    saveQueue.current = saveQueue.current.then(() =>
      performSave(targetLocalizationID, values),
    );
    return saveQueue.current;
  }
  async function publish() {
    setBusy(true);
    if (timer.current) clearTimeout(timer.current);
    try {
      const saved = await save();
      if (!saved) return;
      await api.seal(workspace, localization.id, saved.version);
      setRevisionRefresh(v=>v+1);
      await api.ready(workspace, localization.id);
      await api.publish(workspace, content.id, localization.locale);
      setStatus("已进入发布队列");
      const published = await waitUntilPublished();
      setStatus(
        published
          ? "发布成功"
          : "已排队；若长时间未发布，请确认 Worker 已启动",
      );
      onChanged();
    } catch (err) {
      setStatus(err instanceof Error ? err.message : "发布失败");
    } finally {
      setBusy(false);
    }
  }
  async function preview(){
    setBusy(true); if(timer.current)clearTimeout(timer.current);
    try{const saved=await save();if(!saved)return;const token=await api.preview(workspace,localization.id);const previewOrigin=new URL(publicSite,window.location.origin).origin;window.open(`${previewOrigin}/preview/${token.token}`,"_blank","noopener,noreferrer");setStatus("预览链接已生成，30 分钟内有效")}
    catch(err){setStatus(err instanceof Error?err.message:"预览生成失败")}finally{setBusy(false)}
  }
  function applyRestoredDraft(next:Draft){setDraft(next);draftRef.current=next;setTitle(next.title);titleRef.current=next.title;setSummary(next.summary);summaryRef.current=next.summary;editor?.commands.setContent(next.body,{emitUpdate:false});setStatus("旧版本已恢复到草稿")}
  async function waitUntilPublished() {
    for (let attempt = 0; attempt < 20; attempt += 1) {
      await new Promise((resolve) => setTimeout(resolve, 500));
      const latest = await api.content(workspace, content.id);
      const current = latest.localizations.find(
        (item) => item.id === localization.id,
      );
      if (current?.state === "published") return true;
    }
    return false;
  }
  if (!draft && loadError)
    return (
      <section className="editor-panel welcome-panel" role="alert">
        <h2>草稿加载失败</h2>
        <p>{loadError}</p>
        <button className="secondary" onClick={() => setReloadToken((v) => v + 1)}>
          重新加载
        </button>
      </section>
    );
  if (!draft)
    return <section className="editor-panel skeleton" aria-busy="true" />;
  return (
    <section className="editor-panel">
      <header className="editor-header">
        <div>
          <span className="eyebrow">
            {content.type} · {localization.locale}
          </span>
          <input
            className="title-input"
            aria-label="内容标题"
            value={title}
            disabled={!canEdit}
            onChange={(e) => {
              setTitle(e.target.value);
              titleRef.current = e.target.value;
              scheduleSave();
            }}
          />
        </div>
        <div className="editor-actions">
          <span className="save-state" aria-live="polite">
            {status}
          </span>
          <button className="secondary" onClick={()=>setHistoryOpen(v=>!v)}><History aria-hidden size={17}/>版本</button>
          <button className="secondary" disabled={busy||!canEdit} onClick={() => void preview()}><Eye aria-hidden size={17}/>预览</button>
          <button className="secondary" disabled={busy||!canEdit} onClick={() => void save()}>
            <Save aria-hidden size={17} />
            保存草稿
          </button>
          <button className="primary compact" disabled={busy||!canEdit} onClick={publish}>
            <Send aria-hidden size={17} />
            {busy ? "发布中" : "发布"}
          </button>
        </div>
      </header>
      <input
        className="summary-input"
        aria-label="内容摘要"
        placeholder="添加一段简洁摘要…"
        value={summary}
        disabled={!canEdit}
        onChange={(e) => {
          setSummary(e.target.value);
          summaryRef.current = e.target.value;
          scheduleSave();
        }}
      />
      <div className="toolbar" aria-label="编辑工具栏">
        <button
          aria-label="粗体"
          aria-pressed={editor?.isActive("bold")}
          onClick={() => editor?.chain().focus().toggleBold().run()}
        >
          <Bold />
        </button>
        <button
          aria-label="斜体"
          aria-pressed={editor?.isActive("italic")}
          onClick={() => editor?.chain().focus().toggleItalic().run()}
        >
          <Italic />
        </button>
        <button
          aria-label="二级标题"
          aria-pressed={editor?.isActive("heading", { level: 2 })}
          onClick={() =>
            editor?.chain().focus().toggleHeading({ level: 2 }).run()
          }
        >
          <Heading2 />
        </button>
        <button
          aria-label="项目列表"
          onClick={() => editor?.chain().focus().toggleBulletList().run()}
        >
          <List />
        </button>
        <button
          aria-label="引用"
          onClick={() => editor?.chain().focus().toggleBlockquote().run()}
        >
          <Quote />
        </button>
      </div>
      <EditorContent editor={editor} className="prose-editor" />
      {historyOpen&&<RevisionPanel workspace={workspace} localizationID={localization.id} draftVersion={draft.version} canEdit={canEdit} refresh={revisionRefresh} onClose={()=>setHistoryOpen(false)} onRestored={applyRestoredDraft}/>}
    </section>
  );
}
