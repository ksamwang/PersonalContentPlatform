export type ContentFilter = {type:string;locale:string;state:string;visibility:string;tag:string;sort:"updated"|"title"};
export const emptyContentFilter:ContentFilter={type:"",locale:"",state:"",visibility:"",tag:"",sort:"updated"};

export function ContentFilters({value,onChange}:{value:ContentFilter;onChange:(value:ContentFilter)=>void}){
  const set=(key:keyof ContentFilter,next:string)=>onChange({...value,[key]:next});
  return <div className="content-filters" aria-label="内容筛选">
    <div className="filter-primary">
      <select aria-label="内容类型" value={value.type} onChange={e=>set("type",e.target.value)}><option value="">全部类型</option><option value="article">文章</option><option value="note">笔记</option><option value="page">页面</option></select>
      <select aria-label="内容语言" value={value.locale} onChange={e=>set("locale",e.target.value)}><option value="">全部语言</option><option value="zh-CN">中文</option><option value="en">English</option></select>
      <select aria-label="发布状态" value={value.state} onChange={e=>set("state",e.target.value)}><option value="">全部状态</option><option value="draft">草稿</option><option value="ready">待发布</option><option value="published">已发布</option><option value="archived">已归档</option></select>
    </div>
    <details className="filter-more">
      <summary><SlidersHorizontal aria-hidden size={17}/><span>更多筛选</span></summary>
      <div>
        <select aria-label="可见性" value={value.visibility} onChange={e=>set("visibility",e.target.value)}><option value="">全部可见性</option><option value="private">私有</option><option value="unlisted">不公开列出</option><option value="public">公开</option></select>
        <input aria-label="标签筛选" value={value.tag} onChange={e=>set("tag",e.target.value)} placeholder="标签"/>
        <select aria-label="内容排序" value={value.sort} onChange={e=>set("sort",e.target.value)}><option value="updated">最近更新</option><option value="title">按标题</option></select>
      </div>
    </details>
  </div>;
}
import { SlidersHorizontal } from "lucide-react";
