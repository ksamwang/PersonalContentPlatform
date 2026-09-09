import { useState } from "react";
import { X } from "lucide-react";
import { api, Collection } from "../../lib/api";

export function CreateCollectionDialog({workspace,onClose,onCreated}:{workspace:string;onClose:()=>void;onCreated:(value:Collection)=>void}){
 const [title,setTitle]=useState(""),[slug,setSlug]=useState(""),[visibility,setVisibility]=useState("private"),[busy,setBusy]=useState(false),[error,setError]=useState("");
 async function submit(e:React.FormEvent){e.preventDefault();setBusy(true);setError("");try{onCreated(await api.createCollection(workspace,{title,slug,visibility}))}catch(err){setError(err instanceof Error?err.message:"创建合集失败")}finally{setBusy(false)}}
 return <div className="modal-backdrop" role="presentation"><section className="modal" role="dialog" aria-modal="true" aria-labelledby="collection-create-title"><button className="icon-button close" aria-label="关闭" onClick={onClose}><X/></button><h2 id="collection-create-title">创建合集</h2><p>用分区和明确顺序组织一组相关内容。</p><form onSubmit={e=>void submit(e)}>
  <label><span>合集名称</span><input required value={title} onChange={e=>setTitle(e.target.value)}/></label><label><span>固定链接</span><input required pattern="[a-z0-9]+(?:-[a-z0-9]+)*" placeholder="reading-list" value={slug} onChange={e=>setSlug(e.target.value.toLowerCase())}/></label><label><span>可见性</span><select value={visibility} onChange={e=>setVisibility(e.target.value)}><option value="private">私有</option><option value="unlisted">不公开列出</option><option value="public">公开</option></select></label>{error&&<p className="form-error" role="alert">{error}</p>}<button className="primary" disabled={busy}>{busy?"创建中…":"创建合集"}</button>
 </form></section></div>
}
