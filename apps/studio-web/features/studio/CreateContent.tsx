"use client";
import { FormEvent, useState } from "react";
import { Plus, X } from "lucide-react";
import { api, Content } from "../../lib/api";
export function CreateContent({
  workspace,
  onCreated,
}: {
  workspace: string;
  onCreated: (c: Content) => void;
}) {
  const [open, setOpen] = useState(false),
    [busy, setBusy] = useState(false),
    [error, setError] = useState("");
  async function submit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setBusy(true);
    setError("");
    const d = new FormData(e.currentTarget);
    try {
      const c = await api.create(workspace, {
        type: d.get("type"),
        locale: d.get("locale"),
        slug: d.get("slug"),
        title: d.get("title"),
      });
      onCreated(c);
      setOpen(false);
    } catch (err) {
      setError(err instanceof Error ? err.message : "创建失败");
    } finally {
      setBusy(false);
    }
  }
  return (
    <>
      {
        <button className="primary compact" onClick={() => setOpen(true)}>
          <Plus aria-hidden size={18} />
          新建内容
        </button>
      }
      {open && (
        <div className="modal-backdrop" role="presentation">
          <section
            className="modal"
            role="dialog"
            aria-modal="true"
            aria-labelledby="create-title"
          >
            <button
              className="icon-button close"
              aria-label="关闭"
              onClick={() => setOpen(false)}
            >
              <X aria-hidden />
            </button>
            <h2 id="create-title">创建新内容</h2>
            <p>先建立内容身份，正文会自动保存为草稿。</p>
            <form onSubmit={submit}>
              <label>
                类型
                <select name="type">
                  <option value="article">文章</option>
                  <option value="note">笔记</option>
                  <option value="page">页面</option>
                </select>
              </label>
              <label>
                语言
                <select name="locale">
                  <option value="zh-CN">中文</option>
                  <option value="en">English</option>
                </select>
              </label>
              <label>
                标题
                <input name="title" required autoFocus />
              </label>
              <label>
                URL Slug
                <input
                  name="slug"
                  pattern="[a-z0-9]+(?:-[a-z0-9]+)*"
                  required
                />
                <small>使用小写字母、数字和连字符</small>
              </label>
              {error && (
                <p role="alert" className="form-error">
                  {error}
                </p>
              )}
              <button className="primary" disabled={busy}>
                {busy ? "创建中…" : "创建并编辑"}
              </button>
            </form>
          </section>
        </div>
      )}
    </>
  );
}
