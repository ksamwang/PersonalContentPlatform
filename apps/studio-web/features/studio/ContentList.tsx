import { Archive, ArchiveRestore, FileText, StickyNote, PanelTop, Trash2 } from "lucide-react";
import { Content } from "../../lib/api";
const icons = { article: FileText, note: StickyNote, page: PanelTop };
const stateNames:Record<string,string>={draft:"草稿",ready:"待发布",published:"已发布",archived:"已归档"};
export function ContentList({
  items,
  selected,
  onSelect,
  searching,
  canEdit,
  onArchive,
  onRestore,
  onDelete,
}: {
  items: Content[];
  selected?: string;
  onSelect: (c: Content) => void;
  searching?: boolean;
  canEdit?: boolean;
  onArchive?: (content:Content)=>void;
  onRestore?: (content:Content)=>void;
  onDelete?: (content:Content)=>void;
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
        const archived=item.localizations.every(value=>value.state==="archived");
        return (
          <article key={item.id} className={selected === item.id ? "content-row selected" : "content-row"}>
          <button className="content-row-open" onClick={() => onSelect(item)}>
            <span className="type-icon">
              <Icon aria-hidden size={18} />
            </span>
            <span className="row-main">
              <strong title={revision?.title || "无标题"}>{revision?.title || "无标题"}</strong>
              <span className="row-meta">
                <small title={locale?.slug}>{locale?.slug || "未设置链接"} · {locale?.locale}</small>
                <span className={`status ${locale?.state}`}>{stateNames[locale?.state]??locale?.state}</span>
              </span>
            </span>
          </button>
          {canEdit&&<span className="content-row-actions">{archived?<button aria-label={`恢复 ${revision?.title||"无标题"}`} title="恢复" onClick={()=>onRestore?.(item)}><ArchiveRestore/></button>:<button aria-label={`归档 ${revision?.title||"无标题"}`} title="归档" onClick={()=>onArchive?.(item)}><Archive/></button>}<button className="danger-icon" aria-label={`删除 ${revision?.title||"无标题"}`} title="移至回收站" onClick={()=>onDelete?.(item)}><Trash2/></button></span>}
          </article>
        );
      })}
    </div>
  );
}
