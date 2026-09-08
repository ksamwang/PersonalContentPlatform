import {
  BookOpen,
  Inbox,
  Library,
  Search,
  Settings2,
  Shapes,
} from "lucide-react";
const items = [
  [Library, "内容"],
  [Inbox, "收件箱"],
  [Shapes, "合集"],
  [Search, "发现"],
  [BookOpen, "发布"],
] as const;
export function Sidebar() {
  return (
    <aside className="sidebar">
      <div className="brand">
        <span className="brand-mark">P</span>
        <span>Content Studio</span>
      </div>
      <nav aria-label="主要导航">
        {items.map(([Icon, label], i) => (
          <button className={i === 0 ? "active" : ""} key={label}>
            <Icon aria-hidden size={19} />
            <span>{label}</span>
          </button>
        ))}
      </nav>
      <button className="settings">
        <Settings2 aria-hidden size={19} />
        <span>设置</span>
      </button>
    </aside>
  );
}
