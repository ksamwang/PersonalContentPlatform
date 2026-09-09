import { FileText } from "lucide-react";
import { SearchHit } from "../../lib/api";

export function SearchResults({items,onOpen}:{items:SearchHit[];onOpen:(id:string)=>void}) {
  if (!items.length) return <div className="discover-empty"><FileText aria-hidden/><h2>没有匹配内容</h2><p>可以换个表达，或重建索引后再试。</p></div>;
  return <div className="hybrid-results">{items.map(item=><button key={`${item.content_id}-${item.locale}`} onClick={()=>onOpen(item.content_id)}>
    <span className="hybrid-result-meta"><b>{item.type}</b><span>{item.locale}</span><span>相关度 {(item.score*100).toFixed(0)}%</span></span>
    <strong>{item.title||"未命名内容"}</strong>
    {item.summary&&<small>{item.summary}</small>}
    <p>{item.excerpt}</p>
  </button>)}</div>;
}
