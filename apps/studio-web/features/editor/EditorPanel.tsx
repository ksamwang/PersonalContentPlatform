"use client";
import { useEffect, useRef, useState } from "react";
import { useEditor, EditorContent } from "@tiptap/react";
import StarterKit from "@tiptap/starter-kit";
import { Bold, Heading2, Italic, List, Quote, Save, Send } from "lucide-react";
import { api, Content, Draft } from "../../lib/api";
export function EditorPanel({
  workspace,
  content,
  onChanged,
}: {
  workspace: string;
  content: Content;
  onChanged: () => void;
}) {
  const localization = content.localizations[0];
  const [draft, setDraft] = useState<Draft | null>(null),
    [title, setTitle] = useState(""),
    [summary, setSummary] = useState(""),
    [status, setStatus] = useState("加载中"),
    [busy, setBusy] = useState(false);
  const timer = useRef<ReturnType<typeof setTimeout> | null>(null);
  const draftRef = useRef<Draft | null>(null);
  const titleRef = useRef("");
  const summaryRef = useRef("");
  const editor = useEditor({
    extensions: [StarterKit],
    content: "",
    immediatelyRender: false,
    onUpdate: () => scheduleSave(),
  });
  useEffect(() => {
    let active = true;
    setDraft(null);
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
        editor?.commands.setContent(d.body);
        setStatus("已保存");
      })
      .catch(() => setStatus("草稿加载失败"));
    return () => {
      active = false;
    };
  }, [workspace, localization.id, editor]);
  function scheduleSave() {
    setStatus("等待保存");
    if (timer.current) clearTimeout(timer.current);
    timer.current = setTimeout(() => void save(), 900);
  }
  async function save() {
    if (!draftRef.current || !editor) return null;
    setStatus("正在保存");
    try {
      const next = await api.saveDraft(workspace, localization.id, {
        ...draftRef.current,
        title: titleRef.current,
        summary: summaryRef.current,
        body: editor.getJSON(),
      });
      setDraft(next);
      draftRef.current = next;
      setStatus("已保存");
      return next;
    } catch {
      setStatus("保存冲突，请刷新");
      return null;
    }
  }
  async function publish() {
    setBusy(true);
    try {
      const saved = await save();
      if (!saved) return;
      await api.seal(workspace, localization.id, saved.version);
      await api.ready(workspace, localization.id);
      await api.publish(workspace, content.id, localization.locale);
      setStatus("已进入发布队列");
      onChanged();
    } catch (err) {
      setStatus(err instanceof Error ? err.message : "发布失败");
    } finally {
      setBusy(false);
    }
  }
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
          <button className="secondary" onClick={() => void save()}>
            <Save aria-hidden size={17} />
            保存版本
          </button>
          <button className="primary compact" disabled={busy} onClick={publish}>
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
    </section>
  );
}
