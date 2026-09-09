import { FormEvent, useEffect, useState } from "react";
import { Image as ImageIcon, X } from "lucide-react";
import { api, Asset } from "../../lib/api";

export function AssetPickerDialog({workspace,onInsert,onClose}:{workspace:string;onInsert:(asset:Asset,alt:string)=>void;onClose:()=>void}){
  const [items,setItems]=useState<Asset[]>([]),[selected,setSelected]=useState(""),[alt,setAlt]=useState(""),[status,setStatus]=useState("正在加载资产…");
  useEffect(()=>{void api.assets(workspace).then(value=>{const images=value.items.filter(item=>item.media_type==="image");setItems(images);setSelected(images[0]?.id??"");setStatus(images.length?"":"资产库中还没有图片")}).catch(error=>setStatus(error instanceof Error?error.message:"资产加载失败"))},[workspace]);
  function submit(event:FormEvent){event.preventDefault();const asset=items.find(item=>item.id===selected);if(asset)onInsert(asset,alt.trim())}
  return <div className="modal-backdrop" role="presentation" onMouseDown={event=>{if(event.target===event.currentTarget)onClose()}}><section className="modal asset-picker" role="dialog" aria-modal="true" aria-labelledby="asset-picker-title"><button className="icon-button close" aria-label="关闭" onClick={onClose}><X/></button><h2 id="asset-picker-title">从资产库插入图片</h2><p>选择已上传图片，并填写替代文字。</p>{status&&<p className="settings-status" role="status">{status}</p>}<form onSubmit={submit}><div className="asset-picker-list" role="radiogroup" aria-label="图片资产">{items.map(item=><label key={item.id} className={selected===item.id?"selected":""}><input type="radio" name="asset" value={item.id} checked={selected===item.id} onChange={()=>setSelected(item.id)}/><ImageIcon aria-hidden/><span><strong>{item.filename}</strong><small>{item.mime} · {Math.ceil(item.size/1024)} KB</small></span></label>)}</div><label><span>图片替代文字</span><input required value={alt} onChange={e=>setAlt(e.target.value)} placeholder="描述图片中的关键信息"/></label><button className="primary" disabled={!selected}>插入正文</button></form></section></div>;
}
