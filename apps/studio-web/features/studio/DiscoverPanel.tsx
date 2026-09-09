import { FormEvent, useState } from "react";
import { Search } from "lucide-react";
import { api, Content } from "../../lib/api";
import { ContentList } from "./ContentList";

export function DiscoverPanel({workspace,onOpen}:{workspace:string;onOpen:(content:Content)=>void}){
  const [query,setQuery]=useState(""),[locale,setLocale]=useState(""),[items,setItems]=useState<Content[]>([]),[searched,setSearched]=useState(false),[status,setStatus]=useState("");
  async function search(event:FormEvent){event.preventDefault();if(!query.trim())return;setStatus("正在搜索…");try{const result=await api.search(workspace,query,locale);setItems(result.items);setSearched(true);setStatus(`找到 ${result.items.length} 项内容`)}catch(error){setStatus(error instanceof Error?error.message:"搜索失败")}}
  return <section className="discover-page"><header className="page-heading"><div><span className="eyebrow">DISCOVER</span><h1>发现内容</h1><p>搜索标题、摘要和正文，可限定语言。</p></div></header><form className="discover-search" onSubmit={search}><Search aria-hidden/><input autoFocus value={query} onChange={e=>setQuery(e.target.value)} placeholder="输入标题、正文或关键词…"/><select aria-label="搜索语言" value={locale} onChange={e=>setLocale(e.target.value)}><option value="">所有语言</option><option value="zh-CN">中文</option><option value="en">English</option></select><button className="primary">搜索</button></form><p className="discover-status" role="status">{status}</p><div className="discover-results">{searched?<ContentList items={items} searching onSelect={onOpen}/>:<div className="discover-empty"><Search aria-hidden/><h2>从内容中找回需要的信息</h2><p>输入关键词后会同时检索标题、摘要和正文。</p></div>}</div></section>;
}
