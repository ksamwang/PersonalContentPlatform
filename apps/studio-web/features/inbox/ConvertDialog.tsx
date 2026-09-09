import { useState } from "react";
import { X } from "lucide-react";
import { api, InboxItem } from "../../lib/api";

function suggestedTitle(item:InboxItem){
  if(item.raw_text.trim()) return item.raw_text.trim().split(/\r?\n/)[0].slice(0,60);
  try{return new URL(item.source_url??"").hostname}catch{return "收件箱内容"}
}
export function ConvertDialog({workspace,item,onClose,onConverted}:{workspace:string;item:InboxItem;onClose:()=>void;onConverted:(contentID:string)=>void}){
  const [title,setTitle]=useState(suggestedTitle(item)),[slug,setSlug]=useState(""),[type,setType]=useState<"article"|"note"|"page">("note"),[locale,setLocale]=useState("zh-CN"),[busy,setBusy]=useState(false),[error,setError]=useState("");
  async function submit(event:React.FormEvent){event.preventDefault();setBusy(true);setError("");try{const result=await api.convertInbox(workspace,item.id,{type,locale,slug,title});onConverted(result.content_id)}catch(err){setError(err instanceof Error?err.message:"转换失败，请重试")}finally{setBusy(false)}}
  return <div className="modal-backdrop" role="presentation"><section className="modal" role="dialog" aria-modal="true" aria-labelledby="convert-title">
    <button className="icon-button close" aria-label="关闭" onClick={onClose}><X aria-hidden/></button><h2 id="convert-title">整理为内容</h2><p>原始收集内容会复制到草稿中，收件箱记录将保留为已转换状态。</p>
    <form onSubmit={e=>void submit(e)}>
      <label><span>标题</span><input required value={title} onChange={e=>setTitle(e.target.value)}/></label>
      <label><span>固定链接</span><input required pattern="[a-z0-9]+(?:-[a-z0-9]+)*" value={slug} onChange={e=>setSlug(e.target.value.toLowerCase())} placeholder="my-note"/><small>仅使用小写字母、数字和连字符。</small></label>
      <label><span>内容类型</span><select value={type} onChange={e=>setType(e.target.value as typeof type)}><option value="note">笔记</option><option value="article">文章</option><option value="page">页面</option></select></label>
      <label><span>语言</span><select value={locale} onChange={e=>setLocale(e.target.value)}><option value="zh-CN">简体中文</option><option value="en">English</option></select></label>
      {error&&<p className="form-error" role="alert">{error}</p>}<button className="primary" disabled={busy}>{busy?"转换中…":"创建草稿"}</button>
    </form>
  </section></div>
}
