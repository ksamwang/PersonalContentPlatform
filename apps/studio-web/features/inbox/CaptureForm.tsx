import { useState } from "react";
import { FileUp, Link2, NotebookPen } from "lucide-react";
import { api, InboxItem } from "../../lib/api";

export function CaptureForm({workspace,onCaptured}:{workspace:string;onCaptured:(item:InboxItem)=>void}) {
  const [kind,setKind]=useState<"text"|"link"|"file">("text"),[raw,setRaw]=useState(""),[url,setURL]=useState(""),[file,setFile]=useState<File>(),[busy,setBusy]=useState(false),[error,setError]=useState("");
  async function submit(event:React.FormEvent){
    event.preventDefault(); setBusy(true); setError("");
    try { let assetID:string|undefined,captureKind:"text"|"link"|"image"|"audio"|"file"=kind;if(kind==="file"){if(!file)throw new Error("请选择文件");const asset=await api.uploadAsset(workspace,file);assetID=asset.id;captureKind=file.type.startsWith("image/")?"image":file.type.startsWith("audio/")?"audio":"file"}const item=await api.captureInbox(workspace,{kind:captureKind,raw_text:kind==="file"?(raw||file?.name||""):raw,source_url:kind==="link"?url:undefined,asset_id:assetID}); onCaptured(item); setRaw(""); setURL("");setFile(undefined); }
    catch(err){setError(err instanceof Error?err.message:"捕获失败，请重试")} finally{setBusy(false)}
  }
  return <form className="capture-card" onSubmit={e=>void submit(e)}>
    <div className="capture-heading"><div><span className="eyebrow">QUICK CAPTURE</span><h2>先记下来，稍后整理</h2></div><div className="capture-kind" aria-label="捕获类型">
      <button type="button" aria-pressed={kind==="text"} onClick={()=>setKind("text")}><NotebookPen aria-hidden size={17}/>文字</button>
      <button type="button" aria-pressed={kind==="link"} onClick={()=>setKind("link")}><Link2 aria-hidden size={17}/>链接</button>
      <button type="button" aria-pressed={kind==="file"} onClick={()=>setKind("file")}><FileUp aria-hidden size={17}/>文件</button>
    </div></div>
    {kind==="link"&&<label><span>链接地址</span><input type="url" required value={url} onChange={e=>setURL(e.target.value)} placeholder="https://example.com/article"/></label>}
    {kind==="file"&&<label><span>图片、音频或其他文件</span><input type="file" required onChange={e=>setFile(e.target.files?.[0])}/></label>}
    <label><span>{kind==="text"?"想法或摘录":"备注（可选）"}</span><textarea required={kind==="text"} value={raw} onChange={e=>setRaw(e.target.value)} placeholder={kind==="text"?"输入一段想法、待办或摘录…":kind==="link"?"为什么保存这个链接？":"补充文件说明…"}/></label>
    <div className="capture-footer"><span className="form-error" role="alert">{error}</span><button className="primary compact" disabled={busy}>{busy?"保存中…":"放入收件箱"}</button></div>
  </form>
}
