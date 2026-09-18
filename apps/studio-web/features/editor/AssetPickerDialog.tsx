import { FormEvent, useEffect, useState } from "react";
import { X } from "lucide-react";
import { api, Asset } from "../../lib/api";
import { ImageAssetBrowser } from "./ImageAssetBrowser";

export function AssetPickerDialog({workspace,onInsert,onClose}:{workspace:string;onInsert:(asset:Asset,alt:string)=>void;onClose:()=>void}){
  const [items,setItems]=useState<Asset[]>([]),[selected,setSelected]=useState(""),[alt,setAlt]=useState(""),[status,setStatus]=useState("正在加载资产…");
  useEffect(()=>{void api.assets(workspace).then(value=>{const images=value.items.filter(item=>item.media_type==="image");setItems(images);setSelected("");setStatus(images.length?"":"资产库中还没有图片")}).catch(error=>setStatus(error instanceof Error?error.message:"资产加载失败"))},[workspace]);
  function submit(event:FormEvent){event.preventDefault();const asset=items.find(item=>item.id===selected);if(asset)onInsert(asset,alt.trim())}
  return <div className="modal-backdrop" role="presentation" onMouseDown={event=>{if(event.target===event.currentTarget)onClose()}}><section className="modal asset-picker" role="dialog" aria-modal="true" aria-labelledby="asset-picker-title"><button className="icon-button close" aria-label="关闭" onClick={onClose}><X/></button><h2 id="asset-picker-title">从资产库插入图片</h2><p>先查看并选择图片，再填写用于无障碍阅读和搜索的替代文字。</p>{status&&<p className="settings-status" role="status">{status}</p>}<form onSubmit={submit}><ImageAssetBrowser items={items} selected={selected} onSelect={setSelected}/><label><span>图片替代文字</span><input required value={alt} onChange={e=>setAlt(e.target.value)} placeholder="描述图片中的关键信息"/></label><button className="primary" disabled={!selected}>插入选中的图片</button></form></section></div>;
}
