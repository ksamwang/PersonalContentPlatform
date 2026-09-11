import { FormEvent, useState } from "react";
import { RefreshCw, Search } from "lucide-react";
import { api, Content, SearchHit } from "../../lib/api";
import { RagPanel } from "./RagPanel";
import { SearchResults } from "./SearchResults";
import { WorkspacePage } from "./WorkspacePage";

export function DiscoverPanel({workspace,role,onOpen}:{workspace:string;role:string;onOpen:(content:Content)=>void}){
  const [query,setQuery]=useState(""),[locale,setLocale]=useState(""),[items,setItems]=useState<SearchHit[]>([]),[searched,setSearched]=useState(false),[status,setStatus]=useState(""),[indexing,setIndexing]=useState(false);
  const canEdit=role==="owner"||role==="editor";
  async function open(id:string){try{onOpen(await api.content(workspace,id))}catch(error){setStatus(error instanceof Error?error.message:"内容加载失败")}}
  async function search(event:FormEvent){event.preventDefault();if(!query.trim())return;setStatus("正在进行全文与语义混合检索…");try{const result=await api.hybridSearch(workspace,query.trim(),locale);setItems(result.items);setSearched(true);setStatus(`找到 ${result.items.length} 项相关内容`)}catch(error){setStatus(error instanceof Error?error.message:"搜索失败")}}
  async function rebuild(){setIndexing(true);setStatus("正在为最新内容重建索引…");try{const result=await api.rebuildSearchIndex(workspace);setStatus(`索引已更新，共 ${result.chunks} 个内容分块`)}catch(error){setStatus(error instanceof Error?error.message:"索引重建失败")}finally{setIndexing(false)}}
  return <WorkspacePage className="discover-page" eyebrow="DISCOVER" title="发现与问答" description="结合 PostgreSQL 全文检索与向量语义，找回分散在内容中的信息。" actions={canEdit&&<button className="secondary" disabled={indexing} onClick={()=>void rebuild()}><RefreshCw aria-hidden size={16}/>{indexing?"重建中…":"重建索引"}</button>}>
    <form className="discover-search" onSubmit={search}><Search aria-hidden/><input autoFocus value={query} onChange={e=>setQuery(e.target.value)} placeholder="输入标题、正文、概念或自然语言问题…"/><select aria-label="搜索语言" value={locale} onChange={e=>setLocale(e.target.value)}><option value="">所有语言</option><option value="zh-CN">中文</option><option value="en">English</option></select><button className="primary">混合搜索</button></form>
    <p className="discover-status" role="status">{status}</p><div className="discover-layout"><div className="discover-results">{searched?<SearchResults items={items} onOpen={id=>void open(id)}/>:<div className="discover-empty"><Search aria-hidden/><h2>从内容中找回需要的信息</h2><p>首次使用请先配置 Embedding Provider 并重建索引。</p></div>}</div><RagPanel workspace={workspace} locale={locale} onOpen={id=>void open(id)}/></div>
  </WorkspacePage>;
}
