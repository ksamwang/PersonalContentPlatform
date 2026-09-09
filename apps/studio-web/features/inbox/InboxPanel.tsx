import { useCallback, useEffect, useState } from "react";
import { Archive, FileText, Inbox as InboxIcon, Link2, LoaderCircle, NotebookPen, Sparkles } from "lucide-react";
import { api, InboxItem } from "../../lib/api";
import { CaptureForm } from "./CaptureForm";
import { ConvertDialog } from "./ConvertDialog";

export function InboxPanel({workspace,role,onOpenContent}:{workspace:string;role:string;onOpenContent:(id:string)=>void}){
  const [items,setItems]=useState<InboxItem[]>([]),[state,setState]=useState("pending"),[loading,setLoading]=useState(true),[error,setError]=useState(""),[converting,setConverting]=useState<InboxItem>();
  const canEdit=role==="owner"||role==="editor";
  const load=useCallback(async()=>{setLoading(true);setError("");try{setItems((await api.inbox(workspace,state)).items)}catch(err){setError(err instanceof Error?err.message:"收件箱加载失败")}finally{setLoading(false)}},[workspace,state]);
  useEffect(()=>{void load()},[load]);
  async function archive(item:InboxItem){if(!window.confirm("确认归档这条收集内容？归档后不会出现在待整理列表。"))return;try{await api.archiveInbox(workspace,item.id);await load()}catch(err){setError(err instanceof Error?err.message:"归档失败")}}
  async function process(item:InboxItem){setItems(v=>v.map(candidate=>candidate.id===item.id?{...candidate,processing_state:"processing"}:candidate));try{const saved=await api.processInbox(workspace,item.id);setItems(v=>v.map(candidate=>candidate.id===saved.id?saved:candidate))}catch(err){setError(err instanceof Error?err.message:"自动处理失败");await load()}}
  return <section className="inbox-page">
    <header className="page-heading"><div><span className="eyebrow">INBOX</span><h1>收件箱</h1><p>快速捕获零散想法和链接，再把值得保留的内容整理成草稿。</p></div><label className="state-filter"><span>显示状态</span><select value={state} onChange={e=>setState(e.target.value)}><option value="pending">待整理</option><option value="converted">已转换</option><option value="archived">已归档</option></select></label></header>
    {canEdit&&<CaptureForm workspace={workspace} onCaptured={item=>{if(state==="pending")setItems(v=>[item,...v])}}/>}
    {error&&<div className="inline-error" role="alert"><span>{error}</span><button onClick={()=>void load()}>重试</button></div>}
    <div className="inbox-list" aria-busy={loading}>
      {loading?<div className="inbox-empty">正在加载收件箱…</div>:items.length===0?<div className="inbox-empty"><InboxIcon aria-hidden size={36}/><h2>这里还没有内容</h2><p>{state==="pending"?"把刚出现的灵感先放进来，不必立即整理。":"当前状态下没有收集记录。"}</p></div>:items.map(item=><article className="inbox-row" key={item.id}>
        <span className="type-icon">{item.kind==="link"?<Link2 aria-hidden size={18}/>:item.kind==="text"?<NotebookPen aria-hidden size={18}/>:<FileText aria-hidden size={18}/>}</span><div className="inbox-copy"><strong>{item.title||item.raw_text||item.source_url}</strong>{item.source_url&&<a href={item.source_url} target="_blank" rel="noreferrer">{item.source_url}</a>}{item.extracted_text&&<p>{item.extracted_text.slice(0,220)}{item.extracted_text.length>220?"…":""}</p>}{item.duplicate_of&&<small className="duplicate-note">可能与已有收集内容重复</small>}{item.processing_error&&<small className="form-error">{item.processing_error}</small>}<small>{new Date(item.created_at).toLocaleString("zh-CN")}</small></div>
        <div className="inbox-actions">{item.converted_content_id&&<button className="secondary compact" onClick={()=>onOpenContent(item.converted_content_id!)}>打开内容</button>}{canEdit&&item.state==="pending"&&<>{item.kind!=="text"&&item.kind!=="file"&&<button className="secondary compact" disabled={item.processing_state==="processing"} onClick={()=>void process(item)}>{item.processing_state==="processing"?<LoaderCircle className="spin" aria-hidden size={15}/>:<Sparkles aria-hidden size={15}/>}智能处理</button>}<button className="secondary compact" onClick={()=>setConverting(item)}>整理</button><button className="icon-button" aria-label="归档" title="归档" onClick={()=>void archive(item)}><Archive aria-hidden size={18}/></button></>}</div>
      </article>)}
    </div>
    {converting&&<ConvertDialog workspace={workspace} item={converting} onClose={()=>setConverting(undefined)} onConverted={id=>{setConverting(undefined);onOpenContent(id)}}/>}
  </section>
}
