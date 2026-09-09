import { useState } from "react";
import { Link2, NotebookPen } from "lucide-react";
import { api, InboxItem } from "../../lib/api";

export function CaptureForm({workspace,onCaptured}:{workspace:string;onCaptured:(item:InboxItem)=>void}) {
  const [kind,setKind]=useState<"text"|"link">("text"),[raw,setRaw]=useState(""),[url,setURL]=useState(""),[busy,setBusy]=useState(false),[error,setError]=useState("");
  async function submit(event:React.FormEvent){
    event.preventDefault(); setBusy(true); setError("");
    try { const item=await api.captureInbox(workspace,{kind,raw_text:raw,source_url:kind==="link"?url:undefined}); onCaptured(item); setRaw(""); setURL(""); }
    catch(err){setError(err instanceof Error?err.message:"捕获失败，请重试")} finally{setBusy(false)}
  }
  return <form className="capture-card" onSubmit={e=>void submit(e)}>
    <div className="capture-heading"><div><span className="eyebrow">QUICK CAPTURE</span><h2>先记下来，稍后整理</h2></div><div className="capture-kind" aria-label="捕获类型">
      <button type="button" aria-pressed={kind==="text"} onClick={()=>setKind("text")}><NotebookPen aria-hidden size={17}/>文字</button>
      <button type="button" aria-pressed={kind==="link"} onClick={()=>setKind("link")}><Link2 aria-hidden size={17}/>链接</button>
    </div></div>
    {kind==="link"&&<label><span>链接地址</span><input type="url" required value={url} onChange={e=>setURL(e.target.value)} placeholder="https://example.com/article"/></label>}
    <label><span>{kind==="text"?"想法或摘录":"备注（可选）"}</span><textarea required={kind==="text"} value={raw} onChange={e=>setRaw(e.target.value)} placeholder={kind==="text"?"输入一段想法、待办或摘录…":"为什么保存这个链接？"}/></label>
    <div className="capture-footer"><span className="form-error" role="alert">{error}</span><button className="primary compact" disabled={busy}>{busy?"保存中…":"放入收件箱"}</button></div>
  </form>
}
