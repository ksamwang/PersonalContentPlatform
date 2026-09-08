import { FileText, StickyNote, PanelTop } from "lucide-react";
import { Content } from "../../lib/api";
const icons = { article: FileText, note: StickyNote, page: PanelTop };
export function ContentList({
  items,
  selected,
  onSelect,
  searching,
}: {
  items: Content[];
  selected?: string;
  onSelect: (c: Content) => void;
  searching?: boolean;
}) {
  if (!items.length)
    return (
      <div className="empty">
        <FileText aria-hidden />
        <h2>{searching ? "没有匹配内容" : "还没有内容"}</h2>
        <p>
          {searching
            ? "请尝试其他关键词。"
            : "创建第一篇内容，开始建立你的长期内容资产。"}
        </p>
      </div>
    );
  return (
    <div className="content-list">
      {items.map((item) => {
        const locale = item.localizations[0],
          revision = locale?.current_revision,
          Icon = icons[item.type];
        return (
          <button
            key={item.id}
            className={
              selected === item.id ? "content-row selected" : "content-row"
            }
            onClick={() => onSelect(item)}
          >
            <span className="type-icon">
              <Icon aria-hidden size={18} />
            </span>
            <span className="row-main">
              <strong>{revision?.title || "无标题"}</strong>
              <small>
                {locale?.slug} · {locale?.locale}
              </small>
            </span>
            <span className={`status ${locale?.state}`}>{locale?.state}</span>
          </button>
        );
      })}
    </div>
  );
}
