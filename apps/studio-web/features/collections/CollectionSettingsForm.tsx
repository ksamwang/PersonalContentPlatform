import { useEffect, useState } from "react";
import { Save } from "lucide-react";
import { Collection } from "../../lib/api";

type Props = {
  value: Collection;
  busy: boolean;
  onSave: (input: Pick<Collection, "title" | "slug" | "visibility">) => Promise<void>;
};

export function CollectionSettingsForm({ value, busy, onSave }: Props) {
  const [title, setTitle] = useState(value.title);
  const [slug, setSlug] = useState(value.slug);
  const [visibility, setVisibility] = useState(value.visibility);

  useEffect(() => {
    setTitle(value.title);
    setSlug(value.slug);
    setVisibility(value.visibility);
  }, [value.id, value.title, value.slug, value.visibility]);

  const changed = title !== value.title || slug !== value.slug || visibility !== value.visibility;

  return (
    <form
      className="collection-settings"
      onSubmit={(event) => {
        event.preventDefault();
        void onSave({ title, slug, visibility });
      }}
    >
      <label>
        <span>合集名称</span>
        <input required value={title} onChange={(event) => setTitle(event.target.value)} />
      </label>
      <label>
        <span>固定链接</span>
        <input
          required
          pattern="[a-z0-9]+(?:-[a-z0-9]+)*"
          value={slug}
          onChange={(event) => setSlug(event.target.value.toLowerCase())}
        />
      </label>
      <label>
        <span>公开状态</span>
        <select value={visibility} onChange={(event) => setVisibility(event.target.value as Collection["visibility"])}>
          <option value="private">私有</option>
          <option value="unlisted">不公开列出</option>
          <option value="public">公开到网站</option>
        </select>
      </label>
      <button className="secondary compact" disabled={busy || !changed}>
        <Save aria-hidden size={16} />
        {busy ? "保存中…" : "保存合集设置"}
      </button>
      <p>只有“公开到网站”的合集会出现在 Public Web；合集内仍只展示已发布内容。</p>
    </form>
  );
}
