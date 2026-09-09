import { useEffect, useState } from "react";
import { History, RotateCcw, X } from "lucide-react";
import { api, Draft, Revision } from "../../lib/api";

export function RevisionPanel({workspace,localizationID,draftVersion,canEdit,refresh,onClose,onRestored}:{workspace:string;localizationID:string;draftVersion:number;canEdit:boolean;refresh:number;onClose:()=>void;onRestored:(draft:Draft)=>void}){
 const [items,setItems]=useState<Revision[]>([]),[loading,setLoading]=useState(true),[error,setError]=useState(""),[busy,setBusy]=useState("");
 useEffect(()=>{setLoading(true);api.revisions(workspace,localizationID).then(v=>{setItems(v.items);setError("")}).catch(err=>setError(err instanceof Error?err.message:"历史版本加载失败")).finally(()=>setLoading(false))},[workspace,localizationID,refresh]);
 async function restore(revision:Revision){if(!window.confirm(`确认将版本 ${revision.seq} 恢复到当前草稿？当前未封存的草稿内容会被替换。`))return;setBusy(revision.id);setError("");try{onRestored(await api.restoreRevision(workspace,localizationID,revision.id,draftVersion));onClose()}catch(err){setError(err instanceof Error?err.message:"恢复失败，请重新加载草稿后重试")}finally{setBusy("")}}
 return <aside className="revision-panel" aria-label="版本历史"><header><div><History aria-hidden/><div><h2>版本历史</h2><p>已封存版本不可修改</p></div></div><button className="icon-button" aria-label="关闭版本历史" onClick={onClose}><X/></button></header>{error&&<p className="form-error" role="alert">{error}</p>}{loading?<p className="revision-loading">正在加载历史版本…</p>:items.length===0?<p className="revision-loading">尚未封存任何版本。</p>:<ol>{items.map(item=><li key={item.id}><div><strong>版本 {item.seq} · {item.title||"无标题"}</strong><span>{new Date(item.created_at).toLocaleString("zh-CN")}</span><p>{item.summary||"无摘要"}</p></div>{canEdit&&<button className="secondary compact" disabled={busy!==""} onClick={()=>void restore(item)}><RotateCcw aria-hidden size={16}/>{busy===item.id?"恢复中…":"恢复到草稿"}</button>}</li>)}</ol>}</aside>
}
